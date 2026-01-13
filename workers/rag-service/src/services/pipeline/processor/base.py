"""
Pipeline 基础组件
"""

from abc import ABC, abstractmethod
from typing import Optional, List, Any
from dataclasses import dataclass, field
from uuid import UUID
from enum import Enum

from langchain_core.documents import Document


class ProcessStatus(str, Enum):
    """处理状态"""

    PENDING = "pending"
    IN_PROGRESS = "in_progress"
    COMPLETED = "completed"
    FAILED = "failed"
    SKIPPED = "skipped"


@dataclass
class ChunkData:
    """分片数据结构"""

    content: str
    chunk_index: int
    start_char: Optional[int] = None
    end_char: Optional[int] = None
    metadata: dict = field(default_factory=dict)
    embedding: Optional[List[float]] = None


@dataclass
class AtomQuestionData:
    """原子问题数据结构"""

    question: str
    chunk_index: int
    question_type: Optional[str] = None
    metadata: dict = field(default_factory=dict)
    embedding: Optional[List[float]] = None


@dataclass
class ProcessContext:
    """
    处理上下文,在各个步骤间传递数据
    """

    # 文档标识
    document_id: UUID
    knowledge_base_id: UUID

    # MinIO 信息
    minio_path: str
    document_type: str
    filename: str

    # 本地文件路径（下载后）
    local_path: Optional[str] = None

    # 原始加载的文档（langchain Document 格式）
    raw_documents: Optional[List[Document]] = None

    # 解析后的文档（结构化内容）
    parsed_documents: Optional[List[Document]] = None

    # 分片结果
    chunks: Optional[List[ChunkData]] = None

    # 原子问题
    atom_questions: Optional[List[AtomQuestionData]] = None

    # 处理状态
    status: ProcessStatus = ProcessStatus.PENDING
    error_message: Optional[str] = None

    # 扩展元数据
    metadata: dict = field(default_factory=dict)

    def to_dict(self) -> dict:
        """转换为字典（用于日志等）"""
        return {
            "document_id": str(self.document_id),
            "knowledge_base_id": str(self.knowledge_base_id),
            "filename": self.filename,
            "document_type": self.document_type,
            "status": self.status.value,
            "chunk_count": len(self.chunks) if self.chunks else 0,
            "atom_question_count": (
                len(self.atom_questions) if self.atom_questions else 0
            ),
        }


class PipelineStep(ABC):
    """管道步骤基类"""

    # 步骤名称（子类可覆盖）
    name: str = "BaseStep"

    @abstractmethod
    async def process(self, context: ProcessContext) -> ProcessContext:
        """
        处理逻辑

        Args:
            context: 处理上下文

        Returns:
            更新后的上下文
        """
        pass

    @abstractmethod
    def can_handle(self, context: ProcessContext) -> bool:
        """
        判断是否可以处理该上下文

        Args:
            context: 处理上下文

        Returns:
            是否可以处理
        """
        pass

    def __repr__(self):
        return f"<{self.__class__.__name__}>"


class PipelineOrchestrator:
    """管道编排器"""

    def __init__(self, steps: List[PipelineStep], logger=None):
        self.steps = steps
        self.logger = logger

    async def execute(self, context: ProcessContext) -> ProcessContext:
        """
        执行完整的处理管道

        Args:
            context: 初始上下文

        Returns:
            处理完成的上下文
        """
        context.status = ProcessStatus.IN_PROGRESS

        for step in self.steps:
            step_name = step.__class__.__name__

            if not step.can_handle(context):
                if self.logger:
                    self.logger.debug(
                        f"Step {step_name} skipped - cannot handle context"
                    )
                continue

            try:
                if self.logger:
                    self.logger.info(f"Executing step: {step_name}")

                context = await step.process(context)

                if self.logger:
                    self.logger.info(f"Step {step_name} completed")

            except Exception as e:
                context.status = ProcessStatus.FAILED
                context.error_message = f"Step {step_name} failed: {str(e)}"

                if self.logger:
                    self.logger.error(f"Step {step_name} failed: {e}", exc_info=True)

                raise

        context.status = ProcessStatus.COMPLETED
        return context
