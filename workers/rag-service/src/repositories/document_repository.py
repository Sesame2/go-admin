"""
Document Repository - 文档数据访问层
"""

from typing import List, Optional
from uuid import UUID
from datetime import datetime

from sqlalchemy import select, delete, update, func
from sqlalchemy.ext.asyncio import AsyncSession

from models.document import Document, DocumentStatus
from core.logger import get_logger

logger = get_logger("repository")


class DocumentRepository:
    """文档数据访问层"""

    def __init__(self, db_session: AsyncSession):
        self.db = db_session

    async def create(self, document: Document) -> Document:
        """创建文档"""
        self.db.add(document)
        await self.db.flush()
        return document

    async def get_by_id(self, document_id: UUID) -> Optional[Document]:
        """根据 ID 获取文档"""
        stmt = select(Document).where(Document.id == document_id)
        result = await self.db.execute(stmt)
        return result.scalar_one_or_none()

    async def get_by_knowledge_base(
        self,
        knowledge_base_id: UUID,
        status: Optional[DocumentStatus] = None,
        skip: int = 0,
        limit: int = 100,
    ) -> List[Document]:
        """获取知识库的文档列表"""
        stmt = (
            select(Document)
            .where(Document.knowledge_base_id == knowledge_base_id)
            .order_by(Document.created_at.desc())
            .offset(skip)
            .limit(limit)
        )

        if status:
            stmt = stmt.where(Document.status == status)

        result = await self.db.execute(stmt)
        return list(result.scalars().all())

    async def update_status(
        self,
        document_id: UUID,
        status: DocumentStatus,
        error_message: Optional[str] = None,
    ) -> bool:
        """更新文档状态"""
        values = {"status": status}

        if error_message:
            values["error_message"] = error_message

        if status == DocumentStatus.COMPLETED:
            values["processed_at"] = datetime.utcnow()

        stmt = update(Document).where(Document.id == document_id).values(**values)

        result = await self.db.execute(stmt)
        return result.rowcount > 0

    async def update_chunk_stats(
        self, document_id: UUID, chunk_count: int, atom_question_count: int
    ) -> bool:
        """更新分片统计信息"""
        stmt = (
            update(Document)
            .where(Document.id == document_id)
            .values(chunk_count=chunk_count, atom_question_count=atom_question_count)
        )

        result = await self.db.execute(stmt)
        return result.rowcount > 0

    async def delete(self, document_id: UUID) -> bool:
        """删除文档"""
        stmt = delete(Document).where(Document.id == document_id)
        result = await self.db.execute(stmt)
        return result.rowcount > 0

    async def count_by_knowledge_base(
        self, knowledge_base_id: UUID, status: Optional[DocumentStatus] = None
    ) -> int:
        """统计知识库的文档数量"""
        stmt = select(func.count(Document.id)).where(
            Document.knowledge_base_id == knowledge_base_id
        )

        if status:
            stmt = stmt.where(Document.status == status)

        result = await self.db.execute(stmt)
        return result.scalar() or 0

    async def get_pending_documents(self, limit: int = 10) -> List[Document]:
        """获取待处理的文档（用于后台任务）"""
        stmt = (
            select(Document)
            .where(Document.status == DocumentStatus.PENDING)
            .order_by(Document.created_at)
            .limit(limit)
        )

        result = await self.db.execute(stmt)
        return list(result.scalars().all())
