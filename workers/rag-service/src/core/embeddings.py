from typing import List, Dict, Optional, Tuple
from sqlalchemy import select, text, delete
from sqlalchemy.ext.asyncio import AsyncSession
from pgvector.sqlalchemy import Vector

from core.database import engine, Base, AsyncSessionLocal
from models import DocumentChunk


class AsyncPostgresVectorStore:
    def __init__(self):
        # 无需在 init 中做 IO 操作
        pass

    async def init_db(self):
        """
        初始化数据库表结构和 Vector 扩展
        注意：生产环境建议使用 Alembic 做迁移，不要用 create_all
        """
        async with engine.begin() as conn:
            # 启用 pgvector 扩展 (需要超级用户权限，如果报错请手动在 DB 执行)
            await conn.execute(text("CREATE EXTENSION IF NOT EXISTS vector"))
            # 创建表
            await conn.run_sync(Base.metadata.create_all)

    async def add_documents(
        self,
        texts: List[str],
        embeddings: List[List[float]],
        metadatas: Optional[List[Dict]] = None,
    ) -> List[str]:
        """
        异步批量添加文档
        """
        if not metadatas:
            metadatas = [{} for _ in texts]

        async with AsyncSessionLocal() as session:
            async with session.begin():  # 自动管理事务
                chunks = []
                ids = []
                for t, e, m in zip(texts, embeddings, metadatas):
                    chunk = DocumentChunk(content=t, embedding=e, meta_data=m)
                    chunks.append(chunk)

                session.add_all(chunks)
                await session.flush()  # 刷新以获取 ID，但不提交

                ids = [str(c.id) for c in chunks]

            # session.begin() 上下文结束时会自动 commit
            return ids

    async def similarity_search(
        self,
        query_embedding: List[float],
        top_k: int = 5,
        filter_meta: Optional[Dict] = None,
    ) -> List[Dict]:
        """
        异步相似度搜索 (使用 Cosine 距离)
        返回格式: [{'content': str, 'metadata': dict, 'score': float}, ...]
        """
        async with AsyncSessionLocal() as session:
            # 构建距离计算表达式 (cosine_distance: 1 - cosine_similarity)
            # 也就是距离越小越相似
            distance_expr = DocumentChunk.embedding.cosine_distance(
                query_embedding
            ).label("distance")

            stmt = (
                select(DocumentChunk, distance_expr)
                .order_by(distance_expr)
                .limit(top_k)
            )

            # 添加元数据过滤
            if filter_meta:
                for k, v in filter_meta.items():
                    # 简单的 JSONB 键值匹配
                    stmt = stmt.filter(DocumentChunk.meta_data[k].astext == str(v))

            result = await session.execute(stmt)
            rows = result.all()

            # 格式化输出
            results = []
            for chunk, distance in rows:
                # 将距离转换为相似度分数 (0~1)
                # Cosine distance 范围通常是 0(完全相同) ~ 2(完全相反)
                # 这里的 score 定义视业务而定，这里简单用 1 - distance
                score = 1 - distance
                results.append(
                    {
                        "id": str(chunk.id),
                        "content": chunk.content,
                        "metadata": chunk.meta_data,
                        "score": score,
                    }
                )

            return results

    async def delete_document_by_file(self, filename: str):
        """
        根据 metadata 中的 filename 删除相关切片
        """
        async with AsyncSessionLocal() as session:
            async with session.begin():
                await session.execute(
                    delete(DocumentChunk).where(
                        DocumentChunk.meta_data["filename"].astext == filename
                    )
                )
