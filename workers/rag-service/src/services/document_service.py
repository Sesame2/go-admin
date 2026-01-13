from schemas.document import DocumentAnalysisRequest
from core.logger import get_logger
from core.dependencies import DbSessionDep, MinioClientDep


class DocumentService:
    def __init__(
        self,
        db=None,
        minio_client=None,
        llm_client=None,
        embedding_client=None,
        vector_store=None,
    ):
        self.db = db
        self.logger = get_logger("service")

    async def analysis_document(self, req: DocumentAnalysisRequest):
        """
        analysis_document
        1. 根据传来的Minio路径，从Minio中下载文档
        2. 判断文档类型，把需要版面解析的文档调用MinerU中解析：第一版本先只支持pdf，word，ppt，markdown常见的格式
        3. 将结构化后的文档进行切片，根据pike-rag的workflow进行原子化处理
        4.
        """
        self.logger.info(
            "开始分析文档", extra={"document_id": getattr(req, "document_id", None)}
        )


async def get_document_service(
    db: DbSessionDep, minio_client: MinioClientDep
) -> DocumentService:
    return DocumentService(db=db, minio_client=minio_client)
