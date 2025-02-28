package handlers

import (
	"github.com/hairutdin/url-shortener/internal/config"
	"github.com/hairutdin/url-shortener/internal/service"
	"go.uber.org/zap"
)

type BaseHandler struct {
	service service.IURLService
	logger  *zap.Logger
	cfg     *config.Config
}

func NewBaseHandler(service service.IURLService, logger *zap.Logger, cfg *config.Config) *BaseHandler {
	return &BaseHandler{
		service: service,
		logger:  logger,
		cfg:     cfg,
	}
}
