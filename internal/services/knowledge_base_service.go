package services

import (
	"context"
	"fmt"

	"github.com/Sesame2/go-admin/internal/dao"
	"github.com/Sesame2/go-admin/internal/models/dto"
	"github.com/Sesame2/go-admin/internal/models/ent"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type KnowledgeBaseService struct {
	dao    *dao.KnowledgeBaseDAO
	logger *zap.Logger
}

func NewKnowledgeBaseService(dao *dao.KnowledgeBaseDAO, logger *zap.Logger) *KnowledgeBaseService {
	logger = logger.With(zap.String("component", "KnowledgeBaseService"))
	return &KnowledgeBaseService{
		dao:    dao,
		logger: logger,
	}
}

func (s *KnowledgeBaseService) CreateKnowledgeBase(ctx context.Context, input *dto.CreateKnowledgeBaseInput) (*ent.KnowledgeBase, error) {
	kb, err := s.dao.Create(ctx, input)
	if err != nil {
		s.logger.Error("创建知识库失败", zap.Error(err))
		return nil, err
	}
	s.logger.Info(fmt.Sprintln("创建知识库成功 kb_name:", input.KBName))
	return kb, nil
}

func (s *KnowledgeBaseService) GetAllKnowledgBase(ctx context.Context, page, pageSize int) (dto.KnowledgeBaseListResult, error) {
	count, err := s.dao.Count(ctx)
	if err != nil {
		s.logger.Error("获取知识库数量失败：", zap.Error(err))
		return dto.KnowledgeBaseListResult{}, err
	}
	s.logger.Info(fmt.Sprintln("知识库总数为：", count))
	kbs, err := s.dao.GetPaginated(ctx, pageSize*(page-1), pageSize)
	if err != nil {
		s.logger.Error("查询知识库失败：", zap.Error(err))
		return dto.KnowledgeBaseListResult{}, err
	}
	s.logger.Info(fmt.Sprintln("分页查询知识库成功:", kbs))
	return dto.KnowledgeBaseListResult{
		TotalCount:  count,
		TotalPages:  (count + pageSize - 1) / pageSize,
		CurrentPage: page,
		PageSize:    pageSize,
		Result:      kbs,
	}, nil
}

func (s *KnowledgeBaseService) GetKnowledgeBaseByID(ctx context.Context, id uuid.UUID) (*ent.KnowledgeBase, error) {
	kb, err := s.dao.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("查询知识库失败：", zap.String("id", id.String()), zap.Error(err))
		return nil, err
	}
	s.logger.Info(fmt.Sprintln("查询知识库成功", kb))
	return kb, nil
}
