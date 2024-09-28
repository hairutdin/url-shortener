package main

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/hairutdin/url-shortener/config"
)

var urlStore = struct {
	sync.RWMutex
	m map[string]string
}{
	m: make(map[string]string),
}

func generateShortURL() string {
	b := make([]byte, 6)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	return base64.URLEncoding.EncodeToString(b)
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

	shortURL := generateShortURL()

	urlStore.Lock()
	urlStore.m[shortURL] = requestBody.URL
	urlStore.Unlock()

	c.JSON(http.StatusCreated, gin.H{"short_url": cfg.BaseURL + shortURL})
}

func handleGet(c *gin.Context) {
	id := c.Param("id")

	urlStore.RLock()
	originalURL, ok := urlStore.m[id]
	urlStore.RUnlock()

	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "URL not found"})
		return
	}

	c.Redirect(http.StatusTemporaryRedirect, originalURL)
}

func main() {
	cfg := config.LoadConfig()

	r := gin.Default()

	r.POST("/shorten", handlePost)
	r.GET("/:id", handleGet)

	if err := r.Run(cfg.ServerAddress); err != nil {
		panic(err)
	}
}
