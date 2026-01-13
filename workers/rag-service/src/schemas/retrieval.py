"""
检索相关的数据模型
"""

from typing import List, Optional, Dict, Any
from pydantic import BaseModel, Field
from uuid import UUID
from enum import Enum


class RetrievalEventType(str, Enum):
    """SSE事件类型"""

    # 流程状态
    START = "start"
    COMPLETE = "complete"
    ERROR = "error"

    # 检索步骤
    DECOMPOSE_START = "decompose_start"
    DECOMPOSE_RESULT = "decompose_result"

    RETRIEVE_START = "retrieve_start"
    RETRIEVE_RESULT = "retrieve_result"

    SELECT_START = "select_start"
    SELECT_RESULT = "select_result"

    ANSWER_START = "answer_start"
    ANSWER_CHUNK = "answer_chunk"  # 流式回答
    ANSWER_COMPLETE = "answer_complete"


class AtomInfo(BaseModel):
    """原子问题检索信息"""

    atom_query: str = Field(..., description="用于检索的查询")
    atom_question: str = Field(..., description="原子问题")
    source_chunk_id: str = Field(..., description="来源chunk ID")
    source_chunk_content: str = Field(..., description="来源chunk内容")
    source_chunk_title: Optional[str] = Field(None, description="来源chunk标题")
    retrieval_score: float = Field(..., description="检索相似度分数")


class DecomposeResult(BaseModel):
    """问题分解结果"""

    should_decompose: bool = Field(..., description="是否需要分解")
    thinking: str = Field(..., description="分解思考过程")
    sub_questions: List[str] = Field(default_factory=list, description="子问题列表")


class RetrieveResult(BaseModel):
    """检索结果"""

    atom_candidates: List[AtomInfo] = Field(
        default_factory=list, description="候选原子信息"
    )
    retrieval_method: str = Field("atom", description="检索方式: atom/chunk/backup")


class SelectResult(BaseModel):
    """选择结果"""

    selected: bool = Field(..., description="是否选中")
    thinking: str = Field(..., description="选择思考过程")
    chosen_info: Optional[AtomInfo] = Field(None, description="选中的信息")


class AnswerResult(BaseModel):
    """回答结果"""

    thinking: str = Field("", description="回答思考过程")
    answer: str = Field(..., description="最终答案")
    references: List[AtomInfo] = Field(default_factory=list, description="引用的上下文")


class SSEEvent(BaseModel):
    """SSE事件数据"""

    event_type: RetrievalEventType
    step_id: str = Field("", description="步骤标识，如 Sub1, Sub2")
    data: Dict[str, Any] = Field(default_factory=dict)
    message: str = Field("", description="可读消息")


class RetrievalRequest(BaseModel):
    """检索请求"""

    question: str = Field(..., description="用户问题", min_length=1)
    knowledge_base_id: str = Field(..., description="知识库ID")
    max_iterations: int = Field(5, ge=1, le=10, description="最大迭代次数")
    stream_answer: bool = Field(True, description="是否流式输出答案")


class RetrievalResponse(BaseModel):
    """检索响应（非流式）"""

    question: str
    answer: str
    thinking: str
    references: List[AtomInfo]
    decomposition_steps: List[Dict[str, Any]]
