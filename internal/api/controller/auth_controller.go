package controller

import (
	"net/http"

	"github.com/Sesame2/go-admin/internal/logger"
	"github.com/Sesame2/go-admin/internal/models/dto"
	"github.com/Sesame2/go-admin/internal/services"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type AuthController struct {
	service *services.AuthService
	log     *zap.Logger
}

func NewAuthController(service *services.AuthService) *AuthController {
	return &AuthController{
		service: service,
		log:     logger.NewModuleLogger("AuthController"),
	}
}

// Login godoc
// @Summary      用户登录
// @Description  验证用户凭据并返回JWT令牌
// @Tags         认证
// @Accept       json
// @Produce      json
// @Param        credentials  body      dto.LoginInput  true  "登录凭据"
// @Success      200  {object}  object{token=string}    "登录成功返回令牌"
// @Failure      400  {object}  object{error=string}    "请求参数错误"
// @Failure      401  {object}  object{error=string}    "认证失败"
// @Router       /auth/login [post]
func (c *AuthController) Login(ctx *gin.Context) {
	var input dto.LoginInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的输入"})
		return
	}
	token, err := c.service.Login(ctx, input.Username, input.Password)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"token": token,
	})
}

// Refresh godoc
// @Summary      刷新令牌
// @Description  刷新JWT令牌以延长会话
// @Tags         认证
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Success      200  {object}  object{token=string}    "刷新成功返回新令牌"
// @Failure      400  {object}  object{error=string}    "请求参数错误"
// @Failure      401  {object}  object{error=string}    "令牌无效或过期"
// @Failure      500  {object}  object{error=string}    "服务器内部错误"
// @Router       /auth/refresh [post]
func (c *AuthController) Refresh(ctx *gin.Context) {
	tokenStr := ctx.GetHeader("Authorization")
	if tokenStr == "" {
		c.log.Error("缺少令牌")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "缺少令牌"})
		return
	}
	if len(tokenStr) < 7 || tokenStr[:7] != "Bearer " {
		c.log.Error("令牌格式错误", zap.String("token", tokenStr))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "令牌格式错误"})
		return
	}
	tokenStr = tokenStr[7:]

	newToken, err := c.service.Refresh(ctx, tokenStr)
	if err != nil {
		c.log.Error("刷新令牌失败", zap.Error(err))
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"token": newToken,
	})
	c.log.Info("令牌刷新成功")
}
