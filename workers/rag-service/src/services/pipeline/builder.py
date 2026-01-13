"""
Pipeline Builder - 管道构建器
提供便捷的管道组装方式
"""

from typing import Optional, Any, List

from sqlalchemy.ext.asyncio import AsyncSession

from core.minio_client import MinioClient
from core.logger import get_logger
from services.pipeline.processor.base import PipelineStep, PipelineOrchestrator
from services.pipeline.processor.downloader import DocumentDownloader
from services.pipeline.processor.loader import DocumentLoader
from services.pipeline.processor.chunker import DocumentChunker
from services.pipeline.processor.tagger import AtomQuestionTagger
from services.pipeline.processor.embedder import EmbeddingGenerator
from services.pipeline.processor.persister import VectorStorePersister, LocalFileCleanup

logger = get_logger("pipeline")


class PipelineBuilder:
    """
    管道构建器

    使用示例:
        pipeline = (
            PipelineBuilder()
            .with_minio(minio_client)
            .with_db(db_session)
            .with_chunking(lang="zh", chunk_size=12)
            .with_embedding(embedding_client)
            .with_cleanup()
            .build()
        )

        result = await pipeline.execute(context)
    """

    def __init__(self):
        self._steps: List[PipelineStep] = []
        self._minio_client: Optional[MinioClient] = None
        self._db_session: Optional[AsyncSession] = None
        self._llm_client: Optional[Any] = None
        self._embedding_client: Optional[Any] = None

    def with_minio(self, minio_client: MinioClient) -> "PipelineBuilder":
        """设置 MinIO 客户端"""
        self._minio_client = minio_client
        return self

    def with_db(self, db_session: AsyncSession) -> "PipelineBuilder":
        """设置数据库会话"""
        self._db_session = db_session
        return self

    def with_llm(self, llm_client: Any) -> "PipelineBuilder":
        """设置 LLM 客户端"""
        self._llm_client = llm_client
        return self

    def with_embedding(self, embedding_client: Any) -> "PipelineBuilder":
        """设置 Embedding 客户端"""
        self._embedding_client = embedding_client
        return self

    def add_step(self, step: PipelineStep) -> "PipelineBuilder":
        """添加自定义步骤"""
        self._steps.append(step)
        return self

    def with_download(self) -> "PipelineBuilder":
        """添加文档下载步骤"""
        if not self._minio_client:
            raise ValueError("MinIO client not set. Call with_minio() first.")
        self._steps.append(DocumentDownloader(self._minio_client))
        return self

    def with_loading(self) -> "PipelineBuilder":
        """添加文档加载步骤"""
        self._steps.append(DocumentLoader())
        return self

    def with_chunking(
        self,
        lang: str = "zh",
        chunk_size: int = 12,
        chunk_overlap: int = 4,
        use_spacy: bool = True,
    ) -> "PipelineBuilder":
        """添加文档分片步骤"""
        self._steps.append(
            DocumentChunker(
                lang=lang,
                chunk_size=chunk_size,
                chunk_overlap=chunk_overlap,
                use_spacy=use_spacy,
            )
        )
        return self

    def with_tagging(self, num_parallel: int = 1) -> "PipelineBuilder":
        """添加原子问题生成步骤"""
        self._steps.append(
            AtomQuestionTagger(
                llm_client=self._llm_client,
                num_parallel=num_parallel,
            )
        )
        return self

    def with_embedding_generation(
        self, embed_chunks: bool = True, embed_questions: bool = True
    ) -> "PipelineBuilder":
        """添加向量嵌入生成步骤"""
        self._steps.append(
            EmbeddingGenerator(
                embedding_client=self._embedding_client,
                embed_chunks=embed_chunks,
                embed_questions=embed_questions,
            )
        )
        return self

    def with_persistence(self, batch_size: int = 100) -> "PipelineBuilder":
        """添加数据持久化步骤"""
        if not self._db_session:
            raise ValueError("DB session not set. Call with_db() first.")
        self._steps.append(
            VectorStorePersister(
                db_session=self._db_session,
                batch_size=batch_size,
            )
        )
        return self

    def with_cleanup(self) -> "PipelineBuilder":
        """添加清理步骤"""
        self._steps.append(LocalFileCleanup())
        return self

    def build(self) -> PipelineOrchestrator:
        """构建管道"""
        if not self._steps:
            raise ValueError("No steps added to pipeline")

        return PipelineOrchestrator(steps=self._steps, logger=logger)

    @classmethod
    def create_default_pipeline(
        cls,
        minio_client: MinioClient,
        db_session: AsyncSession,
        embedding_client: Optional[Any] = None,
        llm_client: Optional[Any] = None,
        lang: str = "zh",
        chunk_size: int = 12,
        chunk_overlap: int = 4,
        enable_tagging: bool = True,
    ) -> PipelineOrchestrator:
        """
        创建默认的完整处理管道

        包含: 下载 -> 加载 -> 分片 -> [原子问题生成] -> 向量化 -> 持久化 -> 清理
        """
        builder = (
            cls()
            .with_minio(minio_client)
            .with_db(db_session)
            .with_embedding(embedding_client)
            .with_llm(llm_client)
            .with_download()
            .with_loading()
            .with_chunking(
                lang=lang, chunk_size=chunk_size, chunk_overlap=chunk_overlap
            )
        )

        if enable_tagging and llm_client:
            builder = builder.with_tagging()

        builder = (
            builder.with_embedding_generation(
                embed_chunks=True,
                embed_questions=enable_tagging and llm_client is not None,
            )
            .with_persistence()
            .with_cleanup()
        )

        return builder.build()

    @classmethod
    def create_chunking_only_pipeline(
        cls,
        minio_client: MinioClient,
        db_session: AsyncSession,
        embedding_client: Optional[Any] = None,
        lang: str = "zh",
        chunk_size: int = 12,
        chunk_overlap: int = 4,
    ) -> PipelineOrchestrator:
        """
        创建仅分片的管道（不生成原子问题）

        包含: 下载 -> 加载 -> 分片 -> 向量化 -> 持久化 -> 清理
        """
        return (
            cls()
            .with_minio(minio_client)
            .with_db(db_session)
            .with_embedding(embedding_client)
            .with_download()
            .with_loading()
            .with_chunking(
                lang=lang, chunk_size=chunk_size, chunk_overlap=chunk_overlap
            )
            .with_embedding_generation(embed_chunks=True, embed_questions=False)
            .with_persistence()
            .with_cleanup()
            .build()
        )
