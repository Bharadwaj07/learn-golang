package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: scraper <url>")
		os.Exit(1)
	}

	targetURL := os.Args[1]

	fmt.Printf("Fetching %s...\n", targetURL)
	links, err := extractLinks(targetURL)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("Found %d links. Checking...\n\n", len(links))

	results := checkLinks(links, 10) // 10 concurrent requests

	dead := 0
	for _, r := range results {
		if !r.IsAlive() {
			fmt.Printf("❌ DEAD  %s\n", r.URL)
			dead++
		} else {
			fmt.Printf("✅ %d    %s\n", r.Status, r.URL)
		}
	}

	fmt.Printf("\n%d/%d links are dead\n", dead, len(results))
}
