package main

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hairutdin/url-shortener/config"
	"github.com/hairutdin/url-shortener/internal/middleware"
	"go.uber.org/zap"
)

func main() {
	cfg := config.LoadConfig()

	logger, _ := zap.NewProduction()
	defer logger.Sync()

	r := gin.Default()

	r.Use(middleware.Logger(logger))

	r.POST("/", shortenURL(cfg))
	if err := r.Run(cfg.ServerAddress); err != nil {
		panic(err)
	}
}

func shortenURL(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var requestBody struct {
			URL string `json:"url" binding:"required"`
		}

		if err := c.ShouldBindJSON(&requestBody); err != nil || requestBody.URL == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid URL"})
			return
		}

		shortenedURL, err := generateShortURL()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not generate short URL"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"long_url":  requestBody.URL,
			"short_url": cfg.BaseURL + shortenedURL,
		})
	}
}

func generateShortURL() (string, error) {
	byteLength := 6

	randomBytes := make([]byte, byteLength)

	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}

	shortURL := base64.RawURLEncoding.EncodeToString(randomBytes)
	return shortURL, nil
}
