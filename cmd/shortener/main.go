package main

import (
	"net/http"

	"github.com/AlexeySalamakhin/URLShortener/internal/handler"
	"github.com/AlexeySalamakhin/URLShortener/internal/service"
	"github.com/AlexeySalamakhin/URLShortener/internal/store"
)

func main() {
	db := store.NewInMemoryStore()
	urlShortener := service.NewURLShortener(db)
	urlHandler := handler.NewURLHandler(urlShortener)
	r := urlHandler.SetupRouter()
	err := http.ListenAndServe("localhost:8080", r)
	if err != nil {
		panic(err)
	}
}
