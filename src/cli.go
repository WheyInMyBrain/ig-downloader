package main

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
)

type DownloadConfig struct {
	Username           string
	DownloadPosts      bool
	DownloadHighlights bool
	DownloadReels      bool
	Concurrency        int
	ServeUI            bool
}

func ParseCLICookies() {
	for i := 1; i < len(os.Args); i++ {
		arg := os.Args[i]

		// Format 1: --cookies "raw string or json"
		if arg == "--cookies" && i+1 < len(os.Args) {
			err := ParseAndSaveCookies(os.Args[i+1])
			if err != nil {
				fmt.Printf("[Cookie Error] %v\n", err)
				os.Exit(1)
			}
			os.Exit(0)
		}

		// Format 2: --cookies="raw string or json"
		if strings.HasPrefix(arg, "--cookies=") {
			cookieVal := strings.TrimPrefix(arg, "--cookies=")
			err := ParseAndSaveCookies(cookieVal)
			if err != nil {
				fmt.Printf("[Cookie Error] %v\n", err)
				os.Exit(1)
			}
			os.Exit(0)
		}
	}
}

func ParseCLIProfileURL() (DownloadConfig, error) {
	var config DownloadConfig
	config.Concurrency = 10

	if len(os.Args) < 2 {
		return config, errors.New("missing profile target link.\nUsage: go run . <instagram-profile-url> [--p] [--h] [--r] [--serve] [--cookies '<data>']")
	}

	var inputURL string
	hasExplicitMode := false

	for i := 1; i < len(os.Args); i++ {
		arg := os.Args[i]

		switch arg {
		case "--serve":
			config.ServeUI = true
		case "--p":
			if !hasExplicitMode {
				config.DownloadPosts = true
				config.DownloadHighlights = false
				config.DownloadReels = false
				hasExplicitMode = true
			} else {
				config.DownloadPosts = true
			}
		case "--h":
			if !hasExplicitMode {
				config.DownloadHighlights = true
				config.DownloadPosts = false
				config.DownloadReels = false
				hasExplicitMode = true
			} else {
				config.DownloadHighlights = true
			}
		case "--r":
			if !hasExplicitMode {
				config.DownloadReels = true
				config.DownloadPosts = false
				config.DownloadHighlights = false
				hasExplicitMode = true
			} else {
				config.DownloadReels = true
			}
		case "--workers":
			if i+1 < len(os.Args) {
				val, err := strconv.Atoi(os.Args[i+1])
				if err == nil && val > 0 {
					config.Concurrency = val
				}
				i++
			}
		default:
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

	if config.ServeUI {
		return config, nil 
	}

	if !hasExplicitMode {
		config.DownloadPosts = true
		config.DownloadHighlights = true
		config.DownloadReels = true
	}

	if inputURL == "" {
		return config, errors.New("could not find a valid target link in your arguments")
	}

	if !strings.HasPrefix(inputURL, "http://") && !strings.HasPrefix(inputURL, "https://") {
		inputURL = "https://" + inputURL
	}

	parsed, err := url.Parse(inputURL)
	if err != nil {
		return config, fmt.Errorf("invalid URL syntax: %w", err)
	}

	host := strings.ToLower(parsed.Host)
	if !strings.Contains(host, "instagram.com") {
		return config, errors.New("link does not target instagram.com")
	}

	pathSegments := strings.Split(strings.Trim(parsed.Path, "/"), "/")
	if len(pathSegments) == 0 || pathSegments[0] == "" {
		return config, errors.New("could not isolate username from URL path")
	}

	config.Username = pathSegments[0]

	reservedKeywords := map[string]bool{
		"explore": true, "reels": true, "stories": true, "direct": true, "p": true, "accounts": true,
	}
	if reservedKeywords[strings.ToLower(config.Username)] {
		return config, fmt.Errorf("'%s' is an internal route, not an account profile", config.Username)
	}

	return config, nil
}