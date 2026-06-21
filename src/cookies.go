package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// BrowserCookie JSON scheme mapping for extensions like EditThisCookie
type BrowserCookie struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// unescapeOctalCookies translates cookie string sequences like \054 cleanly into actual character values
func unescapeOctalCookies(val string) string {
	// First strip wrapping double-quotes completely if present
	val = strings.Trim(val, "\"")

	// Look for octal sequences (e.g., \054) and swap them out safely
	pos := 0
	for {
		idx := strings.Index(val[pos:], "\\")
		if idx == -1 {
			break
		}
		start := pos + idx
		if start+4 <= len(val) {
			octalCandidate := val[start+1 : start+4]
			if parsedChar, err := strconv.ParseInt(octalCandidate, 8, 8); err == nil {
				val = val[:start] + string(rune(parsedChar)) + val[start+4:]
				pos = start + 1
				continue
			}
		}
		pos = start + 1
	}
	return val
}

// ParseAndSaveCookies reads either a JSON array or a raw header string and commits it to .env
func ParseAndSaveCookies(rawInput string) error {
	rawInput = strings.TrimSpace(rawInput)
	if rawInput == "" {
		return fmt.Errorf("cookie input string cannot be empty")
	}

	var cookiesMap = make(map[string]string)

	// Route 1: Handle JSON Array configuration strings
	if strings.HasPrefix(rawInput, "[") {
		var list []BrowserCookie
		if err := json.Unmarshal([]byte(rawInput), &list); err != nil {
			return fmt.Errorf("failed decoding cookie JSON structure: %w", err)
		}
		for _, c := range list {
			if c.Name != "" {
				cookiesMap[c.Name] = unescapeOctalCookies(c.Value)
			}
		}
	} else {
		// Route 2: Handle raw key=value; semicolon-separated header format strings
		segments := strings.Split(rawInput, ";")
		for _, segment := range segments {
			segment = strings.TrimSpace(segment)
			if segment == "" {
				continue
			}
			parts := strings.SplitN(segment, "=", 2)
			if len(parts) == 2 {
				name := strings.TrimSpace(parts[0])
				val := strings.TrimSpace(parts[1])
				if name != "" {
					cookiesMap[name] = unescapeOctalCookies(val)
				}
			}
		}
	}

	if len(cookiesMap) == 0 {
		return fmt.Errorf("no valid target cookie tokens could be parsed from input context")
	}

	// Format tokens directly into standard KEY=VALUE environment layout lines
	var envLines []string
	for k, v := range cookiesMap {
		envLines = append(envLines, fmt.Sprintf("%s=%s", k, v))
	}

	// Isolate the true binary runtime destination folder bounds
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to map execution tracking path: %w", err)
	}
	
	envFilePath := filepath.Join(filepath.Dir(exePath), ".env")
	err = os.WriteFile(envFilePath, []byte(strings.Join(envLines, "\n")+"\n"), 0644)
	if err != nil {
		return fmt.Errorf("failed writing configuration file to disk target: %w", err)
	}

	fmt.Printf("[SUCCESS] Saved %d sanitized session keys directly to: %s\n", len(cookiesMap), envFilePath)
	return nil
}