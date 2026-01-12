package models

import (
	"time"

	"github.com/pgvector/pgvector-go"
	"gorm.io/gorm"
)

// KnowledgeChunk 知识块模型
type KnowledgeChunk struct {
	ID        string          `gorm:"type:varchar(255);primary_key" json:"id"`
	Embedding pgvector.Vector `gorm:"type:vector(768)" json:"embedding,omitempty"`
	Text      string          `gorm:"type:text;not null" json:"text"`
	KBID      string          `gorm:"column:kb_id;type:varchar(255);not null;index:idx_kb_doc" json:"kb_id"`
	DocID     string          `gorm:"column:doc_id;type:varchar(255);not null;index:idx_kb_doc" json:"doc_id"`
	ChunkID   int             `gorm:"column:chunk_id;not null" json:"chunk_id"`
	Source    string          `gorm:"type:varchar(255);not null" json:"source"`
	MetaData  JSONMap         `gorm:"column:meta_data;type:jsonb" json:"meta_data,omitempty"`
	CreatedAt time.Time       `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time       `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt  `gorm:"index" json:"-"`
}

// TableName 指定表名
func (KnowledgeChunk) TableName() string {
	return "knowledge_chunks"
}
