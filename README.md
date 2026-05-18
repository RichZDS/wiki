# aisearch Gin 项目骨架

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

可用接口：

```text
GET /health
GET /api/v1/wikis
GET /api/v1/wikis/:id
```

## 目录说明

```text
cmd/server
```

项目启动入口。负责加载配置、初始化路由、启动 HTTP 服务。后续如果有多个程序入口，比如后台任务、命令行工具，也可以继续放在 `cmd` 下面。

```text
internal/config
```

配置层。负责读取环境变量和默认配置，例如应用名、运行环境、端口。后续可以扩展数据库地址、Redis 地址、JWT 密钥等配置。

```text
internal/router
```

路由层。负责创建 Gin Engine、注册全局中间件、挂载接口分组、组合 handler/service/repository 依赖。接口路径和版本分组优先放在这里管理。

```text
internal/middleware
```

中间件层。负责处理请求进入业务之前或响应返回之后的通用逻辑，比如日志、鉴权、跨域、限流、异常恢复等。

```text
internal/handler
```

接口处理层，也可以理解为 Controller。负责接收 HTTP 请求、读取路径参数或请求体、调用 service、返回 JSON 响应。这里不放复杂业务逻辑。

```text
internal/service
```

业务逻辑层。负责组织业务规则、校验业务状态、组合多个 repository 的数据。这里不直接处理 Gin 的 `Context`，这样业务逻辑更容易测试。

```text
internal/repository
```

数据访问层。负责和数据库、缓存、外部存储交互。当前先用内存数据作为示例，后续可以替换成 MySQL、MongoDB、PostgreSQL 等实现。

```text
internal/model
```

数据模型层。负责定义业务实体结构，例如 Wiki、User、Article 等。模型上的 JSON 标签用于控制接口返回字段。

```text
pkg/response
```

公共响应工具。负责统一 API 返回格式，避免每个 handler 手写重复的 JSON 结构。

```text
pkg/logger
```

日志工具。负责日志初始化和写入，按日期自动分割日志文件（`log/YYYY-MM-DD.log`）。

```text
log/
```

日志目录。存放按日期分割的日志文件，如 `log/2026-05-18.log`，已加入 `.gitignore`。

```text
.env.example
```

环境变量示例文件。部署或本地开发时可以按这个文件配置 `APP_NAME`、`APP_ENV`、`APP_PORT`。

## 环境变量

参考 `.env.example` 文件，配置说明：

| 变量 | 说明 | dev 默认值 | prod 默认值 |
|------|------|-----------|------------|
| APP_NAME | 应用名称 | aisearch | aisearch |
| APP_ENV | 环境 | dev | prod |
| APP_PORT | 服务端口 | 8080 | 8081 |
| LOG_LEVEL | 日志级别 | debug | info |
| DB_HOST | 数据库地址 | localhost | prod-db.example.com |
| DB_PORT | 数据库端口 | 5432 | 5432 |
| DB_USER | 数据库用户 | dev_user | prod_user |
| DB_PASSWORD | 数据库密码 | dev_pass | (需设置) |
| DB_NAME | 数据库名 | aisearch_dev | aisearch |

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
