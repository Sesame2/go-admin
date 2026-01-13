from uuid import uuid4, UUID
from datetime import datetime

from fastapi import APIRouter, Depends, HTTPException, UploadFile, File, Form

from core.logger import get_logger
from core.dependencies import DbSessionDep
from schemas.document import (
    DocumentAnalysisRequest,
    MarkdownTestRequest,
    MarkdownTestResponse,
    MarkdownUploadResponse,
    ChunkResponse,
    AtomQuestionResponse,
)
from services.document_service import DocumentService, get_document_service

router = APIRouter()
logger = get_logger("controller")


@router.post("/analysis")
async def analysis_document(
    req: DocumentAnalysisRequest,
    service: DocumentService = Depends(get_document_service),
):
    """
    分析文档（完整流程）

    :param req: 文档分析请求
    :type req: DocumentAnalysisRequest
    """
    logger.info("/analysis called")
    # 实际逻辑留空占位
    return {"status": "ok"}


@router.post("/upload/markdown", response_model=MarkdownUploadResponse)
async def upload_markdown(
    db: DbSessionDep,
    file: UploadFile = File(..., description="Markdown 文件"),
    knowledge_base_id: str = Form(..., description="知识库 ID"),
    title: str = Form(None, description="文档标题（可选）"),
    chunk_size: int = Form(12, ge=1, le=50, description="每个 chunk 的句子数"),
    chunk_overlap: int = Form(4, ge=0, le=20, description="重叠句子数"),
    lang: str = Form("zh", description="语言: zh/en"),
    use_spacy: bool = Form(False, description="是否使用 spaCy 分句"),
    max_atom_questions: int = Form(
        10, ge=1, le=20, description="每个 chunk 生成的最大原子问题数"
    ),
):
    """
    上传 Markdown 文档并完整处理

    此接口会：
    1. 读取上传的 Markdown 文件内容
    2. 进行文档切片
    3. 使用 Qwen LLM 生成原子问题
    4. 使用 Qwen Embedding 生成向量
    5. 存储到 PostgreSQL 数据库

    - 支持配置分片参数
    - 使用真实的 LLM 和 Embedding 服务
    - 持久化到数据库
    """
    from langchain_core.documents import Document
    from services.pipeline.processor import (
        ProcessContext,
        PipelineOrchestrator,
    )
    from services.pipeline.processor.chunker import DocumentChunker
    from services.pipeline.processor.tagger import AtomQuestionTagger
    from services.pipeline.processor.embedder import EmbeddingGenerator
    from services.pipeline.processor.persister import VectorStorePersister
    from core.qwen_client import QwenLLMClient
    from core.qwen_embedding import QwenEmbeddingClient
    from core.config import get_settings
    from models.document import DocumentStatus
    from models.document import Document as Document_model
    from repositories.document_repository import DocumentRepository
    from repositories.chunk_repository import ChunkRepository

    logger.info(f"/upload/markdown called, filename: {file.filename}")

    # 验证文件类型
    if not file.filename.endswith((".md", ".markdown")):
        raise HTTPException(
            status_code=400, detail="只支持 Markdown 文件（.md 或 .markdown）"
        )

    try:
        # 读取文件内容
        content = await file.read()
        content_text = content.decode("utf-8")
        logger.info(f"Read file content, length: {len(content_text)}")

        # 生成文档 ID
        document_id = uuid4()

        # 将 knowledge_base_id 转换为 UUID
        try:
            kb_id = (
                UUID(knowledge_base_id)
                if isinstance(knowledge_base_id, str)
                else knowledge_base_id
            )
        except ValueError:
            # 如果不是有效的 UUID，生成一个新的
            kb_id = uuid4()
            logger.warning(
                f"Invalid knowledge_base_id '{knowledge_base_id}', using generated: {kb_id}"
            )

        # 创建文档记录
        doc_repo = DocumentRepository(db)
        document = await doc_repo.create(
            Document_model(
                id=document_id,
                knowledge_base_id=kb_id,
                filename=file.filename or "untitled.md",
                file_type="markdown",
                minio_path="",  # 本地上传不需要 MinIO 路径
                status=DocumentStatus.CHUNKING,
                meta_data={
                    "upload_type": "local",
                    "content_length": len(content_text),
                },
            )
        )

        logger.info(f"Created document record: {document_id}")

        # 初始化客户端
        settings = get_settings()
        llm_client = QwenLLMClient(
            api_key=settings.QWEN_API_KEY,
            model=settings.QWEN_LLM_MODEL,
        )
        embedding_client = QwenEmbeddingClient(
            api_key=settings.QWEN_API_KEY,
            model=settings.QWEN_EMBEDDING_MODEL,
            dimensions=settings.QWEN_EMBEDDING_DIMENSIONS,
        )

        # 创建处理上下文
        context = ProcessContext(
            document_id=document_id,
            knowledge_base_id=kb_id,
            minio_path="",
            document_type="markdown",
            filename=file.filename or "untitled.md",
            metadata={
                "title": title,
                "raw_content": content_text,
            },
        )

        # 设置 raw_documents
        context.raw_documents = [
            Document(
                page_content=content_text,
                metadata={
                    "filename": file.filename,
                    "document_id": str(document_id),
                    "knowledge_base_id": str(kb_id),
                    "source": "upload_api",
                    "title": title,
                },
            )
        ]

        # 更新状态：chunking
        await doc_repo.update_status(document_id, DocumentStatus.CHUNKING)

        # 构建处理管道
        steps = [
            # 1. 文档切片
            DocumentChunker(
                lang=lang,
                chunk_size=chunk_size,
                chunk_overlap=chunk_overlap,
                use_spacy=use_spacy,
            ),
            # 2. 生成原子问题
            AtomQuestionTagger(
                llm_client=llm_client,
                num_parallel=1,
            ),
            # 3. 生成向量
            EmbeddingGenerator(
                embedding_client=embedding_client,
                batch_size=10,
            ),
            # 4. 存储到数据库
            VectorStorePersister(
                db_session=db,
            ),
        ]

        # 执行管道
        pipeline = PipelineOrchestrator(steps=steps, logger=logger)
        result = await pipeline.execute(context)

        # 更新文档状态和统计
        chunk_count = len(result.chunks or [])
        atom_question_count = len(result.atom_questions or [])

        await doc_repo.update_status(document_id, DocumentStatus.COMPLETED)
        await doc_repo.update_chunk_stats(document_id, chunk_count, atom_question_count)

        logger.info(
            f"Document {document_id} processed successfully. "
            f"Chunks: {chunk_count}, Atom Questions: {atom_question_count}"
        )

        return MarkdownUploadResponse(
            document_id=document_id,
            knowledge_base_id=str(kb_id),
            filename=file.filename or "untitled.md",
            status="completed",
            total_chars=len(content_text),
            chunk_count=chunk_count,
            atom_question_count=atom_question_count,
            message="文档处理成功",
            created_at=datetime.now().isoformat(),
        )

    except Exception as e:
        logger.error(f"Upload markdown processing failed: {e}", exc_info=True)

        # 如果已创建文档，更新状态为失败
        if "document_id" in locals():
            try:
                doc_repo = DocumentRepository(db)
                await doc_repo.update_status(document_id, DocumentStatus.FAILED)
            except:
                pass

        raise HTTPException(status_code=500, detail=f"处理失败: {str(e)}")
