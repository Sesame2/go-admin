package controller

import (
	"net/http"
	"strconv"

	"github.com/Sesame2/go-admin/internal/models/dto"
	"github.com/Sesame2/go-admin/internal/models/ent"
	"github.com/Sesame2/go-admin/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type KnowledgeBaseController struct {
	service *services.KnowledgeBaseService
	logger  *zap.Logger
}

func NewKnowledgeBaseController(service *services.KnowledgeBaseService, logger *zap.Logger) *KnowledgeBaseController {
	logger = logger.With(zap.String("component", "KnowledgeBaseController"))
	return &KnowledgeBaseController{
		service: service,
		logger:  logger,
	}
}

// CreateKnowledgeBase godoc
// @Summary      创建知识库
// @Description  创建一个新的知识库
// @Tags         知识库模块
// @Accept       json
// @Produce      json
// @Param        input  body      dto.CreateKnowledgeBaseInput  true  "知识库创建输入"
// @Success      200  {object}  ent.KnowledgeBase  "创建成功的知识库信息"
// @Failure      400  {object}  object{error=string, details=string}  "请求参数错误"
// @Failure      500  {object}  object{error=string, detail=string}  "系统内部错误"
// @Router       /knowledge_bases [post]
func (c *KnowledgeBaseController) CreateKnowledgeBase(ctx *gin.Context) {
	var input *dto.CreateKnowledgeBaseInput

	if err := ctx.ShouldBindJSON(&input); err != nil {
		c.logger.Error("创建知识库验证参数失败", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的输入数据", "details": err.Error(),
		})
		return
	}
	kb, err := c.service.CreateKnowledgeBase(ctx, input)
	if err != nil {
		c.logger.Error("创建知识库失败", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "系统内部错误", "detail": err.Error(),
		})
		return
	}
	ctx.JSON(http.StatusOK, kb)
}

// GetAllKnowledgBase godoc
// @Summary      获取所有知识库
// @Description  分页获取知识库列表
// @Tags         知识库模块
// @Accept       json
// @Produce      json
// @Param        page        query    int     false  "页码，默认为1"
// @Param        page_size   query    int     false  "每页数量，默认为10"
// @Success      200  {object}  dto.KnowledgeBaseListResult  "知识库列表"
// @Failure      400  {object}  object{error=string}  "参数错误"
// @Failure      500  {object}  object{error=string}  "服务器错误"
// @Router       /knowledge_bases [get]
func (c *KnowledgeBaseController) GetAllKnowledgBase(ctx *gin.Context) {
	pageStr := ctx.DefaultQuery("page", "1")
	pageSizeStr := ctx.DefaultQuery("page_size", "10")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		c.logger.Warn("无效的页码参数", zap.String("page", pageStr), zap.Error(err))
		page = 1
	}
	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 {
		c.logger.Warn("无效的每页数量参数",
			zap.String("page_size", pageSizeStr),
			zap.Error(err))
		pageSize = 10
	}
	c.logger.Info("查询知识库列表",
		zap.Int("page", page),
		zap.Int("page_size", pageSize))

	kbs, err := c.service.GetAllKnowledgBase(ctx, page, pageSize)
	if err != nil {
		c.logger.Error("获取知识库列表失败", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取知识库列表失败",
		})
		return
	}
	ctx.JSON(http.StatusOK, kbs)
}

// GetKnowledgeBaseByID godoc
// @Summary      获取知识库详情
// @Description  根据ID获取知识库的详细信息
// @Tags         知识库模块
// @Accept       json
// @Produce      json
// @Param        id    path     string  true  "知识库ID (UUID格式)"
// @Success      200  {object}  ent.KnowledgeBase  "知识库详情"
// @Failure      400  {object}  object{error=string}  "请求参数错误"
// @Failure      404  {object}  object{error=string}  "知识库不存在"
// @Failure      500  {object}  object{error=string}  "服务器错误"
// @Router       /knowledge_bases/{id} [get]
func (c *KnowledgeBaseController) GetKnowledgeBaseByID(ctx *gin.Context) {
	idStr := ctx.Param("id")
	if idStr == "" {
		c.logger.Error("知识库ID不能为空")
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "知识库ID不能为空",
		})
		return
	}
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.logger.Error("无效的ID格式")
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的ID格式",
		})
		return
	}
	kb, err := c.service.GetKnowledgeBaseByID(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			c.logger.Warn("知识库不存在", zap.String("id", idStr))
			ctx.JSON(http.StatusNotFound, gin.H{
				"error": "知识库不存在",
			})
			return
		}
		c.logger.Error("获取知识库错误", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取知识库错误",
		})
		return
	}
	ctx.JSON(http.StatusOK, kb)
}
