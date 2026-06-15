package model

import (
	"fmt"

	"gorm.io/gorm"
)

// InsertUser 向数据库插入一条新用户记录。
func InsertUser(db *gorm.DB, user *User) error {
	if err := db.Create(user).Error; err != nil {
		return fmt.Errorf("insert user: %w", err)
	}
	return nil
}

// CountUsers 根据筛选条件统计符合条件的用户总数。
func CountUsers(db *gorm.DB, f UserListFilter) (int64, error) {
	var total int64
	q := db.Model(&User{}).Where("is_deleted = 0")
	q = applyUserFilters(q, f)
	if err := q.Count(&total).Error; err != nil {
		return 0, fmt.Errorf("count users: %w", err)
	}
	return total, nil
}

// ListUsers 分页查询符合条件的用户列表，并按指定排序字段返回。
func ListUsers(db *gorm.DB, f UserListFilter) ([]*User, error) {
	q := db.Model(&User{}).Where("is_deleted = 0")
	q = applyUserFilters(q, f)

	sort, order := normalizeSortAndOrder(f.Sort, f.Order)
	q = q.Order(sort + " " + order)

	var users []*User
	if err := q.Offset((f.Page - 1) * f.Size).Limit(f.Size).Find(&users).Error; err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	return users, nil
}

// GetUserByID 根据 ID 和可选的名称模糊匹配查询单个用户。
func GetUserByID(db *gorm.DB, id int64, name string) (*User, error) {
	q := db.Model(&User{}).Where("is_deleted = 0").Where("id = ?", id)
	if name != "" {
		q = q.Where("name LIKE ?", "%"+name+"%")
	}
	var user User
	if err := q.First(&user).Error; err != nil {
		return nil, fmt.Errorf("get user by id (id=%d): %w", id, err)
	}
	return &user, nil
}

// UpdateUser 更新指定用户的字段，返回受影响行数。
func UpdateUser(db *gorm.DB, id int64, updates map[string]any) (int64, error) {
	result := db.Model(&User{}).Where("id = ? AND is_deleted = 0", id).Updates(updates)
	if result.Error != nil {
		return 0, fmt.Errorf("update user (id=%d): %w", id, result.Error)
	}
	return result.RowsAffected, nil
}

// SoftDeleteUser 对指定用户执行软删除（is_deleted = 1）。
func SoftDeleteUser(db *gorm.DB, id int64) (int64, error) {
	result := db.Model(&User{}).Where("id = ? AND is_deleted = 0", id).Update("is_deleted", 1)
	if result.Error != nil {
		return 0, fmt.Errorf("delete user (id=%d): %w", id, result.Error)
	}
	return result.RowsAffected, nil
}

// RefreshUser 重新从数据库加载用户数据。
func RefreshUser(db *gorm.DB, id int64) (*User, error) {
	var user User
	if err := db.Where("id = ?", id).First(&user).Error; err != nil {
		return nil, fmt.Errorf("refresh user (id=%d): %w", id, err)
	}
	return &user, nil
}

// applyUserFilters 将筛选条件应用到 GORM 查询。
func applyUserFilters(db *gorm.DB, f UserListFilter) *gorm.DB {
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

// normalizeSortAndOrder 对排序字段和方向提供默认值。
func normalizeSortAndOrder(sort, order string) (string, string) {
	if sort == "" {
		sort = "id"
	}
	if order != "asc" {
		order = "desc"
	}
	return sort, order
}
