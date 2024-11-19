package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"
)

var (
	verbose    bool
	concurrent int
	subsOnly   bool
	onlyIPs    bool
)

var cloudURLs = map[string]string{
	"Amazon":      "http://kaeferjaeger.gay/sni-ip-ranges/amazon/ipv4_merged_sni.txt",
	"DigitalOcean": "http://kaeferjaeger.gay/sni-ip-ranges/digitalocean/ipv4_merged_sni.txt",
	"Microsoft":   "http://kaeferjaeger.gay/sni-ip-ranges/microsoft/ipv4_merged_sni.txt",
	"Google":      "http://kaeferjaeger.gay/sni-ip-ranges/google/ipv4_merged_sni.txt",
	"Oracle":      "http://kaeferjaeger.gay/sni-ip-ranges/oracle/ipv4_merged_sni.txt",
}

func log(level string, message string) {
	if !verbose && level != "ERROR" {
		return
	}
	switch level {
	case "INFO":
		fmt.Printf("[INFO] %s\n", message)
	case "OK":
		fmt.Printf("[OK] %s\n", message)
	case "ERROR":
		fmt.Printf("[ERROR] %s\n", message)
	default:
		fmt.Println("Invalid log level")
	}
}

func cleanInput(input string) string {
	return strings.TrimPrefix(input, "*.")
}

func extractSubdomains(line string) []string {
	subdomainRegex := regexp.MustCompile(`\[(.*?)\]`)
	matches := subdomainRegex.FindStringSubmatch(line)
	if len(matches) < 2 {
		return nil
	}
	subdomains := strings.Fields(matches[1])
	for i := range subdomains {
		subdomains[i] = cleanInput(subdomains[i])
	}
	return subdomains
}

func extractIPs(line string) string {
	ipRegex := regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+`)
	ip := ipRegex.FindString(line)
	return ip
}

func fetchAndSearchCloudData(query string) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	var wg sync.WaitGroup
	dataChan := make(chan string, concurrent)

	for i := 0; i < concurrent; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for url := range dataChan {
				if err := fetchAndProcessURL(client, url, query); err != nil {
					log("ERROR", fmt.Sprintf("Failed to process %s: %v", url, err))
				}
			}
		}()
	}

	for _, url := range cloudURLs {
		dataChan <- url
	}
	close(dataChan)
	wg.Wait()
}

func fetchAndProcessURL(client *http.Client, url, query string) error {
	const maxRetries = 3
	for attempt := 1; attempt <= maxRetries; attempt++ {
		err := fetchAndProcess(client, url, query)
		if err == nil {
			return nil
		}
		log("INFO", fmt.Sprintf("Retrying %s (attempt %d/%d)", url, attempt, maxRetries))
		time.Sleep(2 * time.Second)
	}
	return errors.New("all retries failed")
}

func fetchAndProcess(client *http.Client, url, query string) error {
	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("HTTP request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected HTTP status: %s", resp.Status)
	}

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if query == "" || strings.Contains(line, query) {
			processLine(line)
		}
	}
	return scanner.Err()
}

func processLine(line string) {
	if subsOnly {
		subdomains := extractSubdomains(line)
		if subdomains != nil {
			for _, subdomain := range subdomains {
				fmt.Println(subdomain)
			}
		}
	} else if onlyIPs {
		ip := extractIPs(line)
		if ip != "" {
			fmt.Println(ip)
		}
	} else {
		fmt.Println(line)
	}
}

func main() {
	flag.BoolVar(&verbose, "v", false, "Enable verbose output")
	flag.IntVar(&concurrent, "c", 10, "Concurrency level for search")
	flag.BoolVar(&subsOnly, "subs", false, "Output only subdomains")
	flag.BoolVar(&onlyIPs, "only-ips", false, "Output only IP addresses")
	flag.Parse()

	scanner := bufio.NewScanner(os.Stdin)
	var input []string
	for scanner.Scan() {
		input = append(input, scanner.Text())
	}

	if len(input) > 0 {
		for _, domain := range input {
			fetchAndSearchCloudData(cleanInput(domain))
		}
	} else {
		fmt.Println("No input provided. Use stdin or specify flags.")
	}
}
