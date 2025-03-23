package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/AlexeySalamakhin/URLShortener/internal/store"
	"github.com/AlexeySalamakhin/URLShortener/internal/utils"
)

const serverURL = `localhost:8080`

var (
	db = store.NewInMemoryStore()
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc(`/`, HandleShorten)
	err := http.ListenAndServe(serverURL, mux)
	if err != nil {
		panic(err)
	}
}

func HandleShorten(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		PostURLHandler(w, r, db)
	} else if r.Method == http.MethodGet {
		GetURLHandler(r, w, db)
	} else {
		w.WriteHeader(http.StatusBadRequest)
	}
}

func PostURLHandler(w http.ResponseWriter, r *http.Request, store store.URLStore) {
	w.Header().Set("Content-Type", "text/plain")
	originalURL, err := io.ReadAll(r.Body)
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}
	if len(originalURL) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	shortKey := utils.GenerateShortUrl()
	store.Save(string(originalURL), shortKey)
	w.WriteHeader(201)
	w.Write(fmt.Appendf(nil, "%s/%s", serverURL, shortKey))

}

func GetURLHandler(r *http.Request, w http.ResponseWriter, store store.URLStore) {
	shortUrl := r.URL.Path[1:]
	found, originalURL := store.Get(shortUrl)
	if !found {
		http.Error(w, "Not found", http.StatusBadRequest)
	}
	http.Redirect(w, r, originalURL, http.StatusTemporaryRedirect)
}
