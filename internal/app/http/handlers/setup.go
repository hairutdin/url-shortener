package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/hairutdin/url-shortener/internal/app/http/middleware"
	"github.com/hairutdin/url-shortener/internal/config"
	"go.uber.org/zap"
)

func SetupRouter(cfg *config.Config, logger *zap.Logger, handler *BaseHandler) *gin.Engine {
	r := gin.Default()

	r.Use(middleware.Logger(logger))
	r.Use(middleware.GzipMiddleware)

	r.POST("/", handler.HandleShortenPost)
	r.POST("/api/shorten", handler.HandleShortenPost)
	r.POST("/api/shorten/batch", handler.handleBatchShortenPost)
	r.GET("/:id", handler.handleGet)
	r.GET("/ping", handler.handlePing)

	return r
}
