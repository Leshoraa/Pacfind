package models

// Package represents a software package in either Official Repos or AUR.
type Package struct {
	Repo        string  // "core", "extra", "multilib", or "aur"
	Name        string  // package name
	Version     string  // package version
	Description string  // package description
	Votes       int     // AUR votes
	Popularity  float64 // AUR popularity
	IsInstalled bool    // whether package is locally installed
	OutOfDate   bool    // whether AUR package is flagged out of date
}

// AURPackage represents the JSON structure returned by AUR RPC v5.
type AURPackage struct {
	Name        string   `json:"Name"`
	Version     string   `json:"Version"`
	Description *string  `json:"Description"`
	NumVotes    int      `json:"NumVotes"`
	Popularity  float64  `json:"Popularity"`
	OutOfDate   *int64   `json:"OutOfDate"`
}

// AURSearchResponse represents the top-level AUR RPC response.
type AURSearchResponse struct {
	Version     int          `json:"version"`
	Type        string       `json:"type"`
	ResultCount int          `json:"resultcount"`
	Results     []AURPackage `json:"results"`
}
