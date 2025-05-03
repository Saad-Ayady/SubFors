package core

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/spaolacci/murmur3"
)

const (
	faviconConcurrency = 50
	faviconTimeout     = 5 * time.Second
)

var (
	faviconPaths = []string{
		"/favicon.ico",
		"/favicon.png",
		"/static/favicon.ico",
		"/img/favicon.ico",
		"/assets/favicon.ico",
	}

	commonSubdomains = []string{
		"www", "mail", "dev", "test", "api", "blog",
		"admin", "portal", "static", "assets", "cdn",
		"login", "auth", "payment", "checkout",
	}
)

func GetSubsFromFavicon(domain string) ([]string, error) {
	targetHash, err := getTargetFaviconHash(domain)
	if err != nil {
		return nil, fmt.Errorf("favicon error: %v", err)
	}

	subdomains := generateSubdomains(domain)
	return findMatchingSubdomains(subdomains, targetHash), nil
}

func getTargetFaviconHash(domain string) (uint32, error) {
	client := createHTTPClient()
	
	for _, path := range faviconPaths {
		for _, scheme := range []string{"https", "http"} {
			url := fmt.Sprintf("%s://%s%s", scheme, domain, path)
			data, err := fetchFavicon(client, url)
			if err == nil && data != nil {
				return murmur3.Sum32WithSeed(data, 0), nil
			}
		}
	}
	return 0, fmt.Errorf("no favicon found")
}

func generateSubdomains(domain string) []string {
	var generated []string
	
	// Basic subdomains
	for _, sub := range commonSubdomains {
		generated = append(generated, fmt.Sprintf("%s.%s", sub, domain))
	}
	
	// Cloud patterns
	cloudPrefixes := []string{"aws", "azure", "gcp", "lb", "alb"}
	for _, prefix := range cloudPrefixes {
		generated = append(generated,
			fmt.Sprintf("%s-%s", prefix, domain),
			fmt.Sprintf("%s.%s", prefix, domain),
		)
	}
	
	return generated
}

func findMatchingSubdomains(subdomains []string, targetHash uint32) []string {
	var (
		wg       sync.WaitGroup
		results  []string
		mu       sync.Mutex
		jobs     = make(chan string, len(subdomains))
	)

	client := createHTTPClient()
	ctx, cancel := context.WithTimeout(context.Background(), faviconTimeout)
	defer cancel()

	// Worker pool
	for i := 0; i < faviconConcurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for sub := range jobs {
				if verifyFaviconMatch(ctx, client, sub, targetHash) {
					mu.Lock()
					results = append(results, sub)
					mu.Unlock()
				}
			}
		}()
	}

	// Feed jobs
	go func() {
		for _, sub := range subdomains {
			if exists, _ := dnsLookup(sub); exists {
				jobs <- sub
			}
		}
		close(jobs)
	}()

	wg.Wait()
	return results
}

func verifyFaviconMatch(ctx context.Context, client *http.Client, sub string, targetHash uint32) bool {
	for _, path := range faviconPaths {
		for _, scheme := range []string{"https", "http"} {
			select {
			case <-ctx.Done():
				return false
			default:
				url := fmt.Sprintf("%s://%s%s", scheme, sub, path)
				data, err := fetchFavicon(client, url)
				if err == nil && murmur3.Sum32WithSeed(data, 0) == targetHash {
					return true
				}
			}
		}
	}
	return false
}

func createHTTPClient() *http.Client {
	return &http.Client{
		Timeout: faviconTimeout,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			IdleConnTimeout: faviconTimeout,
		},
	}
}

func fetchFavicon(client *http.Client, url string) ([]byte, error) {
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

func dnsLookup(host string) (bool, error) {
	resolver := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{Timeout: faviconTimeout}
			return d.DialContext(ctx, network, "8.8.8.8:53")
		},
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), faviconTimeout)
	defer cancel()
	
	_, err := resolver.LookupIPAddr(ctx, host)
	return err == nil, err
}