# CLAUDE.md

## 项目概述

- 模块名: `aisearch`
- Go 版本: 1.25
- Web 框架: Gin
- ORM: GORM (MySQL)
- 缓存: go-redis

## 架构

```
HTTP request -> router -> controller -> service -> model
```

- `internal/router/` — 路由注册，初始化 controller 实例并绑定路由
- `internal/controller/` — HTTP 处理层，解析请求参数，调用 service 层，返回响应
- `internal/service/` — 业务逻辑层，纯 Go 逻辑，不依赖 Gin，通过 GORM 操作数据库
- `internal/model/` — GORM 模型定义
- `pkg/database/` — 全局 `database.DB` (*gorm.DB) 和 `database.RDB` (*redis.Client)
- `pkg/response/` — 统一 JSON 响应: `response.Success(c, status, data)` / `response.Error(c, status, msg)`
- `pkg/snowflake/` — 雪花算法 ID 生成器，通过 `snowflake.Next()` 生成全局唯一 int64 ID
- `pkg/auth/` — 密码哈希与验证，bcrypt + pepper 双重保护

### Controller 和 Service 层规范

所有 controller 和 service 必须使用结构体 + `NewXxx()` 构造函数模式：

```go
// Service 层
type UserService struct{}

func NewUserService() *UserService {
    return &UserService{}
}

// Controller 层
type UserController struct {
    svc *service.UserService
}

func NewUserController() *UserController {
    return &UserController{svc: service.NewUserService()}
}
```

所有 CRUD 方法必须大写开头（导出方法），方法接收者为结构体指针：

```go
func (ctl *UserController) Create(c *gin.Context) { ... }
func (s *UserService) Create(name, password string, quota int64, remark string) (*model.User, error) { ... }
```

Controller 方法签名: `func (ctl *XxxController) Action(c *gin.Context)`
Service 方法签名: `func (s *XxxService) Action(params...) (result, error)`

## 数据库规范

### 所有 SQL 查询必须使用 GORM

禁止使用 `database.DB.Raw()` 或 `database.DB.Exec()` 拼接 SQL。所有查询通过 GORM 链式调用构建。

示例：
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

### GORM 模型

- 模型文件放在 `internal/model/` 目录
- 一个文件一个模型
- 使用 `gorm:"..."` 标签定义字段约束
- JSON 标签控制序列化，密码字段使用 `json:"-"` 隐藏

## 接口规范

### List 接口必须支持查询参数过滤

分页列表接口根据业务需求支持以下查询参数：
- `page` — 页码，默认 1
- `size` — 每页条数，默认 20，最大 100
- 字段过滤参数（如 `name`、`remark`）— 模糊匹配
- 范围参数（如 `quota_min`/`quota_max`、`created_after`/`created_before`）
- `sort` / `order` — 排序字段和方向

所有过滤条件通过 `applyXxxFilters(c, db)` 辅助函数应用到 GORM 查询链。

### View 接口必须支持查询参数

详情接口支持通过查询参数辅助筛选（如 `?name=xxx`），配合主键 ID 进行定位。

### 请求/响应结构体

- 请求体结构体定义在 controller 文件内，使用 `json:"..." binding:"..."` 标签
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

## 配置文件

- `config.yaml` — 开发环境配置（包含数据库和 Redis 连接信息）
- `config.prod.yaml` — 生产环境覆盖配置
- 敏感字段支持环境变量覆盖: `MYSQL_PASSWORD`、`REDIS_PASSWORD`、`APP_ENV`、`APP_CONFIG`
- 配置加载自动从当前目录向上查找 `config.yaml`

## SQL 脚本

- `sql/init.sql` — 建表语句和初始化数据
- 表名使用 `` ` ` `` 反引号，字符集 utf8mb4
- 时间字段: `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP, `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
- 软删除字段: `is_deleted` TINYINT(1) DEFAULT 0
- 主键 ID 使用雪花算法（int64），不使用数据库自增。DDL 中 id 为 `BIGINT NOT NULL`（无 AUTO_INCREMENT）

## ID 生成规范

所有实体 ID 使用雪花算法生成，调用 `pkg/snowflake/snowflake.Next()` 返回 int64：

```go
import "aisearch/pkg/snowflake"

user := model.User{
    ID: snowflake.Next(),
    ...
}
```

- 雪花 ID 是时间递增的 int64，数据库索引友好
- Worker ID 通过环境变量 `SNOWFLAKE_WORKER_ID` 配置（0~1023，默认 1）
- 单机部署无需修改，分布式部署给每个实例分配不同的 worker ID

## 密码安全规范

密码使用 **bcrypt + pepper** 双重保护：

```go
import "aisearch/pkg/auth"

// 加密
hash, err := auth.Hash(plainPassword)

// 验证
if auth.Verify(plainPassword, hash) { ... }
```

- `auth.Hash()` 内部自动执行 `bcrypt(password + pepper, cost=10)`，盐值由 bcrypt 自动生成并嵌入哈希结果
- Pepper（环境盐值）通过环境变量 `PASSWORD_PEPPER` 配置，默认值 `wiki-default-pepper`
- 生产环境必须覆盖 `PASSWORD_PEPPER`，且不可泄露、不可修改（否则所有密码失效）
- Service 层的 Create 和 Update 方法自动调用 `auth.Hash()`，controller 层传入的是明文密码
- 用户模型中 Password 字段使用 `json:"-"` 标签，永远不会序列化返回给客户端
