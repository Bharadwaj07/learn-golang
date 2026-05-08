package main

import (
	"fmt"
	"net/http"
	"net/url"

	"golang.org/x/net/html"
)

// extractLinks fetches the page at rawURL and returns all href values found
func extractLinks(rawURL string) ([]string, error) {
	resp, err := http.Get(rawURL)
	if err != nil {
		return nil, fmt.Errorf("fetching page: %w", err)
	}
	defer resp.Body.Close()

	// parse the HTML body into a node tree
	doc, err := html.Parse(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("parsing html: %w", err)
	}

	base, _ := url.Parse(rawURL)
	var links []string

	// walkNode visits every node in the HTML tree recursively
	var walkNode func(*html.Node)
	walkNode = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, attr := range n.Attr {
				if attr.Key == "href" {
					link := resolveLink(base, attr.Val)
					if link != "" {
						links = append(links, link)
					}
				}
			}
		}
		// visit every child node
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			walkNode(child)
		}
	}

	walkNode(doc)
	return links, nil
}

// resolveLink turns relative links like /about into absolute ones
// like https://example.com/about. Returns "" for non-http links.
func resolveLink(base *url.URL, href string) string {
	u, err := url.Parse(href)
	if err != nil {
		return ""
	}
	resolved := base.ResolveReference(u)
	if resolved.Scheme != "http" && resolved.Scheme != "https" {
		return "" // skip mailto:, javascript:, etc.
	}
	return resolved.String()
}
