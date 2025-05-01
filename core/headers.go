package core

import (
	"net/http"
	"regexp"
)

// Extract subdomains from HTTP Headers
func GetSubsFromHeaders(domain string) ([]string, error) {
	urls := []string{
		"http://" + domain,
		"https://" + domain,
	}

	unique := make(map[string]bool)
	var results []string

	// Iterate over different protocol versions (HTTP, HTTPS)
	for _, url := range urls {
		resp, err := http.Get(url)
		if err != nil {
			continue
		}
		defer resp.Body.Close()

		// Extract headers
		headers := resp.Header

		// Iterate over headers and extract subdomains
		for _, values := range headers {
			for _, value := range values {
				// Extract subdomains from URL-like strings in headers
				re := regexp.MustCompile(`(?:https?:\/\/)?([a-zA-Z0-9_\-\.]+\.` + regexp.QuoteMeta(domain) + `)`)
				matches := re.FindAllStringSubmatch(value, -1)

				// Add unique subdomains to the result
				for _, match := range matches {
					sub := match[1]
					if !unique[sub] {
						unique[sub] = true
						results = append(results, sub)
					}
				}
			}
		}
	}

	return results, nil
}
