package main

import (
	"bufio"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// NewHTTPClient returns a configured client with dynamic cookies loaded from .env
func NewHTTPClient() *http.Client {
	jar, _ := cookiejar.New(nil)
	targetURL, _ := url.Parse("https://www.instagram.com")

	// Locate the .env file in the true binary execution location folder bounds
	exePath, err := os.Executable()
	if err == nil {
		envFilePath := filepath.Join(filepath.Dir(exePath), ".env")
		
		if file, err := os.Open(envFilePath); err == nil {
			defer file.Close()
			var cookies []*http.Cookie
			scanner := bufio.NewScanner(file)

			// Read line-by-line to prevent structural memory parsing issues
			for scanner.Scan() {
				line := strings.TrimSpace(scanner.Text())
				if line == "" || strings.HasPrefix(line, "#") {
					continue
				}

				parts := strings.SplitN(line, "=", 2)
				if len(parts) == 2 {
					name := strings.TrimSpace(parts[0])
					val := strings.TrimSpace(parts[1])
					
					if name != "" {
						cookies = append(cookies, &http.Cookie{
							Name:   name,
							Value:  val,
							Domain: ".instagram.com",
							Path:   "/",
						})
					}
				}
			}

			if len(cookies) > 0 {
				jar.SetCookies(targetURL, cookies)
			}
		}
	}

	return &http.Client{
		Jar:     jar,
		Timeout: 15 * time.Second,
	}
}

// SetDefaultHeaders applies required signatures along with dynamic verification token interceptions
func SetDefaultHeaders(req *http.Request, username string) {
	req.Header.Set("User-Agent", UserAgent)
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("X-IG-App-ID", AppID)
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("Referer", "https://www.instagram.com/"+username+"/")

	// Intercept and assign the active transaction frame authentication token if present
	if csrfCookie, err := req.Cookie("csrftoken"); err == nil && csrfCookie.Value != "" {
		req.Header.Set("X-CSRFToken", csrfCookie.Value)
	}
}