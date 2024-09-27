package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func main() {
	r := gin.Default()
	r.POST("/shorten", shortenURL)
	if err := r.Run(":8080"); err != nil {
		panic(err)
	}
}

func shortenURL(c *gin.Context) {
	longURL := c.PostForm("url")
	if longURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Empty URL"})
		return
	}

	shortenedURL := "http://localhost:8080/short123"
	c.JSON(http.StatusCreated, gin.H{
		"long_url":  longURL,
		"short_url": shortenedURL,
	})
}
