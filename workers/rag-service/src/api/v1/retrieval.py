"""
SSE 检索 API
实现类似 pike-rag QaDecompositionWorkflow 的检索能力
"""

from uuid import UUID
from typing import Optional

from fastapi import APIRouter, Depends, HTTPException, Query
from fastapi.responses import StreamingResponse

from core.logger import get_logger
from core.dependencies import DbSessionDep
from core.config import get_settings
from core.qwen_client import QwenLLMClient
from core.qwen_embedding import QwenEmbeddingClient
from services.qa_workflow import QaDecompositionWorkflow
from schemas.retrieval import RetrievalRequest

router = APIRouter()
logger = get_logger("retrieval_api")


@router.post("/retrieve/stream")
async def retrieve_stream(
    request: RetrievalRequest,
    db: DbSessionDep,
):
    """
    SSE流式检索接口

    实现类似 pike-rag QaDecompositionWorkflow 的检索能力：
    1. 问题分解：分析问题，生成子问题
    2. 检索：用子问题检索相关原子问题和chunk
    3. 选择：选择最相关的信息
    4. 回答：基于收集的上下文生成最终答案

    所有中间状态通过SSE事件通知前端
    """
    logger.info(f"SSE retrieve request: question={request.question[:50]}...")

    try:
        # 验证知识库ID
        try:
            kb_id = UUID(request.knowledge_base_id)
        except ValueError:
            raise HTTPException(status_code=400, detail="无效的知识库ID")

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

        # 创建工作流
        workflow = QaDecompositionWorkflow(
            db_session=db,
            llm_client=llm_client,
            embedding_client=embedding_client,
            knowledge_base_id=kb_id,
            max_iterations=request.max_iterations,
        )

        # 返回SSE流
        async def event_generator():
            async for event in workflow.run_stream(
                question=request.question,
                stream_answer=request.stream_answer,
            ):
                yield event

        return StreamingResponse(
            event_generator(),
            media_type="text/event-stream",
            headers={
                "Cache-Control": "no-cache",
                "Connection": "keep-alive",
                "X-Accel-Buffering": "no",
            },
        )

    except Exception as e:
        logger.error(f"SSE retrieve failed: {e}", exc_info=True)
        raise HTTPException(status_code=500, detail=f"检索失败: {str(e)}")


@router.get("/retrieve/stream")
async def retrieve_stream_get(
    db: DbSessionDep,
    question: str = Query(..., description="用户问题"),
    knowledge_base_id: str = Query(..., description="知识库ID"),
    max_iterations: int = Query(5, ge=1, le=10, description="最大迭代次数"),
    stream_answer: bool = Query(True, description="是否流式输出答案"),
):
    """
    SSE流式检索接口（GET方式，方便浏览器测试）
    """
    request = RetrievalRequest(
        question=question,
        knowledge_base_id=knowledge_base_id,
        max_iterations=max_iterations,
        stream_answer=stream_answer,
    )
    return await retrieve_stream(request, db)
