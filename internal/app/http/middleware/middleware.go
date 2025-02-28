package middleware

import (
	"compress/gzip"
	"io"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type GzipResponseWriter struct {
	gin.ResponseWriter
	Writer io.Writer
}

func GzipMiddleware(c *gin.Context) {
	if strings.Contains(c.GetHeader("Content-Encoding"), "gzip") {
		gz, err := gzip.NewReader(c.Request.Body)
		if err != nil {
			c.JSON(400, gin.H{"error": "Invalid gzip encoding"})
			c.Abort()
			return
		}
		defer gz.Close()
		c.Request.Body = io.NopCloser(gz)
	}

	if strings.Contains(c.GetHeader("Accept-Encoding"), "gzip") {
		gz := gzip.NewWriter(c.Writer)
		defer gz.Close()

		c.Header("Content-Encoding", "gzip")
		c.Writer = &GzipResponseWriter{Writer: gz, ResponseWriter: c.Writer}
	}

	c.Next()
}

func (w *GzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func Logger(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		duration := time.Since(start)
		status := c.Writer.Status()
		size := c.Writer.Size()

		logger.Info("Request info",
			zap.String("uri", c.Request.RequestURI),
			zap.String("method", c.Request.Method),
			zap.Duration("duration", duration),
			zap.Int("status", status),
			zap.Int("size", size),
		)
	}
}
