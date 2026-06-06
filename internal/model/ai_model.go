package model

type AIModel struct {
	ID        int64  `gorm:"primaryKey" json:"id"`
	ModelName string `gorm:"column:model_name;type:varchar(128);uniqueIndex;not null" json:"model_name"`
	IsUsed    int8   `gorm:"column:is_used;not null;default:1" json:"is_used"`
}

// TableName 返回当前模型对应的数据库表名。
func (AIModel) TableName() string {
	return "ai_model"
}
