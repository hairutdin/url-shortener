package shortener

import (
	"crypto/rand"
	"encoding/base64"
	"io"
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

func createShortURL(originalURL string) (string, error) {
	shortURL, err := generateShortURL()
	if err != nil {
		return "", err
	}
	urlStore.Lock()
	defer urlStore.Unlock()
	urlStore.m[shortURL] = originalURL

	return shortURL, nil
}

func handleShortenPost(c *gin.Context) {
	cfg := config.LoadConfig()

	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
		return
	}

	var requestBody ShortenRequest
	if err := requestBody.UnmarshalJSON(bodyBytes); err != nil || requestBody.URL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid URL"})
		return
	}

	shortURL, err := createShortURL(requestBody.URL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate short URL"})
		return
	}

	response := ShortenResponse{Result: cfg.BaseURL + shortURL}
	responseJSON, err := response.MarshalJSON()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal response"})
		return
	}

	c.Data(http.StatusCreated, "application/json", responseJSON)
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

	r.POST("/", handleShortenPost)
	r.POST("/api/shorten", handleShortenPost)
	r.GET("/:id", handleGet)

	if err := r.Run(cfg.ServerAddress); err != nil {
		logger.Fatal("Failed to start server", zap.Error(err))
	}
}
