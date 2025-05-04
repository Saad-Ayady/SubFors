# SubFors 

![SubFors Banner](./images/pic.png)

**SubFors** is a fast, modular subdomain discovery tool that combines multiple enumeration techniques to uncover hidden attack surfaces. Now with **API integrations** for enhanced reconnaissance.

[![Go Version](https://img.shields.io/badge/go-1.20+-blue.svg)](https://golang.org/dl/)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](./LICENSE)
[![API Ready](https://img.shields.io/badge/API%20Integrations-VirusTotal%2FDNSDumpster%2FGitHub-orange)]()

## Features 

- **Multi-engine enumeration** (12 discovery methods)
- **API integrations** (VirusTotal, DNSDumpster, GitHub)
- **Certificate Transparency** monitoring
- **Brute-force** with custom wordlists
- **Web Archives** analysis
- **JavaScript** file scanning
- **GitHub subdomain extraction**
- **Smart rate-limiting** to avoid detection
- **Multiple output formats** (TXT/JSON/XML)
- **Bulk domain processing**

## Installation 

### From Source
```bash
git clone https://github.com/saad-ayady/SubFors
cd SubFors
go build -o subfors main.go
sudo mv subfors /usr/local/bin/
```

## Using Go 

```bash 
go install github.com/saad-ayady/SubFors/cmd/subfors@latest
```

# Usage 

## Basic Scan 

```bash 
subfors -d example.com
```

## Advanced Scan with APIs

```bash 
subfors -d example.com \
  -vt YOUR_VIRUSTOTAL_API_KEY \
  -dn YOUR_DNSDUMPSTER_API_KEY \
  -gt YOUR_GITHUB_TOKEN \
  -oJ results.json


subfors -dL Scope.txt \
  -w custom_wordlist.txt \
  -vt YOUR_VIRUSTOTAL_API_KEY \
  -dn YOUR_DNSDUMPSTER_API_KEY \
  -gt YOUR_GITHUB_TOKEN \
  -oJ results.json

subfors -dL Scope.txt \
  -w custom_wordlist.txt \
  -oJ results.json
```

## Bulk Scanning 

```bash 
subfors -dL domains.txt -w custom_wordlist.txt -oX results.xml
```

# Full Options 

| Flag      | Description                       | Example               |
|-----------|-----------------------------------|-----------------------|
| `-d`      | Target domain                     | `-d example.com`      |
| `-dL`     | File containing domains           | `-dL domains.txt`     |
| `-vt`     | VirusTotal API key                | `-vt abc123def456`    |
| `-dn`     | DNSDumpster API key               | `-dn xyz789uvw012`    |
| `-gt`     | GitHub Personal Access Token      | `-gt ghp_abcd1234xyz` |
| `-w`      | Custom wordlist path              | `-w wordlist.txt`     |
| `-o`      | Text output file                  | `-o results.txt`      |
| `-oJ`     | JSON output file                  | `-oJ results.json`    |
| `-oX`     | XML output file                   | `-oX results.xml`     |
| `-t`      | Threads (default: `10`)           | `-t 20`               |
| `-timeout`| Timeout in seconds (default: `30`)| `-timeout 60`         |

## API Modules Guide :
<div>
  <img src="https://img.shields.io/badge/API_Version-v3.0-0078ff?style=flat&logo=virustotal&logoColor=white" alt="VirusTotal API">
</div>

**VirusTotal Integration**
  1. **Get your API key from [VirusTotal](https://www.virustotal.com/gui/home/upload)**
  2. **Use with `-vt` flag:**
```bash
subfors -d target.com -vt YOUR_API_KEY
```
  . **Queries VirusTotal's subdomains database**<br>
  . **Handles pagination automatically**<br>
  . **Rate-limited to comply with API restrictions**

<div>
  <img src="https://img.shields.io/badge/API_Version-v1.0-28a745?style=flat&logo=namecheap&logoColor=white" alt="DNSDumpster API">
</div>

**DNSDumpster Integration**
  1. **Get your API key from [DNSDumpster](https://dnsdumpster.com/)**
  2. **Use with `-dn` flag:**
```bash
subfors -d target.com -dn YOUR_API_KEY
```
  . **Retrieves DNS records including historical data**<br>
  . **Processes A records for subdomains**

<div> <img src="https://img.shields.io/badge/API_Version-v1.0-6e5494?style=flat&logo=github&logoColor=white" alt="GitHub API"> </div>

**GitHub Integration**
  1. **Get your API key from [GitHub Personal Access Token](https://github.com/settings/tokens)**
  2. **Use with `-gt` flag:**
```bash
subfors -d target.com -gt YOUR_GITHUB_TOKEN
```
  . **Searches public code and repositories for subdomains**<br>
  . **Extracts leaked endpoints and configs containing domains**<br>
  . **Avoids GitHub rate-limits using your token**


# Output Example 

```text
[•] Starting SubFors v0.2 scan for example.com
[✓] VirusTotal API connected (Quota: 498/500)
[✓] GitHub token authenticated
[•] Running 12 discovery modules...

[+] admin.example.com       (Certificate Transparency)
[+] api.dev.example.com     (VirusTotal)
[+] devops.example.com      (GitHub)
[+] legacy.example.com      (DNSDumpster)
[+] beta.example.com        (Web Archives)

[✓] Scan completed in 2m18s
[✓] Found 612 unique subdomains
[✓] JSON results saved to: results.json


```

# Comparison 

| Feature        | SubFors | SubFinder | AssetFinder |
|---------------|---------|-----------|-------------|
| API Integrations  | ✅ (VT+DNS)     | ❌        | ❌          |
| Multi-engine  | ✅ (11)     | ✅ (8)       | ❌          |
| CT Logs       | ✅      | ✅        | ✅          |
| Web Archives  | ✅      | ❌        | ❌          |
| JS Analysis   | ✅      | ❌        | ❌          |
| GitHub Leaks   | ✅      | ❌        | ❌          |
| Rate Limiting | ✅      | ❌        | ❌          |
| Bulk Processing | ✅    | ✅        | ❌          |

## Contribution 

1. **Fork the repository**  

2. **Create your feature branch**  

3. **Commit your changes**  

4. **Push to the branch**  

5. **Open a pull request**  

# Developed by [0xS22d](https://saad-ayady.github.io/myWEBSITE/) - Happy Hunting! 🎯🚀
