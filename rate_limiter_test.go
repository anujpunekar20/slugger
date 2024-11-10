package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
)

func TestRateLimiting(t *testing.T) {
	// Setup server
	rdb := NewRedisClient()
	defer rdb.Close()

	// Create test server
	handler := setupHandler(rdb)
	server := httptest.NewServer(handler)
	defer server.Close()

	// Test data
	payload := map[string]string{
		"url": "https://serverfault.com/questions/281979/how-to-save-close-file-when-editing-in-bash",
	}
	jsonData, _ := json.Marshal(payload)

	// Send requests rapidly
	for i := 0; i < 150; i++ {
		resp, err := http.Post(
			server.URL+"/shorten",
			"application/json",
			bytes.NewBuffer(jsonData),
		)
		if err != nil {
			t.Fatalf("Failed to send request: %v", err)
		}

		if i < 100 {
			if resp.StatusCode != http.StatusOK {
				t.Errorf("Expected OK status for request %d, got %d", i, resp.StatusCode)
			}
		} else {
			if resp.StatusCode != http.StatusTooManyRequests {
				t.Errorf("Expected rate limit status for request %d, got %d", i, resp.StatusCode)
			}
		}
		resp.Body.Close()
	}
}

func TestRateLimitReset(t *testing.T) {
	// Setup server
	rdb := NewRedisClient()
	defer rdb.Close()

	handler := setupHandler(rdb)
	server := httptest.NewServer(handler)
	defer server.Close()

	payload := map[string]string{
		"url": "https://serverfault.com/questions/281979/how-to-save-close-file-when-editing-in-bash",
	}
	jsonData, _ := json.Marshal(payload)

	// Send 90 requests
	for i := 0; i < 90; i++ {
		resp, _ := http.Post(
			server.URL+"/shorten",
			"application/json",
			bytes.NewBuffer(jsonData),
		)
		resp.Body.Close()
	}

	// Wait for rate limit window to reset (use small window for testing)
	time.Sleep(time.Second * 2)

	// Should be able to send requests again
	resp, err := http.Post(
		server.URL+"/shorten",
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Expected Created status after reset, got %d", resp.StatusCode)
	}
}

func setupHandler(rdb *redis.Client) http.Handler {
	mux := http.NewServeMux()

	// Add CORS headers to prevent redirect to YouTube
	addCorsHeaders := func(w http.ResponseWriter) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	}

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		addCorsHeaders(w)
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintf(w, "Welcome to the URL Shortener!")
	})

	mux.HandleFunc("/shorten", func(w http.ResponseWriter, r *http.Request) {
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
			// Read and parse the JSON body
			var payload map[string]string
			decoder := json.NewDecoder(r.Body)
			err := decoder.Decode(&payload)
			if err != nil {
				http.Error(w, "Invalid JSON", http.StatusBadRequest)
				return
			}

			url, ok := payload["url"]
			if !ok || url == "" {
				http.Error(w, "URL is required", http.StatusBadRequest)
				return
			}

			shortURL, err := shortenURL(rdb, url)
			if err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			jsonResponse, err := json.Marshal(map[string]string{"short_url": shortURL})
			if err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			w.Write(jsonResponse)
		}
	})

	mux.HandleFunc("/r/", func(w http.ResponseWriter, r *http.Request) {
		shortURL := r.URL.Path[len("/r/"):]
		if shortURL == "" {
			http.Error(w, "Invalid short URL", http.StatusBadRequest)
			return
		}

		url, err := redirectToURL(rdb, shortURL)
		if err != nil {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}

		http.Redirect(w, r, url, http.StatusSeeOther)
	})

	return mux
}
