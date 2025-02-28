package box

import (
	"fmt"
	"sync"

	"github.com/hairutdin/url-shortener/internal/config"
	"github.com/hairutdin/url-shortener/internal/repository"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

const (
	envDev = "develop"
)

var (
	instance *Env
	once     sync.Once
)

type Env struct {
	Config  *config.Config
	Logger  *zap.Logger
	Storage repository.Storage
}

func New() (*Env, error) {
	var err error
	once.Do(func() {
		_ = godotenv.Load("config/.env")

		cfg := config.LoadConfig()

		logger, loggerErr := SetupLogger(cfg.Env)
		if loggerErr != nil {
			err = fmt.Errorf("failed to setup logger: %w", loggerErr)
			return
		}

		storage, storageErr := initializeStorage(cfg)
		if storageErr != nil {
			err = fmt.Errorf("failed to initialize storage: %w", storageErr)
			return
		}

		instance = &Env{
			Config:  cfg,
			Logger:  logger,
			Storage: storage,
		}
	})

	if instance == nil {
		return nil, err
	}
	return instance, nil
}

func SetupLogger(env string) (*zap.Logger, error) {
	if env == envDev {
		return zap.NewDevelopment()
	}
	return zap.NewProduction()
}

func initializeStorage(cfg *config.Config) (repository.Storage, error) {
	switch cfg.StorageType {
	case "postgres":
		return repository.NewPostgresStorage(cfg.DatabaseDSN)
	case "file":
		return repository.NewFileStorage(cfg.FileStoragePath)
	default:
		return repository.NewInMemoryStorage(), nil
	}
}
