package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/pgvector/pgvector-go"
)

// 存储文档切分后的片段，每个 chunk 包含原文内容和对应的向量嵌入
type DocumentChunk struct {
	ID              uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	DocumentID      uuid.UUID `gorm:"type:uuid;not null;index" json:"document_id"`
	KnowledgeBaseID uuid.UUID `gorm:"type:uuid;not null;index" json:"knowledge_base_id"`

	// 分片内容
	Content string `gorm:"type:text;not null" json:"content"`

	// 分片位置信息
	ChunkIndex int  `gorm:"not null" json:"chunk_index"`              // 分片序号（从0开始）
	StartChar  *int `gorm:"type:integer" json:"start_char,omitempty"` // 在原文中的起始位置
	EndChar    *int `gorm:"type:integer" json:"end_char,omitempty"`   // 在原文中的结束位置

	// 分片向量嵌入（用于语义检索）- 使用 1024 维度匹配 Qwen embedding
	Embedding pgvector.Vector `gorm:"type:vector(1024)" json:"embedding,omitempty"`

	// 元数据（页码、标题层级等）
	MetaData JSONMap `gorm:"type:jsonb;default:'{}'" json:"meta_data,omitempty"`

	// 时间戳
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

// TableName 指定表名
func (DocumentChunk) TableName() string {
	return "document_chunks"
}

// AtomQuestion 原子问题表 - 与 Python 端保持一致
// Pike-RAG 核心：通过原子问题作为检索的中间层，提升检索准确性
type AtomQuestion struct {
	ID              uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ChunkID         uuid.UUID `gorm:"type:uuid;not null;index" json:"chunk_id"`
	DocumentID      uuid.UUID `gorm:"type:uuid;not null;index" json:"document_id"`
	KnowledgeBaseID uuid.UUID `gorm:"type:uuid;not null;index" json:"knowledge_base_id"`

	// 原子问题内容
	Question string `gorm:"type:text;not null" json:"question"`

	// 原子问题的向量嵌入（用于问题-问题匹配）- 使用 1024 维度匹配 Qwen embedding
	Embedding pgvector.Vector `gorm:"type:vector(1024)" json:"embedding,omitempty"`

	// 问题类型/标签（可选）
	QuestionType *string `gorm:"type:varchar(50)" json:"question_type,omitempty"`

	// 元数据
	MetaData JSONMap `gorm:"type:jsonb;default:'{}'" json:"meta_data,omitempty"`

	// 时间戳
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`

	// 关联
	Chunk *DocumentChunk `gorm:"foreignKey:ChunkID" json:"chunk,omitempty"`
}

// TableName 指定表名
func (AtomQuestion) TableName() string {
	return "atom_questions"
}
