"""
FastAPI 依赖注入类型别名
使用 Annotated 简化路由参数声明
"""
from typing import Annotated
from fastapi import Depends
from sqlalchemy.ext.asyncio import AsyncSession

from core.database import get_db
from core.minio_client import MinioClient, get_minio_client

# 数据库会话依赖
DbSessionDep = Annotated[AsyncSession, Depends(get_db)]

# MinIO 客户端依赖
MinioClientDep = Annotated[MinioClient, Depends(get_minio_client)]
