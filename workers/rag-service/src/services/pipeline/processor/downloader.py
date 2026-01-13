"""
文档下载步骤
"""

from core.logger import get_logger
from core.minio_client import MinioClient
from services.pipeline.processor.base import PipelineStep, ProcessContext

logger = get_logger("pipeline")


class DocumentDownloader(PipelineStep):
    """从 MinIO 下载文档到本地"""

    name = "DocumentDownloader"

    def __init__(self, minio_client: MinioClient):
        self.minio_client = minio_client

    async def process(self, context: ProcessContext) -> ProcessContext:
        """从 MinIO 下载文档"""
        logger.info(f"Downloading document: {context.minio_path}")

        local_path = await self.minio_client.download_to_local(context.minio_path)
        context.local_path = local_path

        logger.info(f"Document downloaded to: {local_path}")
        return context

    def can_handle(self, context: ProcessContext) -> bool:
        return context.minio_path is not None and context.local_path is None
