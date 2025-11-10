//go:build windows

package main

import (
	"fmt"
	"strings"

	"golang.org/x/sys/windows/registry"
)

// findAllInstalledVersions is the Windows implementation.
// It checks the registry for all known versions and returns a map
// of {majorVersion: executablePath}.
func findAllInstalledVersions() (map[uint32]string, error) {
	found := make(map[uint32]string)

	// Loop through our known version map (from versions.go)
	// We check for each one.
	for major, name := range versionMap {
		// Use the name to find the ProgID
		// (e.g., "InDesign.Application.19")
		progIDKeyPath := fmt.Sprintf(`InDesign.Application.%s\CLSID`, name)

		progIDKey, err := registry.OpenKey(registry.CLASSES_ROOT, progIDKeyPath, registry.QUERY_VALUE)
		if err != nil {
			continue // This version isn't installed
		}
		defer progIDKey.Close()

		clsid, _, err := progIDKey.GetStringValue("")
		if err != nil {
			continue // CLSID not found
		}

		// Now find the server path from the CLSID
		serverKeyPath := fmt.Sprintf(`CLSID\%s\LocalServer32`, clsid)
		serverKey, err := registry.OpenKey(registry.CLASSES_ROOT, serverKeyPath, registry.QUERY_VALUE)
		if err != nil {
			continue // Server not found
		}
		defer serverKey.Close()

		command, _, err := serverKey.GetStringValue("")
		if err != nil {
			continue // Path not found
		}

		var appPath string
		parts := strings.Split(command, "\"")
		if len(parts) >= 2 {
			appPath = parts[1] // e.g., "C:\..."
		} else {
			continue // Invalid path
		}

		// filter out unwanted versions based on keywords
		lowerPath := strings.ToLower(appPath)
		shouldIgnore := false
		for _, keyword := range ignoreKeywords {
			if strings.Contains(lowerPath, keyword) {
				shouldIgnore = true
				break // Found a keyword, no need to check others
			}
		}

		if shouldIgnore {
			continue // Skip this version
		}

		found[major] = appPath // Add to our map, e.g., found[19] = "C:\..."
	}

	return found, nil
}
