package model

type AIModel struct {
	ID         int64  `gorm:"primaryKey" json:"id"`
	ModelName  string `gorm:"column:model_name;type:varchar(128);uniqueIndex;not null" json:"model_name"`
	ModelId    string `gorm:"column:model_id;type:varchar(128);not null" json:"model_id"`
	APIKey     string `gorm:"column:api_key;type:varchar(512);not null;default:''" json:"-"`
	Provider   string `gorm:"column:provider;type:varchar(32);not null;default:''" json:"provider"`
	IsUsed     int8   `gorm:"column:is_used;not null;default:1" json:"is_used"`
	FailReason string `gorm:"column:fail_reason;type:varchar(512);not null;default:''" json:"fail_reason"`
}

// TableName 返回当前模型对应的数据库表名。
func (AIModel) TableName() string {
	return "ai_model"
}
