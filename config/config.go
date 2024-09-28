package config

import (
	"flag"
)

type Config struct {
	ServerAddress string
	BaseURL       string
}

var (
	serverAddressFlag = "localhost:8080"
	baseURLFlag       = "http://localhost:8080/"
)

var flagParsed = false

func LoadConfig() *Config {
	if !flagParsed {
		serverAddress := flag.String("a", serverAddressFlag, "HTTP server address")
		baseURL := flag.String("b", baseURLFlag, "Base URL for short URLs")
		flag.Parse()
		flagParsed = true

		return &Config{
			ServerAddress: *serverAddress,
			BaseURL:       *baseURL,
		}
	}

	return &Config{
		ServerAddress: serverAddressFlag,
		BaseURL:       baseURLFlag,
	}
}
