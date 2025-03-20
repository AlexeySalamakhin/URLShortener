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

func HandleShorten(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		w.Header().Set("Content-Type", "text/plain")
		originalURL, err := io.ReadAll(r.Body)
		if err != nil {
			w.Write([]byte(err.Error()))
			return
		}
		shortKey := getShortURL()

		db[shortKey] = string(originalURL)
		w.WriteHeader(201)
		w.Write([]byte(fmt.Sprintf("http://localhost:8080/%s", shortKey)))
	} else if r.Method == http.MethodGet {
		id := r.URL.Path[1:]
		originalURL, found := db[id]
		if !found {
			http.Error(w, "Not found", http.StatusBadRequest)
		}
		http.Redirect(w, r, originalURL, http.StatusTemporaryRedirect)
	}

	w.WriteHeader(http.StatusBadRequest)

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
