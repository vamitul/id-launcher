//go:build windows

package main

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"golang.org/x/sys/windows/registry"
)

var standardLocations = map[uint32][]string{
	// CS
	3: {
		`C:\Program Files\Adobe\Adobe InDesign CS\InDesign.exe`,
		`C:\Program Files (x86)\Adobe\Adobe InDesign CS\InDesign.exe`,
	},
	4: {
		`C:\Program Files\Adobe\Adobe InDesign CS2\InDesign.exe`,
		`C:\Program Files (x86)\Adobe\Adobe InDesign CS2\InDesign.exe`,
	},
	5: {
		`C:\Program Files\Adobe\Adobe InDesign CS3\InDesign.exe`,
		`C:\Program Files (x86)\Adobe\Adobe InDesign CS3\InDesign.exe`,
	},
	6: {
		`C:\Program Files\Adobe\Adobe InDesign CS4\InDesign.exe`,
		`C:\Program Files (x86)\Adobe\Adobe InDesign CS4\InDesign.exe`,
	},
	7: {
		`C:\Program Files\Adobe\Adobe InDesign CS5\InDesign.exe`,
		// CS5 is normally 64-bit, but include x86 as fallback
		`C:\Program Files (x86)\Adobe\Adobe InDesign CS5\InDesign.exe`,
	},
	8: {
		`C:\Program Files\Adobe\Adobe InDesign CS6\InDesign.exe`,
		// Rare but possible on 64-bit systems
		`C:\Program Files (x86)\Adobe\Adobe InDesign CS6\InDesign.exe`,
	},

	// CC – Yearless
	9: {
		`C:\Program Files\Adobe\Adobe InDesign CC\InDesign.exe`,
	},
	10: {
		`C:\Program Files\Adobe\Adobe InDesign CC 2014\InDesign.exe`,
	},
	11: {
		`C:\Program Files\Adobe\Adobe InDesign CC 2015\InDesign.exe`,
	},

	// CC – Year-based (always 64-bit)
	12: {
		`C:\Program Files\Adobe\Adobe InDesign CC 2017\InDesign.exe`,
	},
	13: {
		`C:\Program Files\Adobe\Adobe InDesign CC 2018\InDesign.exe`,
	},
	14: {
		`C:\Program Files\Adobe\Adobe InDesign CC 2019\InDesign.exe`,
	},
	15: {
		`C:\Program Files\Adobe\Adobe InDesign 2020\InDesign.exe`,
	},
	16: {
		`C:\Program Files\Adobe\Adobe InDesign 2021\InDesign.exe`,
	},
	17: {
		`C:\Program Files\Adobe\Adobe InDesign 2022\InDesign.exe`,
	},
	18: {
		`C:\Program Files\Adobe\Adobe InDesign 2023\InDesign.exe`,
	},
	19: {
		`C:\Program Files\Adobe\Adobe InDesign 2024\InDesign.exe`,
	},
	20: {
		`C:\Program Files\Adobe\Adobe InDesign 2025\InDesign.exe`,
	},
	21: {
		`C:\Program Files\Adobe\Adobe InDesign 2026\InDesign.exe`,
	},
	22: {
		`C:\Program Files\Adobe\Adobe InDesign 2027\InDesign.exe`,
	},
	23: {
		`C:\Program Files\Adobe\Adobe InDesign 2028\InDesign.exe`,
	},
}

// findAllInstalledVersions is the Windows implementation.
// It checks the registry for all known versions and returns a map
// of {majorVersion: executablePath}.
func findAllInstalledVersions() (map[uint32]string, error) {
	found := make(map[uint32]string)

	// Sort versions to make behavior deterministic
	majors := make([]int, 0, len(versionMap))
	for m := range versionMap {
		majors = append(majors, int(m))
	}
	sort.Ints(majors)

	for _, mv := range majors {
		major := uint32(mv)
		name := versionMap[major]

		// ---- 1. Check standard locations
		if defaults, ok := standardLocations[major]; ok {
			for _, defaultPath := range defaults {
				if _, err := os.Stat(defaultPath); err == nil {
					if !shouldIgnore(defaultPath) {
						found[major] = defaultPath
						break
					}
				}
			}
		}

		if _, ok := found[major]; ok {
			continue
		}

		// ---- 2. Registry lookup via ProgID → CLSID → LocalServer32
		progIDKeyPath := fmt.Sprintf(`InDesign.Application.%s\CLSID`,
			strings.ReplaceAll(name, " ", "."))

		progIDKey, err := registry.OpenKey(
			registry.CLASSES_ROOT, progIDKeyPath, registry.QUERY_VALUE)
		if err != nil {
			continue
		}
		clsid, _, err := progIDKey.GetStringValue("")
		progIDKey.Close()
		if err != nil {
			continue
		}

		serverKeyPath := fmt.Sprintf(`CLSID\%s\LocalServer32`, clsid)

		serverKey, err := registry.OpenKey(
			registry.CLASSES_ROOT, serverKeyPath, registry.QUERY_VALUE)
		if err != nil {
			continue
		}
		command, _, err := serverKey.GetStringValue("")
		serverKey.Close()
		if err != nil {
			continue
		}

		if shouldIgnore(command) {
			continue
		}

		found[major] = command
	}

	return found, nil
}

// filter out unwanted versions based on keywords
func shouldIgnore(command string) bool {
	lowerPath := strings.ToLower(command)
	for _, keyword := range ignoreKeywords {
		if strings.Contains(lowerPath, keyword) {
			return true // Found a keyword, no need to check others
		}
	}
	return false
}
