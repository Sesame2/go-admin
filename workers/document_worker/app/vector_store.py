import os
from typing import List, Optional

from sqlalchemy import String, Integer, JSON, select, delete
from sqlalchemy.ext.asyncio import create_async_engine, AsyncSession, async_sessionmaker
from sqlalchemy.orm import declarative_base, Mapped, mapped_column
from pgvector.sqlalchemy import Vector

from config import Settings

DATABASE_URL = os.getenv(
    "DATABASE_URL",
    f"postgresql+asyncpg://{Settings.postgres_username}:{Settings.postgres_password}@{Settings.postgres_host}:{Settings.postgres_port}/{Settings.postgres_db}",
)

engine = create_async_engine(DATABASE_URL, echo=True)

AsyncSessionLocal = async_sessionmaker(bind=engine, expire_on_commit=False)

Base = declarative_base()


class KnowledgeChunk(Base):
    __tablename__ = "knowledge_chunks"

    # 自定义主键，例如 "doc123-0"
    id: Mapped[str] = mapped_column(String, primary_key=True)   

    # 向量（使用 pgvector，维度根据实际设置 768 或 1536）
    embedding: Mapped[list[float]] = mapped_column(Vector(768))  # 或 Vector(1536)

    # 原始文本
    text: Mapped[str] = mapped_column(String)

    # 文档标识
    doc_id: Mapped[str] = mapped_column(String, index=True)

    # 第几个块
    chunk_id: Mapped[int] = mapped_column(Integer)

    # 来源文件
    source: Mapped[str] = mapped_column(String)

    # 可选的附加信息，作为 JSON 存储
    meta_data: Mapped[dict] = mapped_column(JSON, nullable=True)


async def init_db():
    """初始化数据库，创建表"""
    async with engine.begin() as conn:
        await conn.run_sync(Base.metadata.create_all)


class VectorStore:
    """向量存储类，封装对数据库的操作"""

    def __init__(self):
        self.session_maker = AsyncSessionLocal

    async def add_chunk(self, chunk: KnowledgeChunk):
        """添加一个向量块到数据库"""
        async with self.session_maker() as session:
            async with session.begin():
                session.add(chunk)

    async def add_chunks(self, chunks: List[KnowledgeChunk]):
        """批量添加向量块到数据库"""
        async with self.session_maker() as session:
            async with session.begin():
                session.add_all(chunks)

    async def get_chunk_by_id(self, chunk_id: str) -> Optional[KnowledgeChunk]:
        """根据 ID 获取向量块"""
        async with self.session_maker() as session:
            result = await session.execute(select(KnowledgeChunk).where(KnowledgeChunk.id == chunk_id))
            return result.scalar_one_or_none()

    async def get_chunks_by_doc_id(self, doc_id: str) -> List[KnowledgeChunk]:
        """根据文档 ID 获取所有向量块"""
        async with self.session_maker() as session:
            result = await session.execute(select(KnowledgeChunk).where(KnowledgeChunk.doc_id == doc_id))
            return result.scalars().all()

    async def delete_chunk_by_id(self, chunk_id: str):
        """根据 ID 删除向量块"""
        async with self.session_maker() as session:
            async with session.begin():
                await session.execute(delete(KnowledgeChunk).where(KnowledgeChunk.id == chunk_id))

    async def delete_chunks_by_doc_id(self, doc_id: str):
        """根据文档 ID 删除所有向量块"""
        async with self.session_maker() as session:
            async with session.begin():
                await session.execute(delete(KnowledgeChunk).where(KnowledgeChunk.doc_id == doc_id))

    async def search_by_embedding(self, embedding: List[float], top_k: int = 5) -> List[KnowledgeChunk]:
        """
        根据向量搜索最相似的向量块
        Args:
            embedding: 查询的向量
            top_k: 返回的最相似结果数量
        Returns:
            最相似的向量块列表
        """
        async with self.session_maker() as session:
            query = (
                select(KnowledgeChunk)
                .order_by(KnowledgeChunk.embedding.l2_distance(embedding))
                .limit(top_k)
            )
            result = await session.execute(query)
            return result.scalars().all()

