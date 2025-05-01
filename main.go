package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"github.com/saad-ayady/SubFors/banner"
	"github.com/saad-ayady/SubFors/core" 
	"github.com/saad-ayady/SubFors/output"
	"sync"
	"time"
)

var (
	wordlist    string
	outputText  string
	outputJSON  string
	outputXML   string
	domain      string
	domainList  string
)

func init() {
	flag.StringVar(&wordlist, "w", "./db/wordlist.txt", "Path to wordlist file for brute-force")
	flag.StringVar(&outputText, "o", "", "Save results in text format")
	flag.StringVar(&outputJSON, "oJ", "", "Save results in JSON format")
	flag.StringVar(&outputXML, "oX", "", "Save results in XML format")
	flag.StringVar(&domain, "d", "", "Single domain to scan")
	flag.StringVar(&domainList, "dL", "", "Path to file containing list of domains")
	
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Subdomains Discovery Tool\n\n")
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s [options]\n\n", os.Args[0])
		fmt.Println("Options:")
		flag.PrintDefaults()
		fmt.Println("\nExamples:")
		fmt.Print("-> subfors -d example.com -o results.txt\n")
		fmt.Print("-> subfors -dL domains.txt -oJ results.json\n")
		fmt.Print("-> subfors -d facebook.com -w wordlist.txt -oX results.xml\n")
	}
}

func main() {
	flag.Parse()
	
	// Validate input
	if domain == "" && domainList == "" {
		fmt.Println("Error: You must specify either -d or -dL")
		flag.Usage()
		os.Exit(1)
	}
	banner.PrintBanner()  // Call the exported function
	// Process domains
	if domain != "" {
		processDomain(domain)
	} else if domainList != "" {
		processDomainList(domainList)
	}
}

func processDomain(domain string) {
	startTime := time.Now()

	// Configure search methods
	searchers := []struct {
		name   string
		search func(string) ([]string, error)
	}{
		{"Google", core.GoogleDork},
		{"DuckDuckGo", core.DuckDork},
		{"Bing", core.BingDork},
		{"Certificate Transparency", core.GetSubsFromCRT},
		{"Web Archives", core.GetSubsFromArchive},
		{"Website Crawler", core.GetSubsFromCrawling},
		{"JavaScript Analysis", core.GetSubsFromJS},
		{"Brute Force", func(d string) ([]string, error) {
			return core.BruteSubdomains(d, wordlist)
		}},
	}

	// Processing setup
	results := make(chan string, 1000)
	uniqueSubs := sync.Map{}
	var wg sync.WaitGroup
	var resultsWg sync.WaitGroup

	// Results collector
	resultsWg.Add(1)
	go func() {
		defer resultsWg.Done()
		for sub := range results {
			uniqueSubs.Store(sub, true)
			fmt.Printf("[+] %s\n", sub)
		}
	}()

	// Run all scanners
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	rateLimiter := time.Tick(500 * time.Millisecond)
	
	for _, engine := range searchers {
		<-rateLimiter
		wg.Add(1)
		
		go func(name string, searchFn func(string) ([]string, error)) {
			defer wg.Done()
			
			fmt.Printf("[*] Scanning %s...\n", name)
			subs, err := searchFn(domain)
			if err != nil {
				fmt.Printf("[-] %s error: %v\n", name, err)
				return
			}
			
			for _, sub := range subs {
				select {
				case results <- sub:
				case <-ctx.Done():
					return
				}
			}
		}(engine.name, engine.search)
	}

	wg.Wait()
	close(results)
	resultsWg.Wait()

	// Convert results to slice
	var subdomains []string
	uniqueSubs.Range(func(key, value interface{}) bool {
		subdomains = append(subdomains, key.(string))
		return true
	})

	// Final output
	elapsed := time.Since(startTime).Round(time.Second)
	fmt.Printf("\n[+] Scan completed in %s\n", elapsed)
	fmt.Printf("[+] Found %d unique subdomains\n", len(subdomains))

	// Save results
	saveResults(domain, subdomains)
}

func processDomainList(listPath string) {
	file, err := os.Open(listPath)
	if err != nil {
		fmt.Printf("[-] Error opening domain list: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		domain := strings.TrimSpace(scanner.Text())
		if domain != "" {
			fmt.Printf("\n=== Processing domain: %s ===\n", domain)
			processDomain(domain)
		}
	}
}

func saveResults(domain string, subs []string) {
	if outputText != "" || outputJSON != "" || outputXML != "" {
		if err := output.SaveResults(domain, subs, outputText, outputJSON, outputXML); err != nil {
			fmt.Printf("\n[-] Error saving results: %v\n", err)
		} else {
			if outputText != "" {
				fmt.Printf("[+] Text results saved to: %s\n", outputText)
			}
			if outputJSON != "" {
				fmt.Printf("[+] JSON results saved to: %s\n", outputJSON)
			}
			if outputXML != "" {
				fmt.Printf("[+] XML results saved to: %s\n", outputXML)
			}
		}
	}
}