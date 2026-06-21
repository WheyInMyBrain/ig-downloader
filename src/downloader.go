package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
)

type UniversalDownloadAsset struct {
	DownloadURL string
	LocalPath   string
	Category    string // "posts", "highlights", "stories", etc.
}

// ProgressEvent broadcasts real-time tracking metrics to the SSE server web client
type ProgressEvent struct {
	Category string `json:"category"`
	Type     string `json:"type"` // "init" (total discovered) or "progress" (increment)
	Value    int    `json:"value"`
}

var WebProgressChan = make(chan ProgressEvent, 100)

func ConcurrentDownloadPool(client *http.Client, assets []UniversalDownloadAsset, concurrencyLimit int) {
	var wg sync.WaitGroup
	assetsChan := make(chan UniversalDownloadAsset, len(assets))

	for workerID := 1; workerID <= concurrencyLimit; workerID++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for asset := range assetsChan {
				err := executeAssetDownload(client, asset)
				if err == nil {
					// Broadcast an increment step to the live UI
					select {
					case WebProgressChan <- ProgressEvent{Category: asset.Category, Type: "progress", Value: 1}:
					default:
					}
				} else {
					fmt.Printf("[Worker #%d] Failed downloading %s: %v\n", id, asset.LocalPath, err)
				}
			}
		}(workerID)
	}

	for _, asset := range assets {
		if asset.DownloadURL != "" && asset.LocalPath != "" {
			assetsChan <- asset
		}
	}
	close(assetsChan)
	wg.Wait()
}

func executeAssetDownload(client *http.Client, asset UniversalDownloadAsset) error {
	outputDir := filepath.Dir(asset.LocalPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed creating directory layers: %w", err)
	}

	req, err := http.NewRequest("GET", asset.DownloadURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", UserAgent)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("CDN endpoint dropped link: %d", resp.StatusCode)
	}

	out, err := os.Create(asset.LocalPath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	fmt.Printf("[+] Saved Local Target: %s\n", asset.LocalPath)
	return nil
}