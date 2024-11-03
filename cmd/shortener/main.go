package shortener

import (
	"compress/gzip"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/hairutdin/url-shortener/config"
	"github.com/hairutdin/url-shortener/internal/middleware"
	"github.com/hairutdin/url-shortener/internal/storage"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

var storageInstance storage.Storage

func initializeStorage(cfg *config.Config) (storage.Storage, error) {
	if cfg.DatabaseDSN != "" {
		return storage.NewPostgresStorage(cfg.DatabaseDSN)
	} else if cfg.FileStoragePath != "" {
		return storage.NewFileStorage(cfg.FileStoragePath)
	}
	return storage.NewInMemoryStorage(), nil
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

func handleShortenPost(c *gin.Context, cfg *config.Config) {
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

	var requestBody struct {
		URL string `json:"url" binding:"required"`
	}
	if err := json.NewDecoder(bodyReader).Decode(&requestBody); err != nil || requestBody.URL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid URL"})
		return
	}

	shortURL, err := generateShortURL()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate short URL"})
		return
	}

	err = storageInstance.CreateShortURL(generateUUID(), shortURL, requestBody.URL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save URL"})
		return
	}

	response := gin.H{"result": cfg.BaseURL + "/" + shortURL}
	if strings.Contains(c.Request.Header.Get("Accept-Encoding"), "gzip") {
		c.Header("Content-Encoding", "gzip")
		gz := gzip.NewWriter(c.Writer)
		defer gz.Close()
		c.Writer = &middleware.GzipResponseWriter{Writer: gz, ResponseWriter: c.Writer}
	}

	c.JSON(http.StatusCreated, response)
}

func handleBatchShortenPost(c *gin.Context, cfg *config.Config) {
	var batchRequest []BatchShortenRequest
	if err := c.BindJSON(&batchRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	if len(batchRequest) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Empty batch not allowed"})
		return
	}

	var batchResponse []BatchShortenResponse
	for _, req := range batchRequest {
		shortUUID := uuid.New().String()
		shortURL := cfg.BaseURL + "/" + shortUUID

		err := storageInstance.CreateShortURL(shortUUID, shortUUID, req.OriginalURL)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create short URL"})
			return
		}
		batchResponse = append(batchResponse, BatchShortenResponse{
			CorrelationID: req.CorrelationID,
			ShortURL:      shortURL,
		})
	}

	c.JSON(http.StatusCreated, batchResponse)
}

func handleGet(c *gin.Context) {
	shortURL := c.Param("id")
	originalURL, err := storageInstance.GetOriginalURL(shortURL)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "URL not found"})
		return
	}
	c.Redirect(http.StatusTemporaryRedirect, originalURL)
}

func handlePing(c *gin.Context) {
	if err := storageInstance.Ping(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "Database connection failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "Database connection OK"})
}

func setupRouter(cfg *config.Config, logger *zap.Logger) *gin.Engine {
	r := gin.Default()

	r.Use(middleware.Logger(logger))
	r.Use(middleware.GzipMiddleware)

	r.POST("/", func(c *gin.Context) { handleShortenPost(c, cfg) })
	r.POST("/api/shorten", func(c *gin.Context) { handleShortenPost(c, cfg) })
	r.POST("/api/shorten/batch", func(c *gin.Context) { handleBatchShortenPost(c, cfg) })
	r.GET("/:id", handleGet)
	r.GET("/ping", handlePing)

	return r
}

func main() {
	cfg := config.LoadConfig()
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	var err error
	storageInstance, err = initializeStorage(cfg)
	if err != nil {
		logger.Fatal("Failed to initialize storage", zap.Error(err))
	}
	defer storageInstance.Close()

	if err := storageInstance.Ping(); err != nil {
		logger.Fatal("Storage connection failed", zap.Error(err))
	}
	logger.Info("Storage initialized successfully")

	r := setupRouter(cfg, logger)
	if err := r.Run(cfg.ServerAddress); err != nil {
		logger.Fatal("Failed to start server", zap.Error(err))
	}
}
