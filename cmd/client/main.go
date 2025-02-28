package main

import (
	"compress/gzip"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/hairutdin/url-shortener/internal/app/http/middleware"
	"github.com/hairutdin/url-shortener/internal/config"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func main() {
	cfg := config.LoadConfig()

	logger, _ := zap.NewProduction()
	defer func() {
		if err := logger.Sync(); err != nil {
			fmt.Printf("logger sync failed: %v\n", err)
		}
	}()
	r := gin.Default()

	r.Use(middleware.Logger(logger))

	r.Use(middleware.GzipMiddleware)

	r.POST("/", shortenURL(cfg))
	if err := r.Run(cfg.HTTP.Address); err != nil {
		panic(err)
	}
}

func shortenURL(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {

		var bodyReader io.Reader = c.Request.Body
		if strings.Contains(c.Request.Header.Get("Content-Encoding"), "gzip") {
			gzipReader, err := gzip.NewReader(c.Request.Body)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid gzip encoding"})
				return
			}
			defer gzipReader.Close()
			bodyReader = gzipReader
		}

		var requestBody struct {
			URL string `json:"url" binding:"required"`
		}

		decoder := json.NewDecoder(bodyReader)
		if err := decoder.Decode(&requestBody); err != nil || requestBody.URL == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid URL"})
			return
		}

		shortenedURL, err := generateShortURL()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not generate short URL"})
			return
		}

		c.Header("Content-Type", "application/json")
		response := gin.H{
			"long_url":  requestBody.URL,
			"short_url": cfg.BaseURL + shortenedURL,
		}

		if strings.Contains(c.Request.Header.Get("Accept-Encoding"), "gzip") {
			c.Header("Content-Encoding", "gzip")
			gzipWriter := gzip.NewWriter(c.Writer)
			defer gzipWriter.Close()
			c.Writer = &middleware.GzipResponseWriter{Writer: gzipWriter, ResponseWriter: c.Writer}
		}
		c.JSON(http.StatusCreated, response)
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
