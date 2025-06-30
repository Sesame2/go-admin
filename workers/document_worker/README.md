# 📄 Document Worker

该服务是 `go-admin` 知识库平台的异步任务子模块，专用于执行 **文档解析 → 分片 → Embedding → 写入向量数据库** 的后台任务。

Python 实现该服务的原因是：
- 更适合使用 NLP / ML 工具链（如 Transformers、OpenAI、PDF 解析等）
- 和 Go 主服务解耦，具备良好的任务独立性和可扩展性
- 通过 RabbitMQ 与主服务异步通信，完成任务闭环

---

## 📦 功能职责

- 消费来自 RabbitMQ 的任务消息（包含文档 URL 和任务参数）
- 下载/读取文档内容（PDF、DOCX、TXT等）
- 执行内容清洗、结构化处理、分片（chunking）
- 对分片内容进行向量化（embedding），如使用 OpenAI 或 Sentence-Transformers
- 将向量写入指定向量数据库（如 Qdrant / Weaviate / Milvus）
- 可选：将任务完成状态通过 Webhook / 回调方式通知 Go 服务

---

## 📁 项目结构

```
document_worker/
├── app/
│   ├── api.py                # FastAPI 路由定义
│   ├── consumer.py           # RabbitMQ 消费逻辑
│   ├── embedding.py          # embedding 逻辑（OpenAI/HuggingFace等）
│   ├── parser.py             # 文档解析与分片逻辑
│   └── vector_store.py       # 向量库写入
├── pyproject.toml            # Poetry 配置
├── README.md                 # 使用说明
├── .env                      # 环境变量（可选）
└── main.py                   # 启动 FastAPI 应用 + 消费任务

```

---

## 🚀 快速启动

### 1. 安装依赖（本地开发）

建议使用 poetry 或 conda：

```bash
cd workers/document_worker
poetry install
```

### 2. 启动 Worker 服务

```bash
python main.py
```

或通过 Docker：

```bash
docker build -t doc-worker .
docker run --env-file .env doc-worker
```

---

## ⚙️ 环境变量说明

可通过 `.env` 或 Docker Compose 传入：

| 变量名 | 说明 | 示例 |
|--------|------|------|
| `RABBITMQ_HOST` | RabbitMQ 服务地址 | `localhost` / `rabbitmq` |
| `VECTOR_DB_HOST` | 向量数据库地址 | `http://qdrant:6333` |
| `EMBEDDING_PROVIDER` | 使用的 embedding 引擎 | `openai` / `sbert` |
| `OPENAI_API_KEY` | 如使用 OpenAI 时的密钥 | `sk-xxxx...` |

---

## 🧩 与 Go 主服务的配合方式

1. Go 服务将文档上传后，将任务信息（文档 URL、知识库 ID 等）写入 RabbitMQ 队列。
2. 本服务消费任务 → 解析 → 向量化 → 写入向量库。
3. 可选：向 Go 的 API 回调任务完成状态（用于展示进度或二次处理）。

---

## 📌 扩展建议

- 支持多格式文档：PDF、DOCX、Markdown 等
- 加入任务重试/失败通知机制
- 支持分布式部署（多个 worker 并发处理）
- 加入任务日志上报、指标采集（Prometheus、OpenTelemetry）

---