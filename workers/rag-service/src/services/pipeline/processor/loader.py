"""
文档加载步骤
支持: PDF, DOCX, PPTX, Markdown, TXT
"""

from typing import List

from langchain_core.documents import Document

from core.logger import get_logger
from services.pipeline.processor.base import PipelineStep, ProcessContext

logger = get_logger("pipeline")


class DocumentLoader(PipelineStep):
    """
    加载文档内容
    支持: PDF, DOCX, PPTX, Markdown, TXT
    """

    name = "DocumentLoader"

    SUPPORTED_TYPES = {"pdf", "docx", "pptx", "markdown", "md", "txt"}

    async def process(self, context: ProcessContext) -> ProcessContext:
        """加载文档"""
        file_path = context.local_path
        file_type = context.document_type.lower()

        logger.info(f"Loading document: {file_path} (type: {file_type})")

        documents = await self._load_document(file_path, file_type)

        # 添加元数据
        for doc in documents:
            doc.metadata.update(
                {
                    "filename": context.filename,
                    "document_id": str(context.document_id),
                    "knowledge_base_id": str(context.knowledge_base_id),
                }
            )

        context.raw_documents = documents
        logger.info(f"Loaded {len(documents)} document(s)")

        return context

    def can_handle(self, context: ProcessContext) -> bool:
        return (
            context.local_path is not None
            and context.document_type.lower() in self.SUPPORTED_TYPES
            and context.raw_documents is None
        )

    async def _load_document(self, file_path: str, file_type: str) -> List[Document]:
        """根据文件类型加载文档"""

        if file_type == "pdf":
            return await self._load_pdf(file_path)
        elif file_type == "docx":
            return await self._load_docx(file_path)
        elif file_type == "pptx":
            return await self._load_pptx(file_path)
        elif file_type in ("markdown", "md"):
            return await self._load_markdown(file_path)
        elif file_type == "txt":
            return await self._load_txt(file_path)
        else:
            raise ValueError(f"Unsupported file type: {file_type}")

    async def _load_pdf(self, file_path: str) -> List[Document]:
        """加载 PDF 文档"""
        # TODO: 集成 MinerU 进行版面解析
        try:
            from langchain_community.document_loaders import PyPDFLoader

            loader = PyPDFLoader(file_path)
            return loader.load()
        except ImportError:
            logger.warning("PyPDFLoader not available, using basic loader")
            return [Document(page_content="", metadata={"source": file_path})]

    async def _load_docx(self, file_path: str) -> List[Document]:
        """加载 Word 文档"""
        try:
            from langchain_community.document_loaders import Docx2txtLoader

            loader = Docx2txtLoader(file_path)
            return loader.load()
        except ImportError:
            logger.warning("Docx2txtLoader not available")
            return [Document(page_content="", metadata={"source": file_path})]

    async def _load_pptx(self, file_path: str) -> List[Document]:
        """加载 PPT 文档"""
        try:
            from langchain_community.document_loaders import (
                UnstructuredPowerPointLoader,
            )

            loader = UnstructuredPowerPointLoader(file_path)
            return loader.load()
        except ImportError:
            logger.warning("UnstructuredPowerPointLoader not available")
            return [Document(page_content="", metadata={"source": file_path})]

    async def _load_markdown(self, file_path: str) -> List[Document]:
        """加载 Markdown 文档"""
        with open(file_path, "r", encoding="utf-8") as f:
            content = f.read()
        return [Document(page_content=content, metadata={"source": file_path})]

    async def _load_txt(self, file_path: str) -> List[Document]:
        """加载纯文本文档"""
        with open(file_path, "r", encoding="utf-8") as f:
            content = f.read()
        return [Document(page_content=content, metadata={"source": file_path})]


class MarkdownContentLoader(PipelineStep):
    """
    直接从字符串内容加载 Markdown
    用于测试场景，无需从文件加载
    """

    name = "MarkdownContentLoader"

    async def process(self, context: ProcessContext) -> ProcessContext:
        """从 raw_content 字段加载 markdown 内容"""
        content = context.metadata.get("raw_content", "")

        if not content:
            logger.warning("No raw_content found in metadata")
            return context

        logger.info(f"Loading markdown content from memory ({len(content)} chars)")

        doc = Document(
            page_content=content,
            metadata={
                "filename": context.filename,
                "document_id": str(context.document_id),
                "knowledge_base_id": str(context.knowledge_base_id),
                "source": "memory",
            },
        )

        context.raw_documents = [doc]
        logger.info("Loaded 1 document from memory")

        return context

    def can_handle(self, context: ProcessContext) -> bool:
        has_content = context.metadata.get("raw_content") is not None
        return has_content and context.raw_documents is None
