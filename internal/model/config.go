package model

// Config 配置表
type Config struct {
	Key   string `json:"key" gorm:"primaryKey;column:key;type:varchar(255);not null"`
	Value string `json:"value" gorm:"column:value;type:json;not null"`
}
