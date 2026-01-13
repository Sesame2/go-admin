from pydantic_settings import BaseSettings
from functools import lru_cache
from pathlib import Path

# 获取项目根目录的 .env 文件路径
_ENV_FILE = Path(__file__).resolve().parent.parent.parent / ".env"


class Settings(BaseSettings):
    """应用配置"""

    # 应用配置
    APP_NAME: str = "RAG Service"
    APP_VERSION: str = "1.0.0"
    DEBUG: bool = True
    SERVER_HOST: str = "0.0.0.0"
    SERVER_PORT: int = 8000

    # PostgreSQL 数据库配置（与 Go 端保持一致）
    POSTGRES_USER: str = "admin"
    POSTGRES_PASSWORD: str = "password"
    POSTGRES_DB: str = "go_admin"
    POSTGRES_HOST: str = "localhost"
    POSTGRES_PORT: int = 5432

    @property
    def DATABASE_URL(self) -> str:
        """动态生成数据库连接 URL"""
        return f"postgresql+asyncpg://{self.POSTGRES_USER}:{self.POSTGRES_PASSWORD}@{self.POSTGRES_HOST}:{self.POSTGRES_PORT}/{self.POSTGRES_DB}"

    @property
    def SYNC_DATABASE_URL(self) -> str:
        """同步数据库连接 URL（用于 RabbitMQ worker）"""
        return f"postgresql://{self.POSTGRES_USER}:{self.POSTGRES_PASSWORD}@{self.POSTGRES_HOST}:{self.POSTGRES_PORT}/{self.POSTGRES_DB}"

    # MinIO 配置（与 Go 端保持一致）
    MINIO_ENDPOINT: str = "localhost:9000"
    MINIO_ACCESS_KEY: str = "root"
    MINIO_SECRET_KEY: str = "123456789"
    MINIO_SECURE: bool = False
    MINIO_BUCKET: str = "documents"
    MINIO_REGION: str = "us-east-1"

    # RabbitMQ 配置（与 Go 端保持一致）
    RABBITMQ_HOST: str = "localhost"
    RABBITMQ_PORT: int = 5672
    RABBITMQ_USER: str = "admin"
    RABBITMQ_PASSWORD: str = "123456789"
    RABBITMQ_VHOST: str = "/"
    RABBITMQ_EXCHANGE: str = "go_admin_exchange"
    RABBITMQ_QUEUE: str = "document_parse_queue"
    RABBITMQ_ROUTING_KEY: str = "document.parse"

    @property
    def RABBITMQ_URL(self) -> str:
        """RabbitMQ 连接 URL"""
        return f"amqp://{self.RABBITMQ_USER}:{self.RABBITMQ_PASSWORD}@{self.RABBITMQ_HOST}:{self.RABBITMQ_PORT}/{self.RABBITMQ_VHOST}"

    # 临时文件配置
    TEMP_DIR: str = "/tmp/rag-service"

    # 向量数据库配置
    VECTOR_DIMENSION: int = 1024

    # 通义千问配置
    QWEN_API_KEY: str = ""
    QWEN_LLM_MODEL: str = "qwen-max"
    QWEN_EMBEDDING_MODEL: str = "text-embedding-v3"
    QWEN_EMBEDDING_DIMENSIONS: int = 1024

    # 文档分块配置
    CHUNK_SIZE: int = 12  # 每个 chunk 包含的句子数
    CHUNK_OVERLAP: int = 4  # chunk 之间的重叠句子数
    MAX_ATOM_QUESTIONS: int = 10  # 每个 chunk 生成的最大原子问题数

    class Config:
        env_file = str(_ENV_FILE) if _ENV_FILE.exists() else None
        env_file_encoding = "utf-8"
        case_sensitive = True


@lru_cache()
def get_settings() -> Settings:
    """获取配置实例（单例）"""
    settings = Settings()
    import logging
    logging.getLogger("core").info(f"Database URL: {settings.DATABASE_URL}")
    return settings
