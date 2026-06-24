package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"time"
)

var (
	excludeDirs = []string{".git", ".github", ".vscode", "build", "dist", "public", "docs"}
	excludeURLs = []string{
		"%s/jsonschema",
		"%s/jsonschema/",
	}
)

func main() {
	err := checklinks()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func checklinks() error {
	var (
		verboseFlag   bool
		sourceURLFlag string
		targetURLFlag string
	)
	flag.BoolVar(&verboseFlag, "v", false, "display more information")
	flag.StringVar(&sourceURLFlag, "source", "https://creativeprojects.github.io/resticprofile", "base URL to find links from")
	flag.StringVar(&targetURLFlag, "target", "", "base URL to check links against (leave empty to check against source URL)")
	flag.Parse()

	if sourceURLFlag == "" {
		return fmt.Errorf("source URL must be specified")
	}
	linkPattern, err := regexp.Compile(regexp.QuoteMeta(sourceURLFlag) + `[^\s"\)>]*`)
	if err != nil {
		return fmt.Errorf("compiling regex: %w", err)
	}
	for i, url := range excludeURLs {
		excludeURLs[i] = fmt.Sprintf(url, sourceURLFlag)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	allLinks := make([]string, 0)
	err = fs.WalkDir(os.DirFS("."), ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			fmt.Fprintf(os.Stderr, "error opening %s: %v\n", path, err)
			return nil
		}
		if d.IsDir() {
			if contains(excludeDirs, d.Name()) {
				return fs.SkipDir
			}
			return nil
		}
		links, err := findPatternInFile(linkPattern, path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error reading %s: %v\n", path, err)
			return nil
		}
		for _, link := range links {
			if !contains(allLinks, link) && !contains(excludeURLs, link) {
				allLinks = append(allLinks, link)
			}
			if verboseFlag {
				fmt.Printf("%s: %s\n", path, link)
			}
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("walking directory: %w", err)
	}

	source, err := hostname(sourceURLFlag)
	if err != nil {
		return fmt.Errorf("parsing source URL: %w", err)
	}
	target := ""
	if targetURLFlag != "" {
		target, err = hostname(targetURLFlag)
		if err != nil {
			return fmt.Errorf("parsing target URL: %w", err)
		}
	}
	httpClient := newHTTPClient(source, target)

	hasErrors := false
	fmt.Printf("Checking %d links\n", len(allLinks))
	for _, link := range allLinks {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		targetLink := link
		if targetURLFlag != "" {
			targetLink = strings.Replace(link, sourceURLFlag, targetURLFlag, 1)
		}
		err := checkLink(ctx, httpClient, targetLink)
		if err != nil {
			fmt.Printf("[ Error ] %s => %v\n", targetLink, err)
			hasErrors = true
			continue
		}
		fmt.Printf("[  OK   ] %s\n", targetLink)
	}
	if hasErrors {
		return fmt.Errorf("some links failed to check")
	}
	return nil
}

func findPatternInFile(pattern *regexp.Regexp, path string) ([]string, error) {
	rawBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return pattern.FindAllString(string(rawBytes), -1), nil
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func checkLink(ctx context.Context, client *http.Client, link string) error {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, link, http.NoBody)
	if err != nil {
		return err
	}
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode < 200 || response.StatusCode >= 400 {
		return fmt.Errorf("HTTP %d", response.StatusCode)
	}

	content, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}
	if _, id, found := strings.Cut(link, "#"); found {
		if !strings.Contains(string(content), `id="`+id+`"`) {
			return fmt.Errorf("fragment %q not found in page", id)
		}
	}
	return nil
}

func newHTTPClient(source, target string) *http.Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()

	if target != "" {
		fmt.Printf("Redirecting requests from %s to %s\n", source, target)
		dialer := &net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}
		transport.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
			if addr == source {
				addr = target
			}
			fmt.Printf("Dialing %s\n", addr)
			return dialer.DialContext(ctx, network, addr)
		}
	}

	httpClient := &http.Client{Transport: transport, Timeout: 30 * time.Second}
	return httpClient
}

func hostname(rawURL string) (string, error) {
	parts, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("parsing URL: %w", err)
	}
	port := parts.Port()
	if port == "" {
		if parts.Scheme == "https" {
			port = "443"
		} else {
			port = "80"
		}
	}
	return fmt.Sprintf("%s:%s", parts.Hostname(), port), nil
}
