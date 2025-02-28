package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hairutdin/url-shortener/internal/app/http/handlers"
	"github.com/hairutdin/url-shortener/internal/box"
	"github.com/hairutdin/url-shortener/internal/service"
)

var testServer *gin.Engine

func TestMain(m *testing.M) {
	os.Setenv("DATABASE_DSN", "postgres://testuser:testpassword@localhost:5433/testdb?sslmode=disable")

	time.Sleep(5 * time.Second)

	envBox, err := box.New()
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize test environment: %v", err))
	}

	urlService := service.NewURLService(envBox.Storage, envBox.Logger, envBox.Config.BaseURL)
	baseHandler := handlers.NewBaseHandler(urlService, envBox.Logger, envBox.Config)
	testServer = handlers.SetupRouter(envBox.Config, envBox.Logger, baseHandler)

	code := m.Run()
	os.Exit(code)
}

func TestIntegration_ShorteningAndRetrievingURL(t *testing.T) {
	requestBody := `{"url": "https://example.com"}`
	req, _ := http.NewRequest(http.MethodPost, "/api/shorten", jsonBody(requestBody))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	testServer.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("Expected 201, got %d", recorder.Code)
	}

	var response map[string]string
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response JSON: %v", err)
	}

	shortURL, ok := response["result"]
	if !ok {
		t.Fatalf("Expected 'result' field in response")
	}

	req, _ = http.NewRequest(http.MethodGet, shortURL[len("http://localhost:8080/"):], nil)
	recorder = httptest.NewRecorder()

	testServer.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusTemporaryRedirect {
		t.Fatalf("Expected 307, got %d", recorder.Code)
	}

	originalURL := recorder.Header().Get("Location")
	if originalURL != "https://example.com" {
		t.Errorf("Expected redirect to 'https://example.com', got '%s'", originalURL)
	}
}

func TestIntegration_PingDatabase(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "/ping", nil)
	recorder := httptest.NewRecorder()

	testServer.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("Expected 200, got %d", recorder.Code)
	}

	expectedResponse := `{"status":"Database connection OK"}`
	if recorder.Body.String() != expectedResponse {
		t.Errorf("Expected '%s', got '%s'", expectedResponse, recorder.Body.String())
	}
}

func jsonBody(body string) *bytes.Reader {
	return bytes.NewReader([]byte(body))
}
