package controller

import (
	"net/http"
	"strconv"

	"aisearch/internal/service"
	"aisearch/pkg/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type createUserReq struct {
	Name     string `json:"name"     binding:"required"`
	Password string `json:"password" binding:"required"`
	Quota    int64  `json:"quota"`
	Remark   string `json:"remark"`
}

type updateUserReq struct {
	Name     string  `json:"name"`
	Password string  `json:"password"`
	Quota    *int64  `json:"quota"`
	Remark   *string `json:"remark"`
}

type UserController struct {
	svc *service.UserService
}

func NewUserController() *UserController {
	return &UserController{svc: service.NewUserService()}
}

// Create POST /api/v1/users
func (ctl *UserController) Create(c *gin.Context) {
	var req createUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	user, err := ctl.svc.Create(req.Name, req.Password, req.Quota, req.Remark)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "创建失败: "+err.Error())
		return
	}
	response.Success(c, http.StatusCreated, user)
}

// List GET /api/v1/users
func (ctl *UserController) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))

	f := service.UserListFilter{
		Name:          c.Query("name"),
		Remark:        c.Query("remark"),
		Sort:          c.DefaultQuery("sort", "id"),
		Order:         c.DefaultQuery("order", "desc"),
		Page:          page,
		Size:          size,
		CreatedAfter:  stringPtr(c.Query("created_after")),
		CreatedBefore: stringPtr(c.Query("created_before")),
	}

	if v := c.Query("quota_min"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			f.QuotaMin = &n
		}
	}
	if v := c.Query("quota_max"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			f.QuotaMax = &n
		}
	}

	result, err := ctl.svc.List(f)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "查询失败: "+err.Error())
		return
	}
	response.Success(c, http.StatusOK, result)
}

// Get GET /api/v1/users/:id
func (ctl *UserController) Get(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "无效的ID")
		return
	}

	user, err := ctl.svc.GetByID(id, c.Query("name"))
	if err != nil {
		response.Error(c, http.StatusNotFound, "用户不存在")
		return
	}
	response.Success(c, http.StatusOK, user)
}

// Update PUT /api/v1/users/:id
func (ctl *UserController) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "无效的ID")
		return
	}

	var req updateUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	updates := map[string]any{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Password != "" {
		updates["password"] = req.Password
	}
	if req.Quota != nil {
		updates["quota"] = *req.Quota
	}
	if req.Remark != nil {
		updates["remark"] = *req.Remark
	}
	if len(updates) == 0 {
		response.Error(c, http.StatusBadRequest, "无更新字段")
		return
	}

	user, err := ctl.svc.Update(id, updates)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			response.Error(c, http.StatusNotFound, "用户不存在")
			return
		}
		response.Error(c, http.StatusInternalServerError, "更新失败: "+err.Error())
		return
	}
	response.Success(c, http.StatusOK, user)
}

// Delete DELETE /api/v1/users/:id
func (ctl *UserController) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "无效的ID")
		return
	}

	if err := ctl.svc.Delete(id); err != nil {
		if err == gorm.ErrRecordNotFound {
			response.Error(c, http.StatusNotFound, "用户不存在")
			return
		}
		response.Error(c, http.StatusInternalServerError, "删除失败: "+err.Error())
		return
	}
	response.Success(c, http.StatusOK, nil)
}

func stringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
