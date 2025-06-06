package handler

import (
	"auth_service/internal/repository"
	"auth_service/internal/service"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

type AuthHandler struct {
	service service.AuthService
}

func NewAuthHandler(service *service.AuthService) *AuthHandler {
	return &AuthHandler{service: *service}
}

func (Handler *AuthHandler) Login(c *gin.Context) {
	var user repository.UserInfoDBO

	err := c.ShouldBind(&user)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный json"})
		return
	}

	if user.Login == "" || user.Password == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Логин или пароль не может быть пустым"})
		return
	}

	token, err := Handler.service.Login(c.Request.Context(), &user)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": token})
}

func (Handler *AuthHandler) Register(c *gin.Context) {
	var user repository.UserInfoDBO
	err := c.ShouldBind(&user)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if user.Login == "" || user.Password == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Логин или пароль не может быть пустым"})
		return
	}

	err = Handler.service.Register(c.Request.Context(), &user)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Пользователь %s успешно зарегистрирован", user.Login)})
}
