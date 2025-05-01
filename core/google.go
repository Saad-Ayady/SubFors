package core

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

// DuckDork extracts subdomains from DuckDuckGo
func DuckDork(domain string) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	searchURL := fmt.Sprintf("https://duckduckgo.com/html/?q=site:*.%s+-site:www.%s", domain, domain)

	req, err := http.NewRequestWithContext(ctx, "GET", searchURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	req.Header.Set("Accept", "text/html")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	return extractSubdomains(string(body), domain), nil
}

// BingDork extracts subdomains from Bing
func BingDork(domain string) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	searchURL := fmt.Sprintf("https://www.bing.com/search?q=site:*.%s+-site:www.%s&count=50", domain, domain)

	req, err := http.NewRequestWithContext(ctx, "GET", searchURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	req.Header.Set("Accept", "text/html")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	return extractSubdomains(string(body), domain), nil
}

// GoogleDork extracts subdomains from Google without API
func GoogleDork(domain string) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	searchURL := fmt.Sprintf("https://www.google.com/search?q=site:*.%s+-site:www.%s&num=100", domain, domain)

	req, err := http.NewRequestWithContext(ctx, "GET", searchURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	// Use realistic headers to avoid blocking
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Referer", "https://www.google.com/")

	// Use cookies from a real browser session if possible
	req.AddCookie(&http.Cookie{Name: "CONSENT", Value: "YES+cb.20210720-07-p0.en+FX+410"})

	client := &http.Client{
		Timeout: 45 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// Don't follow redirects to captcha pages
			if strings.Contains(req.URL.String(), "sorry/index") {
				return fmt.Errorf("google captcha triggered")
			}
			return nil
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	// Check for captcha
	if strings.Contains(resp.Request.URL.String(), "sorry/index") {
		return nil, fmt.Errorf("google captcha triggered")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	content := string(body)
	if strings.Contains(content, "Our systems have detected unusual traffic") {
		return nil, fmt.Errorf("google detected automated requests")
	}

	return extractSubdomains(content, domain), nil
}

// extractSubdomains extracts unique subdomains from content
func extractSubdomains(content, domain string) []string {
	// Improved regex to better match subdomains in search results
	pattern := `(?:https?:\/\/)?([a-zA-Z0-9\-\.]+\.` + regexp.QuoteMeta(domain) + `)(?:\/|")`
	re := regexp.MustCompile(pattern)
	matches := re.FindAllStringSubmatch(content, -1)

	unique := make(map[string]bool)
	var results []string

	for _, match := range matches {
		if len(match) < 2 {
			continue
		}
		sub := strings.ToLower(match[1])
		sub = strings.TrimPrefix(sub, "www.")
		
		// Validate the subdomain actually belongs to our domain
		if strings.HasSuffix(sub, "."+domain) && !unique[sub] {
			unique[sub] = true
			results = append(results, sub)
		}
	}

	return results
}