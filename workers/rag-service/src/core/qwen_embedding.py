"""
通义千问 Embedding 客户端
用于生成文本向量
"""

from typing import List, Optional
import numpy as np
from openai import AsyncOpenAI

from core.logger import get_logger
from core.config import get_settings

logger = get_logger("qwen_embedding")
settings = get_settings()


class QwenEmbeddingClient:
    """
    通义千问 Embedding 客户端（使用 OpenAI 兼容接口）
    支持 text-embedding-v3 和 text-embedding-v4
    """

    def __init__(
        self,
        api_key: Optional[str] = None,
        base_url: str = "https://dashscope.aliyuncs.com/compatible-mode/v1",
        model: str = "text-embedding-v3",
        dimensions: int = 1024,
    ):
        """
        初始化 Qwen Embedding 客户端

        Args:
            api_key: API Key
            base_url: API 基础 URL
            model: 模型名称 (text-embedding-v3 或 text-embedding-v4)
            dimensions: 向量维度 (64, 128, 256, 512, 768, 1024, 1536, 2048)
        """
        self.api_key = api_key or settings.QWEN_API_KEY
        self.base_url = base_url
        self.model = model
        self.dimensions = dimensions

        # 验证维度
        valid_dims = [64, 128, 256, 512, 768, 1024, 1536, 2048]
        if dimensions not in valid_dims:
            raise ValueError(
                f"Invalid dimensions: {dimensions}. Must be one of {valid_dims}"
            )

        self.client = AsyncOpenAI(
            api_key=self.api_key,
            base_url=self.base_url,
        )

        logger.info(
            f"Initialized QwenEmbeddingClient with model: {model}, dimensions: {dimensions}"
        )

    async def embed_query(self, text: str) -> List[float]:
        """
        为单个查询文本生成向量

        Args:
            text: 查询文本

        Returns:
            向量（float 列表）
        """
        try:
            response = await self.client.embeddings.create(
                model=self.model,
                input=text,
                dimensions=self.dimensions,
                encoding_format="float",
            )

            embedding = response.data[0].embedding
            logger.debug(
                f"Generated embedding for query text, dimension: {len(embedding)}"
            )
            return embedding

        except Exception as e:
            logger.error(f"Failed to generate embedding for query: {e}")
            raise

    async def embed_documents(self, texts: List[str]) -> List[List[float]]:
        """
        为多个文档生成向量（批量处理）

        Args:
            texts: 文本列表

        Returns:
            向量列表
        """
        if not texts:
            return []

        # text-embedding-v3/v4 最多支持 10 条文本
        max_batch_size = 10

        all_embeddings = []

        for i in range(0, len(texts), max_batch_size):
            batch = texts[i : i + max_batch_size]

            try:
                response = await self.client.embeddings.create(
                    model=self.model,
                    input=batch,
                    dimensions=self.dimensions,
                    encoding_format="float",
                )

                # 按索引排序，确保顺序正确
                embeddings = [
                    item.embedding
                    for item in sorted(response.data, key=lambda x: x.index)
                ]
                all_embeddings.extend(embeddings)

                logger.debug(
                    f"Generated embeddings for batch {i // max_batch_size + 1}, size: {len(batch)}"
                )

            except Exception as e:
                logger.error(
                    f"Failed to generate embeddings for batch {i // max_batch_size + 1}: {e}"
                )
                raise

        logger.info(f"Generated embeddings for {len(all_embeddings)} documents")
        return all_embeddings

    async def embed_batch(
        self,
        texts: List[str],
        batch_size: int = 10,
    ) -> List[List[float]]:
        """
        批量生成向量（与 embed_documents 相同，保持接口兼容性）

        Args:
            texts: 文本列表
            batch_size: 批次大小（最大 10）

        Returns:
            向量列表
        """
        return await self.embed_documents(texts)

    def get_dimensions(self) -> int:
        """获取向量维度"""
        return self.dimensions
