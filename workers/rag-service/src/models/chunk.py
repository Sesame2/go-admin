import uuid
from sqlalchemy import Column, Text, DateTime, Integer, String, func, Index, ForeignKey
from sqlalchemy.dialects.postgresql import JSONB, UUID, ARRAY
from sqlalchemy.orm import mapped_column, relationship
from pgvector.sqlalchemy import Vector

from core.database import Base


class DocumentChunk(Base):
    """
    文档分片表 - 存储文档切分后的片段
    每个 chunk 包含原文内容和对应的向量嵌入
    """

    __tablename__ = "document_chunks"

    id = mapped_column(UUID(as_uuid=True), primary_key=True, default=uuid.uuid4)

    # 关联文档
    document_id = mapped_column(UUID(as_uuid=True), nullable=False, index=True)
    knowledge_base_id = mapped_column(UUID(as_uuid=True), nullable=False, index=True)

    # 分片内容
    content = mapped_column(Text, nullable=False)

    # 分片位置信息
    chunk_index = mapped_column(Integer, nullable=False)  # 分片序号（从0开始）
    start_char = mapped_column(Integer, nullable=True)  # 在原文中的起始位置
    end_char = mapped_column(Integer, nullable=True)  # 在原文中的结束位置

    # 分片向量嵌入（用于语义检索）- 使用 1024 维度匹配 Qwen embedding
    embedding = mapped_column(Vector(1024), nullable=True)

    # 元数据（页码、标题层级等）
    meta_data = mapped_column(JSONB, nullable=True, server_default="{}")

    # 时间戳
    created_at = mapped_column(DateTime(timezone=True), server_default=func.now())

    # 关系
    # document = relationship("Document", back_populates="chunks")
    # atom_questions = relationship("AtomQuestion", back_populates="chunk", cascade="all, delete-orphan")

    __table_args__ = (
        # 复合索引：按文档ID过滤后的向量检索
        Index("idx_chunk_doc_id", "document_id"),
        Index("idx_chunk_kb_id", "knowledge_base_id"),
    )

    def __repr__(self):
        return f"<DocumentChunk(id={self.id}, doc={self.document_id}, index={self.chunk_index})>"


class AtomQuestion(Base):
    """
    原子问题表 - 存储为每个 chunk 生成的原子化问题
    Pike-RAG 核心：通过原子问题作为检索的中间层，提升检索准确性
    """

    __tablename__ = "atom_questions"

    id = mapped_column(UUID(as_uuid=True), primary_key=True, default=uuid.uuid4)

    # 关联
    chunk_id = mapped_column(
        UUID(as_uuid=True),
        ForeignKey("document_chunks.id", ondelete="CASCADE"),
        nullable=False,
        index=True,
    )
    document_id = mapped_column(UUID(as_uuid=True), nullable=False, index=True)
    knowledge_base_id = mapped_column(UUID(as_uuid=True), nullable=False, index=True)

    # 原子问题内容
    question = mapped_column(Text, nullable=False)

    # 原子问题的向量嵌入（用于问题-问题匹配）- 使用 1024 维度匹配 Qwen embedding
    embedding = mapped_column(Vector(1024), nullable=True)

    # 问题类型/标签（可选）
    question_type = mapped_column(String(50), nullable=True)

    # 元数据
    meta_data = mapped_column(JSONB, nullable=True, server_default="{}")

    # 时间戳
    created_at = mapped_column(DateTime(timezone=True), server_default=func.now())

    # 关系
    # chunk = relationship("DocumentChunk", back_populates="atom_questions")

    __table_args__ = (
        Index("idx_aq_chunk_id", "chunk_id"),
        Index("idx_aq_doc_id", "document_id"),
        Index("idx_aq_kb_id", "knowledge_base_id"),
    )

    def __repr__(self):
        return f"<AtomQuestion(id={self.id}, chunk={self.chunk_id}, q={self.question[:30]}...)>"
