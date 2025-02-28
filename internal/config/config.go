package config

import (
	"flag"
	"os"
	"time"
)

type Config struct {
	Env             string           `env:"ENVIRONMENT" envDefault:"development"`
	HTTP            HTTPServerConfig `envPrefix:"HTTP_"`
	BaseURL         string
	FileStoragePath string
	DatabaseDSN     string
	StorageType     string
}

type HTTPServerConfig struct {
	Address       string        `env:"HTTP_SERVER_ADDRESS" envDefault:"0.0.0.0:8080"`
	ReadTimeout   time.Duration `env:"HTTP_READ_TIMEOUT" envDefault:"10s"`
	WriteTimeout  time.Duration `env:"HTTP_WRITE_TIMEOUT" envDefault:"10s"`
	IdleTimeout   time.Duration `env:"HTTP_IDLE_TIMEOUT" envDefault:"15s"`
	HeaderTimeout time.Duration `env:"HTTP_HEADER_TIMEOUT" envDefault:"5s"`
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

		storageType := "memory"
		if databaseDSN != "" {
			storageType = "postgres"
		} else if fileStoragePath != "" {
			storageType = "file"
		}

		return &Config{
			HTTP: HTTPServerConfig{
				Address:       serverAddress,
				ReadTimeout:   10 * time.Second,
				WriteTimeout:  10 * time.Second,
				IdleTimeout:   15 * time.Second,
				HeaderTimeout: 5 * time.Second,
			},
			BaseURL:         baseURL,
			FileStoragePath: fileStoragePath,
			DatabaseDSN:     databaseDSN,
			StorageType:     storageType,
		}
	}

	return &Config{
		HTTP: HTTPServerConfig{
			Address:       defaultServerAddress,
			ReadTimeout:   10 * time.Second,
			WriteTimeout:  10 * time.Second,
			IdleTimeout:   15 * time.Second,
			HeaderTimeout: 5 * time.Second,
		},
		BaseURL:         defaultBaseURL,
		FileStoragePath: defaultFileStorage,
		DatabaseDSN:     defaultDatabaseDSN,
		StorageType:     "memory",
	}
}
