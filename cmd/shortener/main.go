package main

import (
	"crypto/rand"
	"encoding/base64"
	"io"
	"net/http"
	"strings"
	"sync"
)

// For saving URL we will use map
var urlStore = struct {
	sync.RWMutex
	m map[string]string
}{
	m: make(map[string]string),
}

const baseURL = "http://localhost:8080/"

func generateShortURL() string {
	b := make([]byte, 6)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	return base64.URLEncoding.EncodeToString(b)
}

func handleRequest(res http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodPost {
		handlePost(res, req)
	} else if req.Method == http.MethodGet {
		handleGet(res, req)
	} else {
		http.Error(res, "Invalid method", http.StatusBadRequest)
	}
}

func handlePost(res http.ResponseWriter, req *http.Request) {
	// Reading URL from request body
	body, err := io.ReadAll(req.Body)
	if err != nil || len(body) == 0 {
		http.Error(res, "Invalid request body", http.StatusBadRequest)
		return
	}

	originalURL := string(body)

	// Generate a short URL
	shortURL := generateShortURL()

	// Save the URL in the map
	urlStore.Lock()
	urlStore.m[shortURL] = originalURL
	urlStore.Unlock()

	// Respond with the short URL
	res.Header().Set("Content-Type", "text/plain")
	res.WriteHeader(http.StatusCreated)
	res.Write([]byte(baseURL + shortURL))
}

func handleGet(res http.ResponseWriter, req *http.Request) {
	// Reading ID from the URL path
	id := strings.TrimPrefix(req.URL.Path, "/")

	// Retrieve the original URL
	urlStore.RLock()
	originalURL, ok := urlStore.m[id]
	urlStore.RUnlock()

	if !ok {
		http.Error(res, "URL not found", http.StatusBadRequest)
		return
	}

	// Redirect to the original URL
	http.Redirect(res, req, originalURL, http.StatusTemporaryRedirect)
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", handleRequest)

	err := http.ListenAndServe(`:8080`, mux)
	if err != nil {
		panic(err)
	}
}
