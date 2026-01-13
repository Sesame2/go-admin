import uvicorn
from pathlib import Path
from contextlib import asynccontextmanager
from fastapi import FastAPI
from fastapi.staticfiles import StaticFiles
from fastapi.middleware.cors import CORSMiddleware

from api.router import api_router_v1
from core.logger import configure_logging, get_logger
from core.database import init_db


configure_logging()
logger = get_logger("core")


@asynccontextmanager
async def lifespan(app: FastAPI):
    """应用生命周期管理"""
    # 启动时初始化数据库
    logger.info("Initializing database...")
    await init_db()
    logger.info("Database initialized successfully")
    yield
    # 关闭时清理资源
    logger.info("Shutting down...")


app = FastAPI(lifespan=lifespan)

# 添加 CORS 中间件（允许前端跨域访问）
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

app.include_router(api_router_v1)



logger.info("FastAPI app initialized and routers registered")

if __name__ == "__main__":
    logger.info("Starting Uvicorn server on 0.0.0.0:8858")
    uvicorn.run(app, host="0.0.0.0", port=8858)
