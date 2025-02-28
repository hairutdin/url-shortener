package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hairutdin/url-shortener/internal/models"
	"github.com/hairutdin/url-shortener/internal/repository"
	"go.uber.org/zap"
)

func (h *BaseHandler) HandleShortenPost(c *gin.Context) {
	var requestBody struct {
		URL string `json:"url" binding:"required"`
	}
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		h.logger.Warn("Invalid request format", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	shortURL, err := h.service.ShortenURL(requestBody.URL)
	if err != nil {
		if errors.Is(err, repository.ErrDuplicateURL) {
			c.JSON(http.StatusConflict, gin.H{"short_url": h.cfg.BaseURL + "/" + shortURL})
			return
		}
		h.logger.Error("Failed to generate short URL", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate short URL"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"result": h.cfg.BaseURL + "/" + shortURL})
}

func (h *BaseHandler) handleBatchShortenPost(c *gin.Context) {
	var batchRequest []models.BatchShortenRequest
	if err := c.BindJSON(&batchRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	if len(batchRequest) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Empty batch not allowed"})
		return
	}

	batchResponse, err := h.service.ShortenBatchURLs(batchRequest)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create batch URLs"})
		return
	}

	c.JSON(http.StatusCreated, batchResponse)
}

func (h *BaseHandler) handleGet(c *gin.Context) {
	shortURL := c.Param("id")
	originalURL, err := h.service.GetOriginalURL(shortURL)
	if err != nil {
		h.logger.Error("URL not found", zap.String("shortURL", shortURL), zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "URL not found"})
		return
	}
	c.Redirect(http.StatusTemporaryRedirect, originalURL)
}

func (h *BaseHandler) handlePing(c *gin.Context) {
	if err := h.service.Ping(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "Database connection failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "Database connection OK"})
}
