"""
Chunk Repository - 分片数据访问层
"""

from typing import List, Optional, Dict, Any
from uuid import UUID

from sqlalchemy import select, delete, func
from sqlalchemy.ext.asyncio import AsyncSession

from models.chunk import DocumentChunk, AtomQuestion
from core.logger import get_logger

logger = get_logger("repository")


class ChunkRepository:
    """文档分片数据访问层"""

    def __init__(self, db_session: AsyncSession):
        self.db = db_session

    # ==================== DocumentChunk CRUD ====================

    async def create_chunk(self, chunk: DocumentChunk) -> DocumentChunk:
        """创建单个分片"""
        self.db.add(chunk)
        await self.db.flush()
        return chunk

    async def create_chunks_batch(
        self, chunks: List[DocumentChunk]
    ) -> List[DocumentChunk]:
        """批量创建分片"""
        self.db.add_all(chunks)
        await self.db.flush()
        return chunks

    async def get_chunk_by_id(self, chunk_id: UUID) -> Optional[DocumentChunk]:
        """根据 ID 获取分片"""
        stmt = select(DocumentChunk).where(DocumentChunk.id == chunk_id)
        result = await self.db.execute(stmt)
        return result.scalar_one_or_none()

    async def get_chunks_by_document(
        self, document_id: UUID, skip: int = 0, limit: int = 100
    ) -> List[DocumentChunk]:
        """获取文档的所有分片"""
        stmt = (
            select(DocumentChunk)
            .where(DocumentChunk.document_id == document_id)
            .order_by(DocumentChunk.chunk_index)
            .offset(skip)
            .limit(limit)
        )
        result = await self.db.execute(stmt)
        return list(result.scalars().all())

    async def get_chunks_by_knowledge_base(
        self, knowledge_base_id: UUID, skip: int = 0, limit: int = 100
    ) -> List[DocumentChunk]:
        """获取知识库的所有分片"""
        stmt = (
            select(DocumentChunk)
            .where(DocumentChunk.knowledge_base_id == knowledge_base_id)
            .offset(skip)
            .limit(limit)
        )
        result = await self.db.execute(stmt)
        return list(result.scalars().all())

    async def delete_chunks_by_document(self, document_id: UUID) -> int:
        """删除文档的所有分片"""
        stmt = delete(DocumentChunk).where(DocumentChunk.document_id == document_id)
        result = await self.db.execute(stmt)
        return result.rowcount

    async def count_chunks_by_document(self, document_id: UUID) -> int:
        """统计文档的分片数量"""
        stmt = select(func.count(DocumentChunk.id)).where(
            DocumentChunk.document_id == document_id
        )
        result = await self.db.execute(stmt)
        return result.scalar() or 0

    # ==================== 向量检索 ====================

    async def similarity_search(
        self,
        query_embedding: List[float],
        knowledge_base_id: Optional[UUID] = None,
        document_id: Optional[UUID] = None,
        top_k: int = 5,
        score_threshold: float = 0.0,
    ) -> List[Dict[str, Any]]:
        """
        向量相似度搜索

        Args:
            query_embedding: 查询向量
            knowledge_base_id: 知识库 ID（可选过滤）
            document_id: 文档 ID（可选过滤）
            top_k: 返回数量
            score_threshold: 相似度阈值（0-1）

        Returns:
            [{'chunk': DocumentChunk, 'score': float}, ...]
        """
        # 计算余弦距离
        distance_expr = DocumentChunk.embedding.cosine_distance(query_embedding).label(
            "distance"
        )

        stmt = select(DocumentChunk, distance_expr).order_by(distance_expr).limit(top_k)

        # 添加过滤条件
        if knowledge_base_id:
            stmt = stmt.where(DocumentChunk.knowledge_base_id == knowledge_base_id)
        if document_id:
            stmt = stmt.where(DocumentChunk.document_id == document_id)

        result = await self.db.execute(stmt)
        rows = result.all()

        results = []
        for chunk, distance in rows:
            # 余弦距离转相似度: score = 1 - distance
            score = 1 - distance
            if score >= score_threshold:
                results.append(
                    {
                        "chunk": chunk,
                        "score": score,
                    }
                )

        return results

    # ==================== AtomQuestion CRUD ====================

    async def create_atom_question(self, aq: AtomQuestion) -> AtomQuestion:
        """创建原子问题"""
        self.db.add(aq)
        await self.db.flush()
        return aq

    async def create_atom_questions_batch(
        self, questions: List[AtomQuestion]
    ) -> List[AtomQuestion]:
        """批量创建原子问题"""
        self.db.add_all(questions)
        await self.db.flush()
        return questions

    async def get_atom_questions_by_chunk(self, chunk_id: UUID) -> List[AtomQuestion]:
        """获取分片的所有原子问题"""
        stmt = select(AtomQuestion).where(AtomQuestion.chunk_id == chunk_id)
        result = await self.db.execute(stmt)
        return list(result.scalars().all())

    async def get_atom_questions_by_document(
        self, document_id: UUID
    ) -> List[AtomQuestion]:
        """获取文档的所有原子问题"""
        stmt = select(AtomQuestion).where(AtomQuestion.document_id == document_id)
        result = await self.db.execute(stmt)
        return list(result.scalars().all())

    async def delete_atom_questions_by_document(self, document_id: UUID) -> int:
        """删除文档的所有原子问题"""
        stmt = delete(AtomQuestion).where(AtomQuestion.document_id == document_id)
        result = await self.db.execute(stmt)
        return result.rowcount

    async def count_atom_questions_by_document(self, document_id: UUID) -> int:
        """统计文档的原子问题数量"""
        stmt = select(func.count(AtomQuestion.id)).where(
            AtomQuestion.document_id == document_id
        )
        result = await self.db.execute(stmt)
        return result.scalar() or 0

    async def similarity_search_atom_questions(
        self,
        query_embedding: List[float],
        knowledge_base_id: Optional[UUID] = None,
        top_k: int = 5,
        score_threshold: float = 0.0,
    ) -> List[Dict[str, Any]]:
        """
        在原子问题上进行向量检索
        Pike-RAG 核心：通过匹配问题来检索相关内容

        Returns:
            [{'atom_question': AtomQuestion, 'chunk': DocumentChunk, 'score': float}, ...]
        """
        distance_expr = AtomQuestion.embedding.cosine_distance(query_embedding).label(
            "distance"
        )

        stmt = (
            select(AtomQuestion, DocumentChunk, distance_expr)
            .join(DocumentChunk, AtomQuestion.chunk_id == DocumentChunk.id)
            .order_by(distance_expr)
            .limit(top_k)
        )

        if knowledge_base_id:
            stmt = stmt.where(AtomQuestion.knowledge_base_id == knowledge_base_id)

        result = await self.db.execute(stmt)
        rows = result.all()

        results = []
        for aq, chunk, distance in rows:
            score = 1 - distance
            if score >= score_threshold:
                results.append(
                    {
                        "atom_question": aq,
                        "chunk": chunk,
                        "score": score,
                    }
                )

        return results
