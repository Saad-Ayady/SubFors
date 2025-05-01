package core

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	defaultWorkers = 200
	defaultTimeout = 2 * time.Second
)

func BruteSubdomains(domain, wordlistPath string) ([]string, error) {
	words, err := loadWordlist(wordlistPath)
	if err != nil {
		return nil, err
	}

	results := make(chan string)
	done := make(chan struct{})
	var wg sync.WaitGroup
	var mutex sync.Mutex
	var subs []string

	// Start result collector
	go func() {
		for sub := range results {
			mutex.Lock()
			subs = append(subs, sub)
			mutex.Unlock()
		}
		close(done)
	}()

	// Create worker pool
	jobs := make(chan string, len(words))
	for i := 0; i < defaultWorkers; i++ {
		wg.Add(1)
		go worker(domain, jobs, results, &wg)
	}

	// Feed jobs
	for _, word := range words {
		jobs <- word
	}
	close(jobs)
	wg.Wait()
	close(results)
	<-done

	return subs, nil
}

func worker(domain string, jobs <-chan string, results chan<- string, wg *sync.WaitGroup) {
	defer wg.Done()
	resolver := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{Timeout: defaultTimeout}
			return d.DialContext(ctx, "udp", "8.8.8.8:53")
		},
	}

	for word := range jobs {
		subdomain := fmt.Sprintf("%s.%s", word, domain)
		if checkSubdomain(resolver, subdomain) {
			results <- subdomain
		}
	}
}

func checkSubdomain(resolver *net.Resolver, subdomain string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	_, err := resolver.LookupHost(ctx, subdomain)
	return err == nil
}

func loadWordlist(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("error opening wordlist: %v", err)
	}
	defer file.Close()

	var words []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		words = append(words, strings.TrimSpace(scanner.Text()))
	}
	return words, scanner.Err()
}