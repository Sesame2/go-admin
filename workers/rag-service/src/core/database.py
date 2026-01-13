from sqlalchemy import text
from sqlalchemy.ext.asyncio import create_async_engine, async_sessionmaker, AsyncSession
from sqlalchemy.orm import declarative_base

from core.config import get_settings

settings = get_settings()

# 从配置读取数据库 URL
DATABASE_URL = settings.DATABASE_URL

# 创建异步引擎
engine = create_async_engine(
    DATABASE_URL,
    echo=settings.DEBUG,  # DEBUG 模式下打印 SQL
    pool_size=20,
    max_overflow=10,
)

# 创建异步 Session 工厂
AsyncSessionLocal = async_sessionmaker(
    bind=engine,
    class_=AsyncSession,
    expire_on_commit=False,  # 防止 commit 后属性过期导致再次触发 IO
    autoflush=False,
)

Base = declarative_base()


async def init_db():
    """
    初始化数据库：
    1. 启用 pgvector 扩展
    2. 创建所有表
    """
    # 导入所有模型以确保它们被注册到 Base.metadata
    from models import Document, DocumentChunk, AtomQuestion  # noqa: F401

    async with engine.begin() as conn:
        # 启用 pgvector 扩展
        await conn.execute(text("CREATE EXTENSION IF NOT EXISTS vector"))
        # 创建所有表
        await conn.run_sync(Base.metadata.create_all)


async def get_db():
    """依赖注入用的生成器"""
    async with AsyncSessionLocal() as session:
        try:
            yield session
            await session.commit()
        except Exception:
            await session.rollback()
            raise
        finally:
            await session.close()
