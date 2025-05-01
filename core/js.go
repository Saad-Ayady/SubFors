package core

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

// JSSubdomainExtractor configuration
type JSSubdomainExtractor struct {
	Domain string
	Client *http.Client
}

func NewJSSubdomainExtractor(domain string) *JSSubdomainExtractor {
	return &JSSubdomainExtractor{
		Domain: domain,
		Client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetSubsFromJS analyzes JavaScript files and extracts subdomains
func GetSubsFromJS(domain string) ([]string, error) {
	extractor := NewJSSubdomainExtractor(domain)
	return extractor.AnalyzeDomainJS()
}

func (j *JSSubdomainExtractor) AnalyzeDomainJS() ([]string, error) {
	// First find JS files from main domain
	jsFiles, err := j.findJSFiles()
	if err != nil {
		return nil, err
	}

	// Analyze each JS file
	uniqueSubs := make(map[string]struct{})
	for _, jsFile := range jsFiles {
		subs, err := j.analyzeJSFile(jsFile)
		if err != nil {
			continue
		}
		
		for _, sub := range subs {
			uniqueSubs[sub] = struct{}{}
		}
	}

	// Convert map to slice
	result := make([]string, 0, len(uniqueSubs))
	for sub := range uniqueSubs {
		result = append(result, sub)
	}

	return result, nil
}

func (j *JSSubdomainExtractor) findJSFiles() ([]string, error) {
	// Implement logic to find JS files from page sources
	// This could be done by crawling the website first
	// For simplicity, we'll return a direct JS URL
	return []string{
		fmt.Sprintf("https://www.%s/main.js", j.Domain),
		fmt.Sprintf("https://%s/static/app.js", j.Domain),
	}, nil
}

func (j *JSSubdomainExtractor) analyzeJSFile(url string) ([]string, error) {
	resp, err := j.Client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return j.extractSubdomainsFromJS(string(content)), nil
}

func (j *JSSubdomainExtractor) extractSubdomainsFromJS(content string) []string {
	patterns := []string{
		// Match URLs containing subdomains
		`(?i)(?:https?:\\/\\/|['"])?([a-zA-Z0-9-]+\.%s)`, 
		// Match subdomain variables
		`(?i)subdomain['"]?\\s*[:=]\\s*['"]([a-zA-Z0-9-]+)`,
		// Match API endpoints
		`(?i)apiUrl['"]?\\s*[:=]\\s*['"](?:https?:\\/\\/)?([a-zA-Z0-9-]+\.%s)`,
		// Match any string containing .domain.com
		`(?i)([a-zA-Z0-9-]+\.%s)`,
	}

	rootDomain := strings.Replace(j.Domain, ".", "\\.", -1)
	uniqueSubs := make(map[string]struct{})

	for _, pattern := range patterns {
		re := regexp.MustCompile(fmt.Sprintf(pattern, rootDomain))
		matches := re.FindAllStringSubmatch(content, -1)
		
		for _, match := range matches {
			if len(match) > 1 {
				sub := strings.ToLower(strings.Trim(match[1], "'\""))
				if strings.HasSuffix(sub, j.Domain) {
					uniqueSubs[sub] = struct{}{}
				}
			}
		}
	}

	// Filter and format results
	result := make([]string, 0, len(uniqueSubs))
	for sub := range uniqueSubs {
		// Remove protocol if present
		cleanSub := strings.Split(sub, "//")[0]
		result = append(result, cleanSub)
	}

	return result
}