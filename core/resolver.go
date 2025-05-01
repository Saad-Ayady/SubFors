package core

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"
)

func GetSubsFromCrawling(domain string) ([]string, error) {
    crawler := NewWebCrawler()
    return crawler.CrawlDomain(domain)
}

type WebCrawler struct {
	MaxDepth       int
	MaxConcurrency int
	Timeout        time.Duration
	UserAgent      string
}

func NewWebCrawler() *WebCrawler {
	return &WebCrawler{
		MaxDepth:       2,
		MaxConcurrency: 20,
		Timeout:        2 * time.Minute,
		UserAgent:      "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
	}
}

func (w *WebCrawler) CrawlDomain(domain string) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), w.Timeout)
	defer cancel()

	client := &http.Client{
		Timeout: w.Timeout,
		Transport: &http.Transport{
			MaxIdleConns:        w.MaxConcurrency,
			IdleConnTimeout:     30 * time.Second,
			DisableCompression:  true,
		},
	}

	discovered := &sync.Map{}
	subdomains := &sync.Map{}
	rootDomain := extractRootDomain(domain)

	queue := make(chan string, 1000)
	results := make(chan string, 10000)

	// Worker pool
	var wg sync.WaitGroup
	for i := 0; i < w.MaxConcurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case urlStr, ok := <-queue:
					if !ok {
						return
					}
					w.processPage(ctx, client, urlStr, rootDomain, results, queue, discovered)
				}
			}
		}()
	}

	// Result collector
	var resultsWg sync.WaitGroup
	resultsWg.Add(1)
	go func() {
		defer resultsWg.Done()
		for sub := range results {
			subdomains.Store(sub, true)
		}
	}()

	// Seed initial URLs
	go func() {
		initialURLs := []string{
			fmt.Sprintf("https://%s", domain),
			fmt.Sprintf("http://%s", domain),
			fmt.Sprintf("https://www.%s", domain),
			fmt.Sprintf("http://www.%s", domain),
		}

		for _, u := range initialURLs {
			select {
			case queue <- u:
				discovered.Store(u, true)
			case <-ctx.Done():
				return
			}
		}
	}()

	wg.Wait()
	close(results)
	resultsWg.Wait()

	return getSyncMapKeys(subdomains), nil
}

func (w *WebCrawler) processPage(ctx context.Context, client *http.Client, urlStr, rootDomain string, 
	results chan<- string, queue chan<- string, discovered *sync.Map) {

	req, err := http.NewRequestWithContext(ctx, "GET", urlStr, nil)
	if err != nil {
		return
	}
	req.Header.Set("User-Agent", w.UserAgent)

	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != 200 {
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	content := string(body)

	// Extract and send subdomains
	for _, sub := range w.extractSubdomains(content, rootDomain) {
		select {
		case results <- sub:
		case <-ctx.Done():
			return
		}
	}

	// Extract and queue links
	for _, link := range w.extractLinks(content, urlStr) {
		if _, loaded := discovered.LoadOrStore(link, true); !loaded {
			select {
			case queue <- link:
			case <-ctx.Done():
				return
			}
		}
	}
}

func (w *WebCrawler) extractSubdomains(content, rootDomain string) []string {
	pattern := `(?i)(?:https?://)?([a-zA-Z0-9\-\.]+\.` + regexp.QuoteMeta(rootDomain) + `)`
	re := regexp.MustCompile(pattern)
	matches := re.FindAllStringSubmatch(content, -1)

	unique := make(map[string]bool)
	for _, match := range matches {
		if len(match) > 1 {
			sub := strings.ToLower(match[1])
			unique[sub] = true
		}
	}

	subs := make([]string, 0, len(unique))
	for sub := range unique {
		subs = append(subs, sub)
	}
	return subs
}

func (w *WebCrawler) extractLinks(content, baseURL string) []string {
	base, _ := url.Parse(baseURL)
	var links []string

	re := regexp.MustCompile(`(href|src)="(.*?)"`)
	matches := re.FindAllStringSubmatch(content, -1)

	for _, match := range matches {
		if len(match) < 3 {
			continue
		}
		parsed, err := url.Parse(match[2])
		if err != nil {
			continue
		}
		links = append(links, base.ResolveReference(parsed).String())
	}

	return links
}

func extractRootDomain(domain string) string {
	parts := strings.Split(domain, ".")
	if len(parts) < 2 {
		return domain
	}
	return strings.Join(parts[len(parts)-2:], ".")
}

func getSyncMapKeys(m *sync.Map) []string {
	var keys []string
	m.Range(func(key, value interface{}) bool {
		keys = append(keys, key.(string))
		return true
	})
	return keys
}