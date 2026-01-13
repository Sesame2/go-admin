"""
通义千问 LLM 客户端
用于生成原子问题等 LLM 任务
"""

from typing import List, Optional, Dict, Any, AsyncGenerator
import json
from openai import AsyncOpenAI

from core.logger import get_logger
from core.config import get_settings

logger = get_logger("qwen_client")
settings = get_settings()


class QwenLLMClient:
    """
    通义千问 LLM 客户端（使用 OpenAI 兼容接口）
    用于生成原子问题等任务
    """

    def __init__(
        self,
        api_key: Optional[str] = None,
        base_url: str = "https://dashscope.aliyuncs.com/compatible-mode/v1",
        model: str = "qwen-max",
        temperature: float = 0.7,
        max_tokens: int = 2000,
    ):
        self.api_key = api_key or settings.QWEN_API_KEY
        self.base_url = base_url
        self.model = model
        self.temperature = temperature
        self.max_tokens = max_tokens

        self.client = AsyncOpenAI(
            api_key=self.api_key,
            base_url=self.base_url,
        )

        logger.info(f"Initialized QwenLLMClient with model: {model}")

    async def generate(
        self,
        prompt: str,
        system_prompt: Optional[str] = None,
        temperature: Optional[float] = None,
        max_tokens: Optional[int] = None,
    ) -> str:
        """
        生成文本

        Args:
            prompt: 用户提示词
            system_prompt: 系统提示词
            temperature: 温度参数
            max_tokens: 最大生成 token 数

        Returns:
            生成的文本
        """
        messages = []

        if system_prompt:
            messages.append({"role": "system", "content": system_prompt})

        messages.append({"role": "user", "content": prompt})

        return await self.chat(messages, temperature, max_tokens)

    async def chat(
        self,
        messages: List[Dict[str, str]],
        temperature: Optional[float] = None,
        max_tokens: Optional[int] = None,
    ) -> str:
        """
        聊天接口（非流式）

        Args:
            messages: 消息列表
            temperature: 温度参数
            max_tokens: 最大生成 token 数

        Returns:
            生成的文本
        """
        try:
            response = await self.client.chat.completions.create(
                model=self.model,
                messages=messages,
                temperature=temperature or self.temperature,
                max_tokens=max_tokens or self.max_tokens,
            )

            content = response.choices[0].message.content
            logger.debug(f"Generated content length: {len(content)}")
            return content

        except Exception as e:
            logger.error(f"Failed to generate text: {e}")
            raise

    async def chat_stream(
        self,
        messages: List[Dict[str, str]],
        temperature: Optional[float] = None,
        max_tokens: Optional[int] = None,
    ) -> AsyncGenerator[str, None]:
        """
        流式聊天接口

        Args:
            messages: 消息列表
            temperature: 温度参数
            max_tokens: 最大生成 token 数

        Yields:
            生成的文本块
        """
        try:
            response = await self.client.chat.completions.create(
                model=self.model,
                messages=messages,
                temperature=temperature or self.temperature,
                max_tokens=max_tokens or self.max_tokens,
                stream=True,
            )

            async for chunk in response:
                if chunk.choices and chunk.choices[0].delta.content:
                    yield chunk.choices[0].delta.content

        except Exception as e:
            logger.error(f"Failed to stream text: {e}")
            raise

    async def generate_atom_questions(
        self,
        content: str,
        title: Optional[str] = None,
        max_questions: int = 10,
    ) -> List[str]:
        """
        为给定内容生成原子问题
        集成 pike-rag 的 atom question tagging protocol

        Args:
            content: 文本内容
            title: 可选的标题
            max_questions: 最大问题数量

        Returns:
            问题列表
        """
        system_prompt = "You are a helpful AI assistant good at content understanding and asking question."

        # 如果有标题，将标题加入内容
        full_content = content
        if title:
            full_content = f"Title: {title}. Content: {content}"

        user_prompt = f"""# Task
Your task is to extract as many questions as possible that are relevant and can be answered by the given content. Please try to be diverse and avoid extracting duplicated or similar questions. Make sure your question contain necessary entity names and avoid to use pronouns like it, he, she, they, the company, the person etc.

# Output Format
Output your answers line by line, with each question on a new line, without itemized symbols or numbers.

# Content
{full_content}

# Output:"""

        try:
            response = await self.generate(
                prompt=user_prompt,
                system_prompt=system_prompt,
                temperature=0.7,
            )

            # 解析响应，按行分割
            questions = response.split("\n")
            questions = [q.strip() for q in questions if q.strip()]

            # 限制问题数量
            if len(questions) > max_questions:
                questions = questions[:max_questions]

            logger.info(f"Generated {len(questions)} atom questions")
            return questions

        except Exception as e:
            logger.error(f"Failed to generate atom questions: {e}")
            return []

    async def batch_generate_atom_questions(
        self,
        contents: List[str],
        titles: Optional[List[str]] = None,
        max_questions: int = 10,
    ) -> List[List[str]]:
        """
        批量生成原子问题

        Args:
            contents: 文本内容列表
            titles: 可选的标题列表
            max_questions: 每个内容的最大问题数量

        Returns:
            问题列表的列表
        """
        if titles is None:
            titles = [None] * len(contents)

        results = []
        for content, title in zip(contents, titles):
            questions = await self.generate_atom_questions(
                content=content,
                title=title,
                max_questions=max_questions,
            )
            results.append(questions)

        return results
