package controller

import (
	"net/http"
	"strconv"

	"github.com/Sesame2/go-admin/internal/logger"
	"github.com/Sesame2/go-admin/internal/models/dto"
	"github.com/Sesame2/go-admin/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type DocumentController struct {
	service *services.DocumentService
	log     *zap.Logger
}

func NewDocumentController(service *services.DocumentService) *DocumentController {
	return &DocumentController{
		service: service,
		log:     logger.NewModuleLogger("DocumentController"),
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

	doc, err := c.service.UploadDocument(ctx, kbID, file)
	if err != nil {
		c.log.Error("上传文档失败", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, doc)
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

// GetAllDocuments godoc
// @Summary      获取知识库下所有文档
// @Description  分页获取指定知识库下的所有文档
// @Tags         文档模块
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        knowledge_base_id  query     string  true   "知识库ID"
// @Param        page               query     int     false  "页码，默认为1"
// @Param        page_size          query     int     false  "每页数量，默认为10"
// @Success      200  {object}  dto.DocumentListResult  "文档列表"
// @Failure      400  {object}  object{error=string}  "请求参数错误"
// @Failure      500  {object}  object{error=string}  "服务器内部错误"
// @Router       /documents [get]
func (c *DocumentController) GetAllDocuments(ctx *gin.Context) {
	kbIDStr := ctx.Query("knowledge_base_id")
	if kbIDStr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "知识库ID不能为空"})
		return
	}
	kbID, err := uuid.Parse(kbIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的知识库ID格式"})
		return
	}

	pageStr := ctx.DefaultQuery("page", "1")
	pageSizeStr := ctx.DefaultQuery("page_size", "10")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}
	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 {
		pageSize = 10
	}

	result, err := c.service.ListDocuments(ctx, kbID, page, pageSize)
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

	err = c.service.DeleteDocument(ctx, id)
	if err != nil {
		c.log.Error("删除文档失败", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "文档删除成功"})
}
