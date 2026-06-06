# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 构建和运行

```bash
# 运行（开发环境，默认）
go run main.go

# 运行（生产环境）
go run main.go prod

# 构建
go build -o aisearch.exe .

# 运行所有测试
go test ./tests/...

# 运行单个包的测试
go test ./tests/chunk/
go test ./tests/agent/
go test ./tests/tools/

# 运行单个测试函数
go test ./tests/chunk/ -run TestFreeChunker
```

服务默认监听 `:8080`（开发）或 `:8081`（生产），可通过 `config.yaml` / `config.prod.yaml` 的 `server.port` 配置。

入口文件是项目根目录的 `main.go`，按顺序初始化：config → logger → MySQL → Redis → router → run。

## 项目概述

- 模块名: `aisearch`
- Go 版本: 1.25
- Web 框架: Gin
- ORM: GORM (MySQL)
- 缓存: go-redis
- AI 框架: CloudWeGo Eino（切块、Embedding、模型调用）

## 架构

```
HTTP request -> router -> middleware -> controller -> service -> model
```

两层核心子系统：

**Web API 层（`internal/`）**
- `internal/router/` — 路由注册，初始化 controller 实例并绑定路由。生产环境自动设置 `gin.ReleaseMode`。
- `internal/controller/` — HTTP 处理层，解析请求参数，调用 service 层，返回统一 JSON 响应。
- `internal/service/` — 业务逻辑层，纯 Go 逻辑，不依赖 Gin，通过 GORM 操作数据库。
- `internal/model/` — 项目结构体的统一定义目录，包括 GORM 模型、请求/响应 DTO、配置、任务及其他业务结构体。
- `internal/middleware/` — Gin 中间件（请求日志）。
- `internal/config/` — YAML 配置加载，支持环境变量覆盖和多环境合并。

**AI/RAG 层（`internal/ai/`）**
- `internal/ai/chunk/` — 文档切块引擎，4 种策略，返回 `[]*schema.Document` 直接对接 Eino Indexer。
- `internal/ai/embedding/` — Embedding 接口定义（`EmbedStrings`），由切块器和检索器注入实现。
- `internal/ai/agent/` — LLM 客户端封装（OpenAI GPT-4o、DeepSeek），基于 Eino ChatModel。

**公共包（`pkg/`）**
- `pkg/database/` — 全局 `database.DB` (*gorm.DB) 和 `database.RDB` (*redis.Client)。初始化失败直接 `os.Exit(1)`。
- `pkg/logger/` — 全局 logger，同时输出到 stdout 和每日滚动的 `log/YYYY-MM-DD.log` 文件。`GetLogger()` 未初始化时自动回退到 dev 模式。
- `pkg/response/` — 统一 JSON 响应: `response.Success(c, status, data)` / `response.Error(c, status, msg)`。响应格式 `{ "code": int, "data": any, "err": "" }`。
- `pkg/snowflake/` — 雪花算法 ID 生成器，`snowflake.Next()` 返回全局唯一 int64 ID。
- `pkg/auth/` — bcrypt + pepper 密码哈希与验证。
- `pkg/utils/` — 泛型工具，如 `Ptr[T](v) *T`。

## 代码组织与注释规范

- 项目中自定义的所有结构体必须定义在 `internal/model/` 目录下，其他目录不得新增结构体定义；使用方通过 `model.Xxx` 引用。
- 新增或修改结构体时，应同步将不在 `internal/model/` 下的相关结构体迁移到该目录，不能继续扩大历史遗留。
- 所有具名函数和方法（包括导出与非导出、构造函数、辅助函数、测试函数）都必须在函数声明正上方写注释，明确说明该函数的用途和职责。
- 函数注释必须紧贴函数声明；导出函数的注释应以函数名开头，遵循 Go Doc 规范。

## Controller 和 Service 层规范

所有 controller 和 service 必须使用结构体 + `NewXxx()` 构造函数模式：

```go
// Service 层
type UserService struct{}

// NewUserService 创建用户服务。
func NewUserService() *UserService {
    return &UserService{}
}

// Controller 层
type UserController struct {
    svc *service.UserService
}

// NewUserController 创建用户控制器并注入用户服务。
func NewUserController() *UserController {
    return &UserController{svc: service.NewUserService()}
}
```

Controller 方法签名: `func (ctl *XxxController) Action(c *gin.Context)`
Service 方法签名: `func (s *XxxService) Action(params...) (result, error)`

## AI 文档切块（Chunking）

切块模块是 RAG 流水线的 Transformer 角色，通过 `chunk.NewChunker(strategy)` 获取实现，返回 `[]*schema.Document`。

### 四种策略

| 策略 | 常量 | 适用场景 | 特殊依赖 |
|------|------|---------|---------|
| 自由切块 | `StrategyFree` | 纯文本、日志 | 无 |
| Markdown 切块 | `StrategyMD` | Markdown 文档 | 无 |
| 语义切块 | `StrategyEino` | 语义边界敏感场景 | 需注入 `Embedder` |
| 分层切块 | `StrategyHierarchical` | 需要父子块关联的精确检索 | 无 |

**语义切块（Eino）** 需要注入 `Embedder`，不可通过 `NewChunker` 获取，必须显式构造：

```go
emb := &myEmbedder{} // 实现 embedding.Embedder 接口
chunker := chunk.NewEinoChunker(emb)
```

**分层切块（Hierarchical）** 返回两层文档：父块（parent，chunkSize × 3）和子块（child，原始 chunkSize）。子块通过 `parent_chunk_id` 和 `parent_content` 关联父块，用于"小子块检索 + 大父块提供上下文"的检索策略。

### ChunkConfig 参数

```go
cfg := chunk.ChunkConfig{
    ChunkSize:    500,   // 每块最大字符数（rune），<=0 默认 500
    ChunkOverlap: 50,    // 块间重叠字符数，0 无重叠
    Separators:   []string{"\n\n", "\n", "。"}, // 仅 StrategyFree 生效
}
```

### 返回的 Document 元数据

每个 Document.MetaData 包含：`chunk_index`（0-based 序号）、`chunk_total`（总块数）、`chunk_strategy`（策略名）。Markdown 切块额外包含 `heading_path`（标题路径如 "Chapter 1 > Section 1.1"）和 `element_types`。

### Embedder 接口

```go
type Embedder interface {
    EmbedStrings(ctx context.Context, texts []string) ([][]float64, error)
}
```

位于 `internal/ai/embedding/`，由调用方实现具体模型（OpenAI、本地模型等）。

## 数据库规范

### 所有 SQL 查询必须使用 GORM

禁止使用 `database.DB.Raw()` 或 `database.DB.Exec()` 拼接 SQL。所有查询通过 GORM 链式调用构建。

```go
// ✅ 正确：GORM 链式调用
db := database.DB.Model(&model.User{}).Where("is_deleted = 0")
db = db.Where("name LIKE ?", "%"+name+"%")
db.Count(&total)
db.Offset(offset).Limit(size).Find(&users)

// ❌ 错误：原始 SQL
database.DB.Raw("SELECT * FROM user WHERE name LIKE '%" + name + "%'").Scan(&users)
```

### 软删除

所有删除操作使用软删除（`is_deleted` 字段设为 1），查询时始终带上 `WHERE is_deleted = 0`。

## 接口规范

### List 接口必须支持查询参数过滤

分页列表接口支持：`page`（默认 1）、`size`（默认 20，最大 100）、字段模糊匹配（`name`、`remark`）、范围参数（`quota_min`/`quota_max`、`created_after`/`created_before`）、`sort`/`order` 排序。

所有过滤条件通过 `applyXxxFilters(c, db)` 辅助函数应用到 GORM 查询链。

### 请求/响应结构体

- 请求体和响应体结构体统一定义在 `internal/model/` 下，使用方通过 `model.Xxx` 引用；请求体使用 `json:"..." binding:"..."` 标签
- 更新接口的数值字段使用指针类型 `*int64` 区分"未传"和"传零值"
- 列表接口返回 `{ "total": int64, "list": [...] }`
- 创建成功返回 HTTP 201，查询成功返回 200，删除成功返回 200

### 路由命名

RESTful 风格，统一前缀 `/api/v1/`：
```
POST   /api/v1/users        — 创建
GET    /api/v1/users        — 列表
GET    /api/v1/users/:id    — 详情
PUT    /api/v1/users/:id    — 更新
DELETE /api/v1/users/:id    — 删除
```

另有 `/health` 端点返回 `{"code":200, "data": {"message": "OK"}}`。

## 配置文件

- `config.yaml` — 开发环境配置
- `config.prod.yaml` — 生产环境覆盖配置（仅覆盖差异字段）
- 敏感字段支持环境变量覆盖: `MYSQL_PASSWORD`、`REDIS_PASSWORD`
- 环境选择: 通过命令行参数（`go run main.go prod`）或 `APP_ENV` 环境变量
- 配置文件查找: 优先 `APP_CONFIG` 环境变量，否则从当前目录向上查找 `config.yaml`

## 日志

日志通过 `pkg/logger` 初始化，同时写入 stdout 和 `log/YYYY-MM-DD.log` 每日滚动文件。开发环境日志前缀为 `[DEV]`。GORM 的 SQL 日志也通过适配器桥接到同一 logger，慢查询阈值 200ms。

## ID 生成

所有实体 ID 使用雪花算法：`snowflake.Next()` 返回 int64。Worker ID 通过 `SNOWFLAKE_WORKER_ID` 配置（0~1023，默认 1）。DDL 中 id 为 `BIGINT NOT NULL`（无 AUTO_INCREMENT）。

## 密码安全

使用 bcrypt + pepper 双重保护。`auth.Hash()` 内部执行 `bcrypt(password + pepper, cost=10)`。Pepper 通过 `PASSWORD_PEPPER` 环境变量配置（默认 `wiki-default-pepper`）。Service 层自动调用 `auth.Hash()`，controller 传入明文。密码字段 `json:"-"` 永不序列化。

## 测试

测试统一放在 `tests/` 目录，按功能分子目录：
- `tests/chunk/` — 切块策略测试（含 mock Embedder）
- `tests/agent/` — Agent 和切块集成测试
- `tests/tools/` — 工具和文件系统测试
- `tests/integration/` — 集成测试（需 MySQL/Redis，目前 skip）

共享测试工具在 `tests/helpers.go`：`TempDir`、`WriteFile`、`AssertNoErr`、`AssertEqual`、`AssertTrue`。

## LLM Agent 环境变量

| 变量 | 用途 | 默认值 |
|------|------|--------|
| `OPENAI_API_KEY` | OpenAI API 密钥 | 必填，否则 Fatal |
| `OPENAI_MODEL_ID` | OpenAI 模型 ID | `gpt-4o` |
| `DEEPSEEK_API_KEY` | DeepSeek API 密钥 | 必填，否则 Fatal |

Agent 创建失败直接 `log.Fatal` 终止进程。
