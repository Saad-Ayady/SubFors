package core

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// WaybackResponse represents the JSON response structure from Wayback Machine
type WaybackResponse [][]string

// GetSubsFromArchive retrieves subdomains from Wayback Machine archives
func GetSubsFromArchive(domain string) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	// Construct the API URL with parameters
	apiURL := fmt.Sprintf("http://web.archive.org/cdx/search/cdx?url=*.%s/*&output=json&limit=30000", domain)

	// Configure HTTP client with custom settings
	client := &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:        10,
			IdleConnTimeout:     30 * time.Second,
			DisableCompression:  false,
		},
		Timeout: 120 * time.Second,
	}

	// Create request with context
	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("request creation failed: %v", err)
	}

	// Set realistic headers to mimic browser behavior
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	req.Header.Set("Accept", "application/json")

	// Execute request with retry logic
	var resp *http.Response
	for attempt := 0; attempt < 3; attempt++ {
		resp, err = client.Do(req)
		if err == nil {
			break
		}
		time.Sleep(time.Duration(attempt+1) * time.Second) // Exponential backoff
	}
	if err != nil {
		return nil, fmt.Errorf("request failed after 3 attempts: %v", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Read and parse response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	var data WaybackResponse
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, fmt.Errorf("JSON parsing failed: %v", err)
	}

	// Process results concurrently
	seen := make(map[string]struct{})
	var results []string
	var mu sync.Mutex // For thread-safe map operations

	// Skip header row if present
	startIndex := 0
	if len(data) > 0 && len(data[0]) > 0 && data[0][0] == "urlkey" {
		startIndex = 1
	}

	var wg sync.WaitGroup
	semaphore := make(chan struct{}, 10) // Limit concurrent goroutines

	for _, row := range data[startIndex:] {
		if len(row) == 0 {
			continue
		}

		wg.Add(1)
		semaphore <- struct{}{}

		go func(rawURL string) {
			defer wg.Done()
			defer func() { <-semaphore }()

			processedURL := processURL(rawURL, domain)
			if processedURL != "" {
				mu.Lock()
				if _, exists := seen[processedURL]; !exists {
					seen[processedURL] = struct{}{}
					results = append(results, processedURL)
					fmt.Printf("[Archive] Found: %s\n", processedURL)
				}
				mu.Unlock()
			}
		}(row[0])
	}

	wg.Wait()
	return results, nil
}

// processURL cleans and validates URLs from archive
func processURL(rawURL, domain string) string {
	// Remove protocol prefixes
	rawURL = strings.TrimPrefix(rawURL, "http://")
	rawURL = strings.TrimPrefix(rawURL, "https://")

	// Extract hostname
	host := strings.Split(rawURL, "/")[0]

	// Validate domain ownership
	if !strings.HasSuffix(host, domain) {
		return ""
	}

	// Handle URL encoded characters
	decodedHost, err := url.QueryUnescape(host)
	if err != nil {
		return host // Return original if decoding fails
	}

	return decodedHost
}
