"""
原子问题生成步骤
集成 pike-rag 的 LLMPoweredTagger 逻辑
"""

from typing import List, Any
from copy import deepcopy

from core.logger import get_logger
from services.pipeline.processor.base import (
    PipelineStep,
    ProcessContext,
    AtomQuestionData,
)

logger = get_logger("pipeline")


class AtomQuestionTagger(PipelineStep):
    """
    原子问题生成器
    集成 pike-rag 的 LLMPoweredTagger 逻辑
    为每个 chunk 生成可能被问到的问题
    """

    name = "AtomQuestionTagger"

    def __init__(
        self,
        llm_client: Any = None,
        num_parallel: int = 1,
        tag_name: str = "atom_questions",
    ):
        self.llm_client = llm_client
        self.num_parallel = num_parallel
        self.tag_name = tag_name

    async def process(self, context: ProcessContext) -> ProcessContext:
        """为每个 chunk 生成原子问题"""
        if not context.chunks:
            logger.warning("No chunks to tag")
            return context

        logger.info(f"Generating atom questions for {len(context.chunks)} chunks")

        atom_questions: List[AtomQuestionData] = []

        for chunk in context.chunks:
            questions = await self._generate_questions(chunk.content, chunk.metadata)

            for q in questions:
                atom_questions.append(
                    AtomQuestionData(
                        question=q,
                        chunk_index=chunk.chunk_index,
                        metadata=deepcopy(chunk.metadata),
                    )
                )

        context.atom_questions = atom_questions
        logger.info(f"Generated {len(atom_questions)} atom questions")

        return context

    def can_handle(self, context: ProcessContext) -> bool:
        return context.chunks is not None and context.atom_questions is None

    async def _generate_questions(self, content: str, metadata: dict) -> List[str]:
        """
        使用 LLM 生成原子问题
        集成 pike-rag 的 atom question tagging protocol
        """
        if not self.llm_client:
            logger.debug("No LLM client, skipping atom question generation")
            return []

        try:
            # 从 metadata 中获取标题（如果有）
            title = metadata.get("title")

            # 使用 Qwen LLM Client 的 generate_atom_questions 方法
            # 该方法已集成 pike-rag 的 protocol
            questions = await self.llm_client.generate_atom_questions(
                content=content,
                title=title,
                max_questions=10,
            )

            return questions

        except Exception as e:
            logger.error(f"Failed to generate questions: {e}")
            return []


class MockAtomQuestionTagger(PipelineStep):
    """
    Mock 原子问题生成器
    用于测试场景，根据 chunk 内容生成简单的模拟问题
    """

    name = "MockAtomQuestionTagger"

    def __init__(self, questions_per_chunk: int = 3):
        self.questions_per_chunk = questions_per_chunk

    async def process(self, context: ProcessContext) -> ProcessContext:
        """为每个 chunk 生成模拟的原子问题"""
        if not context.chunks:
            logger.warning("No chunks to tag")
            return context

        logger.info(f"Generating mock atom questions for {len(context.chunks)} chunks")

        atom_questions: List[AtomQuestionData] = []

        for chunk in context.chunks:
            questions = self._generate_mock_questions(chunk.content, chunk.chunk_index)

            for q in questions:
                atom_questions.append(
                    AtomQuestionData(
                        question=q,
                        chunk_index=chunk.chunk_index,
                        metadata=deepcopy(chunk.metadata),
                    )
                )

        context.atom_questions = atom_questions
        logger.info(f"Generated {len(atom_questions)} mock atom questions")

        return context

    def can_handle(self, context: ProcessContext) -> bool:
        return context.chunks is not None and context.atom_questions is None

    def _generate_mock_questions(self, content: str, chunk_index: int) -> List[str]:
        """生成模拟问题"""
        # 取内容前30个字符作为关键词
        preview = content[:50].replace("\n", " ").strip()
        if len(preview) > 30:
            preview = preview[:30] + "..."

        questions = [
            f"关于'{preview}'的内容是什么？",
            f"第{chunk_index + 1}段内容讲述了什么？",
            f"如何理解'{preview}'这段内容？",
        ]

        return questions[: self.questions_per_chunk]
