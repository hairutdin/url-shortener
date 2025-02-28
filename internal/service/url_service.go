package service

import (
	"errors"

	"github.com/google/uuid"
	"github.com/hairutdin/url-shortener/internal/lib"
	"github.com/hairutdin/url-shortener/internal/models"
	"github.com/hairutdin/url-shortener/internal/repository"
	"go.uber.org/zap"
)

type URLService struct {
	storage repository.Storage
	logger  *zap.Logger
	baseURL string
}

var _ IURLService = (*URLService)(nil)

func NewURLService(storage repository.Storage, logger *zap.Logger, baseURL string) *URLService {
	return &URLService{
		storage: storage,
		logger:  logger,
		baseURL: baseURL,
	}
}

func (s *URLService) GetBaseURL() string {
	return s.baseURL
}

func (s *URLService) ShortenURL(originalURL string) (string, error) {
	shortURL, err := lib.GenerateShortURL()
	if err != nil {
		return "", err
	}

	return s.CreateShortURL(shortURL, originalURL)
}

func (s *URLService) CreateShortURL(shortURL, originalURL string) (string, error) {
	uid := lib.GenerateUUID()
	existingShortURL, err := s.storage.CreateShortURL(uid, shortURL, originalURL)
	if err != nil {
		if errors.Is(err, repository.ErrDuplicateURL) {
			return existingShortURL, repository.ErrDuplicateURL
		}
		return "", err
	}
	return existingShortURL, nil
}

func (s *URLService) ShortenBatchURLs(requests []models.BatchShortenRequest) ([]models.BatchShortenResponse, error) {
	var batchResponse []models.BatchShortenResponse

	for _, req := range requests {
		shortUUID := uuid.New().String()

		shortURL, err := s.CreateShortURL(shortUUID, req.OriginalURL)
		if err != nil {
			s.logger.Error(
				"failed to create batch short URL",
				zap.String("originalURL", req.OriginalURL),
				zap.Error(err),
			)
			return nil, err
		}

		batchResponse = append(batchResponse, models.BatchShortenResponse{
			CorrelationID: req.CorrelationID,
			ShortURL:      s.baseURL + "/" + shortURL,
		})
	}

	return batchResponse, nil
}

func (s *URLService) GetOriginalURL(shortURL string) (string, error) {
	originalURL, err := s.storage.GetOriginalURL(shortURL)
	if err != nil {
		return "", errors.New("URL not found")
	}
	return originalURL, nil
}

func (s *URLService) Ping() error {
	return s.storage.Ping()
}
