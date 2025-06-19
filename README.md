# go-admin

一个基于 **Go + Gin + Ent ORM** 的简单用户管理系统，支持用户认证与管理，后续计划扩展个人知识库功能。

## 🌟 功能特性

- 用户注册与登录（基于 JWT）
- 用户信息管理
- Swagger 文档自动生成
- 日志管理（使用 Zap）
- 数据库管理（基于 Ent ORM + PostgreSQL）
- 支持 Docker Compose 部署
- 良好的模块化项目结构

---

## 🛠️ 技术栈

- **语言**: Go 1.24+
- **Web 框架**: Gin
- **ORM**: Ent
- **数据库**: PostgreSQL
- **日志系统**: Zap
- **接口文档**: Swagger (Swaggo)
- **容器化部署**: Docker Compose
- **配置管理**: YAML

---

## 📂 项目结构

```plaintext
.
├── bin/                # 构建产物
├── cmd/                # 主程序入口
│   └── server/main.go
├── configs/            # 配置文件
├── deployments/        # 部署文件（Docker Compose）
├── docs/               # Swagger 文档
├── internal/           # 核心业务代码
│   ├── api/            # 接口路由与控制器
│   ├── app/            # 应用初始化
│   ├── config/         # 配置管理
│   ├── dao/            # 数据库访问层
│   ├── database/       # 数据库连接与抽象
│   ├── errors/         # 错误定义
│   ├── logger/         # 日志模块
│   ├── middleware/     # 中间件（JWT等）
│   ├── models/         # 数据模型（Ent ORM）
│   └── services/       # 服务层（业务逻辑）
├── Makefile            # 构建与开发脚本
├── go.mod / go.sum     # Go依赖管理
└── LICENSE             # 开源协议
```

---

## 🚀 快速开始

### 1️⃣ 克隆仓库

```bash
git clone https://github.com/Sesame2/go-admin.git
cd go-admin
```

### 2️⃣ 配置数据库

在 `configs/config.yaml` 中修改数据库连接信息，例如：

```yaml
database:
  driver: postgres
  host: localhost
  port: 5432
  user: postgres
  password: your_password
  dbname: go_admin
```

### 3️⃣ 安装依赖

```bash
go mod tidy
```

### 4️⃣ 生成 Ent ORM 代码

```bash
make ent-gen
```

### 5️⃣ 生成 Swagger 文档

```bash
make swag
```

### 6️⃣ 本地运行（热重载）

```bash
make air
```

### ✅ 生产构建

```bash
make build
./bin/server
```

或者使用 Docker：

```bash
docker-compose up --build
```

---

## 🔥 Makefile 常用命令

| 命令      | 功能                       |
|-----------|----------------------------|
| `make swag`    | 生成 Swagger 接口文档      |
| `make air`     | 热重载运行（开发模式）      |
| `make clean`   | 清理构建产物               |
| `make build`   | 构建二进制文件             |
| `make run`     | 运行构建后的应用           |
| `make ent-gen` | 根据 Ent schema 生成模型代码 |

---

## 📜 License

本项目基于 [Apache 2.0 License](./LICENSE) 开源。

---

## 💡 未来规划

- [ ] 增加个人知识库模块
- [ ] 支持多用户、多角色权限管理
- [ ] 丰富API接口，支持更多管理功能
- [ ] 完善单元测试与集成测试

---

## 🤝 欢迎贡献

如果你有好的建议或者发现问题，欢迎提 Issue 或提交 PR。

