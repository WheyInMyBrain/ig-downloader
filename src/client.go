package main

import (
	"net/http"
	"time"
)

// NewHTTPClient returns a configured client with a standard timeout
func NewHTTPClient() *http.Client {
	return &http.Client{
		Timeout: 15 * time.Second,
	}
}

// SetDefaultHeaders applies the required web signatures to an outbound request using global constants
func SetDefaultHeaders(req *http.Request, username string) {
	req.Header.Set("User-Agent", UserAgent)
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("X-IG-App-ID", AppID)
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("Referer", "https://www.instagram.com/"+username+"/")
}