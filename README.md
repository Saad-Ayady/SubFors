# SubFors 🔍

![SubFors Banner](./images/pic.png)

**SubFors** is a fast, modular subdomain discovery tool that combines multiple enumeration techniques to uncover hidden attack surfaces.

[![Go Version](https://img.shields.io/badge/go-1.20+-blue.svg)](https://golang.org/dl/)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](./LICENSE)
[![Contributions Welcome](https://img.shields.io/badge/contributions-welcome-brightgreen.svg)](./CONTRIBUTING.md)

## Features ✨

- **Multi-engine enumeration** (Google, Bing, DuckDuckGo, etc.)
- **Certificate Transparency** monitoring
- **Brute-force** with custom wordlists
- **Web Archives** analysis
- **JavaScript** file scanning
- **Smart rate-limiting** to avoid detection
- **Multiple output formats** (TXT/JSON/XML)
- **Bulk domain processing**

## Installation 📦

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

# Usage 🛠️ 

## Basic Scan 

```bash 
subfors -d example.com
```

## Advanced Scan 

```bash 
subfors -d example.com \
  -w custom_wordlist.txt \
  -o results.txt \
  -oJ results.json
```

## Bulk Scanning 

```bash 
subfors -dL domains.txt -oJ all_results.json
```

# Full Options 📋

| Flag      | Description                       | Example               |
|-----------|-----------------------------------|-----------------------|
| `-d`      | Target domain                     | `-d example.com`      |
| `-dL`     | File containing domains           | `-dL domains.txt`     |
| `-w`      | Custom wordlist path              | `-w wordlist.txt`     |
| `-o`      | Text output file                  | `-o results.txt`      |
| `-oJ`     | JSON output file                  | `-oJ results.json`    |
| `-oX`     | XML output file                   | `-oX results.xml`     |
| `-t`      | Threads (default: `10`)           | `-t 20`               |
| `-timeout`| Timeout in seconds (default: `30`)| `-timeout 60`         |

# Output Example 📄

```text
[*] Starting SubFors scan for example.com
[+] Found 23 unique subdomains

┌──────────────────────┬──────────────────────────┐
│      SUBDOMAIN       │        SOURCE            │
├──────────────────────┼──────────────────────────┤
│ admin.example.com    │ Certificate Transparency │
│ beta.example.com     │ Google Dork              │
│ dev.example.com      │ Brute Force              │
└──────────────────────┴──────────────────────────┘
```

# Comparison 📊

| Feature        | SubFors | SubFinder | AssetFinder |
|---------------|---------|-----------|-------------|
| Multi-engine  | ✅      | ✅        | ❌          |
| CT Logs       | ✅      | ✅        | ✅          |
| Web Archives  | ✅      | ❌        | ❌          |
| JS Analysis   | ✅      | ❌        | ❌          |
| Rate Limiting | ✅      | ❌        | ❌          |
| Bulk Processing | ✅    | ✅        | ❌          |

## Contribution 🤝

1. **Fork the repository**  

2. **Create your feature branch**  

3. **Commit your changes**  

4. **Push to the branch**  

5. **Open a pull request**  

# Developed by [0xS22d](https://saad-ayady.github.io/myWEBSITE/) - Happy Hunting! 🎯🚀
