package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"regexp"
	"strings"
)

var (
	excludeDirs = []string{".git", ".github", ".vscode", "build", "dist", "public", "docs"}
	linkPattern = regexp.MustCompile(`https?://creativeprojects\.github\.io/resticprofile[^\s"\)>]*`)
	verbose     = false
	sourceURL   = "https://creativeprojects.github.io/resticprofile"
	targetURL   = ""
	excludeURLs = []string{
		"https://creativeprojects.github.io/resticprofile/jsonschema",
		"https://creativeprojects.github.io/resticprofile/jsonschema/",
	}
)

func main() {
	flag.BoolVar(&verbose, "v", false, "display more information")
	flag.StringVar(&targetURL, "target", "http://127.0.0.1:1313/resticprofile", "base URL to check links against")
	flag.Parse()

	allLinks := make([]string, 0)
	_ = fs.WalkDir(os.DirFS("."), ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening %s: %v\n", path, err)
			return err
		}
		if d.IsDir() {
			if contains(excludeDirs, d.Name()) {
				return fs.SkipDir
			}
			return nil
		}
		links, err := findLinks(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading %s: %v\n", path, err)
			return nil
		}
		for _, link := range links {
			if !contains(allLinks, link) && !contains(excludeURLs, link) {
				allLinks = append(allLinks, link)
			}
			if verbose {
				fmt.Printf("%s: %s\n", path, link)
			}
		}
		return nil
	})

	hasErrors := false
	fmt.Printf("Checking %d links\n", len(allLinks))
	for _, link := range allLinks {
		targetLink := strings.Replace(link, sourceURL, targetURL, 1)
		err := checkLink(context.Background(), targetLink)
		if err != nil {
			fmt.Printf("[ Error ] %s => %v\n", targetLink, err)
			hasErrors = true
			continue
		}
		fmt.Printf("[  OK   ] %s\n", targetLink)
	}
	if hasErrors {
		os.Exit(1)
	}
}

func findLinks(path string) ([]string, error) {
	rawBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	links := linkPattern.FindAllString(string(rawBytes), -1)

	return links, nil
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func checkLink(ctx context.Context, link string) error {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, link, http.NoBody)
	if err != nil {
		return err
	}
	response, err := http.DefaultClient.Do(request)
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
