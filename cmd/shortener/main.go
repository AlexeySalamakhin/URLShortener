package main

import (
	"net/http"

	"github.com/AlexeySalamakhin/URLShortener/internal/handler"
	"github.com/AlexeySalamakhin/URLShortener/internal/service"
	"github.com/AlexeySalamakhin/URLShortener/internal/store"
)

func main() {
	mux := http.NewServeMux()
	db := store.NewInMemoryStore()
	urlShortener := service.NewURLShortener(db)
	urlHandler := handler.NewURLHandler(urlShortener)
	mux.HandleFunc(`/`, urlHandler.HandleShorten)
	err := http.ListenAndServe("localhost:8080", mux)
	if err != nil {
		panic(err)
	}
}
