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

// DocumentRepository 文档仓储
type DocumentRepository struct {
	db  interfaces.Database
	log *zap.Logger
}

// NewDocumentRepository 创建文档仓储实例
func NewDocumentRepository(db interfaces.Database) *DocumentRepository {
	return &DocumentRepository{
		db:  db,
		log: logger.NewModuleLogger("DocumentRepository"),
	}
}

// Create 创建文档
func (r *DocumentRepository) Create(ctx context.Context, doc *models.Document) error {
	err := r.db.DB().WithContext(ctx).Create(doc).Error
	if err != nil {
		r.log.Error("创建文档失败", zap.String("name", doc.Name), zap.Error(err))
		return err
	}

	r.log.Info("创建文档成功", zap.String("id", doc.ID.String()), zap.String("name", doc.Name))
	return nil
}

// GetByID 根据ID获取文档
func (r *DocumentRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Document, error) {
	var doc models.Document
	err := r.db.DB().WithContext(ctx).First(&doc, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			r.log.Debug("文档不存在", zap.String("id", id.String()))
			return nil, ErrNotFound
		}
		r.log.Error("获取文档失败", zap.String("id", id.String()), zap.Error(err))
		return nil, err
	}
	return &doc, nil
}

// GetAll 获取所有文档
func (r *DocumentRepository) GetAll(ctx context.Context) ([]*models.Document, error) {
	var docs []*models.Document
	err := r.db.DB().WithContext(ctx).Find(&docs).Error
	if err != nil {
		r.log.Error("获取所有文档失败", zap.Error(err))
		return nil, err
	}
	r.log.Info("获取所有文档成功", zap.Int("count", len(docs)))
	return docs, nil
}

// GetByUserID 根据用户ID获取文档
func (r *DocumentRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*models.Document, error) {
	var docs []*models.Document
	err := r.db.DB().WithContext(ctx).Where("user_id = ?", userID).Find(&docs).Error
	if err != nil {
		r.log.Error("根据用户ID获取文档失败", zap.String("userID", userID.String()), zap.Error(err))
		return nil, err
	}
	r.log.Info("根据用户ID获取文档成功", zap.String("userID", userID.String()), zap.Int("count", len(docs)))
	return docs, nil
}

// ListByKnowledgeBase 根据知识库ID分页获取文档
func (r *DocumentRepository) ListByKnowledgeBase(ctx context.Context, kbID uuid.UUID, limit, offset int) ([]*models.Document, error) {
	var docs []*models.Document
	err := r.db.DB().WithContext(ctx).
		Where("knowledge_base_id = ?", kbID).
		Limit(limit).
		Offset(offset).
		Order("created_at DESC").
		Find(&docs).Error
	if err != nil {
		r.log.Error("根据知识库获取文档失败", zap.String("kb_id", kbID.String()), zap.Error(err))
		return nil, err
	}
	return docs, nil
}

// CountByKnowledgeBase 根据知识库ID统计文档数量
func (r *DocumentRepository) CountByKnowledgeBase(ctx context.Context, kbID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.DB().WithContext(ctx).
		Model(&models.Document{}).
		Where("knowledge_base_id = ?", kbID).
		Count(&count).Error
	if err != nil {
		r.log.Error("统计知识库文档数量失败", zap.String("kb_id", kbID.String()), zap.Error(err))
		return 0, err
	}
	return count, nil
}

// Update 更新文档
func (r *DocumentRepository) Update(ctx context.Context, doc *models.Document) error {
	err := r.db.DB().WithContext(ctx).Save(doc).Error
	if err != nil {
		r.log.Error("更新文档失败", zap.String("id", doc.ID.String()), zap.Error(err))
		return err
	}

	r.log.Info("文档更新成功", zap.String("id", doc.ID.String()))
	return nil
}

// Delete 删除文档
func (r *DocumentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	err := r.db.DB().WithContext(ctx).Delete(&models.Document{}, "id = ?", id).Error
	if err != nil {
		r.log.Error("删除文档失败", zap.String("id", id.String()), zap.Error(err))
		return err
	}
	r.log.Info("删除文档成功", zap.String("id", id.String()))
	return nil
}

// Exist 检查文档是否存在
func (r *DocumentRepository) Exist(ctx context.Context, id uuid.UUID) (bool, error) {
	var count int64
	err := r.db.DB().WithContext(ctx).Model(&models.Document{}).Where("id = ?", id).Count(&count).Error
	if err != nil {
		r.log.Error("检查文档存在失败", zap.String("id", id.String()), zap.Error(err))
		return false, err
	}
	return count > 0, nil
}
