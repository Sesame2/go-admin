"""
RabbitMQ 消费者 - 处理文档解析任务
"""

import json
import asyncio
import aio_pika
from typing import Callable, Any
from uuid import UUID

from core.config import get_settings
from core.logger import get_logger

logger = get_logger("rabbitmq")


class RabbitMQConsumer:
    """RabbitMQ 消费者"""

    def __init__(self):
        self.settings = get_settings()
        self.connection = None
        self.channel = None
        self.queue = None

    async def connect(self):
        """建立连接"""
        logger.info(f"连接 RabbitMQ: {self.settings.RABBITMQ_HOST}:{self.settings.RABBITMQ_PORT}")
        
        self.connection = await aio_pika.connect_robust(
            self.settings.RABBITMQ_URL,
            timeout=30,
        )
        self.channel = await self.connection.channel()
        
        # 声明交换机
        exchange = await self.channel.declare_exchange(
            self.settings.RABBITMQ_EXCHANGE,
            aio_pika.ExchangeType.TOPIC,
            durable=True,
        )
        
        # 声明队列
        self.queue = await self.channel.declare_queue(
            self.settings.RABBITMQ_QUEUE,
            durable=True,
        )
        
        # 绑定队列到交换机
        await self.queue.bind(
            exchange,
            routing_key=self.settings.RABBITMQ_ROUTING_KEY,
        )
        
        logger.info("RabbitMQ 连接成功")

    async def consume(self, callback: Callable[[dict], Any]):
        """开始消费消息"""
        if not self.queue:
            await self.connect()

        logger.info(f"开始消费队列: {self.settings.RABBITMQ_QUEUE}")
        
        async with self.queue.iterator() as queue_iter:
            async for message in queue_iter:
                async with message.process():
                    try:
                        body = json.loads(message.body.decode())
                        logger.info(f"收到消息: {body.get('document_id', 'unknown')}")
                        await callback(body)
                    except json.JSONDecodeError as e:
                        logger.error(f"消息解析失败: {e}")
                    except Exception as e:
                        logger.error(f"处理消息失败: {e}", exc_info=True)

    async def close(self):
        """关闭连接"""
        if self.connection:
            await self.connection.close()
            logger.info("RabbitMQ 连接已关闭")


class DocumentParseMessage:
    """文档解析消息"""
    
    def __init__(self, data: dict):
        self.document_id = UUID(data["document_id"])
        self.knowledge_base_id = UUID(data["knowledge_base_id"])
        self.filename = data["filename"]
        self.file_type = data["file_type"]
        self.minio_path = data["minio_path"]

    def __repr__(self):
        return f"<DocumentParseMessage(doc_id={self.document_id}, filename={self.filename})>"
