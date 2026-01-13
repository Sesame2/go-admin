package dto

import (
	"github.com/Sesame2/go-admin/internal/models"
	"github.com/google/uuid"
)

// UploadDocumentInput 上传文档请求
type UploadDocumentInput struct {
	KnowledgeBaseID string `json:"knowledge_base_id" form:"knowledge_base_id" binding:"required,uuid" example:"123e4567-e89b-12d3-a456-426614174000"`
	// 文件通过 multipart form 上传
}

// CreateDocumentInput 创建文档记录（上传后调用）
type CreateDocumentInput struct {
	KnowledgeBaseID uuid.UUID  `json:"knowledge_base_id" binding:"required"`
	Filename        string     `json:"filename" binding:"required,min=1,max=512"`
	FileType        string     `json:"file_type" binding:"required"`
	FileSize        int64      `json:"file_size"`
	MinioPath       string     `json:"minio_path" binding:"required"`
	UserID          *uuid.UUID `json:"user_id,omitempty"`
}

// UpdateDocumentInput 更新文档
type UpdateDocumentInput struct {
	Status            *string `json:"status,omitempty"`
	ErrorMessage      *string `json:"error_message,omitempty"`
	ChunkCount        *int    `json:"chunk_count,omitempty"`
	AtomQuestionCount *int    `json:"atom_question_count,omitempty"`
}

// ParseDocumentInput 解析文档请求
type ParseDocumentInput struct {
	DocumentID string `json:"document_id" binding:"required,uuid"`
}

// DocumentListInput 文档列表查询
type DocumentListInput struct {
	KnowledgeBaseID string `form:"knowledge_base_id" binding:"required,uuid"`
	Page            int    `form:"page,default=1" binding:"min=1"`
	PageSize        int    `form:"page_size,default=10" binding:"min=1,max=100"`
	Status          string `form:"status,omitempty"`
}

// DocumentListResult 文档列表返回结果
type DocumentListResult struct {
	Documents []*models.Document `json:"documents"`
	Total     int64              `json:"total"`
	Page      int                `json:"page"`
	PageSize  int                `json:"page_size"`
}

// DocumentDetail 文档详情
type DocumentDetail struct {
	*models.Document
	KnowledgeBaseName string `json:"knowledge_base_name,omitempty"`
}

// DocumentParseMessage RabbitMQ 文档解析消息
type DocumentParseMessage struct {
	DocumentID      string `json:"document_id"`
	KnowledgeBaseID string `json:"knowledge_base_id"`
	Filename        string `json:"filename"`
	FileType        string `json:"file_type"`
	MinioPath       string `json:"minio_path"`
}
