package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// TargetHighlightAsset binds the direct download source with its localized disk path layout
type TargetHighlightAsset struct {
	DownloadURL string
	LocalPath   string // Structured as: username/highlights/highlight_title/file.ext
}

// Structural mappings for Step 2: GraphQL Highlight List Payload
type GraphQLHighlightResponse struct {
	Data struct {
		User struct {
			EdgeHighlightReels struct {
				Edges []struct {
					Node struct {
						ID    string `json:"id"`
						Title string `json:"title"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"edge_highlight_reels"`
		} `json:"user"`
	} `json:"data"`
}

// ReelsMediaResponse maps Meta's internal media container layouts for story nodes
type ReelsMediaResponse struct {
	Reels map[string]struct {
		Items []struct {
			MediaType     int `json:"media_type"`
			VideoVersions []struct {
				URL string `json:"url"`
			} `json:"video_versions"`
			ImageVersions2 struct {
				Candidates []struct {
					URL string `json:"url"`
				} `json:"candidates"`
			} `json:"image_versions2"`
		} `json:"items"`
	} `json:"reels"`
}

// FetchNumericUserID queries the web profile endpoint to resolve a username to its internal numeric ID
func FetchNumericUserID(client *http.Client, username string) (string, error) {
	url := fmt.Sprintf("https://www.instagram.com/api/v1/users/web_profile_info/?username=%s", username)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	SetDefaultHeaders(req, username)

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("profile info request failed with status: %d", resp.StatusCode)
	}

	var result struct {
		Data struct {
			User struct {
				ID string `json:"id"`
			} `json:"user"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	return result.Data.User.ID, nil
}

// GatherAndStructureHighlights queries the GraphQL tray endpoint and structures the target download paths
func GatherAndStructureHighlights(client *http.Client, username string) ([]TargetHighlightAsset, error) {
	var structuredAssets []TargetHighlightAsset

	// Step 1: Secure numerical profile tracker ID using web profile metrics
	userID, err := FetchNumericUserID(client, username)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve user numeric id: %w", err)
	}
	fmt.Printf("[+] Target User ID Resolved: %s\n", userID)

	// Step 2: Build the static GraphQL request context
	req, err := http.NewRequest("GET", "https://www.instagram.com/graphql/query/", nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("query_id", StoryGraphQLQueryID) // Pulled directly from constants.go binary state
	q.Add("user_id", userID)
	q.Add("include_chaining", "false")
	q.Add("include_reel", "true")
	q.Add("include_suggested_users", "false")
	q.Add("include_logged_out_extras", "true")
	q.Add("include_live_status", "false")
	q.Add("include_highlight_reels", "true")
	req.URL.RawQuery = q.Encode()

	SetDefaultHeaders(req, username)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var gqlResp GraphQLHighlightResponse
	if err := json.NewDecoder(resp.Body).Decode(&gqlResp); err != nil {
		return nil, fmt.Errorf("failed decoding GraphQL response body: %w", err)
	}

	edges := gqlResp.Data.User.EdgeHighlightReels.Edges
	if len(edges) == 0 {
		fmt.Printf("[*] Profile @%s has 0 active highlight trays pinned.\n", username)
		return nil, nil
	}

	fmt.Printf("[+] Located %d highlight bubbles. Requesting media content arrays...\n", len(edges))

	// Step 3: Loop through discovered highlights and call their reels_media payload maps
	for _, edge := range edges {
		title := edge.Node.Title
		reelID := edge.Node.ID
		if reelID == "" {
			continue
		}

		fmt.Printf("  -> Extracting Bubble: '%s' (ID: %s)\n", title, reelID)
		targetSubFolder := fmt.Sprintf("%s/highlights/%s", username, title)
		reelTargetKey := fmt.Sprintf("highlight:%s", reelID)

		// Request the direct media links for this specific highlight cluster
		apiURL := fmt.Sprintf("https://www.instagram.com/api/v1/feed/reels_media/?reel_ids=%s", reelTargetKey)
		mReq, err := http.NewRequest("GET", apiURL, nil)
		if err != nil {
			return nil, err
		}
		SetDefaultHeaders(mReq, username)

		mResp, err := client.Do(mReq)
		if err != nil {
			fmt.Printf("    [!] Error hitting media api for highlight '%s': %v\n", title, err)
			continue
		}

		var mediaData ReelsMediaResponse // Shared decoding schema mapped inside downloader/parser global layout
		decodeErr := json.NewDecoder(mResp.Body).Decode(&mediaData)
		mResp.Body.Close()
		if decodeErr != nil {
			fmt.Printf("    [!] Error parsing JSON streams for highlight '%s': %v\n", title, decodeErr)
			continue
		}

		items := mediaData.Reels[reelTargetKey].Items
		for itemIdx, item := range items {
			if item.MediaType == 2 && len(item.VideoVersions) > 0 {
				structuredAssets = append(structuredAssets, TargetHighlightAsset{
					DownloadURL: SanitizeURL(item.VideoVersions[0].URL),
					LocalPath:   fmt.Sprintf("%s/slide_%d.mp4", targetSubFolder, itemIdx+1),
				})
			} else if len(item.ImageVersions2.Candidates) > 0 {
				structuredAssets = append(structuredAssets, TargetHighlightAsset{
					DownloadURL: SanitizeURL(item.ImageVersions2.Candidates[0].URL),
					LocalPath:   fmt.Sprintf("%s/slide_%d.jpg", targetSubFolder, itemIdx+1),
				})
			}
		}

		// Stream the dynamic running totals live to the web page UI
		select {
		case WebProgressChan <- ProgressEvent{Category: "highlights", Type: "init_update", Value: len(structuredAssets)}:
		default:
		}
	}

	return structuredAssets, nil
}