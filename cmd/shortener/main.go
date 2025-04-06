package main

import (
	"net/http"

	"github.com/AlexeySalamakhin/URLShortener/internal/config"
	"github.com/AlexeySalamakhin/URLShortener/internal/handler"
	"github.com/AlexeySalamakhin/URLShortener/internal/service"
	"github.com/AlexeySalamakhin/URLShortener/internal/store"
)

func main() {
	config := config.NewConfigs()
	db := store.NewInMemoryStore()
	urlShortener := service.NewURLShortener(db)
	urlHandler := handler.NewURLHandler(urlShortener, config.BaseURL)
	r := urlHandler.SetupRouter()
	err := http.ListenAndServe(config.ServerAddr, r)
	if err != nil {
		panic(err)
	}
}
