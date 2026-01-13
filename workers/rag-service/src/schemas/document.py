from typing import Optional, List, Dict, Any
from uuid import UUID
from pydantic import BaseModel, Field


class DocumentAnalysisRequest(BaseModel):
    """文档分析请求"""

    document_id: str
    file_path: str  # 在minio中的文件路径
    meta_data: dict = Field(default_factory=dict)  # 文档元数据


class MarkdownUploadRequest(BaseModel):
    """Markdown 文件上传请求 - 实际上传文件并存储到数据库"""

    knowledge_base_id: str = Field(..., description="知识库 ID")
    title: Optional[str] = Field(None, description="文档标题（可选）")

    # 分片参数
    chunk_size: int = Field(default=12, ge=1, le=50, description="每个 chunk 的句子数")
    chunk_overlap: int = Field(default=4, ge=0, le=20, description="重叠句子数")
    lang: str = Field(default="zh", description="语言: zh/en")
    use_spacy: bool = Field(default=False, description="是否使用 spaCy 分句")

    # 原子问题生成参数
    max_atom_questions: int = Field(
        default=10, ge=1, le=20, description="每个 chunk 生成的最大原子问题数"
    )


class MarkdownTestRequest(BaseModel):
    """Markdown 测试请求 - 用于测试切片和原子问题生成（不存储数据库）"""

    content: str = Field(..., description="Markdown 文本内容")
    knowledge_base_id: str = Field(default="test-kb-001", description="知识库 ID")
    filename: str = Field(default="test.md", description="文件名")

    # 分片参数
    chunk_size: int = Field(default=6, ge=1, le=50, description="每个 chunk 的句子数")
    chunk_overlap: int = Field(default=2, ge=0, le=20, description="重叠句子数")
    lang: str = Field(default="zh", description="语言: zh/en")
    use_spacy: bool = Field(default=False, description="是否使用 spaCy 分句")

    # 原子问题生成参数
    generate_atom_questions: bool = Field(default=True, description="是否生成原子问题")
    questions_per_chunk: int = Field(
        default=3, ge=1, le=10, description="每个 chunk 生成的问题数"
    )


class ChunkResponse(BaseModel):
    """分片响应"""

    chunk_index: int
    content: str
    char_count: int
    metadata: Dict[str, Any] = Field(default_factory=dict)


class AtomQuestionResponse(BaseModel):
    """原子问题响应"""

    chunk_index: int
    question: str


class MarkdownTestResponse(BaseModel):
    """Markdown 测试响应（不存储数据库）"""

    document_id: str
    knowledge_base_id: str
    filename: str

    # 统计信息
    total_chars: int
    chunk_count: int
    atom_question_count: int

    # 详细内容
    chunks: List[ChunkResponse]
    atom_questions: List[AtomQuestionResponse]

    # 处理参数
    settings: Dict[str, Any]


class MarkdownUploadResponse(BaseModel):
    """Markdown 文件上传响应（存储到数据库）"""

    document_id: UUID
    knowledge_base_id: str
    filename: str
    status: str

    # 统计信息
    total_chars: int
    chunk_count: int
    atom_question_count: int

    # 处理信息
    message: str
    created_at: str
