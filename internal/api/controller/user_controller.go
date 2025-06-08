package controller

import (
	"net/http"

	"github.com/Sesame2/go-admin/internal/models/dto"
	"github.com/Sesame2/go-admin/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UserController struct {
	service *services.UserService
}

func NewUserController(service *services.UserService) *UserController {
	return &UserController{service: service}
}

// GetUser 获取用户信息
// @Summary 获取用户信息
// @Description 获取单个用户的信息
// @Tags 用户模块
// @Accept  json
// @Produce  json
// @Router /users [get]
func (c *UserController) GetUser(ctx *gin.Context) {
	idStr := ctx.Param("id")
	if idStr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "用户ID不能为空"})
		return
	}
	// 将字符串ID转换为uuid.UUID类型
	id, err := uuid.Parse(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID格式"})
		return
	}
	user, err := c.service.GetUser(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, user)
}

// CreateUser 创建用户
// @Summary 创建新用户
// @Description 接收用户信息并创建用户记录
// @Tags 用户模块
// @Accept json
// @Produce json
// @Param data body dto.CreateUserInput true "用户信息"
// @Success 200 {object} ent.User "创建成功，返回用户信息"
// @Router /users [post]
func (c *UserController) CreateUser(ctx *gin.Context) {
	// 创建输入结构体
	var input dto.CreateUserInput

	// 从请求体中解析 JSON 到输入结构体
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的输入数据", "details": err.Error()})
		return
	}
	user, err := c.service.CreateUser(ctx, &input)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "服务器内部错误", "details": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, user)
}

func (c *UserController) GetAllUser(ctx *gin.Context) {
	users, err := c.service.ListUsers(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data": users,
	})
}
