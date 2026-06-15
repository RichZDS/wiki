package model

import "time"

type User struct {
	ID        int64     `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"type:varchar(64);uniqueIndex;not null" json:"name"`
	Password  string    `gorm:"type:varchar(256);not null" json:"-"`
	Quota     int64     `gorm:"not null;default:0" json:"quota"`
	Remark    string    `gorm:"type:varchar(512);not null;default:''" json:"remark"`
	IsDeleted int8      `gorm:"not null;default:0" json:"is_deleted"`
	CreatedAt time.Time `gorm:"not null" json:"created_at"`
	UpdatedAt time.Time `gorm:"not null" json:"updated_at"`
}

// TableName 返回当前模型对应的数据库表名。
func (User) TableName() string {
	return "user"
}
