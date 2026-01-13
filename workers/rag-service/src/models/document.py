import uuid
from datetime import datetime
from enum import Enum
from typing import Optional, List

from sqlalchemy import Column, Text, DateTime, String, Integer, func, Enum as SAEnum
from sqlalchemy.dialects.postgresql import JSONB, UUID
from sqlalchemy.orm import mapped_column, relationship

from core.database import Base


class DocumentStatus(str, Enum):
    """文档处理状态"""

    PENDING = "pending"  # 待处理
    DOWNLOADING = "downloading"  # 下载中
    PARSING = "parsing"  # 解析中
    CHUNKING = "chunking"  # 分片中
    TAGGING = "tagging"  # 原子问题生成中
    EMBEDDING = "embedding"  # 向量化中
    COMPLETED = "completed"  # 处理完成
    FAILED = "failed"  # 处理失败


class Document(Base):
    """文档表 - 存储原始文档信息"""

    __tablename__ = "documents"

    id = mapped_column(UUID(as_uuid=True), primary_key=True, default=uuid.uuid4)

    # 业务关联
    knowledge_base_id = mapped_column(UUID(as_uuid=True), nullable=False, index=True)
    user_id = mapped_column(UUID(as_uuid=True), nullable=True, index=True)

    # 文档基本信息
    filename = mapped_column(String(512), nullable=False)
    file_type = mapped_column(String(50), nullable=False)  # pdf, docx, pptx, markdown
    file_size = mapped_column(Integer, nullable=True)  # 文件大小（字节）
    minio_path = mapped_column(String(1024), nullable=False)  # MinIO 存储路径

    # 处理状态
    status = mapped_column(
        SAEnum(DocumentStatus, name="document_status"),
        nullable=False,
        default=DocumentStatus.PENDING,
        index=True,
    )
    error_message = mapped_column(Text, nullable=True)  # 错误信息

    # 处理结果统计
    chunk_count = mapped_column(Integer, default=0)  # 分片数量
    atom_question_count = mapped_column(Integer, default=0)  # 原子问题数量

    # 元数据（自定义扩展字段）
    meta_data = mapped_column(JSONB, nullable=True, server_default="{}")

    # 时间戳
    created_at = mapped_column(DateTime(timezone=True), server_default=func.now())
    updated_at = mapped_column(
        DateTime(timezone=True), server_default=func.now(), onupdate=func.now()
    )
    processed_at = mapped_column(DateTime(timezone=True), nullable=True)  # 处理完成时间

    # 关系（可选，用于 ORM 关联查询）
    # chunks = relationship("DocumentChunk", back_populates="document", cascade="all, delete-orphan")

    def __repr__(self):
        return (
            f"<Document(id={self.id}, filename={self.filename}, status={self.status})>"
        )
