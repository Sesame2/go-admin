package services

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Sesame2/go-admin/internal/logger"
	"go.uber.org/zap"
)

// RetrievalService 检索服务 - 转发SSE请求到Python服务
type RetrievalService struct {
	pythonBaseURL string
	httpClient    *http.Client
	log           *zap.Logger
}

// RetrievalRequest 检索请求
type RetrievalRequest struct {
	Question        string `json:"question"`
	KnowledgeBaseID string `json:"knowledge_base_id"`
	MaxIterations   int    `json:"max_iterations"`
	StreamAnswer    bool   `json:"stream_answer"`
}

// NewRetrievalService 创建检索服务
func NewRetrievalService(pythonBaseURL string) *RetrievalService {
	return &RetrievalService{
		pythonBaseURL: pythonBaseURL,
		httpClient: &http.Client{
			Timeout: 5 * time.Minute, // SSE 长连接，设置较长超时
		},
		log: logger.NewModuleLogger("RetrievalService"),
	}
}

// StreamRetrieve 流式检索 - 转发到Python服务并返回SSE流
func (s *RetrievalService) StreamRetrieve(ctx context.Context, req *RetrievalRequest) (io.ReadCloser, error) {
	s.log.Info("开始流式检索",
		zap.String("question", req.Question[:min(50, len(req.Question))]),
		zap.String("kb_id", req.KnowledgeBaseID))

	// 构建请求URL - Python服务路径: /v1/qa/retrieve/stream (POST + JSON Body)
	url := fmt.Sprintf("%s/v1/qa/retrieve/stream", s.pythonBaseURL)

	// 构建 JSON Body
	body := fmt.Sprintf(`{"question":"%s","knowledge_base_id":"%s","max_iterations":%d,"stream_answer":%t}`,
		req.Question,
		req.KnowledgeBaseID,
		req.MaxIterations,
		req.StreamAnswer,
	)

	// 创建 POST 请求
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(body))
	if err != nil {
		s.log.Error("创建请求失败", zap.Error(err))
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "text/event-stream")
	httpReq.Header.Set("Cache-Control", "no-cache")
	httpReq.Header.Set("Connection", "keep-alive")

	// 发送请求
	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		s.log.Error("请求Python服务失败", zap.Error(err))
		return nil, fmt.Errorf("请求Python服务失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		s.log.Error("Python服务返回错误", zap.Int("status", resp.StatusCode))
		return nil, fmt.Errorf("Python服务返回错误: %d", resp.StatusCode)
	}

	s.log.Info("SSE连接建立成功")
	return resp.Body, nil
}

// SSEEvent SSE事件结构
type SSEEvent struct {
	Event string `json:"event"`
	Data  string `json:"data"`
}

// ParseSSEStream 解析SSE流
func ParseSSEStream(reader io.Reader) <-chan SSEEvent {
	ch := make(chan SSEEvent)

	go func() {
		defer close(ch)
		scanner := bufio.NewScanner(reader)
		var event SSEEvent

		for scanner.Scan() {
			line := scanner.Text()

			if line == "" {
				// 空行表示事件结束
				if event.Data != "" {
					ch <- event
					event = SSEEvent{}
				}
				continue
			}

			if len(line) > 6 && line[:6] == "event:" {
				event.Event = line[7:]
			} else if len(line) > 5 && line[:5] == "data:" {
				event.Data = line[6:]
			}
		}
	}()

	return ch
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
