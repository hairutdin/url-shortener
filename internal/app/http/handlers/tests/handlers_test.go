package tests

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/hairutdin/url-shortener/internal/app/http/handlers"
	"github.com/hairutdin/url-shortener/internal/config"
	"github.com/hairutdin/url-shortener/internal/service/mocks"
	"go.uber.org/zap"
)

func setupTestHandler(mockService *mocks.MockIURLService) *handlers.BaseHandler {
	logger, _ := zap.NewDevelopment()
	return handlers.NewBaseHandler(mockService, logger, &config.Config{BaseURL: "http://localhost:8080"})
}

func TestHandleShortenPost(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockIURLService(ctrl)
	mockService.EXPECT().ShortenURL("https://example.com").Return("short123", nil)

	router := gin.Default()
	handler := setupTestHandler(mockService)
	router.POST("/api/shorten", handler.HandleShortenPost)

	body := `{"url": "https://example.com"}`
	req, _ := http.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", recorder.Code)
	}

	var response map[string]string
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response JSON: %v", err)
	}

	if response["result"] != "http://localhost:8080/short123" {
		t.Errorf("Expected short URL 'http://localhost:8080/short123', got %s", response["result"])
	}
}
