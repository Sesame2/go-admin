package repository

import (
	"context"
	"errors"

	"github.com/Sesame2/go-admin/internal/database/interfaces"
	"github.com/Sesame2/go-admin/internal/logger"
	"github.com/Sesame2/go-admin/internal/models"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// KnowledgeBaseRepository 知识库仓储
type KnowledgeBaseRepository struct {
	db  interfaces.Database
	log *zap.Logger
}

// NewKnowledgeBaseRepository 创建知识库仓储实例
func NewKnowledgeBaseRepository(db interfaces.Database) *KnowledgeBaseRepository {
	return &KnowledgeBaseRepository{
		db:  db,
		log: logger.NewModuleLogger("KnowledgeBaseRepository"),
	}
}

// Create 创建知识库
func (r *KnowledgeBaseRepository) Create(ctx context.Context, kb *models.KnowledgeBase) error {
	err := r.db.DB().WithContext(ctx).Create(kb).Error
	if err != nil {
		r.log.Error("创建知识库失败", zap.String("name", kb.Name), zap.Error(err))
		return err
	}

	r.log.Info("创建知识库成功", zap.String("id", kb.ID.String()), zap.String("name", kb.Name))
	return nil
}

// GetByID 根据ID获取知识库
func (r *KnowledgeBaseRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.KnowledgeBase, error) {
	var kb models.KnowledgeBase
	err := r.db.DB().WithContext(ctx).First(&kb, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			r.log.Debug("知识库不存在", zap.String("id", id.String()))
			return nil, ErrNotFound
		}
		r.log.Error("获取知识库失败", zap.String("id", id.String()), zap.Error(err))
		return nil, err
	}
	return &kb, nil
}

// GetAll 分页获取知识库（limit 和 offset）
func (r *KnowledgeBaseRepository) GetAll(ctx context.Context, limit, offset int) ([]*models.KnowledgeBase, error) {
	var kbs []*models.KnowledgeBase
	err := r.db.DB().WithContext(ctx).Limit(limit).Offset(offset).Order("created_at DESC").Find(&kbs).Error
	if err != nil {
		r.log.Error("获取知识库列表失败", zap.Error(err))
		return nil, err
	}
	r.log.Info("获取知识库列表成功", zap.Int("count", len(kbs)))
	return kbs, nil
}

// Exist 检查知识库是否存在
func (r *KnowledgeBaseRepository) Exist(ctx context.Context, id uuid.UUID) (bool, error) {
	var count int64
	err := r.db.DB().WithContext(ctx).Model(&models.KnowledgeBase{}).Where("id = ?", id).Count(&count).Error
	if err != nil {
		r.log.Error("检查知识库存在失败", zap.String("id", id.String()), zap.Error(err))
		return false, err
	}
	return count > 0, nil
}

// Count 统计知识库总数
func (r *KnowledgeBaseRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.DB().WithContext(ctx).Model(&models.KnowledgeBase{}).Count(&count).Error
	if err != nil {
		r.log.Error("统计知识库总数失败", zap.Error(err))
		return 0, err
	}
	r.log.Info("知识库总数统计完成", zap.Int64("count", count))
	return count, nil
}

// Update 更新知识库
func (r *KnowledgeBaseRepository) Update(ctx context.Context, kb *models.KnowledgeBase) error {
	err := r.db.DB().WithContext(ctx).Save(kb).Error
	if err != nil {
		r.log.Error("更新知识库失败", zap.String("id", kb.ID.String()), zap.Error(err))
		return err
	}

	r.log.Info("知识库更新成功", zap.String("id", kb.ID.String()))
	return nil
}

// Delete 删除知识库
func (r *KnowledgeBaseRepository) Delete(ctx context.Context, id uuid.UUID) error {
	err := r.db.DB().WithContext(ctx).Delete(&models.KnowledgeBase{}, "id = ?", id).Error
	if err != nil {
		r.log.Error("删除知识库失败", zap.String("id", id.String()), zap.Error(err))
		return err
	}

	r.log.Info("知识库删除成功", zap.String("id", id.String()))
	return nil
}

// IncrementDocumentCount 增加文档计数
func (r *KnowledgeBaseRepository) IncrementDocumentCount(ctx context.Context, id uuid.UUID) error {
	err := r.db.DB().WithContext(ctx).
		Model(&models.KnowledgeBase{}).
		Where("id = ?", id).
		UpdateColumn("document_count", gorm.Expr("document_count + ?", 1)).
		Error
	if err != nil {
		r.log.Error("增加文档计数失败", zap.String("id", id.String()), zap.Error(err))
		return err
	}
	return nil
}

// DecrementDocumentCount 减少文档计数
func (r *KnowledgeBaseRepository) DecrementDocumentCount(ctx context.Context, id uuid.UUID) error {
	err := r.db.DB().WithContext(ctx).
		Model(&models.KnowledgeBase{}).
		Where("id = ? AND document_count > 0", id).
		UpdateColumn("document_count", gorm.Expr("document_count - ?", 1)).
		Error
	if err != nil {
		r.log.Error("减少文档计数失败", zap.String("id", id.String()), zap.Error(err))
		return err
	}
	return nil
}
