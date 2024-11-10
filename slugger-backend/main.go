package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	fmt.Println("URL Shortener with Go and Redis")

	rdb := NewRedisClient()
	defer rdb.Close()

	// Add CORS headers to prevent redirect to YouTube
	addCorsHeaders := func(w http.ResponseWriter) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		addCorsHeaders(w)
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintf(w, "Welcome to the URL Shortener!")
	})

	http.HandleFunc("/shorten", func(w http.ResponseWriter, r *http.Request) {
		addCorsHeaders(w)
		
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		
		// Rate limiting: 100 requests per hour per IP
		rateLimiter := NewRateLimiter(rdb)
		ip := r.RemoteAddr
		allowed, err := rateLimiter.IsAllowed(ip, 100, time.Hour)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		if !allowed {
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}
		
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

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default to 8080 if not set
	}
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
