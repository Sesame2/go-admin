package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Document 文档模型
type Document struct {
	ID              uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name            string         `gorm:"type:varchar(255);not null" json:"name"`
	Description     string         `gorm:"type:text;not null" json:"description"`
	DocNum          int64          `gorm:"default:0" json:"doc_num"`
	UserID          uuid.UUID      `gorm:"type:uuid;not null" json:"user_id"`
	KnowledgeBaseID uuid.UUID      `gorm:"type:uuid" json:"knowledge_base_id"`
	FileName        string         `gorm:"type:varchar(255)" json:"file_name"`
	FileSize        int64          `gorm:"default:0" json:"file_size"`
	FileType        string         `gorm:"type:varchar(100)" json:"file_type"`
	Status          string         `gorm:"type:varchar(50);default:'pending'" json:"status"`
	CreatedAt       time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`

	// 关联
	Owner         *User          `gorm:"foreignKey:UserID" json:"owner,omitempty"`
	KnowledgeBase *KnowledgeBase `gorm:"foreignKey:KnowledgeBaseID" json:"knowledge_base,omitempty"`
}

// TableName 指定表名
func (Document) TableName() string {
	return "documents"
}

// BeforeCreate 创建前钩子
func (d *Document) BeforeCreate(tx *gorm.DB) error {
	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}
	return nil
}
