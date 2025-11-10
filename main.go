package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"runtime"
  "path/filepath"
  "sort"
)

//Based on the findings of Theunis de Jong in this Adobe forum post:
//https://community.adobe.com/t5/indesign-discussions/ann-identify-your-indesign-file/td-p/3809701
//

// --- File Header Constants ---

// magicNumber is the 16-byte sequence at the very start of an .indd file
// that identifies it as a valid InDesign document.
var magicNumber = []byte{
	0x06, 0x06, 0xED, 0xF5, 0xD8, 0x1D, 0x46, 0xE5,
	0xBD, 0x31, 0xEF, 0xE7, 0xFE, 0x74, 0xB7, 0x1D,
}

// headerSize is the minimum number of bytes we need to read from the file
// to get all the information we need (Magic Number + Type + Flag + Padding + Versions).
// 16 (Magic) + 8 (Type) + 1 (Flag) + 4 (Padding) + 4 (Major) + 4 (Minor) = 37 bytes
const headerSize = 37


// --- Main Application ---

func main() {
	// 1. Get the file path from command-line arguments
	if len(os.Args) < 2 {
		log.Fatal("Usage: indesign-launcher <path-to-file.indd>")
	}
	filePath, err := filepath.Abs(os.Args[1])
	if err != nil {
		log.Fatalf("Could not get absolute path for file: %v", err)
	}

	// 2. Call our new function to get the version
	versionName, err := getInDesignVersion(filePath)
	if err != nil {
		log.Fatalf("Error reading file '%s': %v", filePath, err)
	}
  fmt.Printf("File: %s\n", filePath)
	fmt.Printf("Detected File Version: %s (Major: %d)\n", versionMap[versionName], versionName)

	// 3. DISCOVER: Find all installed versions
	installedVersions, err := findAllInstalledVersions()
	if err != nil {
		log.Fatalf("Error finding installed versions: %v", err)
	}
	if len(installedVersions) == 0 {
		log.Fatal("Failed: No InDesign versions found on this system.")
	}

	// 4. DECIDE: Select the best version to use
	appPath, launchedVersion := selectVersionToLaunch(versionName, installedVersions)

	if launchedVersion >= versionName {
		fmt.Printf("... launching compatible version: %s (Major: %d)\n", versionMap[launchedVersion], launchedVersion)
	} else {
		fmt.Printf("... WARNING: No compatible version found.\n")
		fmt.Printf("... Launching latest available version: %s (Major: %d)\n", versionMap[launchedVersion], launchedVersion)
	}
	fmt.Printf("Found application: %s\n", appPath)

	// 5. LAUNCH
	if err := launchApp(appPath, filePath); err != nil {
		log.Fatalf("Failed to launch InDesign: %v", err)
	}
	fmt.Println("Successfully launched!")
}

// getInDesignVersion opens the file, reads the header, and returns the version string.
func getInDesignVersion(filePath string) (uint32, error) {

	// --- Step 1: Open the file for reading ---
	file, err := os.Open(filePath)
	if err != nil {
		return 0, fmt.Errorf("could not open file: %w", err)
	}
	// 'defer' ensures this runs right before the function exits,
	// guaranteeing the file is closed.
	defer file.Close()

	// --- Step 2: Read the 37-byte header ---

	// Create a byte slice (a buffer) to hold the header data
	header := make([]byte, headerSize)

	// Read exactly 'headerSize' bytes from the file into our buffer
	// We use io.ReadFull to ensure we get all 37 bytes, or an error.
	if _, err := io.ReadFull(file, header); err != nil {
		return 0, fmt.Errorf("could not read header: %w", err)
	}

	// --- Step 3: Check the Magic Number (Bytes 0-15) ---

	// 'header[0:16]' creates a "slice" pointing to the first 16 bytes.
	// We compare it to our 'magicNumber' constant.
	if !bytes.Equal(header[0:16], magicNumber) {
		return 0, fmt.Errorf("not a valid InDesign file (magic number mismatch)")
	}

	// --- Step 4: Determine Endianness (Byte 24) ---

	// 'header[24]' accesses the 25th byte (index 24).
	// We must know the endianness *before* we can read the version numbers.
	var byteOrder binary.ByteOrder
	// Based on the forum's JavaScript code:
	// Flag 2 = Big Endian
	// Flag 1 (or other) = Little Endian
	switch header[24] {
	case 2:
		byteOrder = binary.BigEndian
	default:
		// Default to LittleEndian for flag 1 or any other value
		byteOrder = binary.LittleEndian
	}

	// --- Step 5: Read the Major Version (Bytes 29-32) ---

	// 'header[29:33]' creates a slice pointing to the 4 bytes for the major version.
	// We use Go's binary package to convert these 4 bytes into a
	// 32-bit unsigned integer (uint32), using the 'byteOrder' we just found.
	majorVersion := byteOrder.Uint32(header[29:33])

  return majorVersion, nil
}

// launchApp executes the command to open the file with the found application.
// This logic is also OS-specific, so we handle it here.
func launchApp(appPath, filePath string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		// On Windows, appPath is the .exe. We pass the file path as an argument.
		cmd = exec.Command(appPath, filePath)
	case "darwin":
		// On macOS, we use the 'open' command.
		// '-a' specifies the application (appPath)
		// and the final argument is the file to open.
		cmd = exec.Command("open", "-a", appPath, filePath)
	default:
		return fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}

	// We use Start() instead of Run() so our launcher can
	// exit immediately without waiting for InDesign to close.
	return cmd.Start()
}

// It returns the app path and the major version it selected.
func selectVersionToLaunch(fileMajor uint32, installed map[uint32]string) (string, uint32) {
	
	// Create a sorted list of all installed major versions
	var keys []uint32
	for k := range installed {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	
	// Case 1: Find the lowest compatible version
	// Loop from low to high
	for _, major := range keys {
		if major >= fileMajor {
			// Found it! This is the oldest, compatible version.
			return installed[major], major
		}
	}
	
	// Case 2: No compatible version found.
	// Fallback to the latest installed version.
	latestVersion := keys[len(keys)-1]
	return installed[latestVersion], latestVersion
}