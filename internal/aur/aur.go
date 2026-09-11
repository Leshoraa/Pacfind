package aur

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os/exec"
	"sort"
	"strings"
	"time"

	"pacfind/internal/models"
)

var client = &http.Client{
	Timeout: 7 * time.Second,
}

// SearchAUR queries the official Arch User Repository RPC v5 API.
func SearchAUR(query string, installed map[string]bool) ([]models.Package, error) {
	apiURL := fmt.Sprintf("https://aur.archlinux.org/rpc/v5/search/%s", url.QueryEscape(query))
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return fallbackYay(query, installed)
	}
	req.Header.Set("User-Agent", "pacfind/1.1 (Arch Linux package search tool; Go)")

	resp, err := client.Do(req)
	if err != nil {
		return fallbackYay(query, installed)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fallbackYay(query, installed)
	}

	var searchResp models.AURSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return fallbackYay(query, installed)
	}

	var results []models.Package
	for _, item := range searchResp.Results {
		desc := "No description provided"
		if item.Description != nil && *item.Description != "" {
			desc = *item.Description
		}

		results = append(results, models.Package{
			Repo:        "aur",
			Name:        item.Name,
			Version:     item.Version,
			Description: desc,
			Votes:       item.NumVotes,
			Popularity:  item.Popularity,
			IsInstalled: installed[item.Name],
			OutOfDate:   item.OutOfDate != nil,
		})
	}

	// Sort results: exact match first, then prefix, then popularity descending
	qLower := strings.ToLower(query)
	sort.Slice(results, func(i, j int) bool {
		ni := strings.ToLower(results[i].Name)
		nj := strings.ToLower(results[j].Name)

		if ni == qLower && nj != qLower {
			return true
		}
		if nj == qLower && ni != qLower {
			return false
		}

		preI := strings.HasPrefix(ni, qLower)
		preJ := strings.HasPrefix(nj, qLower)
		if preI && !preJ {
			return true
		}
		if preJ && !preI {
			return false
		}

		if results[i].Popularity != results[j].Popularity {
			return results[i].Popularity > results[j].Popularity
		}
		return results[i].Votes > results[j].Votes
	})

	return results, nil
}

// fallbackYay searches AUR via yay if direct HTTP fails
func fallbackYay(query string, installed map[string]bool) ([]models.Package, error) {
	cmd := exec.Command("yay", "-Ss", "--aur", query)
	out, err := cmd.Output()
	if err != nil && len(out) == 0 {
		return nil, nil
	}

	var results []models.Package
	var current *models.Package

	scanner := bufio.NewScanner(bytes.NewReader(out))
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) == 0 {
			continue
		}

		if strings.HasPrefix(line, "    ") || strings.HasPrefix(line, "\t") {
			if current != nil {
				current.Description = strings.TrimSpace(line)
				results = append(results, *current)
				current = nil
			}
		} else if strings.HasPrefix(line, "aur/") {
			parts := strings.Fields(line)
			if len(parts) == 0 {
				continue
			}
			name := strings.TrimPrefix(parts[0], "aur/")
			ver := ""
			if len(parts) > 1 {
				ver = parts[1]
			}
			current = &models.Package{
				Repo:        "aur",
				Name:        name,
				Version:     ver,
				IsInstalled: installed[name],
			}
		}
	}

	if current != nil {
		results = append(results, *current)
	}

	return results, nil
}
