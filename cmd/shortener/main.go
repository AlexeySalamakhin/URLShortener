package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/AlexeySalamakhin/URLShortener/internal/config"
	"github.com/AlexeySalamakhin/URLShortener/internal/handler"
	logger "github.com/AlexeySalamakhin/URLShortener/internal/logger"
	"github.com/AlexeySalamakhin/URLShortener/internal/service"
	"github.com/AlexeySalamakhin/URLShortener/internal/store"
	"go.uber.org/zap"
	"golang.org/x/crypto/acme/autocert"
)

var (
	buildVersion, buildDate, buildCommit string
)

func main() {
	// Вывод информации о сборке
	setBuildInfoDefaults()
	printBuildInfo()

	config := config.NewConfigs()
	flag.Parse()
	logger.Initialize("info")

	store, err := store.InitStore(config)
	if err != nil {
		logger.Log.Error("Failed to initialize store: " + err.Error())
		panic(err)
	}
	defer func() {
		if err := store.Close(); err != nil {
			logger.Log.Error("Failed to close store: " + err.Error())
		}
	}()

	urlShortener := service.NewURLShortener(store)
	urlHandler := handler.NewURLHandler(urlShortener, config.BaseURL, config.TrustedSubnet)
	r := urlHandler.SetupRouter()

	server := &http.Server{
		Addr:    config.ServerAddr,
		Handler: r,
	}

	// Канал для сигналов завершения
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	// Канал для ошибок сервера
	errCh := make(chan error, 1)

	go func() {
		if config.EnableHTTPS {
			manager := &autocert.Manager{
				Cache:  autocert.DirCache(".autocert-cache"),
				Prompt: autocert.AcceptTOS,
			}
			server.TLSConfig = manager.TLSConfig()
			logger.Log.Info("Запуск HTTPS-сервера с autocert...", zap.String("addr", config.ServerAddr))
			errCh <- server.ListenAndServeTLS("", "")
		} else {
			logger.Log.Info("Запуск HTTP-сервера...", zap.String("addr", config.ServerAddr))
			errCh <- server.ListenAndServe()
		}
	}()

	select {
	case sig := <-sigCh:
		logger.Log.Info("Получен сигнал завершения", zap.String("signal", sig.String()))
	case err := <-errCh:
		if err != nil && err != http.ErrServerClosed {
			logger.Log.Error("Ошибка сервера", zap.Error(err))
			os.Exit(1)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		logger.Log.Error("Ошибка при завершении сервера", zap.Error(err))
	} else {
		logger.Log.Info("Сервер завершён корректно")
	}
}

func setBuildInfoDefaults() {
	if buildVersion == "" {
		buildVersion = "N/A"
	}
	if buildDate == "" {
		buildDate = "N/A"
	}
	if buildCommit == "" {
		buildCommit = "N/A"
	}
}

func printBuildInfo() {
	fmt.Println("Build version:", buildVersion)
	fmt.Println("Build date:", buildDate)
	fmt.Println("Build commit:", buildCommit)
}
