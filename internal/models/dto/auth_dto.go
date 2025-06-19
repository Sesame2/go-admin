package dto

type LoginInput struct {
    Username string `json:"username" binding:"required,min=3,max=50" example:"test"`
    Password string `json:"password" binding:"required,min=6" example:"test123"`
}
