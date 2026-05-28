package service

import (
	"aisearch/internal/model"
	"aisearch/pkg/auth"
	"aisearch/pkg/database"
	"aisearch/pkg/snowflake"

	"gorm.io/gorm"
)

type UserService struct{}

func NewUserService() *UserService {
	return &UserService{}
}

func (s *UserService) Create(name, password string, quota int64, remark string) (*model.User, error) {
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
	Total int64
	List  []*model.User
}

func (s *UserService) List(f UserListFilter) (*UserListResult, error) {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.Size < 1 || f.Size > 100 {
		f.Size = 20
	}

	db := database.DB.Model(&model.User{}).Where("is_deleted = 0")
	db = s.applyFilters(db, f)

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

	return &UserListResult{Total: total, List: users}, nil
}

func (s *UserService) applyFilters(db *gorm.DB, f UserListFilter) *gorm.DB {
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

func (s *UserService) GetByID(id int64, name string) (*model.User, error) {
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

func (s *UserService) Update(id int64, updates map[string]any) (*model.User, error) {
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

func (s *UserService) Delete(id int64) error {
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
