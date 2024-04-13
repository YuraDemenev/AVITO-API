package handler

import (
	"avito/pkg/service"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *service.Service
}

func NewHandler(service *service.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) InitRoutes() *gin.Engine {
	//Создаем router
	router := gin.New()

	//Определяем пути и какие методы
	//будут вызываться
	router.GET("/user_banner", h.userGetBanner)
	router.GET("/banner", h.adminGetBanners)
	router.POST("/banner", h.adminCreateBanner)
	router.PATCH("/banner/:{id}", h.adminUpdateBanner)
	router.DELETE("/banner/:{id}", h.adminDeleteBanner)

	return router
}
