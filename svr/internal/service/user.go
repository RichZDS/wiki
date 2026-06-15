package service

import (
	"wiki/internal/model"
	"wiki/pkg/auth"
	"wiki/pkg/database"
	"wiki/pkg/snowflake"

	"gorm.io/gorm"
)

type UserService = model.UserService
type UserListFilter = model.UserListFilter
type UserListResult = model.UserListResult

// NewUserService 创建并初始化对应的实例。
func NewUserService() *UserService {
	return &model.UserService{
		CreateFunc:  create,
		ListFunc:    list,
		GetByIDFunc: getByID,
		UpdateFunc:  update,
		DeleteFunc:  deleteUser,
	}
}

// create 创建用户并持久化到数据库。
func create(name, password string, quota int64, remark string) (*model.User, error) {
	hash, err := auth.Hash(password)
	if err != nil {
		return nil, err
	}

	user := model.User{
		ID:       snowflake.Next(),
		Name:     name,
		Password: hash,
		Quota:    quota,
		Remark:   remark,
	}
	if err := model.InsertUser(database.DB, &user); err != nil {
		return nil, err
	}
	return &user, nil
}

// list 查询符合筛选条件的用户列表。
func list(f model.UserListFilter) (*model.UserListResult, error) {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.Size < 1 || f.Size > 100 {
		f.Size = 20
	}

	total, err := model.CountUsers(database.DB, f)
	if err != nil {
		return nil, err
	}

	users, err := model.ListUsers(database.DB, f)
	if err != nil {
		return nil, err
	}

	return &model.UserListResult{Total: total, List: users}, nil
}

// getByID 根据编号和可选名称查询用户。
func getByID(id int64, name string) (*model.User, error) {
	return model.GetUserByID(database.DB, id, name)
}

// update 更新指定用户并返回最新数据。
func update(id int64, updates map[string]any) (*model.User, error) {
	if pw, ok := updates["password"]; ok {
		hash, err := auth.Hash(pw.(string))
		if err != nil {
			return nil, err
		}
		updates["password"] = hash
	}

	affected, err := model.UpdateUser(database.DB, id, updates)
	if err != nil {
		return nil, err
	}
	if affected == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	return model.RefreshUser(database.DB, id)
}

// deleteUser 对指定用户执行软删除。
func deleteUser(id int64) error {
	affected, err := model.SoftDeleteUser(database.DB, id)
	if err != nil {
		return err
	}
	if affected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
