package main

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
)

var urlStore = struct {
	sync.RWMutex
	m map[string]string
}{
	m: make(map[string]string),
}

const baseURL = "http://localhost:8080/"

func generateShortURL() string {
	b := make([]byte, 6)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	return base64.URLEncoding.EncodeToString(b)
}

func handlePost(c *gin.Context) {
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

	c.JSON(http.StatusCreated, gin.H{"short_url": baseURL + shortURL})
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
	r := gin.Default()

	r.POST("/shorten", handlePost)
	r.GET("/:id", handleGet)

	if err := r.Run(":8080"); err != nil {
		panic(err)
	}
}
