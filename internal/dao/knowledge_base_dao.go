package dao

import (
	"context"

	"github.com/Sesame2/go-admin/internal/database/interfaces"
	"github.com/Sesame2/go-admin/internal/models/dto"
	"github.com/Sesame2/go-admin/internal/models/ent"
	"github.com/Sesame2/go-admin/internal/models/ent/knowledgebase"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type KnowledgeBaseDAO struct {
	db     interfaces.Database
	logger *zap.Logger
}

func NewKnowledgeBaseDAO(db interfaces.Database, logger *zap.Logger) *KnowledgeBaseDAO {
	daoLogger := logger.With(zap.String("component", "KnowledgeBaseDAO"))
	return &KnowledgeBaseDAO{
		db:     db,
		logger: daoLogger,
	}
}

func (d *KnowledgeBaseDAO) Create(ctx context.Context, input *dto.CreateKnowledgeBaseInput) (*ent.KnowledgeBase, error) {
	d.logger.Info("创建知识库", zap.String("name", input.KBName))

	builder := d.db.Client().KnowledgeBase.Create().SetName(input.KBName)

	if input.DatasetID != nil {
		d.logger.Debug("设置数据集ID", zap.String("dataset_id", *input.DatasetID))
		builder.SetDatasetID(*input.DatasetID)
	}

	// 保存实体
	kb, err := builder.Save(ctx)
	if err != nil {
		d.logger.Error("创建知识库失败",
			zap.String("name", input.KBName),
			zap.Error(err))
		return nil, err
	}

	d.logger.Info("知识库创建成功", zap.String("id", kb.ID.String()))
	return kb, nil
}

func (d *KnowledgeBaseDAO) GetByID(ctx context.Context, id uuid.UUID) (*ent.KnowledgeBase, error) {
	d.logger.Debug("获取知识库详情", zap.String("id", id.String()))

	kb, err := d.db.Client().KnowledgeBase.Get(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			d.logger.Debug("知识库不存在", zap.String("id", id.String()))
		} else {
			d.logger.Error("获取知识库详情失败", zap.String("id", id.String()), zap.Error(err))
		}
		return nil, err
	}

	d.logger.Debug("获取知识库详情成功", zap.String("id", id.String()), zap.String("name", kb.Name))
	return kb, nil
}

func (d *KnowledgeBaseDAO) GetAll(ctx context.Context) ([]*ent.KnowledgeBase, error) {
	d.logger.Debug("获取所有知识库")

	kbs, err := d.db.Client().KnowledgeBase.Query().All(ctx)
	if err != nil {
		d.logger.Error("获取所有知识库失败", zap.Error(err))
		return nil, err
	}

	d.logger.Debug("获取所有知识库成功", zap.Int("count", len(kbs)))
	return kbs, nil
}

func (d *KnowledgeBaseDAO) Exist(ctx context.Context, id uuid.UUID) (bool, error) {
	d.logger.Debug("检查知识库是否存在", zap.String("id", id.String()))

	exists, err := d.db.Client().KnowledgeBase.Query().Where(knowledgebase.ID(id)).Exist(ctx)
	if err != nil {
		d.logger.Error("检查知识库存在失败", zap.String("id", id.String()), zap.Error(err))
		return false, err
	}

	if exists {
		d.logger.Debug("知识库存在", zap.String("id", id.String()))
	} else {
		d.logger.Debug("知识库不存在", zap.String("id", id.String()))
	}

	return exists, nil
}

func (d *KnowledgeBaseDAO) Count(ctx context.Context) (int, error) {
	d.logger.Debug("统计知识库总数")

	count, err := d.db.Client().KnowledgeBase.Query().Count(ctx)
	if err != nil {
		d.logger.Error("获取知识库总数失败", zap.Error(err))
		return 0, err
	}

	d.logger.Debug("知识库总数统计完成", zap.Int("count", count))
	return count, nil
}

func (d *KnowledgeBaseDAO) Update(ctx context.Context, id uuid.UUID, input *dto.UpdateKnowledgeBaseInput) (*ent.KnowledgeBase, error) {
	d.logger.Info("更新知识库", zap.String("id", id.String()))

	// 检查知识库是否存在
	exists, err := d.Exist(ctx, id)
	if err != nil {
		return nil, err
	}

	if !exists {
		d.logger.Warn("要更新的知识库不存在", zap.String("id", id.String()))
		return nil, &ent.NotFoundError{}
	}

	// 构建更新
	update := d.db.Client().KnowledgeBase.UpdateOneID(id)

	// 有条件地设置字段
	if input.KBName != nil {
		d.logger.Debug("更新知识库名称", zap.String("id", id.String()), zap.String("name", *input.KBName))
		update.SetName(*input.KBName)
	}

	if input.DatasetID != nil {
		d.logger.Debug("更新数据集ID", zap.String("id", id.String()), zap.String("dataset_id", *input.DatasetID))
		update.SetDatasetID(*input.DatasetID)
	}

	// 保存更新
	kb, err := update.Save(ctx)
	if err != nil {
		d.logger.Error("更新知识库失败", zap.String("id", id.String()), zap.Error(err))
		return nil, err
	}

	d.logger.Info("知识库更新成功", zap.String("id", id.String()))
	return kb, nil
}

func (d *KnowledgeBaseDAO) Delete(ctx context.Context, id uuid.UUID) error {
	d.logger.Info("删除知识库", zap.String("id", id.String()))

	// 检查知识库是否存在
	exists, err := d.Exist(ctx, id)
	if err != nil {
		return err
	}

	if !exists {
		d.logger.Warn("要删除的知识库不存在", zap.String("id", id.String()))
		return nil // 不存在则视为删除成功
	}

	// 执行删除
	err = d.db.Client().KnowledgeBase.DeleteOneID(id).Exec(ctx)
	if err != nil {
		d.logger.Error("删除知识库失败", zap.String("id", id.String()), zap.Error(err))
		return err
	}

	d.logger.Info("知识库删除成功", zap.String("id", id.String()))
	return nil
}

func (d *KnowledgeBaseDAO) GetPaginated(ctx context.Context, offset, limit int) ([]*ent.KnowledgeBase, error) {
	d.logger.Debug("分页获取知识库", zap.Int("offset", offset), zap.Int("limit", limit))

	kbs, err := d.db.Client().KnowledgeBase.Query().
		Order(ent.Desc(knowledgebase.FieldCreatedAt)).
		Offset(offset).
		Limit(limit).
		All(ctx)

	if err != nil {
		d.logger.Error("分页获取知识库失败",
			zap.Int("offset", offset),
			zap.Int("limit", limit),
			zap.Error(err))
		return nil, err
	}

	d.logger.Debug("分页获取知识库成功",
		zap.Int("offset", offset),
		zap.Int("limit", limit),
		zap.Int("count", len(kbs)))
	return kbs, nil
}

func (d *KnowledgeBaseDAO) GetByName(ctx context.Context, name string) (*ent.KnowledgeBase, error) {
	d.logger.Debug("通过名称获取知识库", zap.String("name", name))

	kb, err := d.db.Client().KnowledgeBase.Query().
		Where(knowledgebase.Name(name)).
		Only(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			d.logger.Debug("找不到指定名称的知识库", zap.String("name", name))
		} else if ent.IsNotSingular(err) {
			d.logger.Warn("找到多个同名知识库", zap.String("name", name))
		} else {
			d.logger.Error("通过名称获取知识库失败", zap.String("name", name), zap.Error(err))
		}
		return nil, err
	}

	d.logger.Debug("通过名称获取知识库成功", zap.String("name", name), zap.String("id", kb.ID.String()))
	return kb, nil
}
