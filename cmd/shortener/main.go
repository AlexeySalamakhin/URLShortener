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
		logger.Log.Fatal(err.Error())
	}
	urlShortener := service.NewURLShortener(store)
	urlHandler := handler.NewURLHandler(urlShortener, config.BaseURL)
	r := urlHandler.SetupRouter()
	err = http.ListenAndServe(config.ServerAddr, r)
	if err != nil {
		panic(err)
	}
}
