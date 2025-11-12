//go:build windows

package main

import (
	"fmt"
	"os"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

// --- Registry Constants ---
const (
	// Our new Program ID
	progID = "InDesignLauncher.indd"

	// The key for the "Open With" list
	openWithKeyPath = `Software\Microsoft\Windows\CurrentVersion\Explorer\FileExts\.indd\OpenWithProgids`
)

var progIDKeyPath = fmt.Sprintf(`Software\Classes\%s`, progID)

// --- Windows API Constants for SHChangeNotify ---
const (
	SHCNE_ASSOCCHANGED = 0x08000000
	SHCNF_IDLIST       = 0x0000
)

var (
	// Get a lazy-loaded handle to shell32.dll
	shell32 = windows.NewLazySystemDLL("shell32.dll")
	// Find the procedure (function) within that DLL
	shChangeNotifyProc = shell32.NewProc("SHChangeNotify")
)

// RegisterHandler registers our app in the "Open With..." list
func RegisterHandler() error {
	// 1. Get the full path to our own executable
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("could not find own executable path: %w", err)
	}
	openCommand := fmt.Sprintf("\"%s\" \"%%1\"", exePath)

	// 2. Create ProgID definition
	// e.g., HKCU\Software\Classes\InDesignLauncher.indd\shell\open\command
	cmdKey, _, err := registry.CreateKey(registry.CURRENT_USER, progIDKeyPath+`\shell\open\command`, registry.SET_VALUE)
	if err != nil {
		return fmt.Errorf("could not create shell command key: %w", err)
	}
	// Set the (Default) value to our open command
	if err := cmdKey.SetStringValue("", openCommand); err != nil {
		cmdKey.Close()
		return fmt.Errorf("could not set shell command: %w", err)
	}
	cmdKey.Close()

	// 3. Add entry to the OpenWithProgids list
	key, _, err := registry.CreateKey(registry.CURRENT_USER, openWithKeyPath, registry.SET_VALUE)
	if err != nil {
		return fmt.Errorf("could not create/open OpenWithProgids key: %w", err)
	}
	defer key.Close()

	// Set our value. We use SetValue with REG_NONE.
	// We just need to create an empty byte slice for the data.
	if err := key.SetBinaryValue(progID, []byte{}); err != nil {
		return fmt.Errorf("could not add entry to OpenWithProgids: %w", err)
	}

	// 4. Notify the Windows Shell of the change
	notifyWindowsShell()

	// 5. Inform the user as requested
	fmt.Println("Successfully added 'InDesign Launcher' to the 'Open With' list.")
	fmt.Println("To set as default:")
	fmt.Println("  1. Right-click an .indd file")
	fmt.Println("  2. Select 'Open with' > 'Choose another app'")
	fmt.Println("  3. Select 'InDesign Launcher' and check 'Always use this app...'")
	return nil
}

// UnregisterHandler removes our app from the "Open With..." list
func UnregisterHandler() error {
	var warnings []string

	// 1. Remove from OpenWithProgids
	key, err := registry.OpenKey(registry.CURRENT_USER, openWithKeyPath, registry.SET_VALUE)
	if err == nil {
		// Key exists, now delete our value
		if err := key.DeleteValue(progID); err != nil && err != registry.ErrNotExist {
			warnings = append(warnings, fmt.Sprintf("could not delete OpenWithProgids value: %v", err))
		}
		key.Close()
	} else if err != registry.ErrNotExist {
		warnings = append(warnings, fmt.Sprintf("could not open OpenWithProgids key: %v", err))
	}

	// 2. Delete our ProgID
	// We must delete from the "inside out"
	if err := registry.DeleteKey(registry.CURRENT_USER, progIDKeyPath+`\shell\open\command`); err != nil && err != registry.ErrNotExist {
		warnings = append(warnings, fmt.Sprintf("could not delete command key: %v", err))
	}
	if err := registry.DeleteKey(registry.CURRENT_USER, progIDKeyPath+`\shell\open`); err != nil && err != registry.ErrNotExist {
		warnings = append(warnings, fmt.Sprintf("could not delete open key: %v", err))
	}
	if err := registry.DeleteKey(registry.CURRENT_USER, progIDKeyPath+`\shell`); err != nil && err != registry.ErrNotExist {
		warnings = append(warnings, fmt.Sprintf("could not delete shell key: %v", err))
	}
	if err := registry.DeleteKey(registry.CURRENT_USER, progIDKeyPath); err != nil && err != registry.ErrNotExist {
		warnings = append(warnings, fmt.Sprintf("could not delete main ProgID key: %v", err))
	}

	// Print all warnings at the end
	if len(warnings) > 0 {
		fmt.Fprintln(os.Stderr, "Unregister complete with warnings:")
		for _, w := range warnings {
			fmt.Fprintf(os.Stderr, "  - %s\n", w)
		}
	} else {
		fmt.Println("Successfully unregistered.")
	}

	// 3. Notify the Windows Shell
	notifyWindowsShell()
	return nil
}

// notifyWindowsShell tells Explorer that file associations have changed.
func notifyWindowsShell() {
	shChangeNotifyProc.Call(SHCNE_ASSOCCHANGED, SHCNF_IDLIST, 0, 0)
}
