from contextlib import asynccontextmanager
import logging

import aio_pika
from fastapi import FastAPI
import uvicorn


async def start_mq():
    connection = await aio_pika.connect_robust("amqp://admin:123456789@localhost/")
    channel = await connection.channel()

    queue = await channel.declare_queue("simple_queue", durable=True)

    async with queue.iterator() as queue_iter:
        async for message in queue_iter:
            async with message.process():
                print("收到消息:", message.body.decode())


@asynccontextmanager
async def lifespan(app: FastAPI):
    print("开始监听rabbitmq消息")
    await start_mq()
    yield
    print("程序关闭")


app = FastAPI(lifespan=lifespan)


# 保证 FastAPI 持续运行
@app.get("/")
async def root():
    return {"message": "消费者正在运行"}


if __name__ == "__main__":
    uvicorn.run(app=app, host=".0.0.0.0", port=22222)
