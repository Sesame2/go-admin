package controller

import (
	"net/http"

	"github.com/Sesame2/go-admin/internal/models/dto"
	"github.com/Sesame2/go-admin/internal/services"
	"github.com/gin-gonic/gin"
)

type AuthController struct {
	service *services.AuthService
}

func NewAuthController(service *services.AuthService) *AuthController {
	return &AuthController{
		service: service,
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
