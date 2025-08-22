from typing import List, Dict, Any
from llama_index.embeddings.dashscope import (
    DashScopeEmbedding,
    DashScopeTextEmbeddingModels,
    DashScopeTextEmbeddingType,
)

from config import Settings


def get_embedding_model():
    """获取嵌入模型实例"""

    return DashScopeEmbedding(
        model_name=DashScopeTextEmbeddingModels.TEXT_EMBEDDING_V3,
        text_type=DashScopeTextEmbeddingType.TEXT_TYPE_DOCUMENT,
        api_key=Settings.qwen_api_key,
    )


class EmbeddingClient:
    """嵌入服务类，用于处理文档块的嵌入生成"""

    def __init__(self):
        self.embedding_model = get_embedding_model()

    def get_text_embedding(self, text: str) -> List[float]:
        """获取单个文本的嵌入向量"""
        return self.embedding_model.get_text_embedding(text)

    def get_text_embedding_batch(self, texts: List[str]) -> List[List[float]]:
        """批量获取文本的嵌入向量"""
        return self.embedding_model.get_text_embedding_batch(texts)

    def process_chunks(self, chunks: List[Dict[str, Any]]) -> List[Dict[str, Any]]:
        """处理文本块，添加嵌入向量"""
        texts = [chunk["text"] for chunk in chunks]
        embeddings = self.get_text_embedding_batch(texts)

        result_chunks = []
        for i, chunk in enumerate(chunks):
            if embeddings[i] is not None:
                chunk_with_embedding = chunk.copy()
                chunk_with_embedding["embedding"] = embeddings[i]
                result_chunks.append(chunk_with_embedding)
            else:
                print(f"警告: 无法为文本块 {i} 生成嵌入向量")

        return result_chunks


