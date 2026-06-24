// checklinks is a link-checking tool for the resticprofile documentation.
//
// It scans all files in the current directory tree (skipping common non-doc
// directories such as .git, build, and dist) for URLs that begin with the
// configured source base URL, then verifies that every discovered URL returns
// a successful HTTP response.
//
// When checking links, the tool can operate in two modes:
//
//  1. Local mode (default): starts an in-process HTTPS server that serves the
//     compiled documentation from a local directory, then checks all links
//     against that server.  Use -dir to point at the directory.
//
//  2. Remote mode: redirects all requests to an alternative base URL supplied
//     via -target.  Useful for smoke-testing a staging deployment without
//     rewriting every link in the source files.
//
// In addition to checking HTTP status codes, the tool validates URL fragments
// (e.g. /page#section) by searching the response body for the corresponding
// id="…" attribute, so broken anchors are caught as well.
//
// Usage:
//
//	checklinks [-v] [-source <url>] [-target <url>] [-dir <path>]
//
// Flags:
//
//	-v          Print each link as it is discovered (verbose output).
//	-source     Base URL whose occurrences are searched for in files.
//	            Defaults to https://creativeprojects.github.io/resticprofile.
//	-target     Alternative base URL to send requests to.  When omitted, a
//	            local server is started and -dir must be provided.
//	-dir        Directory containing the compiled HTML documentation.
//	            Required when -target is not specified.
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"os/signal"
	"regexp"
	"slices"
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
		dirFlag       string
	)
	flag.BoolVar(&verboseFlag, "v", false, "display more information")
	flag.StringVar(&sourceURLFlag, "source", "https://creativeprojects.github.io/resticprofile", "base URL to find links from")
	flag.StringVar(&targetURLFlag, "target", "", "base URL to check links against (leave empty to check against source URL)")
	flag.StringVar(&dirFlag, "dir", "", "directory with HTML documentation")
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
			if slices.Contains(excludeDirs, d.Name()) {
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
			if !slices.Contains(allLinks, link) && !slices.Contains(excludeURLs, link) {
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

	var transport *http.Transport
	if target == "" {
		// Start a local server to serve the documentation if no target URL is provided
		if dirFlag == "" {
			return fmt.Errorf("directory must be specified when no target URL is provided")
		}
		server := startServer(dirFlag)
		defer server.Close()

		targetURLFlag = server.URL
		target, err = hostname(targetURLFlag)
		if err != nil {
			return fmt.Errorf("parsing local server URL: %w", err)
		}
		fmt.Printf("Started local server at %s serving directory %s\n", targetURLFlag, dirFlag)
		transport = server.Client().Transport.(*http.Transport)
	}
	httpClient := newHTTPClient(source, target, transport)

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

func newHTTPClient(source, target string, transport *http.Transport) *http.Client {
	if transport == nil {
		transport = http.DefaultTransport.(*http.Transport)
	}
	transport = transport.Clone()

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

func startServer(root string) *httptest.Server {
	return httptest.NewTLSServer(http.FileServer(http.Dir(root)))
}
