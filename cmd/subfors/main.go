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
	vtAPIKey    string
	dnAPIKey    string
)

func init() {
	flag.StringVar(&wordlist, "w", "./db/wordlist.txt", "Path to wordlist file for brute-force")
	flag.StringVar(&outputText, "o", "", "Save results in text format")
	flag.StringVar(&outputJSON, "oJ", "", "Save results in JSON format")
	flag.StringVar(&outputXML, "oX", "", "Save results in XML format")
	flag.StringVar(&domain, "d", "", "Single domain to scan")
	flag.StringVar(&domainList, "dL", "", "Path to file containing list of domains")
	flag.StringVar(&vtAPIKey, "vt", "", "VirusTotal API key (optional)")
	flag.StringVar(&dnAPIKey, "dn", "", "DNSDumpster API key (optional)")
	
	flag.Usage = func() {
		banner.PrintBanner()
		fmt.Fprintf(flag.CommandLine.Output(), "\nSubdomains Discovery Tool\n\n")
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s [options]\n\n", os.Args[0])
		fmt.Println("Options:")
		flag.PrintDefaults()
		fmt.Println("\nAPI Modules:")
		fmt.Println("  -vt string   VirusTotal API key (enables VT subdomain lookup)")
		fmt.Println("  -dn string   DNSDumpster API key (enables DNSDumpster lookup)")
		fmt.Println("\nScan Techniques:")
		fmt.Println("  - Brute Force")
		fmt.Println("  - Certificate Transparency")
		fmt.Println("  - Web Archives")
		fmt.Println("  - Favicon Hash Matching")
		fmt.Println("  - Search Engine Dorking")
		fmt.Println("  - JavaScript Analysis")
		fmt.Println("\nExamples:")
		fmt.Println("  Basic scan:")
		fmt.Println("    subfors -d example.com -o results.txt")
		fmt.Println("  Full scan with APIs:")
		fmt.Println("    subfors -d example.com -vt YOUR_VT_KEY -dn YOUR_DNS_KEY -oJ results.json")
		fmt.Println("  Domain list scan:")
		fmt.Println("    subfors -dL domains.txt -w custom_wordlist.txt -oX results.xml")
	}
}

func main() {
	flag.Parse()
	
	if domain == "" && domainList == "" {
		fmt.Println("\n[!] Error: You must specify either -d or -dL")
		flag.Usage()
		os.Exit(1)
	}

	banner.PrintBanner()

	if domain != "" {
		processDomain(domain)
	} else if domainList != "" {
		processDomainList(domainList)
	}
}

func processDomain(domain string) {
	startTime := time.Now()
	fmt.Printf("\n[•] Starting scan for %s\n", domain)

	// Configure all available search methods
	searchers := []struct {
		name   string
		search func(string) ([]string, error)
	}{
		{"Brute Force", func(d string) ([]string, error) {
			return core.BruteSubdomains(d, wordlist)
		}},
		{"Certificate Transparency", core.GetSubsFromCRT},
		{"Web Archives", core.GetSubsFromArchive},
		{"Website Crawler", core.GetSubsFromCrawling},
		{"JavaScript Analysis", core.GetSubsFromJS},
		{"Google Dork", core.GoogleDork},
		{"Bing Dork", core.BingDork},
		{"DuckDuckGo Dork", core.DuckDork},
		{"Headers Analysis", core.GetSubsFromHeaders},
		{"Favicon Matching", core.GetSubsFromFavicon},
		{"VirusTotal", func(d string) ([]string, error) {
			if vtAPIKey == "" {
				return nil, nil
			}
			return core.GetSubsFromVirusTotal(d, vtAPIKey)
		}},
		{"DNSDumpster", func(d string) ([]string, error) {
			if dnAPIKey == "" {
				return nil, nil
			}
			return core.GetSubsFromDNSDumpster(d, dnAPIKey)
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

	// Run all scanners with rate limiting
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	rateLimiter := time.Tick(500 * time.Millisecond)
	
	for _, engine := range searchers {
		<-rateLimiter
		wg.Add(1)
		
		go func(name string, searchFn func(string) ([]string, error)) {
			defer wg.Done()
			
			if searchFn == nil {
				return
			}

			fmt.Printf("[•] Running %s module...\n", name)
			subs, err := searchFn(domain)
			if err != nil {
				fmt.Printf("[!] %s error: %v\n", name, err)
				return
			}
			
			if len(subs) > 0 {
				fmt.Printf("[✓] %s found %d subdomains\n", name, len(subs))
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
	fmt.Printf("\n[✓] Scan completed in %s\n", elapsed)
	fmt.Printf("[✓] Found %d unique subdomains\n", len(subdomains))

	// Save results if any output path specified
	saveResults(domain, subdomains)
}

func processDomainList(listPath string) {
	file, err := os.Open(listPath)
	if err != nil {
		fmt.Printf("\n[!] Error opening domain list: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		domain := strings.TrimSpace(scanner.Text())
		if domain != "" {
			fmt.Printf("\n=== Processing domain: %s ===\n", domain)
			processDomain(domain)
			time.Sleep(2 * time.Second) // Rate limit between domains
		}
	}
}

func saveResults(domain string, subs []string) {
	if outputText != "" || outputJSON != "" || outputXML != "" {
		if err := output.SaveResults(domain, subs, outputText, outputJSON, outputXML); err != nil {
			fmt.Printf("\n[!] Error saving results: %v\n", err)
		} else {
			if outputText != "" {
				fmt.Printf("[✓] Text results saved to: %s\n", outputText)
			}
			if outputJSON != "" {
				fmt.Printf("[✓] JSON results saved to: %s\n", outputJSON)
			}
			if outputXML != "" {
				fmt.Printf("[✓] XML results saved to: %s\n", outputXML)
			}
		}
	}
}