//go:build darwin

package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

//go:embed resources/Mac/id-launcher.icns
var iconData []byte

//go:embed versioninfo.json
var versionInfoData []byte

// RegisterHandler on Mac guides the user to create a script bundle
func RegisterHandler() error {
	// Create a proper macOS .app bundle so Finder can register it as an application.

	// 1. Locate our current executable
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("could not find own executable path: %w", err)
	}
	if p, err := filepath.EvalSymlinks(exePath); err == nil {
		exePath = p
	}

	// 2. Decide install location: prefer CWD, then ~/Desktop
	appName := "indesign-launcher.app"
	cwd, _ := os.Getwd()
	candidates := []string{cwd}
	if home, e := os.UserHomeDir(); e == nil {
		candidates = append(candidates, filepath.Join(home, "Desktop"))
	}

	var installDir string
	var lastErr error
	for _, base := range candidates {
		dest := filepath.Join(base, appName)
		// If it already exists, choose it (we will overwrite)
		if _, err := os.Stat(dest); err == nil {
			installDir = dest
			break
		}
		// Try creating the directory (atomic create test)
		if err := os.MkdirAll(dest, 0755); err == nil {
			installDir = dest
			break
		} else {
			lastErr = err
		}
	}
	if installDir == "" {
		return fmt.Errorf("could not create app bundle in candidate locations: %v", lastErr)
	}

	contents := filepath.Join(installDir, "Contents")
	macosDir := filepath.Join(contents, "MacOS")
	resourcesDir := filepath.Join(contents, "Resources")

	// Ensure directories exist
	if err := os.MkdirAll(macosDir, 0755); err != nil {
		return fmt.Errorf("failed to create MacOS dir: %w", err)
	}
	if err := os.MkdirAll(resourcesDir, 0755); err != nil {
		return fmt.Errorf("failed to create Resources dir: %w", err)
	}

	// 3. Copy our executable into Contents/MacOS/indesign-launcher
	bundleExeName := "indesign-launcher"
	destExePath := filepath.Join(macosDir, bundleExeName)
	if err := copyFile(exePath, destExePath); err != nil {
		return fmt.Errorf("failed to copy executable into app bundle: %w", err)
	}
	if err := os.Chmod(destExePath, 0755); err != nil {
		return fmt.Errorf("failed to set executable permissions: %w", err)
	}

	// 4. Write embedded icon into Resources
	iconDst := filepath.Join(resourcesDir, "id-launcher.icns")
	if len(iconData) > 0 {
		if err := os.WriteFile(iconDst, iconData, 0644); err != nil {
			fmt.Printf("warning: could not write embedded icon to %s: %v\n", iconDst, err)
		}
	} else {
		fmt.Printf("warning: no embedded icon data available\n")
	}

	// 5. Determine version info from embedded versioninfo.json (if present)
	shortVer := "1.0"
	buildVer := "1"
	if len(versionInfoData) > 0 {
		var vi struct {
			Version        string
			FileVersion    string
			ProductVersion string
		}
		if err := json.Unmarshal(versionInfoData, &vi); err == nil {
			if vi.Version != "" {
				shortVer = vi.Version
			}
			// Prefer ProductVersion, then FileVersion, else fallback to Version
			if vi.ProductVersion != "" {
				buildVer = vi.ProductVersion
			} else if vi.FileVersion != "" {
				buildVer = vi.FileVersion
			} else if vi.Version != "" {
				buildVer = vi.Version
			}
		}
	}

	// 6. Write Info.plist
	plist := buildInfoPlist(bundleExeName, shortVer, buildVer)
	plistPath := filepath.Join(contents, "Info.plist")
	if err := os.WriteFile(plistPath, []byte(plist), 0644); err != nil {
		return fmt.Errorf("failed to write Info.plist: %w", err)
	}

	fmt.Printf("Successfully created app bundle at: %s\n\n", installDir)

	// Detailed user instructions
	fmt.Println("Next steps to install and set as default:")
	fmt.Println("  1. Move the app to your preferred location (optional):")
	fmt.Println("     - To move to Applications, use Finder or run:")
	fmt.Println("       sudo mv \"" + installDir + "\" /Applications/")
	fmt.Println("  2. If you keep it on Desktop or another folder, ensure it has execute permissions (done automatically).")
	fmt.Println("  3. To make this the default app for .indd files:")
	fmt.Println("     - Right-click any .indd file -> Get Info -> Open with -> Select 'InDesign Launcher' -> Change All...")
	fmt.Println("  4. If macOS blocks opening the app because it's unsigned/unidentified:")
	fmt.Println("     - Right-click the app and choose 'Open' and then confirm 'Open' in the dialog to bypass Gatekeeper.")
	fmt.Println("     - Or in Terminal, remove the quarantine attribute:")
	fmt.Println("       xattr -d com.apple.quarantine \"" + installDir + "\"")
	fmt.Println("  5. If you moved the app and encounter permission issues, use 'sudo' to move it, then run:")
	fmt.Println("       chmod -R 755 /Applications/indesign-launcher.app/Contents/MacOS/indesign-launcher")
	fmt.Println("  6. To unregister later, remove the bundle from where it was installed.")
	fmt.Println()

	return nil
}

// UnregisterHandler on Mac just prints instructions
func UnregisterHandler() error {
	fmt.Println("To unregister, simply remove the installed \"indesign-launcher.app\"")
	return nil
}

// copyFile copies a file from src to dst. It creates parent directories for dst if needed.
func copyFile(src, dst string) error {
	// Normalize paths: allow tilde? Already expanded earlier.
	if strings.HasPrefix(src, "~") {
		if home, err := os.UserHomeDir(); err == nil {
			src = filepath.Join(home, strings.TrimPrefix(src, "~/"))
		}
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() {
		_ = out.Close()
	}()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Sync()
}

func buildInfoPlist(exeName, shortVer, buildVer string) string {
	// Minimal Info.plist required for Finder to treat this as an app
	return `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>CFBundleName</key>
	<string>InDesign Launcher</string>
	<key>CFBundleDisplayName</key>
	<string>InDesign Launcher</string>
	<key>CFBundleIdentifier</key>
	<string>com.vamitul.indesign-launcher</string>
	<key>CFBundleShortVersionString</key>
	<string>` + shortVer + `</string>
	<key>CFBundleVersion</key>
	<string>` + buildVer + `</string>
	<key>CFBundleExecutable</key>
	<string>` + exeName + `</string>
	<key>CFBundlePackageType</key>
	<string>APPL</string>
	<key>CFBundleIconFile</key>
	<string>id-launcher</string>
	<key>CFBundleDocumentTypes</key>
	<array>
		<dict>
			<key>CFBundleTypeName</key>
			<string>InDesign Document</string>
			<key>CFBundleTypeExtensions</key>
			<array>
				<string>indd</string>
			</array>
			<key>LSItemContentTypes</key>
			<array>
				<string>com.adobe.indesign.indd</string>
			</array>
			<key>CFBundleTypeRole</key>
			<string>Editor</string>
		</dict>
	</array>
</dict>
</plist>`
}
