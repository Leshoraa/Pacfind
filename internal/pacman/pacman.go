package pacman

import (
	"bufio"
	"bytes"
	"os/exec"
	"sort"
	"strings"

	"pacfind/internal/models"
)

// GetInstalled returns a map of all currently installed package names.
func GetInstalled() map[string]bool {
	installed := make(map[string]bool)
	cmd := exec.Command("pacman", "-Qq")
	out, err := cmd.Output()
	if err != nil {
		return installed
	}

	scanner := bufio.NewScanner(bytes.NewReader(out))
	for scanner.Scan() {
		pkg := strings.TrimSpace(scanner.Text())
		if pkg != "" {
			installed[pkg] = true
		}
	}
	return installed
}

// SearchOfficial queries pacman -Ss for official packages.
func SearchOfficial(query string) ([]models.Package, error) {
	cmd := exec.Command("pacman", "-Ss", query)
	cmd.Env = append(cmd.Environ(), "LANG=C")
	out, err := cmd.Output()
	// pacman returns exit code 1 when no packages are found
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
		} else {
			parts := strings.Fields(line)
			if len(parts) == 0 {
				continue
			}

			repoName := parts[0]
			repo := "official"
			name := repoName
			if slashIdx := strings.Index(repoName, "/"); slashIdx != -1 {
				repo = repoName[:slashIdx]
				name = repoName[slashIdx+1:]
			}

			version := ""
			if len(parts) > 1 {
				version = parts[1]
			}

			isInstalled := false
			for _, p := range parts[2:] {
				if strings.Contains(p, "installed") {
					isInstalled = true
					break
				}
			}

			current = &models.Package{
				Repo:        repo,
				Name:        name,
				Version:     version,
				IsInstalled: isInstalled,
			}
		}
	}

	if current != nil {
		results = append(results, *current)
	}

	// Sort results: exact match first, then prefix, then alphabetical
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

		return ni < nj
	})

	return results, nil
}
