package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/hairutdin/url-shortener/internal/app/http/handlers"
	"github.com/hairutdin/url-shortener/internal/box"
	"github.com/hairutdin/url-shortener/internal/service"
	"go.uber.org/zap"
)

// @title			URL Shortener Service
// @version		1.0
// @description	A service for shortening URLs
// @in				header
func main() {
	envBox, err := box.New()
	if err != nil {
		envBox.Logger.Fatal("unable to initialize box", zap.Error(err))
	}

	_ = zap.ReplaceGlobals(envBox.Logger)
	defer func(l *zap.Logger) {
		err := l.Sync()
		if err != nil {
			log.Printf("can't type zap logs: %s", err)
		}
	}(envBox.Logger)

	urlService := service.NewURLService(envBox.Storage, envBox.Logger, envBox.Config.BaseURL)
	baseHandler := handlers.NewBaseHandler(urlService, envBox.Logger, envBox.Config)
	httpHandlers := handlers.SetupRouter(envBox.Config, envBox.Logger, baseHandler)

	envBox.Logger.Info("starting server", zap.String("address", envBox.Config.HTTP.Address))

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	srv := &http.Server{
		Handler:           httpHandlers,
		Addr:              envBox.Config.HTTP.Address,
		ReadTimeout:       envBox.Config.HTTP.ReadTimeout,
		WriteTimeout:      envBox.Config.HTTP.WriteTimeout,
		IdleTimeout:       envBox.Config.HTTP.IdleTimeout,
		ReadHeaderTimeout: envBox.Config.HTTP.HeaderTimeout,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			envBox.Logger.Fatal("failed to start server", zap.Error(err))
		}
	}()

	envBox.Logger.Info("server started")

	<-ctx.Done()
	envBox.Logger.Info("stopping server")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		envBox.Logger.Fatal("failed to shutdown server", zap.Error(err))
	}

	envBox.Logger.Info("server stopped")
}
