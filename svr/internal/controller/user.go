package controller

import (
	"net/http"
	"strconv"

	"wiki/internal/model"
	"wiki/internal/service"
	"wiki/pkg/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type UserController = model.UserController

// NewUserController 创建用户控制器并绑定处理函数。
func NewUserController() *UserController {
	svc := service.NewUserService()
	return &model.UserController{
		CreateFunc: func(ctx *gin.Context) { create(ctx, svc) },
		ListFunc:   func(ctx *gin.Context) { list(ctx, svc) },
		GetFunc:    func(ctx *gin.Context) { get(ctx, svc) },
		UpdateFunc: func(ctx *gin.Context) { update(ctx, svc) },
		DeleteFunc: func(ctx *gin.Context) { deleteUser(ctx, svc) },
	}
}

// create 处理创建用户的 HTTP 请求。
func create(c *gin.Context, svc *service.UserService) {
	var req model.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	user, err := svc.Create(req.Name, req.Password, req.Quota, req.Remark)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "创建失败: "+err.Error())
		return
	}
	response.Success(c, http.StatusCreated, user)
}

// list 处理查询用户列表的 HTTP 请求。
func list(c *gin.Context, svc *service.UserService) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))

	f := model.UserListFilter{
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

	result, err := svc.List(f)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "查询失败: "+err.Error())
		return
	}
	response.Success(c, http.StatusOK, result)
}

// get 处理查询用户详情的 HTTP 请求。
func get(c *gin.Context, svc *service.UserService) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "无效的ID")
		return
	}

	user, err := svc.GetByID(id, c.Query("name"))
	if err != nil {
		response.Error(c, http.StatusNotFound, "用户不存在")
		return
	}
	response.Success(c, http.StatusOK, user)
}

// update 处理更新用户的 HTTP 请求。
func update(c *gin.Context, svc *service.UserService) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "无效的ID")
		return
	}

	var req model.UpdateUserRequest
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

	user, err := svc.Update(id, updates)
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

// deleteUser 处理删除用户的 HTTP 请求。
func deleteUser(c *gin.Context, svc *service.UserService) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "无效的ID")
		return
	}

	if err := svc.Delete(id); err != nil {
		if err == gorm.ErrRecordNotFound {
			response.Error(c, http.StatusNotFound, "用户不存在")
			return
		}
		response.Error(c, http.StatusInternalServerError, "删除失败: "+err.Error())
		return
	}
	response.Success(c, http.StatusOK, nil)
}

// stringPtr 将非空字符串转换为指针。
func stringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
