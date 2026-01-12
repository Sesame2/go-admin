package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/Sesame2/go-admin/internal/logger"
	"github.com/Sesame2/go-admin/internal/models"
	"github.com/Sesame2/go-admin/internal/models/dto"
	"github.com/Sesame2/go-admin/internal/repository"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type KnowledgeBaseService struct {
	repo *repository.KnowledgeBaseRepository
	log  *zap.Logger
}

func NewKnowledgeBaseService(repo *repository.KnowledgeBaseRepository) *KnowledgeBaseService {
	return &KnowledgeBaseService{
		repo: repo,
		log:  logger.NewModuleLogger("KnowledgeBaseService"),
	}
}

func (s *KnowledgeBaseService) CreateKnowledgeBase(ctx context.Context, input *dto.CreateKnowledgeBaseInput) (*models.KnowledgeBase, error) {
	kb := &models.KnowledgeBase{
		Name:      input.KBName,
		DatasetID: input.DatasetID,
	}

	err := s.repo.Create(ctx, kb)
	if err != nil {
		s.log.Error("创建知识库失败", zap.Error(err))
		return nil, err
	}
	s.log.Info("创建知识库成功", zap.String("kb_name", input.KBName))
	return kb, nil
}

func (s *KnowledgeBaseService) GetAllKnowledgeBase(ctx context.Context, page, pageSize int) (dto.KnowledgeBaseListResult, error) {
	count, err := s.repo.Count(ctx)
	if err != nil {
		s.log.Error("获取知识库数量失败", zap.Error(err))
		return dto.KnowledgeBaseListResult{}, err
	}
	s.log.Info("知识库总数", zap.Int64("count", count))

	offset := (page - 1) * pageSize
	kbs, err := s.repo.GetAll(ctx, pageSize, offset)
	if err != nil {
		s.log.Error("查询知识库失败", zap.Error(err))
		return dto.KnowledgeBaseListResult{}, err
	}
	s.log.Info("分页查询知识库成功", zap.Int("count", len(kbs)))

	totalPages := int(count+int64(pageSize)-1) / pageSize

	return dto.KnowledgeBaseListResult{
		TotalCount:  int(count),
		TotalPages:  totalPages,
		CurrentPage: page,
		PageSize:    pageSize,
		Result:      kbs,
	}, nil
}

func (s *KnowledgeBaseService) GetKnowledgeBaseByID(ctx context.Context, id uuid.UUID) (*models.KnowledgeBase, error) {
	kb, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			s.log.Warn("知识库不存在", zap.String("id", id.String()))
			return nil, fmt.Errorf("知识库不存在")
		}
		s.log.Error("查询知识库失败", zap.String("id", id.String()), zap.Error(err))
		return nil, err
	}
	s.log.Info("查询知识库成功", zap.String("id", id.String()))
	return kb, nil
}

func (s *KnowledgeBaseService) UpdateKnowledgeBase(ctx context.Context, id uuid.UUID, input *dto.UpdateKnowledgeBaseInput) (*models.KnowledgeBase, error) {
	kb, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			s.log.Warn("知识库不存在", zap.String("id", id.String()))
			return nil, fmt.Errorf("知识库不存在")
		}
		s.log.Error("查询知识库失败", zap.String("id", id.String()), zap.Error(err))
		return nil, err
	}

	if input.KBName != nil {
		kb.Name = *input.KBName
	}
	if input.DatasetID != nil {
		kb.DatasetID = input.DatasetID
	}

	err = s.repo.Update(ctx, kb)
	if err != nil {
		s.log.Error("更新知识库失败", zap.String("id", id.String()), zap.Error(err))
		return nil, err
	}
	s.log.Info("更新知识库成功", zap.String("id", id.String()))
	return kb, nil
}

func (s *KnowledgeBaseService) DeleteKnowledgeBase(ctx context.Context, id uuid.UUID) error {
	err := s.repo.Delete(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			s.log.Warn("知识库不存在", zap.String("id", id.String()))
			return fmt.Errorf("知识库不存在")
		}
		s.log.Error("删除知识库失败", zap.String("id", id.String()), zap.Error(err))
		return err
	}
	s.log.Info("删除知识库成功", zap.String("id", id.String()))
	return nil
}
