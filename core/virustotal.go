package core

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

type VirusTotalResponse struct {
	Data []struct {
		ID string `json:"id"`
	} `json:"data"`
	Links struct {
		Next string `json:"next"`
	} `json:"links"`
}

func GetSubsFromVirusTotal(domain, apiKey string) ([]string, error) {
	client := &http.Client{Timeout: 15 * time.Second}
	encodedDomain := url.PathEscape(domain)
	urlStr := fmt.Sprintf("https://www.virustotal.com/api/v3/domains/%s/subdomains?limit=40", encodedDomain)

	var subdomains []string

	for {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		req, err := http.NewRequestWithContext(ctx, "GET", urlStr, nil)
		if err != nil {
			return subdomains, fmt.Errorf("request creation failed: %w", err)
		}

		req.Header.Set("x-apikey", apiKey)
		req.Header.Set("Accept", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			return subdomains, fmt.Errorf("request failed: %w", err)
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return subdomains, fmt.Errorf("failed to read response: %w", err)
		}

		if resp.StatusCode == http.StatusTooManyRequests {
			return subdomains, fmt.Errorf("API quota exceeded (captured %d subs)", len(subdomains))
		}

		if resp.StatusCode != http.StatusOK {
			return subdomains, fmt.Errorf("API error (%d): %s", resp.StatusCode, string(body))
		}

		var vtResponse VirusTotalResponse
		if err := json.Unmarshal(body, &vtResponse); err != nil {
			return subdomains, fmt.Errorf("JSON decode failed: %w", err)
		}

		for _, item := range vtResponse.Data {
			subdomains = append(subdomains, item.ID)
		}

		if vtResponse.Links.Next == "" {
			break
		}
		urlStr = vtResponse.Links.Next
	}

	return subdomains, nil
}