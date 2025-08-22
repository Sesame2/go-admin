from dataclasses import dataclass
from langchain.text_splitter import RecursiveCharacterTextSplitter


@dataclass
class TextChunk:
    text: str
    metadata: dict


class TextChunker:
    """
    使用 LangChain 将文档内容分割成适合向量数据库和语言模型使用的块
    """

    def __init__(self, chunk_size=1000, chunk_overlap=200):
        self.splitter = RecursiveCharacterTextSplitter(
            chunk_size=chunk_size,
            chunk_overlap=chunk_overlap,
            length_function=len,
            separators=["\n\n", "\n", " ", ""],
        )

    def chunk_text(self, text: str, metadata: dict) -> list[TextChunk]:
        chunks = self.splitter.split_text(text)
        return [TextChunk(text=chunk, metadata=metadata) for chunk in chunks]


