package controller

import (
	"fmt"
	"github.com/gin-gonic/gin"
)

type UserController struct {

}

func NewUserController() *UserController {
	return &UserController{}
}

func (controller *UserController) GetUser(c *gin.Context) {
	fmt.Println("GetUser is called")
}