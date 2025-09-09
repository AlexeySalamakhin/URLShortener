package main

import (
	"flag"
	"fmt"
	"net/http"

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
	urlHandler := handler.NewURLHandler(urlShortener, config.BaseURL)
	r := urlHandler.SetupRouter()

	if config.EnableHTTPS {
		manager := &autocert.Manager{
			Cache:  autocert.DirCache(".autocert-cache"),
			Prompt: autocert.AcceptTOS,
		}
		server := &http.Server{
			Addr:      config.ServerAddr,
			Handler:   r,
			TLSConfig: manager.TLSConfig(),
		}
		logger.Log.Info("Запуск HTTPS-сервера с autocert...", zap.String("addr", config.ServerAddr))
		if err := server.ListenAndServeTLS("", ""); err != nil {
			panic(err)
		}
	} else {
		logger.Log.Info("Запуск HTTP-сервера...", zap.String("addr", config.ServerAddr))
		if err := http.ListenAndServe(config.ServerAddr, r); err != nil {
			panic(err)
		}
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
