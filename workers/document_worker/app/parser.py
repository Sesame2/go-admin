from abc import ABC, abstractmethod
import os
from typing import Any, Dict

from docling.document_converter import DocumentConverter
from langdetect import detect
from pydantic import BaseModel


class DocumentParseResult(BaseModel):
    metadata: Dict[str, Any]
    parser_result: Dict[str, Any]
    language: str


class DocumentParser(ABC):
    """
    文档解析器基类
    """

    @abstractmethod
    def parse(self, file_path: str) -> DocumentParseResult:
        """
        解析文档并返回结构化数据

        Args:
            file_path: 要解析的文档路径

        Returns:
            包含解析结果的字典
        """
        pass

    def extract_metadata(self, file_path: str) -> Dict[str, Any]:
        """提取文档的基本元数据"""
        file_stats = os.stat(file_path)
        file_name = os.path.basename(file_path)
        file_ext = os.path.splitext(file_name)[1].lower()

        return {
            "filename": file_name,
            "extension": file_ext,
            "size_bytes": file_stats.st_size,
            "last_modified": file_stats.st_mtime,
        }

    def detect_language(self, text: str) -> str:
        """检测文本语言"""
        try:
            return detect(text[:1000])  # 使用前1000个字符检测语言
        except:
            return "unknown"


class WordParser(DocumentParser):
    def parse(self, file_path: str) -> DocumentParseResult:
        metadata = self.extract_metadata(file_path)
        converter = DocumentConverter()
        content = converter.convert(file_path)
        parser_result = content.document.export_to_dict()
        plain_text = ""
        for text_node in content.document.texts:
            plain_text += text_node.text + "\n"
        language = self.detect_language(plain_text)
        return DocumentParseResult(
            metadata=metadata, parser_result=parser_result, language=language
        )
