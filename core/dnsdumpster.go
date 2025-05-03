package core

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
)

type DNSDumpsterResponse struct {
	A []struct {
		Host string `json:"host"`
	} `json:"a"`
}

func GetSubsFromDNSDumpster(domain, apiKey string) ([]string, error) {
	client := &http.Client{}
	apiURL := fmt.Sprintf("https://api.dnsdumpster.com/domain/%s", domain)

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("request creation failed: %v", err)
	}

	// Set headers
	req.Header.Set("X-API-Key", apiKey)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("API request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (%d): %s", resp.StatusCode, string(body))
	}

	var response DNSDumpsterResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	var (
		subdomains []string
		mu         sync.Mutex
	)

	for _, record := range response.A {
		if record.Host != "" {
			mu.Lock()
			subdomains = append(subdomains, record.Host)
			mu.Unlock()
		}
	}

	return subdomains, nil
}