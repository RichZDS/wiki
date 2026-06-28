package model

type AIModel struct {
	ID         int64   `gorm:"primaryKey" json:"id"`
	ModelName  string  `gorm:"column:model_name;type:varchar(128);uniqueIndex;not null" json:"model_name"`
	ModelId    string  `gorm:"column:model_id;type:varchar(255);index:idx_model_id;not null" json:"model_id"`
	IsUsed     int8    `gorm:"column:is_used;not null;default:1;index:idx_used_check,priority:1" json:"is_used"`
	FailReason *string `gorm:"column:fail_reason;type:text" json:"fail_reason,omitempty"`
	APIKey     *string `gorm:"column:api_key;type:varchar(255)" json:"-"`
	Provider   string  `gorm:"column:provider;type:varchar(255);not null;default:''" json:"provider"`
	IsCheck    *int8   `gorm:"column:is_check;index:idx_used_check,priority:2" json:"is_check,omitempty"`
	TestModel  *string `gorm:"column:test_model;type:varchar(255)" json:"test_model,omitempty"`
	BaseURL    *string `gorm:"column:base_url;type:varchar(512)" json:"base_url,omitempty"`
	Ability    *string `gorm:"column:ability;type:varchar(255)" json:"ability,omitempty"`
}

// TableName 返回当前模型对应的数据库表名。
func (AIModel) TableName() string {
	return "ai_model"
}

func (m AIModel) APIKeyValue() string {
	if m.APIKey == nil {
		return ""
	}
	return *m.APIKey
}

func (m AIModel) FailReasonValue() string {
	if m.FailReason == nil {
		return ""
	}
	return *m.FailReason
}

func (m AIModel) BaseURLValue() string {
	if m.BaseURL == nil {
		return ""
	}
	return *m.BaseURL
}

// AbilityValue 返回模型的能力标签，空值时返回空字符串。
func (m AIModel) AbilityValue() string {
	if m.Ability == nil {
		return ""
	}
	return *m.Ability
}
