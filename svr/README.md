# wiki Gin 项目骨架

这是一个基于 Gin 的 Go Web API 基础框架，已经包含入口、配置、路由、中间件、处理器、服务、仓储、模型和统一响应结构。

## 启动项目

```bash
# 开发环境（默认）
go run cmd/server/main.go

# 指定环境
go run cmd/server/main.go dev
go run cmd/server/main.go prod
```

环境配置文件：

| 环境 | 配置文件 | 端口 | 日志级别 |
|------|---------|------|---------|
| dev  | .env.dev | 8080 | debug |
| prod | .env.prod | 8081 | info |

默认服务地址：

```text
http://localhost:8080
```

## 环境变量

参考 `.env.example` 文件，配置说明：

| 变量 | 说明 | dev 默认值 | prod 默认值 |
|------|------|-----------|------------|
| APP_NAME | 应用名称 | wiki | wiki |
| APP_ENV | 环境 | dev | prod |
| APP_PORT | 服务端口 | 8080 | 8081 |
| LOG_LEVEL | 日志级别 | debug | info |
| DB_HOST | 数据库地址 | localhost | prod-db.example.com |
| DB_PORT | 数据库端口 | 5432 | 5432 |
| DB_USER | 数据库用户 | dev_user | prod_user |
| DB_PASSWORD | 数据库密码 | dev_pass | (需设置) |
| DB_NAME | 数据库名 | wiki_dev | wiki |

## 分层调用关系

```text
HTTP 请求
  -> router
  -> middleware
  -> handler
  -> service
  -> repository
  -> model
```

推荐规则：

- `handler` 只处理 HTTP 输入输出。
- `service` 只处理业务逻辑。
- `repository` 只处理数据读写。
- `model` 只定义数据结构。
- `pkg` 放可以跨模块复用的公共能力。
