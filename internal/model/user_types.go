package model

import "github.com/gin-gonic/gin"

type CreateUserRequest struct {
	Name     string `json:"name" binding:"required"`
	Password string `json:"password" binding:"required"`
	Quota    int64  `json:"quota"`
	Remark   string `json:"remark"`
}

type UpdateUserRequest struct {
	Name     string  `json:"name"`
	Password string  `json:"password"`
	Quota    *int64  `json:"quota"`
	Remark   *string `json:"remark"`
}

type UserListFilter struct {
	Name          string
	Remark        string
	QuotaMin      *int64
	QuotaMax      *int64
	CreatedAfter  *string
	CreatedBefore *string
	Sort          string
	Order         string
	Page          int
	Size          int
}

type UserListResult struct {
	Total int64   `json:"total"`
	List  []*User `json:"list"`
}

type UserService struct {
	CreateFunc  func(string, string, int64, string) (*User, error)
	ListFunc    func(UserListFilter) (*UserListResult, error)
	GetByIDFunc func(int64, string) (*User, error)
	UpdateFunc  func(int64, map[string]any) (*User, error)
	DeleteFunc  func(int64) error
}

// Create 创建并保存用户。
func (s *UserService) Create(name, password string, quota int64, remark string) (*User, error) {
	return s.CreateFunc(name, password, quota, remark)
}

// List 查询并返回用户列表。
func (s *UserService) List(filter UserListFilter) (*UserListResult, error) {
	return s.ListFunc(filter)
}

// GetByID 根据编号查询用户。
func (s *UserService) GetByID(id int64, name string) (*User, error) {
	return s.GetByIDFunc(id, name)
}

// Update 更新用户数据。
func (s *UserService) Update(id int64, updates map[string]any) (*User, error) {
	return s.UpdateFunc(id, updates)
}

// Delete 软删除指定用户。
func (s *UserService) Delete(id int64) error {
	return s.DeleteFunc(id)
}

type UserController struct {
	CreateFunc func(*gin.Context)
	ListFunc   func(*gin.Context)
	GetFunc    func(*gin.Context)
	UpdateFunc func(*gin.Context)
	DeleteFunc func(*gin.Context)
}

// Create 处理创建用户请求。
func (c *UserController) Create(ctx *gin.Context) {
	c.CreateFunc(ctx)
}

// List 处理用户列表请求。
func (c *UserController) List(ctx *gin.Context) {
	c.ListFunc(ctx)
}

// Get 处理用户详情请求。
func (c *UserController) Get(ctx *gin.Context) {
	c.GetFunc(ctx)
}

// Update 处理更新用户请求。
func (c *UserController) Update(ctx *gin.Context) {
	c.UpdateFunc(ctx)
}

// Delete 处理删除用户请求。
func (c *UserController) Delete(ctx *gin.Context) {
	c.DeleteFunc(ctx)
}

type ResponseBody struct {
	Code int    `json:"code"`
	Data any    `json:"data"`
	Err  string `json:"err"`
}
