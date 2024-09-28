package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hairutdin/url-shortener/config"
)

func main() {
	cfg := config.LoadConfig()

	r := gin.Default()
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
	return "short123", nil
}
