//go:build darwin

package main

import (
	"fmt"
	"os"
)

// RegisterHandler on Mac guides the user to create a script bundle
func RegisterHandler() error {
	// 1. Get the full path to our own executable
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("could not find own executable path: %w", err)
	}

	// 2. Define the AppleScript content
	scriptContent := fmt.Sprintf(
`on open dropped_files
    -- This script will run your Go program
    -- for each file dropped on it.
    repeat with f in dropped_files
        do shell script "%s " & (quoted form of (POSIX path of f))
    end repeat
end open`, exePath)

	// 3. Create the launcher.applescript file
	scriptName := "indesign_launcher.applescript"
	if err := os.WriteFile(scriptName, []byte(scriptContent), 0644); err != nil {
		return fmt.Errorf("failed to create %s: %w", scriptName, err)
	}

	// 4. Print the instructions for the user
	fmt.Printf("Success! A helper file was created at: %s\n\n", scriptName)
	fmt.Println("To complete registration, you MUST do this manually:")
	fmt.Println("  1. Open 'Script Editor' (it's in your Utilities folder).")
	fmt.Println("  2. Drag the 'indesign_launcher.applescript' file into it.")
	fmt.Println("  3. Go to 'File' > 'Export...'.")
	fmt.Println("  4. Set 'File Format' to 'Application'.")
	fmt.Println("  5. Save it as 'InDesignLauncher.app' in your Applications folder.")
	fmt.Println("\nYou can now right-click any .indd file, choose 'Get Info',")
	fmt.Println("and set 'InDesignLauncher.app' as the new default.")

	return nil
}

// UnregisterHandler on Mac just prints instructions
func UnregisterHandler() error {
	fmt.Println("To unregister, simply reset the default handler in Finder:")
	fmt.Println("  1. Right-click any .indd file.")
	fmt.Println("  2. Select 'Get Info'.")
	fmt.Println("  3. In the 'Open with:' section, select 'Adobe InDesign' (or another app).")
	fmt.Println("  4. Click 'Change All...'.")
	return nil
}