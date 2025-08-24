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
)

var buildVersion string
var buildDate string
var buildCommit string

func main() {
	// Вывод информации о сборке
	if buildVersion == "" {
		buildVersion = "N/A"
	}
	if buildDate == "" {
		buildDate = "N/A"
	}
	if buildCommit == "" {
		buildCommit = "N/A"
	}
	fmt.Println("Build version:", buildVersion)
	fmt.Println("Build date:", buildDate)
	fmt.Println("Build commit:", buildCommit)

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

	if err := http.ListenAndServe(config.ServerAddr, r); err != nil {
		panic(err)
	}
}
