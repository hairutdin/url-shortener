package config

import (
	"flag"
	"os"
)

type Config struct {
	ServerAddress   string
	BaseURL         string
	FileStoragePath string
}

var (
	defaultServerAddress = "localhost:8080"
	defaultBaseURL       = "http://localhost:8080/"
	defaultFileStorage   = "/tmp/short-url-db.json"
)

var flagParsed = false

func LoadConfig() *Config {
	if !flagParsed {
		serverAddress := os.Getenv("SERVER_ADDRESS")
		baseURL := os.Getenv("BASE_URL")
		fileStoragePath := os.Getenv("FILE_STORAGE_PATH")

		serverAddressFlag := flag.String("a", defaultServerAddress, "HTTP server address")
		baseURLFlag := flag.String("b", defaultBaseURL, "Base URL for short URLs")
		fileStorageFlag := flag.String("f", defaultFileStorage, "File storage path for URL data")

		flag.Parse()
		flagParsed = true

		if serverAddress == "" {
			serverAddress = *serverAddressFlag
		}

		if baseURL == "" {
			baseURL = *baseURLFlag
		}

		if fileStoragePath == "" {
			fileStoragePath = *fileStorageFlag
		}

		return &Config{
			ServerAddress:   serverAddress,
			BaseURL:         baseURL,
			FileStoragePath: fileStoragePath,
		}
	}

	return &Config{
		ServerAddress:   defaultServerAddress,
		BaseURL:         defaultBaseURL,
		FileStoragePath: defaultFileStorage,
	}
}
