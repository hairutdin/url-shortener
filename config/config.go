package config

import (
	"flag"
	"os"
)

type Config struct {
	ServerAddress string
	BaseURL       string
}

var (
	defaultServerAddress = "localhost:8080"
	defaultBaseURL       = "http://localhost:8080/"
)

var flagParsed = false

func LoadConfig() *Config {
	if !flagParsed {
		serverAddress := os.Getenv("SERVER_ADDRESS")
		baseURL := os.Getenv("BASE_URL")

		serverAddressFlag := flag.String("a", defaultServerAddress, "HTTP server address")
		baseURLFlag := flag.String("b", defaultBaseURL, "Base URL for short URLs")
		flag.Parse()
		flagParsed = true

		if serverAddress == "" {
			serverAddress = *serverAddressFlag
		}

		if baseURL == "" {
			baseURL = *baseURLFlag
		}

		return &Config{
			ServerAddress: serverAddress,
			BaseURL:       baseURL,
		}
	}

	return &Config{
		ServerAddress: defaultServerAddress,
		BaseURL:       defaultBaseURL,
	}
}
