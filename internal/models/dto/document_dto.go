package dto

import (
	"github.com/Sesame2/go-admin/internal/models"
	"github.com/google/uuid"
)

type DocStatus string

const (
	DocPending    DocStatus = "pending"
	DocProcessing DocStatus = "processing"
	DocCompleted  DocStatus = "completed"
)

type CreateDocumentInput struct {
	Name        string    `json:"name" binding:"required,min=1,max=100" example:"文档标题"`
	Description string    `json:"description" binding:"required,min=1,max=500" example:"文档描述，简要介绍文档内容和目的"`
	KBID        string    `json:"kb_id" binding:"required,uuid" example:"123e4567-e89b-12d3-a456-426614174000"`
	UserID      uuid.UUID `json:"user_id"`
	Status      DocStatus `json:"status"`
}

type UpdateDocumentInput struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	FileName    *string `json:"file_name,omitempty"`
	Status      *string `json:"status,omitempty"`
}

// DocumentListResult 文档列表返回结果
type DocumentListResult struct {
	Documents []*models.Document `json:"documents"`
	Total     int64              `json:"total"`
	Page      int                `json:"page"`
	PageSize  int                `json:"page_size"`
}
