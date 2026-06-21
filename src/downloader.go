package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
)

// UniversalDownloadAsset wraps both post and highlight structured assets into a unified signature for the download workers
type UniversalDownloadAsset struct {
	DownloadURL string
	LocalPath   string
}

// ConcurrentDownloadPool manages a fixed set of persistent consumer goroutines feeding from a task channel
func ConcurrentDownloadPool(client *http.Client, assets []UniversalDownloadAsset, concurrencyLimit int) {
	fmt.Printf("\n[*] Starting Fixed Worker Pool (Spawning Exactly %d Persistent GoRoutines)...\n", concurrencyLimit)

	var wg sync.WaitGroup
	assetsChan := make(chan UniversalDownloadAsset, len(assets))

	// 1. Spawn exactly N fixed consumer worker goroutines
	for workerID := 1; workerID <= concurrencyLimit; workerID++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			// Workers run continuously block-waiting for jobs until assetsChan closes
			for asset := range assetsChan {
				err := executeAssetDownload(client, asset)
				if err != nil {
					fmt.Printf("[Worker #%d] Failed downloading %s: %v\n", id, asset.LocalPath, err)
				}
			}
		}(workerID)
	}

	// 2. Producer: Push all compiled tasks into the queue channel
	for _, asset := range assets {
		if asset.DownloadURL != "" && asset.LocalPath != "" {
			assetsChan <- asset
		}
	}
	
	// 3. Close channel to cleanly notify consumer routines to break their loops when empty
	close(assetsChan)

	// 4. Block wait until the workers clear out the queue completely
	wg.Wait()
	fmt.Println("\n[SUCCESS] Fixed download worker pool queue synchronization complete.")
}

// executeAssetDownload creates deep directories dynamically and commits the raw bytes directly to disk
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
		return fmt.Errorf("CDN endpoint dropped connection frame code: %d", resp.StatusCode)
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