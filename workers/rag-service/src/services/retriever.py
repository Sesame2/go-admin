"""
基于 pgvector 的向量检索器
实现类似 pike-rag ChunkAtomRetriever 的功能
"""

from typing import List, Optional, Tuple
from uuid import UUID
from dataclasses import dataclass, asdict

import numpy as np
from sqlalchemy import select, text
from sqlalchemy.ext.asyncio import AsyncSession

from core.logger import get_logger
from core.qwen_embedding import QwenEmbeddingClient
from models.chunk import DocumentChunk, AtomQuestion

logger = get_logger("retriever")


@dataclass
class AtomRetrievalInfo:
    """原子问题检索信息"""

    atom_query: str  # 用于检索的查询
    atom: str  # 原子问题内容
    source_chunk_id: str  # 来源 chunk ID
    source_chunk: str  # 来源 chunk 内容
    source_chunk_title: Optional[str]  # 来源 chunk 标题
    retrieval_score: float  # 检索相似度分数
    atom_embedding: List[float]  # 原子问题的嵌入向量

    def to_dict(self):
        return asdict(self)


class PgVectorRetriever:
    """
    基于 PostgreSQL + pgvector 的向量检索器

    功能类似 pike-rag 的 ChunkAtomRetriever：
    - 支持通过原子问题检索
    - 支持通过 chunk 内容检索
    - 支持余弦相似度计算
    """

    def __init__(
        self,
        db_session: AsyncSession,
        embedding_client: QwenEmbeddingClient,
        knowledge_base_id: UUID,
        retrieve_k: int = 5,
        atom_retrieve_k: int = 3,
    ):
        self.db = db_session
        self.embedding_client = embedding_client
        self.knowledge_base_id = knowledge_base_id
        self.retrieve_k = retrieve_k
        self.atom_retrieve_k = atom_retrieve_k

    def _cosine_similarity(self, vec1: List[float], vec2: List[float]) -> float:
        """计算余弦相似度"""
        v1 = np.array(vec1)
        v2 = np.array(vec2)
        return float(np.dot(v1, v2) / (np.linalg.norm(v1) * np.linalg.norm(v2)))

    async def retrieve_atom_info_through_atom(
        self,
        queries: List[str],
        retrieve_k: Optional[int] = None,
    ) -> List[AtomRetrievalInfo]:
        """
        通过原子问题向量库检索相关信息

        Args:
            queries: 查询列表（子问题）
            retrieve_k: 每个查询检索的数量

        Returns:
            检索到的原子信息列表
        """
        k = retrieve_k or self.atom_retrieve_k
        logger.info(f"Retrieving through atom store with {len(queries)} queries, k={k}")

        results: List[AtomRetrievalInfo] = []
        seen_atom_ids = set()

        for query in queries:
            # 生成查询向量
            query_embeddings = await self.embedding_client.embed_documents([query])
            query_embedding = query_embeddings[0]

            # 使用 pgvector 的余弦距离检索
            # 注意：pgvector 使用 <=> 操作符计算余弦距离（1 - 余弦相似度）
            sql = text(
                """
                SELECT 
                    aq.id,
                    aq.question,
                    aq.chunk_id,
                    aq.embedding,
                    dc.content as chunk_content,
                    dc.meta_data as chunk_meta,
                    1 - (aq.embedding <=> :query_vec) as similarity
                FROM atom_questions aq
                JOIN document_chunks dc ON aq.chunk_id = dc.id
                WHERE aq.knowledge_base_id = :kb_id
                    AND aq.embedding IS NOT NULL
                ORDER BY aq.embedding <=> :query_vec
                LIMIT :k
            """
            )

            result = await self.db.execute(
                sql,
                {
                    "query_vec": str(query_embedding),
                    "kb_id": str(self.knowledge_base_id),
                    "k": k,
                },
            )

            rows = result.fetchall()

            for row in rows:
                atom_id = str(row.id)
                if atom_id in seen_atom_ids:
                    continue
                seen_atom_ids.add(atom_id)

                chunk_meta = row.chunk_meta or {}
                title = chunk_meta.get("title") or chunk_meta.get("filename")

                results.append(
                    AtomRetrievalInfo(
                        atom_query=query,
                        atom=row.question,
                        source_chunk_id=str(row.chunk_id),
                        source_chunk=row.chunk_content,
                        source_chunk_title=title,
                        retrieval_score=float(row.similarity),
                        atom_embedding=list(row.embedding) if row.embedding else [],
                    )
                )

        logger.info(f"Retrieved {len(results)} atom infos through atom store")
        return results

    async def retrieve_atom_info_through_chunk(
        self,
        query: str,
        retrieve_k: Optional[int] = None,
    ) -> List[AtomRetrievalInfo]:
        """
        通过 chunk 向量库检索，然后找到最相关的原子问题

        Args:
            query: 查询字符串
            retrieve_k: 检索数量

        Returns:
            检索到的原子信息列表
        """
        k = retrieve_k or self.retrieve_k
        logger.info(
            f"Retrieving through chunk store with query: {query[:50]}..., k={k}"
        )

        # 生成查询向量
        query_embeddings = await self.embedding_client.embed_documents([query])
        query_embedding = query_embeddings[0]

        # 检索相关 chunks
        sql = text(
            """
            SELECT 
                dc.id,
                dc.content,
                dc.meta_data,
                dc.embedding,
                1 - (dc.embedding <=> :query_vec) as similarity
            FROM document_chunks dc
            WHERE dc.knowledge_base_id = :kb_id
                AND dc.embedding IS NOT NULL
            ORDER BY dc.embedding <=> :query_vec
            LIMIT :k
        """
        )

        result = await self.db.execute(
            sql,
            {
                "query_vec": str(query_embedding),
                "kb_id": str(self.knowledge_base_id),
                "k": k,
            },
        )

        chunk_rows = result.fetchall()
        results: List[AtomRetrievalInfo] = []

        for chunk_row in chunk_rows:
            chunk_id = str(chunk_row.id)
            chunk_meta = chunk_row.meta_data or {}
            title = chunk_meta.get("title") or chunk_meta.get("filename")

            # 获取该 chunk 关联的所有原子问题
            atom_sql = text(
                """
                SELECT id, question, embedding
                FROM atom_questions
                WHERE chunk_id = :chunk_id
                    AND embedding IS NOT NULL
            """
            )

            atom_result = await self.db.execute(atom_sql, {"chunk_id": chunk_id})
            atom_rows = atom_result.fetchall()

            if not atom_rows:
                # 如果没有原子问题，用 chunk 内容作为"原子问题"
                results.append(
                    AtomRetrievalInfo(
                        atom_query=query,
                        atom=chunk_row.content[:200] + "...",
                        source_chunk_id=chunk_id,
                        source_chunk=chunk_row.content,
                        source_chunk_title=title,
                        retrieval_score=float(chunk_row.similarity),
                        atom_embedding=(
                            list(chunk_row.embedding) if chunk_row.embedding else []
                        ),
                    )
                )
                continue

            # 找到与查询最相似的原子问题
            best_atom = None
            best_score = -1
            best_embedding = []

            for atom_row in atom_rows:
                if atom_row.embedding:
                    score = self._cosine_similarity(
                        query_embedding, list(atom_row.embedding)
                    )
                    if score > best_score:
                        best_score = score
                        best_atom = atom_row.question
                        best_embedding = list(atom_row.embedding)

            if best_atom:
                results.append(
                    AtomRetrievalInfo(
                        atom_query=query,
                        atom=best_atom,
                        source_chunk_id=chunk_id,
                        source_chunk=chunk_row.content,
                        source_chunk_title=title,
                        retrieval_score=best_score,
                        atom_embedding=best_embedding,
                    )
                )

        logger.info(f"Retrieved {len(results)} atom infos through chunk store")
        return results

    def filter_atom_infos(
        self,
        candidates: List[AtomRetrievalInfo],
        chosen_infos: List[AtomRetrievalInfo],
    ) -> List[AtomRetrievalInfo]:
        """
        过滤已选择的 chunk 对应的原子信息

        Args:
            candidates: 候选列表
            chosen_infos: 已选择的列表

        Returns:
            过滤后的候选列表
        """
        if not chosen_infos:
            return candidates

        chosen_chunk_ids = {info.source_chunk_id for info in chosen_infos}
        filtered = [c for c in candidates if c.source_chunk_id not in chosen_chunk_ids]

        logger.debug(
            f"Filtered {len(candidates) - len(filtered)} duplicates, {len(filtered)} remaining"
        )
        return filtered
