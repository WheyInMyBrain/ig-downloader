package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// TargetReelAsset maps the direct video and thumbnail links into the engine configuration signature
type TargetReelAsset struct {
	DownloadURL string
	LocalPath   string
}

// Structural schema mapping for the GraphQL Reels payload response query bounds
type GraphQLReelsResponse struct {
	Data struct {
		XDTClipsUserConnectionV2 struct {
			Edges []struct {
				Node struct {
					Media struct {
						PK   string `json:"pk"`
						Code string `json:"code"`
					} `json:"media"`
				} `json:"node"`
			} `json:"edges"`
			PageInfo struct {
				EndCursor   string `json:"end_cursor"`
				HasNextPage bool   `json:"has_next_page"`
			} `json:"page_info"`
		} `json:"xdt_api__v1__clips__user__connection_v2"`
	} `json:"data"`
}

// MediaInfoResponse leverages parser.go shared components for absolute format sync
type MediaInfoResponse struct {
	Items []struct {
		VideoVersions  []VideoVersion `json:"video_versions"`
		ImageVersions2 ImageVersions  `json:"image_versions2"`
	} `json:"items"`
}

// FetchRealMediaLinks hits the specific media item container endpoint to extract high-res source streams
func FetchRealMediaLinks(client *http.Client, username string, mediaPK string) (string, string, error) {
	infoURL := fmt.Sprintf("https://www.instagram.com/api/v1/media/%s/info/", mediaPK)
	req, err := http.NewRequest("GET", infoURL, nil)
	if err != nil {
		return "", "", err
	}
	SetDefaultHeaders(req, username)

	resp, err := client.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("media info returned frame status code: %d", resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", err
	}

	var infoData MediaInfoResponse
	if err := json.Unmarshal(bodyBytes, &infoData); err != nil {
		return "", "", err
	}

	var videoURL, coverURL string
	if len(infoData.Items) > 0 {
		item := infoData.Items[0]
		
		if len(item.VideoVersions) > 0 {
			// Loop through candidates to locate the fully multiplexed stream tracker
			selectedVideoStr := item.VideoVersions[0].URL
			for _, v := range item.VideoVersions {
				if strings.Contains(v.URL, "vs=") || strings.Contains(v.URL, "_nc_vs=") {
					selectedVideoStr = v.URL
					break
				}
			}
			
			// Format clean slashes 
			rawURL := strings.ReplaceAll(selectedVideoStr, `\/`, "/")
			
			// ESCAPE ENGINE FIX: Parse into a true system url object to seal query properties 
			parsedObj, err := url.Parse(rawURL)
			if err == nil {
				videoURL = parsedObj.String()
			} else {
				videoURL = rawURL // Fallback if parse fails
			}
		}
		
		if len(item.ImageVersions2.Candidates) > 0 {
			rawCoverStr := strings.ReplaceAll(item.ImageVersions2.Candidates[0].URL, `\/`, "/")
			parsedCover, err := url.Parse(rawCoverStr)
			if err == nil {
				coverURL = parsedCover.String()
			} else {
				coverURL = rawCoverStr
			}
		}
	}

	return videoURL, coverURL, nil
}

// GatherAndStructureReels tracks the profile Reels graph network and maps source items out for concurrent worker pools
func GatherAndStructureReels(client *http.Client, username string) ([]TargetReelAsset, error) {
	var structuredAssets []TargetReelAsset

	userID, err := FetchNumericUserID(client, username)
	if err != nil {
		return nil, fmt.Errorf("failed to extract profile tracking metrics layout: %w", err)
	}

	currentCursor := ""
	hasNextPage := true
	page := 1
	targetURL := "https://www.instagram.com/graphql/query"

	for hasNextPage {
		fmt.Printf("[*] Scanning Reels Feed Page %d for @%s...\n", page, username)

		variables := map[string]interface{}{
			"after":  nil,
			"before": nil,
			"first":  12,
			"last":   nil,
			"data": map[string]interface{}{
				"include_feed_video": true,
				"page_size":          12,
				"target_user_id":     userID,
			},
		}
		if currentCursor != "" {
			variables["after"] = currentCursor
		}

		varsBytes, _ := json.Marshal(variables)

		formData := url.Values{}
		formData.Set("lsd", FallbackLSD)
		formData.Set("fb_api_caller_class", "RelayModern")
		formData.Set("fb_api_req_friendly_name", "PolarisProfileReelsTabContentQuery")
		formData.Set("variables", string(varsBytes))
		formData.Set("doc_id", ReelsGraphQLDocID)

		req, err := http.NewRequest("POST", targetURL, bytes.NewBufferString(formData.Encode()))
		if err != nil {
			return nil, err
		}
		
		SetDefaultHeaders(req, username)
		
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("X-FB-Friendly-Name", "PolarisProfileReelsTabContentQuery")
		req.Header.Set("X-FB-LSD", FallbackLSD)
		req.Header.Set("X-ASBD-ID", ASBDID)
		req.Header.Set("Origin", "https://www.instagram.com")
		req.Header.Set("Referer", fmt.Sprintf("https://www.instagram.com/%s/reels/", username))

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

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return nil, fmt.Errorf("reels query configuration rejected code: %d", resp.StatusCode)
		}

		var gqlResp GraphQLReelsResponse
		decodeErr := json.NewDecoder(resp.Body).Decode(&gqlResp)
		resp.Body.Close()
		if decodeErr != nil {
			return nil, fmt.Errorf("failed decoding query stream fields layout: %w", decodeErr)
		}

		connection := gqlResp.Data.XDTClipsUserConnectionV2
		edges := connection.Edges
		if len(edges) == 0 {
			break
		}

		for _, edge := range edges {
			code := edge.Node.Media.Code
			pk := edge.Node.Media.PK
			if pk == "" || code == "" {
				continue
			}

			videoURL, coverURL, err := FetchRealMediaLinks(client, username, pk)
			if err != nil {
				fmt.Printf("    [!] Skipping Reel node frame content tracking for shortcode '%s': %v\n", code, err)
				continue
			}

			targetSubFolder := fmt.Sprintf("%s/reels/%s", username, code)

			if videoURL != "" {
				structuredAssets = append(structuredAssets, TargetReelAsset{
					DownloadURL: videoURL,
					LocalPath:   fmt.Sprintf("%s/%s.mp4", targetSubFolder, code),
				})
			}
			if coverURL != "" {
				structuredAssets = append(structuredAssets, TargetReelAsset{
					DownloadURL: coverURL,
					LocalPath:   fmt.Sprintf("%s/%s.jpg", targetSubFolder, code),
				})
			}

			time.Sleep(500 * time.Millisecond)
		}

		reelsCount := len(structuredAssets) / 2

		select {
		case WebProgressChan <- ProgressEvent{Category: "reels", Type: "init_update", Value: reelsCount}:
		default:
		}

		currentCursor = connection.PageInfo.EndCursor
		hasNextPage = connection.PageInfo.HasNextPage && currentCursor != ""

		if hasNextPage {
			page++
			time.Sleep(500 * time.Millisecond)
		}
	}

	return structuredAssets, nil
}