package main

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/hairutdin/url-shortener/config"
	"github.com/hairutdin/url-shortener/internal/middleware"
	"go.uber.org/zap"
)

var urlStore = struct {
	sync.RWMutex
	m map[string]string
}{
	m: make(map[string]string),
}

func generateShortURL() (string, error) {
	b := make([]byte, 6)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func handlePost(c *gin.Context) {
	cfg := config.LoadConfig()

	var requestBody struct {
		URL string `json:"url"`
	}

	if err := c.ShouldBindJSON(&requestBody); err != nil || requestBody.URL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid URL"})
		return
	}

	shortURL, err := generateShortURL()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not generate short URL"})
		return
	}

	urlStore.Lock()
	defer urlStore.Unlock()
	urlStore.m[shortURL] = requestBody.URL

	c.JSON(http.StatusCreated, gin.H{"short_url": cfg.BaseURL + shortURL})
}

func handleGet(c *gin.Context) {
	id := c.Param("id")

	urlStore.RLock()
	defer urlStore.RUnlock()
	originalURL, exists := urlStore.m[id]

	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "URL not found"})
		return
	}

	c.Redirect(http.StatusTemporaryRedirect, originalURL)
}

func main() {
	cfg := config.LoadConfig()

	logger, _ := zap.NewProduction()
	defer logger.Sync()

	r := gin.Default()

	r.Use(middleware.Logger(logger))

	r.POST("/", handlePost)
	r.GET("/:id", handleGet)

	if err := r.Run(cfg.ServerAddress); err != nil {
		logger.Fatal("Failed to start server", zap.Error(err))
	}
}
