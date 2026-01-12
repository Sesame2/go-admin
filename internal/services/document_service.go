package services

import (
	"context"
	"mime/multipart"

	"github.com/Sesame2/go-admin/internal/logger"
	"github.com/Sesame2/go-admin/internal/models"
	"github.com/Sesame2/go-admin/internal/models/dto"
	"github.com/Sesame2/go-admin/internal/repository"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type DocumentService struct {
	repo   *repository.DocumentRepository
	kbRepo *repository.KnowledgeBaseRepository
	log    *zap.Logger
}

func NewDocumentService(docRepo *repository.DocumentRepository, kbRepo *repository.KnowledgeBaseRepository) *DocumentService {
	return &DocumentService{
		repo:   docRepo,
		kbRepo: kbRepo,
		log:    logger.NewModuleLogger("DocumentService"),
	}
}

// UploadDocument 上传文档到指定知识库
func (s *DocumentService) UploadDocument(ctx context.Context, kbID uuid.UUID, file *multipart.FileHeader) (*models.Document, error) {
	s.log.Info("上传文档",
		zap.String("kb_id", kbID.String()),
		zap.String("filename", file.Filename),
		zap.Int64("size", file.Size))

	// 检查知识库是否存在
	_, err := s.kbRepo.GetByID(ctx, kbID)
	if err != nil {
		s.log.Error("知识库不存在", zap.Error(err))
		return nil, err
	}

	// TODO: 实现文件上传到存储服务的逻辑
	// TODO: 发送消息到消息队列进行文档处理

	doc := &models.Document{
		KnowledgeBaseID: kbID,
		FileName:        file.Filename,
		FileSize:        file.Size,
		FileType:        file.Header.Get("Content-Type"),
		Status:          "pending",
	}

	if err := s.repo.Create(ctx, doc); err != nil {
		s.log.Error("创建文档记录失败", zap.Error(err))
		return nil, err
	}

	return doc, nil
}

// GetDocument 根据ID获取文档
func (s *DocumentService) GetDocument(ctx context.Context, id uuid.UUID) (*models.Document, error) {
	s.log.Debug("获取文档", zap.String("id", id.String()))
	return s.repo.GetByID(ctx, id)
}

// ListDocuments 获取知识库下的文档列表
func (s *DocumentService) ListDocuments(ctx context.Context, kbID uuid.UUID, page, pageSize int) (*dto.DocumentListResult, error) {
	s.log.Debug("获取文档列表",
		zap.String("kb_id", kbID.String()),
		zap.Int("page", page),
		zap.Int("page_size", pageSize))

	offset := (page - 1) * pageSize

	docs, err := s.repo.ListByKnowledgeBase(ctx, kbID, pageSize, offset)
	if err != nil {
		s.log.Error("获取文档列表失败", zap.Error(err))
		return nil, err
	}

	total, err := s.repo.CountByKnowledgeBase(ctx, kbID)
	if err != nil {
		s.log.Error("获取文档总数失败", zap.Error(err))
		return nil, err
	}

	return &dto.DocumentListResult{
		Documents: docs,
		Total:     total,
		Page:      page,
		PageSize:  pageSize,
	}, nil
}

// UpdateDocument 更新文档信息
func (s *DocumentService) UpdateDocument(ctx context.Context, id uuid.UUID, input *dto.UpdateDocumentInput) (*models.Document, error) {
	s.log.Info("更新文档", zap.String("id", id.String()))

	doc, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.log.Error("文档不存在", zap.Error(err))
		return nil, err
	}

	if input.FileName != nil {
		doc.FileName = *input.FileName
	}
	if input.Status != nil {
		doc.Status = *input.Status
	}

	if err := s.repo.Update(ctx, doc); err != nil {
		s.log.Error("更新文档失败", zap.Error(err))
		return nil, err
	}

	return doc, nil
}

// DeleteDocument 删除文档
func (s *DocumentService) DeleteDocument(ctx context.Context, id uuid.UUID) error {
	s.log.Info("删除文档", zap.String("id", id.String()))

	// TODO: 删除存储服务中的文件
	// TODO: 删除相关的知识块

	return s.repo.Delete(ctx, id)
}
