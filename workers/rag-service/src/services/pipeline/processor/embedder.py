"""
向量嵌入生成步骤
"""

from typing import List, Any

from core.logger import get_logger
from services.pipeline.processor.base import PipelineStep, ProcessContext

logger = get_logger("pipeline")


class EmbeddingGenerator(PipelineStep):
    """向量嵌入生成器"""

    name = "EmbeddingGenerator"

    def __init__(
        self,
        embedding_client: Any = None,
        embed_chunks: bool = True,
        embed_questions: bool = True,
        batch_size: int = 10,
    ):
        self.embedding_client = embedding_client
        self.embed_chunks = embed_chunks
        self.embed_questions = embed_questions
        self.batch_size = batch_size

    async def process(self, context: ProcessContext) -> ProcessContext:
        """生成向量嵌入"""

        # 为 chunks 生成嵌入
        if self.embed_chunks and context.chunks:
            logger.info(f"Generating embeddings for {len(context.chunks)} chunks")
            texts = [chunk.content for chunk in context.chunks]
            embeddings = await self._batch_embed(texts)

            for chunk, embedding in zip(context.chunks, embeddings):
                chunk.embedding = embedding

        # 为原子问题生成嵌入
        if self.embed_questions and context.atom_questions:
            logger.info(
                f"Generating embeddings for {len(context.atom_questions)} atom questions"
            )
            questions = [aq.question for aq in context.atom_questions]
            embeddings = await self._batch_embed(questions)

            for aq, embedding in zip(context.atom_questions, embeddings):
                aq.embedding = embedding

        return context

    def can_handle(self, context: ProcessContext) -> bool:
        has_chunks = context.chunks is not None
        has_questions = context.atom_questions is not None
        needs_embedding = False

        if has_chunks and self.embed_chunks:
            needs_embedding = any(c.embedding is None for c in context.chunks)
        if has_questions and self.embed_questions:
            needs_embedding = needs_embedding or any(
                q.embedding is None for q in context.atom_questions
            )

        return needs_embedding

    async def _batch_embed(self, texts: List[str]) -> List[List[float]]:
        """
        批量生成嵌入
        使用 embedding_client 的 embed_documents 方法
        """
        if not self.embedding_client:
            logger.warning("No embedding client available, returning zero vectors")
            # 使用默认维度 1024
            return [[0.0] * 1024 for _ in texts]

        try:
            # embedding_client.embed_documents 内部已经处理了批量逻辑
            embeddings = await self.embedding_client.embed_documents(texts)
            logger.info(f"Generated {len(embeddings)} embeddings")
            return embeddings
        except Exception as e:
            logger.error(f"Failed to generate embeddings: {e}")
            raise


class MockEmbeddingGenerator(PipelineStep):
    """
    Mock 向量嵌入生成器
    用于测试场景，生成固定维度的随机向量
    """

    name = "MockEmbeddingGenerator"

    def __init__(
        self,
        dimension: int = 1536,
        embed_chunks: bool = True,
        embed_questions: bool = True,
    ):
        self.dimension = dimension
        self.embed_chunks = embed_chunks
        self.embed_questions = embed_questions

    async def process(self, context: ProcessContext) -> ProcessContext:
        """生成模拟的向量嵌入"""
        import hashlib

        def text_to_mock_embedding(text: str) -> List[float]:
            """根据文本生成确定性的模拟向量"""
            # 使用文本哈希生成确定性的"伪随机"向量
            hash_bytes = hashlib.sha256(text.encode()).digest()
            # 扩展到目标维度
            embedding = []
            for i in range(self.dimension):
                byte_idx = i % len(hash_bytes)
                # 归一化到 [-1, 1]
                value = (hash_bytes[byte_idx] / 255.0) * 2 - 1
                embedding.append(value)
            return embedding

        # 为 chunks 生成嵌入
        if self.embed_chunks and context.chunks:
            logger.info(f"Generating mock embeddings for {len(context.chunks)} chunks")
            for chunk in context.chunks:
                chunk.embedding = text_to_mock_embedding(chunk.content)

        # 为原子问题生成嵌入
        if self.embed_questions and context.atom_questions:
            logger.info(
                f"Generating mock embeddings for {len(context.atom_questions)} atom questions"
            )
            for aq in context.atom_questions:
                aq.embedding = text_to_mock_embedding(aq.question)

        return context

    def can_handle(self, context: ProcessContext) -> bool:
        has_chunks = context.chunks is not None
        has_questions = context.atom_questions is not None
        needs_embedding = False

        if has_chunks and self.embed_chunks:
            needs_embedding = any(c.embedding is None for c in context.chunks)
        if has_questions and self.embed_questions:
            needs_embedding = needs_embedding or any(
                q.embedding is None for q in context.atom_questions
            )

        return needs_embedding
