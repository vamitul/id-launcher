package main

import (
	"fmt"
	"log"
	"os/exec"
	"runtime"
  "path/filepath"
  "sort"
  "flag"
)




// --- Main Application ---

func main() {

  registerFlag := flag.Bool("register", false, "Register as default .indd handler")
	unregisterFlag := flag.Bool("unregister", false, "Unregister as default .indd handler")

	// Parse the flags
	flag.Parse()

	// --- Route based on flags ---
	if *registerFlag {
		if runtime.GOOS == "windows" {
			if err := RegisterHandler(); err != nil {
				log.Fatalf("Failed to register: %v", err)
			}
			fmt.Println("Successfully registered as default .indd handler.")
		} else {
			fmt.Println("--register is only supported on Windows.")
		}
		return // Exit after task is done
	}

	if *unregisterFlag {
		if runtime.GOOS == "windows" {
			if err := UnregisterHandler(); err != nil {
				log.Fatalf("Failed to unregister: %v", err)
			}
			fmt.Println("Successfully unregistered.")
		} else {
			fmt.Println("--unregister is only supported on Windows.")
		}
		return // Exit after task is done
	}

  // Check if a file path was provided
	if flag.NArg() == 0 {
		log.Println("Usage: indesign-launcher [options] <path-to-file.indd>")
		log.Println("Options:")
		flag.PrintDefaults()
		return
	}

	// Get the file path from the remaining arguments
	filePath := flag.Arg(0)
	if err := openFile(filePath); err != nil {
		log.Fatal(err)
	}

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

func openFile(filePath string) error {
	// 1. Get and clean the file path
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return fmt.Errorf("could not get absolute path for file: %w", err)
	}

	// 2. Get the file's required major version
	fileMajorVersion, err := getInDesignVersion(absPath)
	if err != nil {
		return fmt.Errorf("error reading file '%s': %w", absPath, err)
	}
	fmt.Printf("File: %s\n", absPath)
	fmt.Printf("Detected File Version: %s (Major: %d)\n", versionMap[fileMajorVersion], fileMajorVersion)

	// 3. DISCOVER: Find all installed versions
	installedVersions, err := findAllInstalledVersions()
	if err != nil {
		return fmt.Errorf("error finding installed versions: %w", err)
	}
	if len(installedVersions) == 0 {
		return fmt.Errorf("failed: No InDesign versions found on this system")
	}

	// 4. DECIDE: Select the best version to use
	appPath, launchedVersion := selectVersionToLaunch(fileMajorVersion, installedVersions)

	if launchedVersion >= fileMajorVersion {
		fmt.Printf("... launching compatible version: %s (Major: %d)\n", versionMap[launchedVersion], launchedVersion)
	} else {
		fmt.Printf("... WARNING: No compatible version found.\n")
		fmt.Printf("... Launching latest available version: %s (Major: %d)\n", versionMap[launchedVersion], launchedVersion)
	}
	fmt.Printf("Found application: %s\n", appPath)

	// 5. LAUNCH
	if err := launchApp(appPath, absPath); err != nil {
		return fmt.Errorf("failed to launch InDesign: %w", err)
	}
	
	fmt.Println("Successfully launched!")
	return nil
}