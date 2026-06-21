package main

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

// TargetPostAsset binds the direct download source with its localized disk path layout
type TargetPostAsset struct {
	DownloadURL string
	LocalPath   string // Structured as: username/posts/shortcode/file.ext
}

// GatherAndStructurePosts runs the entire profile grid scraping loop and structures the assets for downloading
func GatherAndStructurePosts(client *http.Client, username string) ([]TargetPostAsset, error) {
	var structuredAssets []TargetPostAsset
	nextMaxID := ""
	hasMore := true
	page := 1

	baseURL := fmt.Sprintf("https://www.instagram.com/api/v1/feed/user/%s/username/", username)

	for hasMore {
		fmt.Printf("[*] Scanning Feed Page %d for @%s...\n", page, username)

		targetURL := fmt.Sprintf("%s?count=12", baseURL)
		if nextMaxID != "" {
			targetURL = fmt.Sprintf("%s&max_id=%s", targetURL, nextMaxID)
		}

		req, err := http.NewRequest("GET", targetURL, nil)
		if err != nil {
			return nil, err
		}
		SetDefaultHeaders(req, username)

		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return nil, fmt.Errorf("instagram rejected request. HTTP Status: %d", resp.StatusCode)
		}

		bodyBytes, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, err
		}

		// Decode the raw JSON bytes using our parser module mapping
		feedData, err := DecodeFeedJSON(bodyBytes)
		if err != nil {
			return nil, fmt.Errorf("failed parsing JSON payload: %w", err)
		}

		if len(feedData.Items) == 0 {
			fmt.Println("[*] No items returned in this frame. Stream boundary reached.")
			break
		}

		// Process each item to build structural paths instead of a flat unmapped list
		pageAssetsCount := 0
		for _, item := range feedData.Items {
			shortcode := item.Code
			if shortcode == "" {
				continue
			}

			// Subfolder destination layout requirement: username/posts/shortcode
			targetSubFolder := fmt.Sprintf("%s/posts/%s", username, shortcode)

			// Scenario 1: Multi-slide Carousels
			if len(item.CarouselMedia) > 0 {
				for slideIdx, slide := range item.CarouselMedia {
					if len(slide.VideoVersions) > 0 {
						structuredAssets = append(structuredAssets, TargetPostAsset{
							DownloadURL: SanitizeURL(slide.VideoVersions[0].URL),
							LocalPath:   fmt.Sprintf("%s/slide_%d.mp4", targetSubFolder, slideIdx+1),
						})
						pageAssetsCount++
					} else if len(slide.ImageVersions2.Candidates) > 0 {
						structuredAssets = append(structuredAssets, TargetPostAsset{
							DownloadURL: SanitizeURL(slide.ImageVersions2.Candidates[0].URL),
							LocalPath:   fmt.Sprintf("%s/slide_%d.jpg", targetSubFolder, slideIdx+1),
						})
						pageAssetsCount++
					}
				}
			// Scenario 2: Single Videos / Reels
			} else if item.MediaType == 2 && len(item.VideoVersions) > 0 {
				structuredAssets = append(structuredAssets, TargetPostAsset{
					DownloadURL: SanitizeURL(item.VideoVersions[0].URL),
					LocalPath:   fmt.Sprintf("%s/%s.mp4", targetSubFolder, shortcode),
				})
				pageAssetsCount++
			// Scenario 3: Standard Static Images
			} else if len(item.ImageVersions2.Candidates) > 0 {
				structuredAssets = append(structuredAssets, TargetPostAsset{
					DownloadURL: SanitizeURL(item.ImageVersions2.Candidates[0].URL),
					LocalPath:   fmt.Sprintf("%s/%s.jpg", targetSubFolder, shortcode),
				})
				pageAssetsCount++
			}
		}

		fmt.Printf("    -> Structured %d assets on this page. (Total Pooled: %d)\n", pageAssetsCount, len(structuredAssets))

		// Stream the structural updates live to the webpage if running via UI server
		select {
		case WebProgressChan <- ProgressEvent{Category: "posts", Type: "init_update", Value: len(structuredAssets)}:
		default:
		}

		nextMaxID = feedData.NextMaxID
		hasMore = nextMaxID != ""

		if hasMore {
			page++
			time.Sleep(2500 * time.Millisecond) // Protective human delay interval
		}
	}

	return structuredAssets, nil
}