package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Sesame2/go-admin/internal/mq"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type DocumentTask struct {
	ID              string                 `json:"id"`
	KnowledgeBaseID string                 `json:"knowledgebase_id"`
	FilePath        string                 `json:"file_path"`
	FileType        string                 `json:"file_type"`
	FileName        string                 `json:"file_name"`
	Status          TaskStatus             `json:"status"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// ProcessResult 表示处理结果
type ProcessResult struct {
	TaskID      string                 `json:"task_id"`
	KnowledgeID string                 `json:"knowledge_id"`
	Status      TaskStatus             `json:"status"`
	CompletedAt time.Time              `json:"completed_at"`
	Error       string                 `json:"error,omitempty"`
	Stats       map[string]interface{} `json:"stats,omitempty"`
}

type DocumentTaskHandler struct {
	producer mq.Producer
	logger   *zap.Logger
	exchange string
}

func NewDocumentTaskHandler(producer mq.Producer, logger *zap.Logger, exchange string) *DocumentTaskHandler {
	if exchange == "" {
		exchange = "go_admin_exchange"
	}

	return &DocumentTaskHandler{
		producer: producer,
		logger:   logger.With(zap.String("component", "DocumentTaskHandler")),
		exchange: exchange,
	}
}

func (h *DocumentTaskHandler) SendProcessTask(ctx context.Context, input *DocumentTask) error {
	if input.ID == "" {
		input.ID = uuid.New().String()
	}

	// 设置任务状态和时间
	input.Status = StatusPending
	now := time.Now()
	if input.CreatedAt.IsZero() {
		input.CreatedAt = now
	}
	input.UpdatedAt = now

	h.logger.Info("发送文档处理任务",
		zap.String("task_id", input.ID),
		zap.String("knowledge_id", input.KnowledgeBaseID),
		zap.String("file_name", input.FileName))

	// 序列化任务
	data, err := json.Marshal(input)
	if err != nil {
		return fmt.Errorf("序列化任务失败")
	}

	// 发送消息
	routingKey := string(DocumentProcess)
	err = h.producer.Send(ctx, h.exchange, routingKey, data)
	if err != nil {
		return fmt.Errorf("发送任务失败：%w", err)
	}
	h.logger.Info("文档处理任务已发送", zap.String("task_id", input.ID))
	return nil
}

func (h *DocumentTaskHandler) HandleResult(result []byte) error {
	var processResult ProcessResult
	if err := json.Unmarshal(result, &processResult); err != nil {
		h.logger.Error("解析处理结果失败", zap.Error(err))
		return fmt.Errorf("解析处理结果失败: %w", err)
	}
	h.logger.Info("收到处理结果",
		zap.String("task_id", processResult.TaskID),
		zap.String("status", string(processResult.Status)))

	return nil
}
