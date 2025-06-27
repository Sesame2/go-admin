package dto

import "github.com/Sesame2/go-admin/internal/models/ent"

type CreateKnowledgeBaseInput struct {
	KBName    string  `json:"knowledgebase_name"`
	DatasetID *string `json:"dataset_id,omitempty"`
}

type UpdateKnowledgeBaseInput struct {
	KBName    *string `json:"knowledgebase_name"`
	DatasetID *string `json:"dataset_id,omitempty"`
}

type KnowledgeBaseListResult struct {
	TotalCount  int                  `json:"total_count"`
	TotalPages  int                  `json:"total_pages"`
	CurrentPage int                  `json:"current_page"`
	PageSize    int                  `json:"page_size"`
	Result      []*ent.KnowledgeBase `json:"result"`
}
