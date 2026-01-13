"""
文档处理 Worker
消费 RabbitMQ 消息并执行文档解析、分片、向量化
"""

import asyncio
import sys
from pathlib import Path
from datetime import datetime, timezone
from uuid import UUID

# 添加 src 目录到 Python 路径
sys.path.insert(0, str(Path(__file__).parent))

from core.config import get_settings
from core.logger import get_logger
from core.rabbitmq import RabbitMQConsumer, DocumentParseMessage
from core.database import AsyncSessionLocal
from core.minio_client import MinioClient
from core.qwen_client import QwenLLMClient
from core.qwen_embedding import QwenEmbeddingClient
from models.document import Document, DocumentStatus
from services.pipeline.builder import PipelineBuilder
from services.pipeline.processor.base import ProcessContext, ProcessStatus

logger = get_logger("worker")


async def update_document_status(
    document_id: str,
    status: DocumentStatus,
    error_message: str = None,
    chunk_count: int = None,
    atom_question_count: int = None,
):
    """更新文档状态 - 使用原生 SQL 绕过 SQLAlchemy ENUM 类型处理"""
    async with AsyncSessionLocal() as session:
        from sqlalchemy import text
        
        # 构建动态 SQL
        set_parts = ["status = :status", "updated_at = :updated_at"]
        params = {
            "status": status.value,  # 小写值: 'downloading', 'completed', etc.
            "updated_at": datetime.now(timezone.utc),
            "document_id": document_id,
        }
        
        if error_message:
            set_parts.append("error_message = :error_message")
            params["error_message"] = error_message
        if chunk_count is not None:
            set_parts.append("chunk_count = :chunk_count")
            params["chunk_count"] = chunk_count
        if atom_question_count is not None:
            set_parts.append("atom_question_count = :atom_question_count")
            params["atom_question_count"] = atom_question_count
        if status == DocumentStatus.COMPLETED:
            set_parts.append("processed_at = :processed_at")
            params["processed_at"] = datetime.now(timezone.utc)
        
        sql = f"UPDATE documents SET {', '.join(set_parts)} WHERE id = :document_id"
        await session.execute(text(sql), params)
        await session.commit()
        
        logger.info(f"文档状态更新: {document_id} -> {status.value}")


async def process_document(message: dict):
    """处理文档"""
    settings = get_settings()
    
    try:
        msg = DocumentParseMessage(message)
        logger.info(f"开始处理文档: {msg.filename}")
        
        # 转换 UUID
        document_id = UUID(str(msg.document_id))
        knowledge_base_id = UUID(str(msg.knowledge_base_id))
        
        # 更新状态为下载中
        await update_document_status(str(document_id), DocumentStatus.DOWNLOADING)
        
        # 初始化客户端
        minio_client = MinioClient(
            settings=settings
        )
        
        llm_client = None
        embedding_client = None
        
        if settings.QWEN_API_KEY:
            llm_client = QwenLLMClient(
                api_key=settings.QWEN_API_KEY,
                model=settings.QWEN_LLM_MODEL,
            )
            embedding_client = QwenEmbeddingClient(
                api_key=settings.QWEN_API_KEY,
                model=settings.QWEN_EMBEDDING_MODEL,
                dimensions=settings.QWEN_EMBEDDING_DIMENSIONS,
            )
        
        # 创建数据库会话
        async with AsyncSessionLocal() as db_session:
            # 创建处理管道
            pipeline = PipelineBuilder.create_default_pipeline(
                minio_client=minio_client,
                db_session=db_session,
                embedding_client=embedding_client,
                llm_client=llm_client,
                lang="zh",
                chunk_size=settings.CHUNK_SIZE,
                chunk_overlap=settings.CHUNK_OVERLAP,
                enable_tagging=llm_client is not None,
            )
            
            # 创建处理上下文
            context = ProcessContext(
                document_id=document_id,
                knowledge_base_id=knowledge_base_id,
                minio_path=msg.minio_path,
                document_type=msg.file_type,
                filename=msg.filename,
            )
            
            # 执行管道
            result = await pipeline.execute(context)
            
            if result.status == ProcessStatus.COMPLETED:
                # 更新文档状态为完成
                chunk_count = len(result.chunks) if result.chunks else 0
                atom_count = len(result.atom_questions) if result.atom_questions else 0
                
                await update_document_status(
                    str(document_id),
                    DocumentStatus.COMPLETED,
                    chunk_count=chunk_count,
                    atom_question_count=atom_count,
                )
                logger.info(f"文档处理完成: {msg.filename}, chunks={chunk_count}, atoms={atom_count}")
            else:
                await update_document_status(
                    str(document_id),
                    DocumentStatus.FAILED,
                    error_message=result.error_message or "处理失败",
                )
                logger.error(f"文档处理失败: {msg.filename}, error={result.error_message}")
                
    except Exception as e:
        logger.error(f"文档处理异常: {e}", exc_info=True)
        await update_document_status(
            message.get("document_id", ""),
            DocumentStatus.FAILED,
            error_message=str(e),
        )


async def main():
    """主函数"""
    logger.info("启动文档处理 Worker...")
    
    consumer = RabbitMQConsumer()
    
    try:
        await consumer.connect()
        await consumer.consume(process_document)
    except KeyboardInterrupt:
        logger.info("收到退出信号")
    except Exception as e:
        logger.error(f"Worker 异常: {e}", exc_info=True)
    finally:
        await consumer.close()


if __name__ == "__main__":
    asyncio.run(main())
