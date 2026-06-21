package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// TargetStoryAsset maps isolated live timeline frame paths to the download coordinator queues
type TargetStoryAsset struct {
	DownloadURL string
	LocalPath   string
}

// Structural layout mappings mirroring the exact reels_media response dictionary
type GraphQLStoriesResponse struct {
	Data struct {
		XDTFeedReelsMedia struct {
			ReelsMedia []struct {
				ID    string `json:"id"`
				Items []struct {
					ID             string         `json:"id"`
					Code           string         `json:"code"`
					MediaType      int            `json:"media_type"`
					VideoVersions  []VideoVersion `json:"video_versions"`
					ImageVersions2 ImageVersions  `json:"image_versions2"`
				} `json:"items"`
			} `json:"reels_media"`
		} `json:"xdt_api__v1__feed__reels_media"`
	} `json:"data"`
}

// GatherAndStructureStories hits Meta's PolarisStories V3 endpoint to map active profile updates
func GatherAndStructureStories(client *http.Client, username string) ([]TargetStoryAsset, error) {
	var structuredAssets []TargetStoryAsset

	// Isolate numeric user identifier tracking index via standard profile mappings
	userID, err := FetchNumericUserID(client, username)
	if err != nil {
		return nil, fmt.Errorf("failed to extract story target metrics layout: %w", err)
	}

	fmt.Printf("[*] Scanning active story timeline tray for @%s...\n", username)
	targetURL := "https://www.instagram.com/graphql/query"

	// Construct structural variables array (FIXED: lowercase false to satisfy strict Go syntax)
	variables := map[string]interface{}{
		"reel_ids_arr": []string{userID},
		"__relay_internal__pv__PolarisCommunityNoteStoriesLabelEnabledrelayprovider": false,
		"__relay_internal__pv__PolarisAIGMMediaWebLabelEnabledrelayprovider":        false,
	}

	varsBytes, _ := json.Marshal(variables)

	// Replicate simplified direct form data mapping structures
	formData := url.Values{}
	formData.Set("lsd", FallbackLSD)
	formData.Set("fb_api_caller_class", "RelayModern")
	formData.Set("fb_api_req_friendly_name", "PolarisStoriesV3ReelPageStandaloneQuery")
	formData.Set("variables", string(varsBytes))
	formData.Set("doc_id", StoryGraphQLDocID)

	req, err := http.NewRequest("POST", targetURL, bytes.NewBufferString(formData.Encode()))
	if err != nil {
		return nil, err
	}

	// Apply default user agent configurations safely from client.go
	SetDefaultHeaders(req, username)

	// Re-verify explicit transaction context metadata headers matching your raw network stream
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("X-FB-Friendly-Name", "PolarisStoriesV3ReelPageStandaloneQuery")
	req.Header.Set("X-Root-Field-Name", "xdt_api__v1__feed__reels_media")
	req.Header.Set("X-FB-LSD", FallbackLSD)
	req.Header.Set("X-ASBD-ID", ASBDID)
	req.Header.Set("Origin", "https://www.instagram.com")
	req.Header.Set("Referer", fmt.Sprintf("https://www.instagram.com/%s/", username))

	// Re-map active authentication strings out of current jar to satisfy CSRF protection parameters
	if client.Jar != nil {
		cookies := client.Jar.Cookies(req.URL)
		for _, cookie := range cookies {
			req.AddCookie(cookie)
			if cookie.Name == "csrftoken" {
				req.Header.Set("X-CSRFToken", cookie.Value)
			}
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("stories query context rejected with state code: %d", resp.StatusCode)
	}

	var gqlResp GraphQLStoriesResponse
	if err := json.NewDecoder(resp.Body).Decode(&gqlResp); err != nil {
		return nil, fmt.Errorf("failed to parse story stream fields layouts: %w", err)
	}

	reelsMedia := gqlResp.Data.XDTFeedReelsMedia.ReelsMedia
	if len(reelsMedia) == 0 {
		return nil, nil // Return empty clean stack slice if profile timeline contains no active stories
	}

	for _, reel := range reelsMedia {
		for _, item := range reel.Items {
			code := item.Code
			if code == "" {
				code = item.ID
			}

			targetSubFolder := fmt.Sprintf("%s/stories", username)
			var downloadURL string

			if item.MediaType == 2 && len(item.VideoVersions) > 0 {
				// Handle video stories progressive multiplex trackers cleanly 
				rawVideoStr := item.VideoVersions[0].URL
				downloadURL = strings.ReplaceAll(rawVideoStr, `\/`, "/")
				
				if downloadURL != "" {
					structuredAssets = append(structuredAssets, TargetStoryAsset{
						DownloadURL: downloadURL,
						LocalPath:   fmt.Sprintf("%s/%s.mp4", targetSubFolder, code),
					})
				}
			} else if len(item.ImageVersions2.Candidates) > 0 {
				// Handle classic photo status items
				rawCoverStr := item.ImageVersions2.Candidates[0].URL
				downloadURL = strings.ReplaceAll(rawCoverStr, `\/`, "/")
				
				if downloadURL != "" {
					structuredAssets = append(structuredAssets, TargetStoryAsset{
						DownloadURL: downloadURL,
						LocalPath:   fmt.Sprintf("%s/%s.jpg", targetSubFolder, code),
					})
				}
			}
		}
	}

	// Update live frontend display progress trackers via your global sync channel
	select {
	case WebProgressChan <- ProgressEvent{Category: "stories", Type: "init_update", Value: len(structuredAssets)}:
	default:
	}

	return structuredAssets, nil
}