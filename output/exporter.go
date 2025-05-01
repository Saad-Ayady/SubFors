package output

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
)

func SaveResults(domain string, subs []string, textPath, jsonPath, xmlPath string) error {
	if len(subs) == 0 {
		return fmt.Errorf("no subdomains found to save")
	}

	var errors []error
	
	if textPath != "" {
		if err := saveText(textPath, subs); err != nil {
			errors = append(errors, err)
		}
	}
	
	if jsonPath != "" {
		if err := saveJSON(jsonPath, domain, subs); err != nil {
			errors = append(errors, err)
		}
	}
	
	if xmlPath != "" {
		if err := saveXML(xmlPath, domain, subs); err != nil {
			errors = append(errors, err)
		}
	}
	
	if len(errors) > 0 {
		return fmt.Errorf("output errors: %v", errors)
	}
	return nil
}

func saveText(path string, subs []string) error {
	if err := ensureDir(path); err != nil {
		return err
	}
	
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create text file: %w", err)
	}
	defer file.Close()
	
	for _, sub := range subs {
		if _, err := file.WriteString(sub + "\n"); err != nil {
			return fmt.Errorf("write error: %w", err)
		}
	}
	return nil
}

func saveJSON(path, domain string, subs []string) error {
	type Output struct {
		Domain     string   `json:"domain"`
		Subdomains []string `json:"subdomains"`
	}
	
	data := Output{
		Domain:     domain,
		Subdomains: subs,
	}
	
	if err := ensureDir(path); err != nil {
		return err
	}
	
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create JSON file: %w", err)
	}
	defer file.Close()
	
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(data); err != nil {
		return fmt.Errorf("JSON encoding error: %w", err)
	}
	return nil
}

func saveXML(path, domain string, subs []string) error {
	type Subdomain struct {
		Name string `xml:"name"`
	}
	
	type Output struct {
		XMLName    xml.Name    `xml:"subdomains"`
		Domain     string      `xml:"domain"`
		Subdomains []Subdomain `xml:"subdomain"`
	}
	
	output := Output{
		Domain: domain,
	}
	
	for _, sub := range subs {
		output.Subdomains = append(output.Subdomains, Subdomain{Name: sub})
	}
	
	if err := ensureDir(path); err != nil {
		return err
	}
	
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create XML file: %w", err)
	}
	defer file.Close()
	
	if _, err := file.WriteString(xml.Header); err != nil {
		return fmt.Errorf("XML header error: %w", err)
	}
	
	encoder := xml.NewEncoder(file)
	encoder.Indent("", "  ")
	if err := encoder.Encode(output); err != nil {
		return fmt.Errorf("XML encoding error: %w", err)
	}
	return nil
}

func ensureDir(path string) error {
	dir := filepath.Dir(path)
	if dir != "" {
		return os.MkdirAll(dir, 0755)
	}
	return nil
}