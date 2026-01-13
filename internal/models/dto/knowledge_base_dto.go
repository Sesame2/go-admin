package dto

import "github.com/Sesame2/go-admin/internal/models"

type CreateKnowledgeBaseInput struct {
	Name        string `json:"name" binding:"required,min=1,max=255"`
	Description string `json:"description,omitempty"`
}

type UpdateKnowledgeBaseInput struct {
	Name        *string `json:"name,omitempty" binding:"omitempty,min=1,max=255"`
	Description *string `json:"description,omitempty"`
}

type KnowledgeBaseListResult struct {
	TotalCount  int                     `json:"total_count"`
	TotalPages  int                     `json:"total_pages"`
	CurrentPage int                     `json:"current_page"`
	PageSize    int                     `json:"page_size"`
	Result      []*models.KnowledgeBase `json:"result"`
}
