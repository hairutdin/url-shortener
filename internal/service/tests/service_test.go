package tests

import (
	"errors"
	"testing"

	"github.com/hairutdin/url-shortener/internal/models"
	"github.com/hairutdin/url-shortener/internal/repository"

	"github.com/golang/mock/gomock"
	"github.com/hairutdin/url-shortener/internal/repository/mocks"
	"github.com/hairutdin/url-shortener/internal/service"
	"go.uber.org/zap"
)

func TestCreateShortURL_Duplicate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)
	logger, _ := zap.NewDevelopment()
	urlService := service.NewURLService(mockStorage, logger, "http://localhost:8080")

	originalURL := "https://example.com"
	shortURL := "short123"

	mockStorage.EXPECT().
		CreateShortURL(gomock.Any(), gomock.Any(), originalURL).
		Return(shortURL, repository.ErrDuplicateURL)

	result, err := urlService.CreateShortURL(shortURL, originalURL)

	if err != repository.ErrDuplicateURL {
		t.Errorf("Expected ErrDuplicateURL, got %v", err)
	}

	if result != shortURL {
		t.Errorf("Expected existing short URL '%s', got '%s'", shortURL, result)
	}
}

func TestShortenBatchURLs_ErrorOnDuplicate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)
	logger, _ := zap.NewDevelopment()
	urlService := service.NewURLService(mockStorage, logger, "http://localhost:8080")

	requests := []models.BatchShortenRequest{
		{CorrelationID: "1", OriginalURL: "https://example1.com"},
		{CorrelationID: "2", OriginalURL: "https://example2.com"},
	}

	mockStorage.EXPECT().CreateShortURL(gomock.Any(), gomock.Any(), requests[0].OriginalURL).Return("short1", nil)
	mockStorage.EXPECT().
		CreateShortURL(gomock.Any(), gomock.Any(), requests[1].OriginalURL).
		Return("", repository.ErrDuplicateURL)

	batchResponse, err := urlService.ShortenBatchURLs(requests)

	if err != repository.ErrDuplicateURL {
		t.Errorf("Expected ErrDuplicateURL, got %v", err)
	}

	if batchResponse != nil {
		t.Errorf("Expected batchResponse to be nil on error, got %+v", batchResponse)
	}
}

func TestGetOriginalURL_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)
	logger, _ := zap.NewDevelopment()
	urlService := service.NewURLService(mockStorage, logger, "http://localhost:8080")

	shortURL := "short-not-exist"

	mockStorage.EXPECT().GetOriginalURL(shortURL).Return("", errors.New("not found"))

	result, err := urlService.GetOriginalURL(shortURL)

	if err == nil || err.Error() != "URL not found" {
		t.Errorf("Expected 'URL not found' error, got %v", err)
	}

	if result != "" {
		t.Errorf("Expected empty result, got %s", result)
	}
}
