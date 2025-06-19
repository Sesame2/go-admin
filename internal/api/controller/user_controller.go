package controller

import (
	"errors"
	"net/http"

	customerrors "github.com/Sesame2/go-admin/internal/errors" // 自定义错误包使用别名
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

// GetUser godoc
// @Summary      获取单个用户信息
// @Description  根据用户ID获取特定用户的详细信息
// @Tags         用户模块
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        id   path      string  true  "用户ID (UUID格式)"
// @Success      200  {object}  ent.User       "用户详细信息"
// @Failure      400  {object}  object{error=string}  "请求参数错误"
// @Failure      404  {object}  object{error=string}  "用户不存在"
// @Failure      500  {object}  object{error=string}  "服务器内部错误"
// @Router       /users/{id} [get]
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
		// 处理用户不存在的情况
		if errors.Is(err, customerrors.ErrUserNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
			return
		}
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
// @Security     Bearer
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

// GetAllUser godoc
// @Summary      获取所有用户
// @Description  获取系统中所有用户的列表
// @Tags         用户模块
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Success      200  {object}  object{data=[]ent.User}  "返回用户列表"
// @Router       /users [get]
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

// UpdateUser godoc
// @Summary      更新用户信息
// @Description  更新用户的各种信息，支持部分字段更新
// @Tags         用户模块
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        id   path      string  true  "用户ID"
// @Param        user body      dto.UpdateUserInput  true  "用户更新信息"
// @Success      200  {object}  ent.User
// @Router       /users/{id} [put]
func (c *UserController) UpdateUser(ctx *gin.Context) {
	userID := ctx.Param("id")
	id, err := uuid.Parse(userID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户格式"})
		return
	}
	var input dto.UpdateUserInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	user, err := c.service.UpdateUser(ctx, id, &input)
	if err != nil {
		if errors.Is(err, customerrors.ErrUserNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, user)
}
