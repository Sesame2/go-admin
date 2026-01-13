package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"github.com/Sesame2/go-admin/internal/logger"
	"github.com/Sesame2/go-admin/internal/models"
	"github.com/Sesame2/go-admin/internal/models/dto"
	"github.com/Sesame2/go-admin/internal/mq/rabbitmq"
	"github.com/Sesame2/go-admin/internal/repository"
	"github.com/Sesame2/go-admin/internal/storage/s3"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const (
	// DocumentRoutingKey RabbitMQ 文档解析路由键
	DocumentRoutingKey = "document.parse"
	// DocumentBucket MinIO 文档存储桶
	DocumentBucket = "documents"
)

type DocumentService struct {
	repo        *repository.DocumentRepository
	kbRepo      *repository.KnowledgeBaseRepository
	minioClient *s3.MinioClient
	mqClient    *rabbitmq.RabbitMQClient
	log         *zap.Logger
}

func NewDocumentService(
	docRepo *repository.DocumentRepository,
	kbRepo *repository.KnowledgeBaseRepository,
	minioClient *s3.MinioClient,
	mqClient *rabbitmq.RabbitMQClient,
) *DocumentService {
	return &DocumentService{
		repo:        docRepo,
		kbRepo:      kbRepo,
		minioClient: minioClient,
		mqClient:    mqClient,
		log:         logger.NewModuleLogger("DocumentService"),
	}
}

// UploadDocument 上传文档到MinIO并创建记录
func (s *DocumentService) UploadDocument(ctx context.Context, kbID uuid.UUID, file *multipart.FileHeader, userID *uuid.UUID) (*models.Document, error) {
	s.log.Info("上传文档",
		zap.String("kb_id", kbID.String()),
		zap.String("filename", file.Filename),
		zap.Int64("size", file.Size))

	// 检查知识库是否存在
	kb, err := s.kbRepo.GetByID(ctx, kbID)
	if err != nil {
		s.log.Error("知识库不存在", zap.Error(err))
		return nil, fmt.Errorf("知识库不存在: %w", err)
	}

	// 获取文件扩展名
	ext := strings.ToLower(filepath.Ext(file.Filename))
	fileType := s.getFileType(ext)
	if fileType == "" {
		return nil, fmt.Errorf("不支持的文件类型: %s", ext)
	}

	// 打开文件
	src, err := file.Open()
	if err != nil {
		s.log.Error("打开上传文件失败", zap.Error(err))
		return nil, fmt.Errorf("打开文件失败: %w", err)
	}
	defer src.Close()

	// 生成MinIO存储路径
	docID := uuid.New()
	minioPath := fmt.Sprintf("%s/%s/%s%s", kbID.String(), docID.String(), docID.String(), ext)

	// 上传到MinIO
	if s.minioClient != nil {
		if err := s.minioClient.Upload(ctx, DocumentBucket, minioPath, src, file.Size); err != nil {
			s.log.Error("上传到MinIO失败", zap.Error(err))
			return nil, fmt.Errorf("上传文件失败: %w", err)
		}
		s.log.Info("文件上传到MinIO成功", zap.String("path", minioPath))
	}

	// 创建文档记录
	fileSize := file.Size
	doc := &models.Document{
		ID:              docID,
		KnowledgeBaseID: kbID,
		UserID:          userID,
		Filename:        file.Filename,
		FileType:        fileType,
		FileSize:        &fileSize,
		MinioPath:       minioPath,
		Status:          models.DocumentStatusPending,
	}

	if err := s.repo.Create(ctx, doc); err != nil {
		s.log.Error("创建文档记录失败", zap.Error(err))
		return nil, fmt.Errorf("创建文档记录失败: %w", err)
	}

	// 更新知识库文档数量
	if err := s.kbRepo.IncrementDocumentCount(ctx, kb.ID); err != nil {
		s.log.Warn("更新知识库文档数量失败", zap.Error(err))
	}

	return doc, nil
}

// ParseDocument 触发文档解析（发送消息到RabbitMQ）
// 注意：状态更新由 Python Worker 负责，确保状态反映实际处理进度
func (s *DocumentService) ParseDocument(ctx context.Context, docID uuid.UUID) error {
	s.log.Info("触发文档解析", zap.String("doc_id", docID.String()))

	// 获取文档信息
	doc, err := s.repo.GetByID(ctx, docID)
	if err != nil {
		s.log.Error("文档不存在", zap.Error(err))
		return fmt.Errorf("文档不存在: %w", err)
	}

	// 检查文档状态
	if !doc.CanProcess() {
		return fmt.Errorf("文档状态不允许解析: %s", doc.Status)
	}

	// 构建消息
	msg := dto.DocumentParseMessage{
		DocumentID:      doc.ID.String(),
		KnowledgeBaseID: doc.KnowledgeBaseID.String(),
		Filename:        doc.Filename,
		FileType:        doc.FileType,
		MinioPath:       doc.MinioPath,
	}

	msgBytes, err := json.Marshal(msg)
	if err != nil {
		s.log.Error("序列化消息失败", zap.Error(err))
		return fmt.Errorf("序列化消息失败: %w", err)
	}

	// 发送到RabbitMQ
	if s.mqClient != nil {
		if err := s.mqClient.Send(ctx, "", DocumentRoutingKey, msgBytes); err != nil {
			s.log.Error("发送消息到RabbitMQ失败", zap.Error(err))
			return fmt.Errorf("发送消息失败: %w", err)
		}
		s.log.Info("文档解析消息已发送", zap.String("doc_id", docID.String()))
	}

	return nil
}

// GetDocument 根据ID获取文档
func (s *DocumentService) GetDocument(ctx context.Context, id uuid.UUID) (*models.Document, error) {
	s.log.Debug("获取文档", zap.String("id", id.String()))
	return s.repo.GetByID(ctx, id)
}

// ListDocuments 获取知识库下的文档列表
func (s *DocumentService) ListDocuments(ctx context.Context, input *dto.DocumentListInput) (*dto.DocumentListResult, error) {
	kbID, err := uuid.Parse(input.KnowledgeBaseID)
	if err != nil {
		return nil, fmt.Errorf("无效的知识库ID: %w", err)
	}

	s.log.Debug("获取文档列表",
		zap.String("kb_id", kbID.String()),
		zap.Int("page", input.Page),
		zap.Int("page_size", input.PageSize))

	offset := (input.Page - 1) * input.PageSize
	var status *string
	if input.Status != "" {
		status = &input.Status
	}

	docs, err := s.repo.ListByKnowledgeBase(ctx, kbID, input.PageSize, offset, status)
	if err != nil {
		s.log.Error("获取文档列表失败", zap.Error(err))
		return nil, err
	}

	total, err := s.repo.CountByKnowledgeBase(ctx, kbID, status)
	if err != nil {
		s.log.Error("获取文档总数失败", zap.Error(err))
		return nil, err
	}

	return &dto.DocumentListResult{
		Documents: docs,
		Total:     total,
		Page:      input.Page,
		PageSize:  input.PageSize,
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

	if input.Status != nil {
		doc.Status = models.DocumentStatus(*input.Status)
	}
	if input.ErrorMessage != nil {
		doc.ErrorMessage = input.ErrorMessage
	}
	if input.ChunkCount != nil {
		doc.ChunkCount = *input.ChunkCount
	}
	if input.AtomQuestionCount != nil {
		doc.AtomQuestionCount = *input.AtomQuestionCount
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

	doc, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// 删除MinIO中的文件
	if s.minioClient != nil && doc.MinioPath != "" {
		// TODO: 实现MinIO删除
	}

	// 更新知识库文档数量
	if err := s.kbRepo.DecrementDocumentCount(ctx, doc.KnowledgeBaseID); err != nil {
		s.log.Warn("更新知识库文档数量失败", zap.Error(err))
	}

	return s.repo.Delete(ctx, id)
}

// GetDocumentDownloadURL 获取文档下载URL
func (s *DocumentService) GetDocumentDownloadURL(ctx context.Context, id uuid.UUID) (string, error) {
	doc, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return "", err
	}

	if s.minioClient == nil {
		return "", fmt.Errorf("存储服务未配置")
	}

	// 生成预签名URL，有效期1小时
	url, err := s.minioClient.SignURL(ctx, DocumentBucket, doc.MinioPath, time.Hour)
	if err != nil {
		s.log.Error("生成下载URL失败", zap.Error(err))
		return "", fmt.Errorf("生成下载URL失败: %w", err)
	}

	return url, nil
}

// getFileType 根据扩展名获取文件类型
func (s *DocumentService) getFileType(ext string) string {
	switch ext {
	case ".pdf":
		return "pdf"
	case ".docx", ".doc":
		return "docx"
	case ".pptx", ".ppt":
		return "pptx"
	case ".md", ".markdown":
		return "markdown"
	case ".txt":
		return "txt"
	default:
		return ""
	}
}

// StreamReader 包装io.Reader以实现流式读取
type StreamReader struct {
	reader io.ReadCloser
}

func (s *StreamReader) Read(p []byte) (n int, err error) {
	return s.reader.Read(p)
}

func (s *StreamReader) Close() error {
	return s.reader.Close()
}
