package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// JSONMap 自定义JSON类型
type JSONMap map[string]interface{}

// Value 实现 driver.Valuer 接口
func (j JSONMap) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan 实现 sql.Scanner 接口
func (j *JSONMap) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, j)
}

// KnowledgeBase 知识库模型
type KnowledgeBase struct {
	ID            uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name          string         `gorm:"type:varchar(255);not null" json:"name"`
	Description   string         `gorm:"type:text" json:"description"`
	Model         string         `gorm:"type:varchar(100)" json:"model"`
	DatasetID     *string        `gorm:"type:varchar(255)" json:"dataset_id,omitempty"`
	Tags          JSONMap        `gorm:"type:jsonb" json:"tags,omitempty"`
	DocNum        int64          `gorm:"default:0" json:"doc_num"`
	DocumentCount int64          `gorm:"default:0" json:"document_count"`
	CreatedAt     time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (KnowledgeBase) TableName() string {
	return "knowledge_bases"
}

// BeforeCreate 创建前钩子
func (kb *KnowledgeBase) BeforeCreate(tx *gorm.DB) error {
	if kb.ID == uuid.Nil {
		kb.ID = uuid.New()
	}
	return nil
}
