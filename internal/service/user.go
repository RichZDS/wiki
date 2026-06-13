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
	if err := database.DB.Create(&user).Error; err != nil {
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

	db := database.DB.Model(&model.User{}).Where("is_deleted = 0")
	db = applyFilters(db, f)

	if f.Sort == "" {
		f.Sort = "id"
	}
	if f.Order != "asc" {
		f.Order = "desc"
	}
	db = db.Order(f.Sort + " " + f.Order)

	var total int64
	db.Count(&total)

	var users []*model.User
	db.Offset((f.Page - 1) * f.Size).Limit(f.Size).Find(&users)

	return &model.UserListResult{Total: total, List: users}, nil
}

// applyFilters 将用户筛选条件应用到数据库查询。
func applyFilters(db *gorm.DB, f model.UserListFilter) *gorm.DB {
	if f.Name != "" {
		db = db.Where("name LIKE ?", "%"+f.Name+"%")
	}
	if f.Remark != "" {
		db = db.Where("remark LIKE ?", "%"+f.Remark+"%")
	}
	if f.QuotaMin != nil {
		db = db.Where("quota >= ?", *f.QuotaMin)
	}
	if f.QuotaMax != nil {
		db = db.Where("quota <= ?", *f.QuotaMax)
	}
	if f.CreatedAfter != nil {
		db = db.Where("created_at >= ?", *f.CreatedAfter)
	}
	if f.CreatedBefore != nil {
		db = db.Where("created_at <= ?", *f.CreatedBefore)
	}
	return db
}

// getByID 根据编号和可选名称查询用户。
func getByID(id int64, name string) (*model.User, error) {
	db := database.DB.Model(&model.User{}).Where("is_deleted = 0")
	if name != "" {
		db = db.Where("name LIKE ?", "%"+name+"%")
	}
	var user model.User
	if err := db.Where("id = ?", id).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
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

	result := database.DB.Model(&model.User{}).Where("id = ? AND is_deleted = 0", id).Updates(updates)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	var user model.User
	database.DB.Where("id = ?", id).First(&user)
	return &user, nil
}

// deleteUser 对指定用户执行软删除。
func deleteUser(id int64) error {
	result := database.DB.Model(&model.User{}).Where("id = ? AND is_deleted = 0", id).
		Update("is_deleted", 1)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
