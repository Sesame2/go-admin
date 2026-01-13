"""
数据持久化步骤
将 chunks 和 atom_questions 保存到 PostgreSQL (pgvector)
"""

from pathlib import Path

from core.logger import get_logger
from services.pipeline.processor.base import PipelineStep, ProcessContext

logger = get_logger("pipeline")


class VectorStorePersister(PipelineStep):
    """
    向量数据库持久化
    将 chunks 和 atom_questions 保存到 PostgreSQL (pgvector)
    """

    name = "VectorStorePersister"

    def __init__(self, db_session, batch_size: int = 100):
        self.db_session = db_session
        self.batch_size = batch_size

    async def process(self, context: ProcessContext) -> ProcessContext:
        """保存到数据库"""
        from models.chunk import DocumentChunk, AtomQuestion

        chunk_id_map = {}  # chunk_index -> chunk_id

        # 保存 chunks
        if context.chunks:
            logger.info(f"Persisting {len(context.chunks)} chunks to database")

            chunk_models = []

            for chunk in context.chunks:
                chunk_model = DocumentChunk(
                    document_id=context.document_id,
                    knowledge_base_id=context.knowledge_base_id,
                    content=chunk.content,
                    chunk_index=chunk.chunk_index,
                    start_char=chunk.start_char,
                    end_char=chunk.end_char,
                    embedding=chunk.embedding,
                    meta_data=chunk.metadata,
                )
                chunk_models.append(chunk_model)

            self.db_session.add_all(chunk_models)
            await self.db_session.flush()  # 获取 ID

            # 建立 chunk_index -> chunk_id 映射
            for model in chunk_models:
                chunk_id_map[model.chunk_index] = model.id

            logger.info(f"Persisted {len(chunk_models)} chunks")

        # 保存原子问题
        if context.atom_questions:
            logger.info(f"Persisting {len(context.atom_questions)} atom questions")

            aq_models = []
            for aq in context.atom_questions:
                chunk_id = chunk_id_map.get(aq.chunk_index)
                if not chunk_id:
                    logger.warning(
                        f"Chunk not found for atom question at index {aq.chunk_index}"
                    )
                    continue

                aq_model = AtomQuestion(
                    chunk_id=chunk_id,
                    document_id=context.document_id,
                    knowledge_base_id=context.knowledge_base_id,
                    question=aq.question,
                    embedding=aq.embedding,
                    question_type=aq.question_type,
                    meta_data=aq.metadata,
                )
                aq_models.append(aq_model)

            self.db_session.add_all(aq_models)
            logger.info(f"Persisted {len(aq_models)} atom questions")

        # 提交事务
        await self.db_session.commit()
        logger.info("Database transaction committed")

        return context

    def can_handle(self, context: ProcessContext) -> bool:
        return context.chunks is not None


class LocalFileCleanup(PipelineStep):
    """清理本地临时文件"""

    name = "LocalFileCleanup"

    async def process(self, context: ProcessContext) -> ProcessContext:
        """删除本地临时文件"""
        if context.local_path:
            try:
                Path(context.local_path).unlink(missing_ok=True)
                logger.info(f"Cleaned up local file: {context.local_path}")
            except Exception as e:
                logger.warning(f"Failed to cleanup local file: {e}")

        return context

    def can_handle(self, context: ProcessContext) -> bool:
        return context.local_path is not None
