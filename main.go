package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

func main() {
	fmt.Println("URL Shortener with Go and Redis")

	rdb := NewRedisClient()
	defer rdb.Close()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Welcome to the URL Shortener!")
	})

	http.HandleFunc("/shorten", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			url := r.FormValue("url")
			log.Printf("Received URL to shorten: %s", url)
			if url == "" {
				log.Printf("Empty URL received")
				http.Error(w, "URL is required", http.StatusBadRequest)
				return
			}
			shortURL, err := shortenURL(rdb, url)
			if err != nil {
				log.Printf("Error shortening URL: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			jsonResponse, err := json.Marshal(map[string]string{"short_url": shortURL})
			if err != nil {
				log.Printf("Error marshaling JSON: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			w.Write(jsonResponse)
		}
	})

	http.HandleFunc("/r/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Full path received: %s", r.URL.Path)

		shortURL := r.URL.Path[len("/r/"):]
		log.Printf("Path length: %d, /r/ length: %d", len(r.URL.Path), len("/r/"))
		log.Printf("Extracted short URL: '%s'", shortURL)

		if shortURL == "" {
			log.Printf("Short URL is empty")
			http.Error(w, "Invalid short URL", http.StatusBadRequest)
			return
		}

		url, err := redirectToURL(rdb, shortURL)
		if err != nil {
			log.Printf("Error retrieving URL: %v", err)
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}

		log.Printf("Redirecting to: %s", url)
		http.Redirect(w, r, url, http.StatusSeeOther)
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}
