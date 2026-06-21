package main

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
)

// DownloadConfig encapsulates the parsed execution states for downstream workers
type DownloadConfig struct {
	Username           string
	DownloadPosts      bool
	DownloadHighlights bool
	Concurrency        int
}

// ParseCLIProfileURL extracts operational parameters and configuration from CLI args
func ParseCLIProfileURL() (DownloadConfig, error) {
	var config DownloadConfig
	config.Concurrency = 10

	if len(os.Args) < 2 {
		return config, errors.New("missing profile target link.\nUsage: ./ig-downloader <instagram-profile-url> [--p] [--h] [--workers 10]")
	}

	var inputURL string
	hasExplicitMode := false

	// Step 1: Parse arguments loop
	for i := 1; i < len(os.Args); i++ {
		arg := os.Args[i]

		switch arg {
		case "--p":
			if !hasExplicitMode {
				config.DownloadPosts = true
				config.DownloadHighlights = false
				hasExplicitMode = true
			} else {
				config.DownloadPosts = true
			}
		case "--h":
			if !hasExplicitMode {
				config.DownloadHighlights = true
				config.DownloadPosts = false
				hasExplicitMode = true
			} else {
				config.DownloadHighlights = true
			}
		case "--workers":
			// Handle space separated block: --workers 12
			if i+1 < len(os.Args) {
				val, err := strconv.Atoi(os.Args[i+1])
				if err == nil && val > 0 {
					config.Concurrency = val
				}
				i++ // Advance index position past value token
			}
		default:
			// Handle inline assignment syntax block: --workers=12
			if strings.HasPrefix(arg, "--workers=") {
				valStr := strings.TrimPrefix(arg, "--workers=")
				val, err := strconv.Atoi(valStr)
				if err == nil && val > 0 {
					config.Concurrency = val
				}
			} else if !strings.HasPrefix(arg, "-") && inputURL == "" {
				inputURL = strings.TrimSpace(arg)
			}
		}
	}

	// Default fallback: if no download targets specified, extract everything
	if !hasExplicitMode {
		config.DownloadPosts = true
		config.DownloadHighlights = true
	}

	if inputURL == "" {
		return config, errors.New("could not find a valid target link in your arguments")
	}

	// Step 2: Protocol fallback handling
	if !strings.HasPrefix(inputURL, "http://") && !strings.HasPrefix(inputURL, "https://") {
		inputURL = "https://" + inputURL
	}

	parsed, err := url.Parse(inputURL)
	if err != nil {
		return config, fmt.Errorf("invalid URL framework syntax: %w", err)
	}

	// Step 3: Enforce domain verification bounds
	host := strings.ToLower(parsed.Host)
	if !strings.Contains(host, "instagram.com") {
		return config, errors.New("provided link does not target a valid instagram.com domain namespace")
	}

	// Step 4: Tokenize and extract target username segments
	pathSegments := strings.Split(strings.Trim(parsed.Path, "/"), "/")
	if len(pathSegments) == 0 || pathSegments[0] == "" {
		return config, errors.New("could not isolate a valid username profile node from the provided URL path")
	}

	config.Username = pathSegments[0]

	// Step 5: Intercept internal web routing parameters
	reservedKeywords := map[string]bool{
		"explore":  true,
		"reels":    true,
		"stories":  true,
		"direct":   true,
		"p":        true,
		"accounts": true,
	}
	
	if reservedKeywords[strings.ToLower(config.Username)] {
		return config, fmt.Errorf("'%s' is an internal application layout route, not an individual user account profile", config.Username)
	}

	return config, nil
}