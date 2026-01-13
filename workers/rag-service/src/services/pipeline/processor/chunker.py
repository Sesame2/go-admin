"""
文档分片步骤
集成 pike-rag 的 RecursiveSentenceSplitter 逻辑
"""

from typing import List
from copy import deepcopy

from langchain_core.documents import Document

from core.logger import get_logger
from services.pipeline.processor.base import PipelineStep, ProcessContext, ChunkData

logger = get_logger("pipeline")


class DocumentChunker(PipelineStep):
    """
    文档分片器
    集成 pike-rag 的 RecursiveSentenceSplitter 逻辑
    按句子数量分片，支持重叠
    """

    name = "DocumentChunker"

    def __init__(
        self,
        lang: str = "zh",
        chunk_size: int = 12,  # 每个 chunk 的句子数
        chunk_overlap: int = 4,  # 重叠句子数
        use_spacy: bool = True,  # 是否使用 spaCy 分句
    ):
        self.lang = lang
        self.chunk_size = chunk_size
        self.chunk_overlap = chunk_overlap
        self.use_spacy = use_spacy
        self._nlp = None
        self._stride = chunk_size - chunk_overlap

    def _ensure_nlp_loaded(self):
        """确保 spaCy 模型已加载"""
        if self._nlp is not None:
            return

        if not self.use_spacy:
            return

        try:
            import spacy

            model_map = {"en": "en_core_web_sm", "zh": "zh_core_web_sm"}
            model_name = model_map.get(self.lang, "en_core_web_sm")

            try:
                self._nlp = spacy.load(model_name)
            except OSError:
                logger.warning(f"Downloading spaCy model: {model_name}")
                spacy.cli.download(model_name)
                self._nlp = spacy.load(model_name)

            self._nlp.max_length = 4000000

        except ImportError:
            logger.warning("spaCy not available, falling back to simple splitting")
            self.use_spacy = False

    async def process(self, context: ProcessContext) -> ProcessContext:
        """执行文档分片"""
        logger.info(
            f"Chunking documents with size={self.chunk_size}, overlap={self.chunk_overlap}"
        )

        documents = context.raw_documents or context.parsed_documents
        if not documents:
            logger.warning("No documents to chunk")
            return context

        self._ensure_nlp_loaded()

        chunks: List[ChunkData] = []
        chunk_index = 0

        for doc in documents:
            doc_chunks = self._split_document(doc)

            for content in doc_chunks:
                chunks.append(
                    ChunkData(
                        content=content,
                        chunk_index=chunk_index,
                        metadata=deepcopy(doc.metadata),
                    )
                )
                chunk_index += 1

        context.chunks = chunks
        logger.info(f"Created {len(chunks)} chunks")

        return context

    def can_handle(self, context: ProcessContext) -> bool:
        has_docs = (
            context.raw_documents is not None or context.parsed_documents is not None
        )
        return has_docs and context.chunks is None

    def _split_document(self, doc: Document) -> List[str]:
        """分割单个文档"""
        text = doc.page_content

        if not text or not text.strip():
            return []

        if self.use_spacy and self._nlp:
            return self._split_with_spacy(text)
        else:
            return self._split_simple(text)

    def _split_with_spacy(self, text: str) -> List[str]:
        """使用 spaCy 按句子分片"""
        doc = self._nlp(text)
        sentences = [sent.text.strip() for sent in doc.sents if sent.text.strip()]

        if not sentences:
            return [text] if text.strip() else []

        segments = []
        for i in range(0, len(sentences), self._stride):
            segment = " ".join(sentences[i : i + self.chunk_size])
            segments.append(segment)
            if i + self.chunk_size >= len(sentences):
                break

        return segments

    def _split_simple(self, text: str) -> List[str]:
        """简单分片（按标点分句）"""
        import re

        sentences = re.split(r"[。！？.!?]+", text)
        sentences = [s.strip() for s in sentences if s.strip()]

        if not sentences:
            return [text] if text.strip() else []

        segments = []
        for i in range(0, len(sentences), self._stride):
            segment = "。".join(sentences[i : i + self.chunk_size])
            segments.append(segment)
            if i + self.chunk_size >= len(sentences):
                break

        return segments
