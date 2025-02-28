package service

import (
	"github.com/hairutdin/url-shortener/internal/models"
)

type IURLService interface {
	ShortenURL(originalURL string) (string, error)
	CreateShortURL(shortURL, originalURL string) (string, error)
	ShortenBatchURLs(requests []models.BatchShortenRequest) ([]models.BatchShortenResponse, error)
	GetOriginalURL(shortURL string) (string, error)
	Ping() error
	GetBaseURL() string
}
