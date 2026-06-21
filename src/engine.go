package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// OrchestrateEngine serves as the routing core for execution bounds
func OrchestrateEngine() {
	// Parse input profiles and flags into our abstracted configuration layer
	config, err := ParseCLIProfileURL()
	if err != nil {
		fmt.Printf("[CLI Error] %v\n", err)
		os.Exit(1)
	}

	startTime := time.Now()
	client := NewHTTPClient() 
	
	fmt.Println("==================================================")
	fmt.Printf("Starting Standalone Engine for @%s\n", config.Username)
	fmt.Println("==================================================")

	var collectiveDownloadQueue []UniversalDownloadAsset

	// -------------------------------------------------------------------------
	// MODULE 1: TIMELINE GRID POSTS (posts.go)
	// -------------------------------------------------------------------------
	if config.DownloadPosts {
		fmt.Printf("[Engine] Initiating standalone timeline sequence for @%s...\n", config.Username)
		postAssets, err := GatherAndStructurePosts(client, config.Username)
		if err != nil {
			fmt.Printf("[Engine Error] Posts extraction failed: %v\n", err)
		} else {
			for _, asset := range postAssets {
				collectiveDownloadQueue = append(collectiveDownloadQueue, UniversalDownloadAsset{
					DownloadURL: asset.DownloadURL,
					LocalPath:   filepath.Join(config.OutputDir, asset.LocalPath),
					Category:    "posts",
				})
			}
		}
	}

	// -------------------------------------------------------------------------
	// MODULE 2: STORIES HIGHLIGHT TRAYS (highlights.go)
	// -------------------------------------------------------------------------
	if config.DownloadHighlights {
		fmt.Printf("\n[Engine] Initiating standalone GraphQL highlight sequence for @%s...\n", config.Username)
		highlightAssets, err := GatherAndStructureHighlights(client, config.Username)
		if err != nil {
			fmt.Printf("[Engine Error] Highlights extraction failed: %v\n", err)
		} else {
			for _, asset := range highlightAssets {
				collectiveDownloadQueue = append(collectiveDownloadQueue, UniversalDownloadAsset{
					DownloadURL: asset.DownloadURL,
					LocalPath:   filepath.Join(config.OutputDir, asset.LocalPath),
					Category:    "highlights",
				})
			}
		}
	}

	// -------------------------------------------------------------------------
	// VERIFY AUTHENTICATION ENVIRONMENT (Required for Reels & Stories)
	// -------------------------------------------------------------------------
	exePath, err := os.Executable()
	hasCookies := false
	if err == nil {
		_, statErr := os.Stat(filepath.Join(filepath.Dir(exePath), ".env"))
		hasCookies = statErr == nil
	}

	if hasCookies {
		// -------------------------------------------------------------------------
		// MODULE 3: REELS EXTRACTION LAYER (reels.go)
		// -------------------------------------------------------------------------
		fmt.Printf("\n[Engine] Authenticated state detected. Initiating Reels extraction for @%s...\n", config.Username)
		reelAssets, err := GatherAndStructureReels(client, config.Username)
		if err != nil {
			fmt.Printf("[Engine Error] Reels extraction failed: %v\n", err)
		} else {
			for _, asset := range reelAssets {
				collectiveDownloadQueue = append(collectiveDownloadQueue, UniversalDownloadAsset{
					DownloadURL: asset.DownloadURL,
					LocalPath:   filepath.Join(config.OutputDir, asset.LocalPath),
					Category:    "reels",
				})
			}
		}

		// -------------------------------------------------------------------------
		// MODULE 4: ACTIVE STORIES EXTRACTION LAYER (stories.go)
		// -------------------------------------------------------------------------
		fmt.Printf("\n[Engine] Initiating active Stories extraction layer for @%s...\n", config.Username)
		storyAssets, err := GatherAndStructureStories(client, config.Username)
		if err != nil {
			fmt.Printf("[Engine Error] Stories extraction failed: %v\n", err)
		} else {
			for _, asset := range storyAssets {
				collectiveDownloadQueue = append(collectiveDownloadQueue, UniversalDownloadAsset{
					DownloadURL: asset.DownloadURL,
					LocalPath:   filepath.Join(config.OutputDir, asset.LocalPath),
					Category:    "stories",
				})
			}
		}
	} else {
		fmt.Println("\n[Engine] Skipping Authenticated Modules (Reels & Stories) because .env file is missing.")
	}

	// -------------------------------------------------------------------------
	// DOWNLOAD HANDOVER: CONCURRENT WORKERS (downloader.go)
	// -------------------------------------------------------------------------
	if len(collectiveDownloadQueue) == 0 {
		fmt.Println("\n[Engine Execution Complete] No valid asset streams gathered for processing.")
		return
	}

	fmt.Printf("\n[Engine] Handing over %d structured assets to Concurrent Downloader...\n", len(collectiveDownloadQueue))
	
	// Pass the dynamically configured concurrency value instead of the hardcoded literal
	ConcurrentDownloadPool(client, collectiveDownloadQueue, config.Concurrency)

	fmt.Println("==================================================")
	fmt.Printf("Execution Completed in: %v\n", time.Since(startTime))
	fmt.Println("==================================================")
}