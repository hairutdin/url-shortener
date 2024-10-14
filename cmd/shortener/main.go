package shortener

import (
	"compress/gzip"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hairutdin/url-shortener/config"
	"github.com/hairutdin/url-shortener/internal/middleware"
	"go.uber.org/zap"
)

type ShortenedURL struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

var urlStore = struct {
	sync.RWMutex
	m map[string]ShortenedURL
}{
	m: make(map[string]ShortenedURL),
}

func generateShortURL() (string, error) {
	b := make([]byte, 6)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func generateUUID() string {
	return uuid.New().String()
}

func createShortURL(originalURL string) (string, error) {
	shortURL, err := generateShortURL()
	if err != nil {
		return "", err
	}
	urlStore.Lock()
	defer urlStore.Unlock()
	urlStore.m[shortURL] = ShortenedURL{
		UUID:        generateUUID(),
		ShortURL:    shortURL,
		OriginalURL: originalURL,
	}
	return shortURL, nil
}

func loadURLsFromFile(filePath string) error {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil
	}

	fileData, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	var urls []ShortenedURL
	err = json.Unmarshal(fileData, &urls)
	if err != nil {
		return err
	}

	urlStore.Lock()
	defer urlStore.Unlock()
	for _, url := range urls {
		urlStore.m[url.ShortURL] = url
	}

	return nil
}

func saveURLsToFile(filePath string) error {
	urlStore.RLock()
	defer urlStore.RUnlock()

	urls := make([]ShortenedURL, 0, len(urlStore.m))
	for _, url := range urlStore.m {
		urls = append(urls, url)
	}

	fileData, err := json.Marshal(urls)
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, fileData, 0644)
}

func handleShortenPost(c *gin.Context) {
	cfg := config.LoadConfig()

	var bodyReader io.Reader = c.Request.Body
	if c.GetHeader("Content-Encoding") == "gzip" {
		gz, err := gzip.NewReader(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read gzip body"})
			return
		}
		defer gz.Close()
		bodyReader = gz
	}

	bodyBytes, err := io.ReadAll(bodyReader)
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

	if cfg.FileStoragePath != "" {
		err := saveURLsToFile(cfg.FileStoragePath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save URL to file"})
			return
		}
	}

	response := ShortenResponse{Result: cfg.BaseURL + shortURL}
	responseJSON, err := response.MarshalJSON()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal response"})
		return
	}

	if strings.Contains(c.Request.Header.Get("Accept-Encoding"), "gzip") {
		c.Header("Content-Encoding", "gzip")
		gz := gzip.NewWriter(c.Writer)
		defer gz.Close()
		c.Writer = &middleware.GzipResponseWriter{Writer: gz, ResponseWriter: c.Writer}
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

	c.Redirect(http.StatusTemporaryRedirect, originalURL.OriginalURL)
}

func main() {
	cfg := config.LoadConfig()

	if cfg.FileStoragePath != "" {
		err := loadURLsFromFile(cfg.FileStoragePath)
		if err != nil {
			zap.S().Fatalf("Failed to load URLs from file: %v", err)
		}
	}
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	r := gin.Default()

	r.Use(middleware.Logger(logger))
	r.Use(middleware.GzipMiddleware)

	r.POST("/", handleShortenPost)
	r.POST("/api/shorten", handleShortenPost)
	r.GET("/:id", handleGet)

	if err := r.Run(cfg.ServerAddress); err != nil {
		logger.Fatal("Failed to start server", zap.Error(err))
	}
}
