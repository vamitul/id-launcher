//go:build darwin

package main

import (
	"fmt"
	"os"
	"strings"
)

// findAllInstalledVersions is the macOS implementation.
// It scans /Applications for "Adobe InDesign XXXX" folders.
func findAllInstalledVersions() (map[uint32]string, error) {
	found := make(map[uint32]string)
	
	// Read the /Applications directory
	entries, err := os.ReadDir("/Applications")
	if err != nil {
		return nil, fmt.Errorf("could not read /Applications: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		
		dirName := entry.Name()
		// Check if it's an InDesign folder
		if strings.HasPrefix(dirName, "Adobe InDesign ") {

			// filter out unwanted versions based on keywords
			lowerName := strings.ToLower(dirName)
			shouldIgnore := false
			for _, keyword := range ignoreKeywords {
				if strings.Contains(lowerName, keyword) {
					shouldIgnore = true
					break // Found a keyword
				}
			}

			if shouldIgnore {
				continue // Skip this version
			}

			// e.g., "Adobe InDesign 2024" -> "2024"
			versionName := strings.TrimPrefix(dirName, "Adobe InDesign ")
			
			// Use our reverseVersionMap (from versions.go)
			// to turn "2024" into 19
			if major, ok := reverseVersionMap[versionName]; ok {
				// Build the full path to the .app
				appPath := fmt.Sprintf("/Applications/%s/%s.app", dirName, dirName)
				
				// Check if the .app file actually exists
				if _, err := os.Stat(appPath); err == nil {
					found[major] = appPath // e.g., found[19] = "/..."
				}
			}
		}
	}

	return found, nil
}