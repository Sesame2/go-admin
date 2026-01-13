package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// DocumentStatus 文档处理状态
type DocumentStatus string

const (
	DocumentStatusPending     DocumentStatus = "pending"     // 待处理
	DocumentStatusDownloading DocumentStatus = "downloading" // 下载中
	DocumentStatusParsing     DocumentStatus = "parsing"     // 解析中
	DocumentStatusChunking    DocumentStatus = "chunking"    // 分片中
	DocumentStatusTagging     DocumentStatus = "tagging"     // 原子问题生成中
	DocumentStatusEmbedding   DocumentStatus = "embedding"   // 向量化中
	DocumentStatusCompleted   DocumentStatus = "completed"   // 处理完成
	DocumentStatusFailed      DocumentStatus = "failed"      // 处理失败
)

// Document 文档模型 - 与 Python 端保持一致
type Document struct {
	ID              uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	KnowledgeBaseID uuid.UUID  `gorm:"type:uuid;not null;index" json:"knowledge_base_id"`
	UserID          *uuid.UUID `gorm:"type:uuid;index" json:"user_id,omitempty"`

	// 文档基本信息
	Filename  string `gorm:"type:varchar(512);not null" json:"filename"`
	FileType  string `gorm:"type:varchar(50);not null" json:"file_type"`    // pdf, docx, pptx, markdown
	FileSize  *int64 `gorm:"type:integer" json:"file_size,omitempty"`       // 文件大小（字节）
	MinioPath string `gorm:"type:varchar(1024);not null" json:"minio_path"` // MinIO 存储路径

	// 处理状态 - 使用 PostgreSQL ENUM 类型（与 Python 模型保持一致）
	Status       DocumentStatus `gorm:"type:document_status;not null;default:'pending';index" json:"status"`
	ErrorMessage *string        `gorm:"type:text" json:"error_message,omitempty"` // 错误信息

	// 处理结果统计
	ChunkCount        int `gorm:"default:0" json:"chunk_count"`         // 分片数量
	AtomQuestionCount int `gorm:"default:0" json:"atom_question_count"` // 原子问题数量

	// 元数据（自定义扩展字段）
	MetaData JSONMap `gorm:"type:jsonb;default:'{}'" json:"meta_data,omitempty"`

	// 时间戳
	CreatedAt   time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	ProcessedAt *time.Time     `json:"processed_at,omitempty"` // 处理完成时间
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

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

// IsProcessing 判断文档是否正在处理中
func (d *Document) IsProcessing() bool {
	return d.Status != DocumentStatusPending &&
		d.Status != DocumentStatusCompleted &&
		d.Status != DocumentStatusFailed
}

// CanProcess 判断文档是否可以处理
func (d *Document) CanProcess() bool {
	return d.Status == DocumentStatusPending || d.Status == DocumentStatusFailed
}
