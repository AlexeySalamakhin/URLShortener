package main

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
)

const serverURL = `localhost:8080`

var (
	db = make(map[string]string)
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc(`/`, HandleShorten)
	err := http.ListenAndServe(serverURL, mux)
	if err != nil {
		panic(err)
	}
}

func PostURLHandler(w http.ResponseWriter, r *http.Request) {
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
	shortKey := getShortURL()

	db[shortKey] = string(originalURL)
	w.WriteHeader(201)
	w.Write(fmt.Appendf(nil, "http://localhost:8080/%s", shortKey))

}
func HandleShorten(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		PostURLHandler(w, r)
	} else if r.Method == http.MethodGet {
		GetURLHandler(r, w, db)
	} else {
		w.WriteHeader(http.StatusBadRequest)
	}
}

func GetURLHandler(r *http.Request, w http.ResponseWriter, urlMapping map[string]string) {
	id := r.URL.Path[1:]
	originalURL, found := urlMapping[id]
	if !found {
		http.Error(w, "Not found", http.StatusBadRequest)
	}
	http.Redirect(w, r, originalURL, http.StatusTemporaryRedirect)
}

func getShortURL() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const keyLength = 6

	shortKey := make([]byte, keyLength)
	for i := range shortKey {
		shortKey[i] = charset[rand.Intn(len(charset))]
	}
	return string(shortKey)
}
