package core

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type CRTResponse []struct {
	NameValue string `json:"name_value"`
}

func GetSubsFromCRT(domain string) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	apiURL := fmt.Sprintf("https://crt.sh/?q=%%25.%s&output=json", domain)

	client := &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:        10,
			IdleConnTimeout:     30 * time.Second,
			DisableCompression:  false,
		},
		Timeout: 60 * time.Second,
	}

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("error in create req  %v", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	req.Header.Set("Accept", "application/json")

	var resp *http.Response
	for attempt := 0; attempt < 3; attempt++ {
		resp, err = client.Do(req)
		if err == nil && resp.StatusCode == http.StatusOK {
			break
		}
		time.Sleep(time.Duration(attempt+1) * time.Second)
	}
	if err != nil {
		return nil, fmt.Errorf("error in connect with crf.sh: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error in responsing from crf.sh: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error in reading respons :  %v", err)
	}

	// التحقق من وجود CAPTCHA
	if strings.Contains(string(body), "captcha") {
		return nil, fmt.Errorf("in response, Fond CAPTCHA, please try again later")
	}

	var data CRTResponse
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, fmt.Errorf("Error in analyse JSON: %v", err)
	}

	var results []string
	seen := make(map[string]struct{})

	for _, entry := range data {
		names := strings.Split(entry.NameValue, "\n")
		for _, name := range names {
			name = strings.TrimSpace(name)
			if name == "" || !strings.Contains(name, domain) {
				continue
			}

			if strings.HasPrefix(name, "*.") {
				name = strings.TrimPrefix(name, "*.")
			}

			if _, exists := seen[name]; !exists {
				seen[name] = struct{}{}
				results = append(results, name)
				fmt.Printf("[crt.sh] ➜ %s\n", name)
			}
		}
	}

	return results, nil
}