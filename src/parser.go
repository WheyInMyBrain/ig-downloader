package main

import (
	"encoding/json"
	"strings"
)

// Candidate represents a single media layout URL resource node
type Candidate struct {
	URL string `json:"url"`
}

// ImageVersions holds the bounding box resolution options for photos
type ImageVersions struct {
	Candidates []Candidate `json:"candidates"`
}

// VideoVersion holds the raw file access path for videos and reels
type VideoVersion struct {
	URL string `json:"url"`
}

// CarouselItem handles the individual nested slide structures
type CarouselItem struct {
	MediaType      int           `json:"media_type"`
	ImageVersions2 ImageVersions `json:"image_versions2"`
	VideoVersions  []VideoVersion `json:"video_versions"`
}

// FeedItem maps the core properties of an Instagram asset layout
type FeedItem struct {
	ID             string         `json:"id"`
	Code           string         `json:"code"`
	MediaType      int            `json:"media_type"`
	ImageVersions2 ImageVersions  `json:"image_versions2"`
	VideoVersions  []VideoVersion `json:"video_versions"`
	CarouselMedia  []CarouselItem `json:"carousel_media"`
}

// InstagramFeedResponse is the top-level wrapper for timeline and media API responses
type InstagramFeedResponse struct {
	Items     []FeedItem `json:"items"`
	NextMaxID string     `json:"next_max_id"`
}

// SanitizeURL fixes the HTML escaping signature mismatch that triggers the "Bad URL Hash" error
func SanitizeURL(rawURL string) string {
	return strings.ReplaceAll(rawURL, "&amp;", "&")
}

// DecodeFeedJSON simply unmarshals raw bytes into our structured response layout object
func DecodeFeedJSON(bodyBytes []byte) (*InstagramFeedResponse, error) {
	var resp InstagramFeedResponse
	if err := json.Unmarshal(bodyBytes, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}