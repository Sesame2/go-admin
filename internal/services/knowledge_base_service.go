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

func (s *KnowledgeBaseService) CreateKnowledgeBase(ctx context.Context, userID uuid.UUID, input *dto.CreateKnowledgeBaseInput) (*models.KnowledgeBase, error) {
	kb := &models.KnowledgeBase{
		Name:        input.Name,
		Description: input.Description,
		UserID:      userID,
	}

	err := s.repo.Create(ctx, kb)
	if err != nil {
		s.log.Error("创建知识库失败", zap.Error(err))
		return nil, err
	}
	s.log.Info("创建知识库成功", zap.String("kb_name", input.Name), zap.String("user_id", userID.String()))
	return kb, nil
}

func (s *KnowledgeBaseService) GetAllKnowledgeBase(ctx context.Context, userID uuid.UUID, page, pageSize int) (dto.KnowledgeBaseListResult, error) {
	count, err := s.repo.CountByUserID(ctx, userID)
	if err != nil {
		s.log.Error("获取用户知识库数量失败", zap.String("user_id", userID.String()), zap.Error(err))
		return dto.KnowledgeBaseListResult{}, err
	}
	s.log.Info("用户知识库总数", zap.String("user_id", userID.String()), zap.Int64("count", count))

	offset := (page - 1) * pageSize
	kbs, err := s.repo.GetByUserID(ctx, userID, pageSize, offset)
	if err != nil {
		s.log.Error("查询用户知识库失败", zap.String("user_id", userID.String()), zap.Error(err))
		return dto.KnowledgeBaseListResult{}, err
	}
	s.log.Info("分页查询用户知识库成功", zap.String("user_id", userID.String()), zap.Int("count", len(kbs)))

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

func (s *KnowledgeBaseService) UpdateKnowledgeBase(ctx context.Context, id uuid.UUID, userID uuid.UUID, input *dto.UpdateKnowledgeBaseInput) (*models.KnowledgeBase, error) {
	kb, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			s.log.Warn("知识库不存在", zap.String("id", id.String()))
			return nil, fmt.Errorf("知识库不存在")
		}
		s.log.Error("查询知识库失败", zap.String("id", id.String()), zap.Error(err))
		return nil, err
	}

	// 验证权限：只有知识库所有者才能更新
	if kb.UserID != userID {
		s.log.Warn("无权限更新知识库", zap.String("kb_id", id.String()), zap.String("user_id", userID.String()), zap.String("owner_id", kb.UserID.String()))
		return nil, fmt.Errorf("无权限操作此知识库")
	}

	if input.Name != nil {
		kb.Name = *input.Name
	}
	if input.Description != nil {
		kb.Description = *input.Description
	}

	err = s.repo.Update(ctx, kb)
	if err != nil {
		s.log.Error("更新知识库失败", zap.String("id", id.String()), zap.Error(err))
		return nil, err
	}
	s.log.Info("更新知识库成功", zap.String("id", id.String()))
	return kb, nil
}

func (s *KnowledgeBaseService) DeleteKnowledgeBase(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	kb, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			s.log.Warn("知识库不存在", zap.String("id", id.String()))
			return fmt.Errorf("知识库不存在")
		}
		s.log.Error("查询知识库失败", zap.String("id", id.String()), zap.Error(err))
		return err
	}

	// 验证权限：只有知识库所有者才能删除
	if kb.UserID != userID {
		s.log.Warn("无权限删除知识库", zap.String("kb_id", id.String()), zap.String("user_id", userID.String()), zap.String("owner_id", kb.UserID.String()))
		return fmt.Errorf("无权限操作此知识库")
	}

	err = s.repo.Delete(ctx, id)
	if err != nil {
		s.log.Error("删除知识库失败", zap.String("id", id.String()), zap.Error(err))
		return err
	}
	s.log.Info("删除知识库成功", zap.String("id", id.String()))
	return nil
}
