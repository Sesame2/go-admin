package controller

import (
	"io"
	"net/http"

	"github.com/Sesame2/go-admin/internal/logger"
	"github.com/Sesame2/go-admin/internal/models/dto"
	"github.com/Sesame2/go-admin/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type DocumentController struct {
	service          *services.DocumentService
	retrievalService *services.RetrievalService
	log              *zap.Logger
}

func NewDocumentController(service *services.DocumentService, retrievalService *services.RetrievalService) *DocumentController {
	return &DocumentController{
		service:          service,
		retrievalService: retrievalService,
		log:              logger.NewModuleLogger("DocumentController"),
	}
}

// UploadDocument godoc
// @Summary      上传文档
// @Description  上传文档到指定知识库
// @Tags         文档模块
// @Accept       multipart/form-data
// @Produce      json
// @Security     Bearer
// @Param        knowledge_base_id  formData  string  true  "知识库ID"
// @Param        file               formData  file    true  "要上传的文件"
// @Success      200  {object}  models.Document  "上传成功的文档信息"
// @Failure      400  {object}  object{error=string}  "请求参数错误"
// @Failure      500  {object}  object{error=string}  "服务器内部错误"
// @Router       /documents/upload [post]
func (c *DocumentController) UploadDocument(ctx *gin.Context) {
	kbIDStr := ctx.PostForm("knowledge_base_id")
	if kbIDStr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "知识库ID不能为空"})
		return
	}
	kbID, err := uuid.Parse(kbIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的知识库ID格式"})
		return
	}

	file, err := ctx.FormFile("file")
	if err != nil {
		c.log.Error("获取上传文件失败", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "获取上传文件失败"})
		return
	}

	// 获取用户ID（从JWT中）
	var userID *uuid.UUID
	if userIDStr, exists := ctx.Get("userID"); exists {
		if id, err := uuid.Parse(userIDStr.(string)); err == nil {
			userID = &id
		}
	}

	doc, err := c.service.UploadDocument(ctx, kbID, file, userID)
	if err != nil {
		c.log.Error("上传文档失败", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, doc)
}

// ParseDocument godoc
// @Summary      解析文档
// @Description  触发文档解析任务（异步处理）
// @Tags         文档模块
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        input  body      dto.ParseDocumentInput  true  "解析请求"
// @Success      200  {object}  object{message=string}  "解析任务已提交"
// @Failure      400  {object}  object{error=string}  "请求参数错误"
// @Failure      500  {object}  object{error=string}  "服务器内部错误"
// @Router       /documents/parse [post]
func (c *DocumentController) ParseDocument(ctx *gin.Context) {
	var input dto.ParseDocumentInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的输入数据", "details": err.Error()})
		return
	}

	docID, err := uuid.Parse(input.DocumentID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的文档ID格式"})
		return
	}

	if err := c.service.ParseDocument(ctx, docID); err != nil {
		c.log.Error("触发文档解析失败", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "解析任务已提交"})
}

// GetDocument godoc
// @Summary      获取文档详情
// @Description  根据ID获取文档的详细信息
// @Tags         文档模块
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        id   path      string  true  "文档ID (UUID格式)"
// @Success      200  {object}  models.Document  "文档详情"
// @Failure      400  {object}  object{error=string}  "请求参数错误"
// @Failure      404  {object}  object{error=string}  "文档不存在"
// @Failure      500  {object}  object{error=string}  "服务器内部错误"
// @Router       /documents/{id} [get]
func (c *DocumentController) GetDocument(ctx *gin.Context) {
	idStr := ctx.Param("id")
	if idStr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "文档ID不能为空"})
		return
	}
	id, err := uuid.Parse(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的文档ID格式"})
		return
	}

	doc, err := c.service.GetDocument(ctx, id)
	if err != nil {
		c.log.Error("获取文档失败", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, doc)
}

// ListDocuments godoc
// @Summary      获取知识库下所有文档
// @Description  分页获取指定知识库下的所有文档
// @Tags         文档模块
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        knowledge_base_id  query     string  true   "知识库ID"
// @Param        page               query     int     false  "页码，默认为1"
// @Param        page_size          query     int     false  "每页数量，默认为10"
// @Param        status             query     string  false  "文档状态筛选"
// @Success      200  {object}  dto.DocumentListResult  "文档列表"
// @Failure      400  {object}  object{error=string}  "请求参数错误"
// @Failure      500  {object}  object{error=string}  "服务器内部错误"
// @Router       /documents [get]
func (c *DocumentController) ListDocuments(ctx *gin.Context) {
	var input dto.DocumentListInput
	if err := ctx.ShouldBindQuery(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的查询参数", "details": err.Error()})
		return
	}

	if input.Page < 1 {
		input.Page = 1
	}
	if input.PageSize < 1 {
		input.PageSize = 10
	}

	result, err := c.service.ListDocuments(ctx, &input)
	if err != nil {
		c.log.Error("获取文档列表失败", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, result)
}

// UpdateDocument godoc
// @Summary      更新文档
// @Description  更新文档信息
// @Tags         文档模块
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        id    path      string                  true  "文档ID (UUID格式)"
// @Param        input body      dto.UpdateDocumentInput true  "更新输入"
// @Success      200  {object}  models.Document  "更新后的文档"
// @Failure      400  {object}  object{error=string}  "请求参数错误"
// @Failure      404  {object}  object{error=string}  "文档不存在"
// @Failure      500  {object}  object{error=string}  "服务器内部错误"
// @Router       /documents/{id} [put]
func (c *DocumentController) UpdateDocument(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的文档ID格式"})
		return
	}

	var input dto.UpdateDocumentInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的输入数据", "details": err.Error()})
		return
	}

	doc, err := c.service.UpdateDocument(ctx, id, &input)
	if err != nil {
		c.log.Error("更新文档失败", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, doc)
}

// DeleteDocument godoc
// @Summary      删除文档
// @Description  删除指定ID的文档
// @Tags         文档模块
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        id   path      string  true  "文档ID (UUID格式)"
// @Success      200  {object}  object{message=string}  "删除成功"
// @Failure      400  {object}  object{error=string}  "请求参数错误"
// @Failure      404  {object}  object{error=string}  "文档不存在"
// @Failure      500  {object}  object{error=string}  "服务器内部错误"
// @Router       /documents/{id} [delete]
func (c *DocumentController) DeleteDocument(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的文档ID格式"})
		return
	}

	if err := c.service.DeleteDocument(ctx, id); err != nil {
		c.log.Error("删除文档失败", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

// GetDocumentDownloadURL godoc
// @Summary      获取文档下载URL
// @Description  获取文档的预签名下载URL
// @Tags         文档模块
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        id   path      string  true  "文档ID (UUID格式)"
// @Success      200  {object}  object{url=string}  "下载URL"
// @Failure      400  {object}  object{error=string}  "请求参数错误"
// @Failure      500  {object}  object{error=string}  "服务器内部错误"
// @Router       /documents/{id}/download [get]
func (c *DocumentController) GetDocumentDownloadURL(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的文档ID格式"})
		return
	}

	url, err := c.service.GetDocumentDownloadURL(ctx, id)
	if err != nil {
		c.log.Error("获取下载URL失败", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"url": url})
}

// StreamRetrieve godoc
// @Summary      流式检索问答
// @Description  通过SSE流式返回检索和问答结果
// @Tags         检索模块
// @Accept       json
// @Produce      text/event-stream
// @Security     Bearer
// @Param        question           query  string  true   "用户问题"
// @Param        knowledge_base_id  query  string  true   "知识库ID"
// @Param        max_iterations     query  int     false  "最大迭代次数，默认5"
// @Param        stream_answer      query  bool    false  "是否流式输出答案，默认true"
// @Success      200  {string}  string  "SSE事件流"
// @Failure      400  {object}  object{error=string}  "请求参数错误"
// @Failure      500  {object}  object{error=string}  "服务器内部错误"
// @Router       /retrieval/stream [get]
func (c *DocumentController) StreamRetrieve(ctx *gin.Context) {
	question := ctx.Query("question")
	kbID := ctx.Query("knowledge_base_id")

	if question == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "问题不能为空"})
		return
	}
	if kbID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "知识库ID不能为空"})
		return
	}

	// 验证UUID格式
	if _, err := uuid.Parse(kbID); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的知识库ID格式"})
		return
	}

	maxIterations := 5
	if maxIterStr := ctx.Query("max_iterations"); maxIterStr != "" {
		if val, err := parseInt(maxIterStr); err == nil && val > 0 && val <= 10 {
			maxIterations = val
		}
	}

	streamAnswer := true
	if streamStr := ctx.Query("stream_answer"); streamStr == "false" {
		streamAnswer = false
	}

	// 调用检索服务
	req := &services.RetrievalRequest{
		Question:        question,
		KnowledgeBaseID: kbID,
		MaxIterations:   maxIterations,
		StreamAnswer:    streamAnswer,
	}

	stream, err := c.retrievalService.StreamRetrieve(ctx.Request.Context(), req)
	if err != nil {
		c.log.Error("流式检索失败", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer stream.Close()

	// 设置SSE响应头
	ctx.Header("Content-Type", "text/event-stream")
	ctx.Header("Cache-Control", "no-cache")
	ctx.Header("Connection", "keep-alive")
	ctx.Header("X-Accel-Buffering", "no")

	// 转发SSE流
	ctx.Stream(func(w io.Writer) bool {
		buf := make([]byte, 4096)
		n, err := stream.Read(buf)
		if err != nil {
			if err != io.EOF {
				c.log.Error("读取SSE流失败", zap.Error(err))
			}
			return false
		}
		if n > 0 {
			w.Write(buf[:n])
			ctx.Writer.Flush()
		}
		return true
	})
}

func parseInt(s string) (int, error) {
	var result int
	_, err := func() (interface{}, error) {
		for _, c := range s {
			if c < '0' || c > '9' {
				return nil, nil
			}
			result = result*10 + int(c-'0')
		}
		return result, nil
	}()
	return result, err
}
