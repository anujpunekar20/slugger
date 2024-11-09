package main

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
    "time"
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
        "url": "https://example.com",
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
    if resp.StatusCode != http.StatusOK {
        t.Errorf("Expected OK status after reset, got %d", resp.StatusCode)
    }
}