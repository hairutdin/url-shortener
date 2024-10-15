package config

import (
	"flag"
	"os"
)

type Config struct {
	ServerAddress   string
	BaseURL         string
	FileStoragePath string
	DatabaseDSN     string
}

var (
	defaultServerAddress = "localhost:8080"
	defaultBaseURL       = "http://localhost:8080/"
	defaultFileStorage   = "/tmp/short-url-db.json"
	defaultDatabaseDSN   = "postgres://postgres:berlin@localhost:5432/testdb?sslmode=disable"
)

var flagParsed = false

func LoadConfig() *Config {
	if !flagParsed {
		serverAddress := os.Getenv("SERVER_ADDRESS")
		baseURL := os.Getenv("BASE_URL")
		fileStoragePath := os.Getenv("FILE_STORAGE_PATH")
		databaseDSN := os.Getenv("DATABASE_DSN")

		serverAddressFlag := flag.String("a", defaultServerAddress, "HTTP server address")
		baseURLFlag := flag.String("b", defaultBaseURL, "Base URL for short URLs")
		fileStorageFlag := flag.String("f", defaultFileStorage, "File storage path for URL data")
		databaseDSNFlag := flag.String("d", defaultDatabaseDSN, "Database DSN")

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

		if databaseDSN == "" {
			databaseDSN = *databaseDSNFlag
		}

		return &Config{
			ServerAddress:   serverAddress,
			BaseURL:         baseURL,
			FileStoragePath: fileStoragePath,
			DatabaseDSN:     databaseDSN,
		}
	}

	return &Config{
		ServerAddress:   defaultServerAddress,
		BaseURL:         defaultBaseURL,
		FileStoragePath: defaultFileStorage,
		DatabaseDSN:     defaultDatabaseDSN,
	}
}
