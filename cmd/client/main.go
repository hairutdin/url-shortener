package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hairutdin/url-shortener/config"
)

func main() {
	cfg := config.LoadConfig()

	r := gin.Default()
	r.POST("/shorten", shortenURL)
	if err := r.Run(cfg.ServerAddress); err != nil {
		panic(err)
	}
}

func shortenURL(c *gin.Context) {
	longURL := c.PostForm("url")
	if longURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Empty URL"})
		return
	}

	cfg := config.LoadConfig()
	shortenedURL := cfg.BaseURL + "short123"

	c.JSON(http.StatusCreated, gin.H{
		"long_url":  longURL,
		"short_url": shortenedURL,
	})
}
