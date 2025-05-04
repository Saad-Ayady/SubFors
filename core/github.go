package core

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

func GetSubsFromGitHub(domain, token string) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	uniqueSubs := make(map[string]struct{})
	subdomainRegex := regexp.MustCompile(fmt.Sprintf(`(?i)(?:[a-z0-9-]+\.)*%s(?:$|/|\s|"|')`, regexp.QuoteMeta(domain)))

	client := &http.Client{
		Timeout: 15 * time.Second,
	}

	page := 1
	maxPages := 10

	for page <= maxPages {
		select {
		case <-ctx.Done():
			return mapToSlice(uniqueSubs), nil
		default:
		}

		req, err := http.NewRequest("GET", 
			fmt.Sprintf("https://api.github.com/search/code?q=%s+in:file&page=%d&per_page=100", 
				url.QueryEscape(domain), page), 
			nil)
		if err != nil {
			return nil, fmt.Errorf("request creation failed: %v", err)
		}

		req.Header.Set("Authorization", "token "+token)
		req.Header.Set("Accept", "application/vnd.github.v3.text-match+json")

		resp, err := client.Do(req)
		if err != nil {
			return mapToSlice(uniqueSubs), err
		}

		if resp.StatusCode == http.StatusForbidden {
			resp.Body.Close()
			return mapToSlice(uniqueSubs), fmt.Errorf("GitHub API rate limit exceeded")
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return mapToSlice(uniqueSubs), fmt.Errorf("GitHub API returned status: %s", resp.Status)
		}

		var result struct {
			Items []struct {
				TextMatches []struct {
					Fragment string `json:"fragment"`
				} `json:"text_matches"`
			} `json:"items"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			resp.Body.Close()
			return mapToSlice(uniqueSubs), err
		}
		resp.Body.Close()

		if len(result.Items) == 0 {
			break
		}

		for _, item := range result.Items {
			for _, match := range item.TextMatches {
				matches := subdomainRegex.FindAllString(match.Fragment, -1)
				for _, m := range matches {
					clean := strings.TrimRight(strings.TrimSpace(m), `/"'`)
					if _, exists := uniqueSubs[clean]; !exists {
						uniqueSubs[clean] = struct{}{}
					}
				}
			}
		}

		page++
		time.Sleep(3 * time.Second)
	}

	return mapToSlice(uniqueSubs), nil
}

func mapToSlice(m map[string]struct{}) []string {
	slice := make([]string, 0, len(m))
	for k := range m {
		slice = append(slice, k)
	}
	return slice
}