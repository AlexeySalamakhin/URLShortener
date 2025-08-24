package main

import (
	"flag"
	"net/http"

	"github.com/AlexeySalamakhin/URLShortener/internal/config"
	"github.com/AlexeySalamakhin/URLShortener/internal/handler"
	logger "github.com/AlexeySalamakhin/URLShortener/internal/logger"
	"github.com/AlexeySalamakhin/URLShortener/internal/service"
	"github.com/AlexeySalamakhin/URLShortener/internal/store"
)

func main() {
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
