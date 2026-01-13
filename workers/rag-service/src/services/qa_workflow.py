"""
QA分解检索工作流
实现类似 pike-rag QaDecompositionWorkflow 的功能
支持SSE流式输出中间状态
"""

import json
import asyncio
from typing import List, Dict, Any, AsyncGenerator, Optional
from uuid import UUID
from dataclasses import asdict

from sqlalchemy.ext.asyncio import AsyncSession

from core.logger import get_logger
from core.qwen_client import QwenLLMClient
from core.qwen_embedding import QwenEmbeddingClient
from services.retriever import PgVectorRetriever, AtomRetrievalInfo
from schemas.retrieval import (
    RetrievalEventType,
    SSEEvent,
    AtomInfo,
    DecomposeResult,
    RetrieveResult,
    SelectResult,
    AnswerResult,
)

logger = get_logger("qa_workflow")


def atom_infos_to_context(
    chosen_infos: List[AtomRetrievalInfo], limit: int = 80000
) -> str:
    """将选中的原子信息转换为上下文字符串"""
    context = ""
    chunk_id_set = set()

    for info in chosen_infos:
        if info.source_chunk_id in chunk_id_set:
            continue
        chunk_id_set.add(info.source_chunk_id)

        if info.source_chunk_title:
            context += f"\n标题: {info.source_chunk_title}\n内容: {info.source_chunk}\n"
        else:
            context += f"\n{info.source_chunk}\n"

        if len(context) >= limit:
            break

    return context.strip()


def atom_info_to_schema(info: AtomRetrievalInfo) -> AtomInfo:
    """将 AtomRetrievalInfo 转换为 API schema"""
    return AtomInfo(
        atom_query=info.atom_query,
        atom_question=info.atom,
        source_chunk_id=info.source_chunk_id,
        source_chunk_content=info.source_chunk,
        source_chunk_title=info.source_chunk_title,
        retrieval_score=info.retrieval_score,
    )


class QaDecompositionWorkflow:
    """
    QA分解检索工作流

    工作流程：
    1. 问题分解（Decompose）：分析问题，生成子问题列表
    2. 检索（Retrieve）：用子问题检索相关原子问题和chunk
    3. 选择（Select）：选择最相关的信息
    4. 迭代：重复1-3直到收集足够信息或达到最大迭代次数
    5. 回答（Answer）：基于收集的上下文生成最终答案
    """

    def __init__(
        self,
        db_session: AsyncSession,
        llm_client: QwenLLMClient,
        embedding_client: QwenEmbeddingClient,
        knowledge_base_id: UUID,
        max_iterations: int = 5,
        retrieve_k: int = 5,
        atom_retrieve_k: int = 3,
    ):
        self.db = db_session
        self.llm_client = llm_client
        self.embedding_client = embedding_client
        self.knowledge_base_id = knowledge_base_id
        self.max_iterations = max_iterations

        self.retriever = PgVectorRetriever(
            db_session=db_session,
            embedding_client=embedding_client,
            knowledge_base_id=knowledge_base_id,
            retrieve_k=retrieve_k,
            atom_retrieve_k=atom_retrieve_k,
        )

    def _create_event(
        self,
        event_type: RetrievalEventType,
        step_id: str = "",
        data: Dict[str, Any] = None,
        message: str = "",
    ) -> str:
        """创建SSE事件"""
        event = SSEEvent(
            event_type=event_type,
            step_id=step_id,
            data=data or {},
            message=message,
        )
        return f"event: {event_type.value}\ndata: {json.dumps(event.model_dump(), ensure_ascii=False)}\n\n"

    async def _propose_decomposition(
        self,
        question: str,
        chosen_infos: List[AtomRetrievalInfo],
    ) -> DecomposeResult:
        """
        问题分解：分析问题并生成子问题列表
        """
        context = atom_infos_to_context(chosen_infos)

        prompt = f"""# 任务
你的任务是分析提供的上下文，然后提出可以帮助你更好地回答问题的原子子问题。从不同角度思考，尽可能提出多样化的问题。

# 输出格式
请以以下JSON格式输出：
{{
    "thinking": "<你对这个任务的思考，包括对问题和给定上下文的分析>",
    "should_decompose": <布尔值，是否需要继续分解问题来获取更多信息>,
    "sub_questions": ["<子问题1>", "<子问题2>", ...]
}}

# 已有上下文
{context if context else "暂无上下文"}

# 问题
{question}

# 你的输出："""

        messages = [
            {"role": "system", "content": "你是一个擅长问题分解的AI助手。"},
            {"role": "user", "content": prompt},
        ]

        response = await self.llm_client.chat(messages)

        try:
            # 解析JSON响应
            import re

            json_match = re.search(r"\{.*\}", response, re.DOTALL)
            if json_match:
                result = json.loads(json_match.group())
                return DecomposeResult(
                    should_decompose=result.get("should_decompose", True),
                    thinking=result.get("thinking", ""),
                    sub_questions=result.get("sub_questions", []),
                )
        except Exception as e:
            logger.error(f"Failed to parse decompose response: {e}")

        return DecomposeResult(
            should_decompose=False,
            thinking="解析响应失败",
            sub_questions=[],
        )

    async def _retrieve_candidates(
        self,
        sub_questions: List[str],
        original_question: str,
        chosen_infos: List[AtomRetrievalInfo],
    ) -> tuple[List[AtomRetrievalInfo], str]:
        """
        检索候选信息

        检索策略：
        1. 首先用子问题检索原子问题库
        2. 如果没有结果，用原问题检索原子问题库
        3. 如果还没有结果，用原问题检索chunk库
        """
        retrieval_method = "atom_by_sub_questions"

        # 策略1：用子问题检索
        candidates = await self.retriever.retrieve_atom_info_through_atom(sub_questions)
        candidates = self.retriever.filter_atom_infos(candidates, chosen_infos)

        if not candidates:
            # 策略2：用原问题检索原子库
            retrieval_method = "atom_by_original_question"
            candidates = await self.retriever.retrieve_atom_info_through_atom(
                [original_question]
            )
            candidates = self.retriever.filter_atom_infos(candidates, chosen_infos)

        if not candidates:
            # 策略3：用原问题检索chunk库
            retrieval_method = "chunk_by_original_question"
            candidates = await self.retriever.retrieve_atom_info_through_chunk(
                original_question
            )
            candidates = self.retriever.filter_atom_infos(candidates, chosen_infos)

        return candidates, retrieval_method

    async def _select_best_info(
        self,
        question: str,
        candidates: List[AtomRetrievalInfo],
        chosen_infos: List[AtomRetrievalInfo],
    ) -> SelectResult:
        """
        从候选中选择最相关的信息
        """
        if not candidates:
            return SelectResult(
                selected=False,
                thinking="没有候选信息可供选择",
                chosen_info=None,
            )

        context = atom_infos_to_context(chosen_infos)

        # 构建候选列表字符串
        candidate_list = ""
        for i, info in enumerate(candidates):
            candidate_list += f"问题 {i + 1}: {info.atom}\n"

        prompt = f"""# 任务
你的任务是分析提供的上下文，然后决定哪个子问题在你回答给定问题之前最需要被回答。从给定的问题列表中选择一个最相关的子问题，避免选择已经可以用给定上下文或你自己的知识回答的子问题。

# 输出格式
请以以下JSON格式输出：
{{
    "thinking": "<你对这个选择任务的思考>",
    "selected": <布尔值，是否选择了一个子问题>,
    "question_idx": <整数，从1到{len(candidates)}的子问题索引，如果不选择则为0>
}}

# 已有上下文
{context if context else "暂无上下文"}

# 可选择的子问题
{candidate_list}

# 问题
{question}

# 你的输出："""

        messages = [
            {"role": "system", "content": "你是一个擅长信息选择的AI助手。"},
            {"role": "user", "content": prompt},
        ]

        response = await self.llm_client.chat(messages)

        try:
            import re

            json_match = re.search(r"\{.*\}", response, re.DOTALL)
            if json_match:
                result = json.loads(json_match.group())
                thinking = result.get("thinking", "")
                selected = result.get("selected", False)
                question_idx = result.get("question_idx", 0)

                if selected and 1 <= question_idx <= len(candidates):
                    chosen = candidates[question_idx - 1]
                    return SelectResult(
                        selected=True,
                        thinking=thinking,
                        chosen_info=atom_info_to_schema(chosen),
                    )

                return SelectResult(
                    selected=False,
                    thinking=thinking,
                    chosen_info=None,
                )
        except Exception as e:
            logger.error(f"Failed to parse selection response: {e}")

        return SelectResult(
            selected=False,
            thinking="解析响应失败",
            chosen_info=None,
        )

    async def _generate_answer(
        self,
        question: str,
        chosen_infos: List[AtomRetrievalInfo],
        stream: bool = True,
    ) -> AsyncGenerator[str, None]:
        """
        生成最终答案
        """
        context = atom_infos_to_context(chosen_infos)

        prompt = f"""# 任务
根据提供的参考信息回答用户的问题。如果参考信息不足以回答问题，请诚实地说明。

# 输出格式
请以以下JSON格式输出：
{{
    "thinking": "<你的思考过程>",
    "answer": "<你的回答>"
}}

# 参考信息
{context if context else "无参考信息"}

# 问题
{question}

# 你的回答："""

        messages = [
            {
                "role": "system",
                "content": "你是一个擅长问答的AI助手。请根据参考信息准确、详细地回答问题。",
            },
            {"role": "user", "content": prompt},
        ]

        if stream:
            # 流式输出
            async for chunk in self.llm_client.chat_stream(messages):
                yield chunk
        else:
            response = await self.llm_client.chat(messages)
            yield response

    async def run_stream(
        self,
        question: str,
        stream_answer: bool = True,
    ) -> AsyncGenerator[str, None]:
        """
        执行工作流并以SSE流式输出中间状态

        Args:
            question: 用户问题
            stream_answer: 是否流式输出答案

        Yields:
            SSE事件字符串
        """
        logger.info(
            f"Starting QA decomposition workflow for question: {question[:50]}..."
        )

        # 开始事件
        yield self._create_event(
            RetrievalEventType.START,
            message="开始检索问答流程",
            data={"question": question},
        )

        chosen_infos: List[AtomRetrievalInfo] = []
        decomposition_steps: List[Dict[str, Any]] = []

        try:
            # 迭代循环：分解 -> 检索 -> 选择
            for iteration in range(self.max_iterations):
                step_id = f"Sub{iteration + 1}"
                step_data: Dict[str, Any] = {"step_id": step_id}

                # Step 1: 问题分解
                yield self._create_event(
                    RetrievalEventType.DECOMPOSE_START,
                    step_id=step_id,
                    message=f"第{iteration + 1}轮：正在分解问题...",
                )

                decompose_result = await self._propose_decomposition(
                    question, chosen_infos
                )
                step_data["decompose"] = decompose_result.model_dump()

                yield self._create_event(
                    RetrievalEventType.DECOMPOSE_RESULT,
                    step_id=step_id,
                    message=f"分解完成，生成{len(decompose_result.sub_questions)}个子问题",
                    data=decompose_result.model_dump(),
                )

                if (
                    not decompose_result.should_decompose
                    or not decompose_result.sub_questions
                ):
                    logger.info(
                        f"No more decomposition needed at iteration {iteration + 1}"
                    )
                    decomposition_steps.append(step_data)
                    break

                # Step 2: 检索
                yield self._create_event(
                    RetrievalEventType.RETRIEVE_START,
                    step_id=step_id,
                    message="正在检索相关信息...",
                )

                candidates, retrieval_method = await self._retrieve_candidates(
                    decompose_result.sub_questions,
                    question,
                    chosen_infos,
                )

                retrieve_result = RetrieveResult(
                    atom_candidates=[atom_info_to_schema(c) for c in candidates],
                    retrieval_method=retrieval_method,
                )
                step_data["retrieve"] = retrieve_result.model_dump()

                yield self._create_event(
                    RetrievalEventType.RETRIEVE_RESULT,
                    step_id=step_id,
                    message=f"检索完成，找到{len(candidates)}个候选信息",
                    data=retrieve_result.model_dump(),
                )

                if not candidates:
                    logger.info(f"No candidates found at iteration {iteration + 1}")
                    decomposition_steps.append(step_data)
                    break

                # Step 3: 选择
                yield self._create_event(
                    RetrievalEventType.SELECT_START,
                    step_id=step_id,
                    message="正在选择最相关的信息...",
                )

                select_result = await self._select_best_info(
                    question, candidates, chosen_infos
                )
                step_data["select"] = select_result.model_dump()

                yield self._create_event(
                    RetrievalEventType.SELECT_RESULT,
                    step_id=step_id,
                    message="选择完成"
                    + (
                        "，已选中相关信息" if select_result.selected else "，未选中信息"
                    ),
                    data=select_result.model_dump(),
                )

                if select_result.selected and select_result.chosen_info:
                    # 将选中的信息加入已选列表
                    # 需要将 AtomInfo 转回 AtomRetrievalInfo
                    for c in candidates:
                        if c.atom == select_result.chosen_info.atom_question:
                            chosen_infos.append(c)
                            break
                else:
                    logger.info(f"No info selected at iteration {iteration + 1}")
                    decomposition_steps.append(step_data)
                    break

                decomposition_steps.append(step_data)

            # Step 4: 生成答案
            yield self._create_event(
                RetrievalEventType.ANSWER_START,
                message="正在生成答案...",
                data={"references_count": len(chosen_infos)},
            )

            full_answer = ""
            async for chunk in self._generate_answer(
                question, chosen_infos, stream=stream_answer
            ):
                full_answer += chunk
                if stream_answer:
                    yield self._create_event(
                        RetrievalEventType.ANSWER_CHUNK,
                        message=chunk,
                        data={"chunk": chunk},
                    )

            # 解析最终答案
            try:
                import re

                json_match = re.search(r"\{.*\}", full_answer, re.DOTALL)
                if json_match:
                    result = json.loads(json_match.group())
                    thinking = result.get("thinking", "")
                    answer = result.get("answer", full_answer)
                else:
                    thinking = ""
                    answer = full_answer
            except:
                thinking = ""
                answer = full_answer

            # 完成事件
            yield self._create_event(
                RetrievalEventType.ANSWER_COMPLETE,
                message="回答生成完成",
                data={
                    "thinking": thinking,
                    "answer": answer,
                    "references": [
                        atom_info_to_schema(info).model_dump() for info in chosen_infos
                    ],
                },
            )

            yield self._create_event(
                RetrievalEventType.COMPLETE,
                message="检索问答流程完成",
                data={
                    "question": question,
                    "answer": answer,
                    "thinking": thinking,
                    "references_count": len(chosen_infos),
                    "iterations": len(decomposition_steps),
                    "decomposition_steps": decomposition_steps,
                },
            )

        except Exception as e:
            logger.error(f"Workflow error: {e}", exc_info=True)
            yield self._create_event(
                RetrievalEventType.ERROR,
                message=f"处理出错: {str(e)}",
                data={"error": str(e)},
            )
