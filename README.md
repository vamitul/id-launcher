# InDesign Launcher

A cross-platform utility to automatically open `.indd` files with the correct version of Adobe InDesign.

- - -

## The Problem

When you have multiple versions of Adobe InDesign installed (e.g., 2023, 2024, and 2025), Windows and macOS will only set **one** of them as the default application for `.indd` files.

This means if InDesign 2025 is your default, double-clicking a file saved in 2023 will try to open it in 2025. This forces an unnecessary conversion, may update file formats, and can cause compatibility warnings.

## The Solution

**InDesign Launcher** acts as a lightweight dispatcher. When you open a file, it performs three steps:

1. **Reads** the file header to detect the specific version required.

2. **Scans** the system to identify installed InDesign versions.

3. **Launches** the correct version immediately.

It handles the complexity invisibly, ensuring you always work in the correct environment.

- - -

## Features

* **Cross-Platform:** Works on both Windows and macOS.

* **Version Detection:** Reads the binary header of `.indd` files to detect the exact version they were saved with (e.g., CS6, CC 2019, 2024, etc.).

* **System Scan:** Automatically finds all installed InDesign versions on your system.

* **Filters Unwanted Apps:** Ignores special versions like "Server", "Debug", "Prerelease", and "Beta".

* **Smart Selection:**

  * It will always try to launch the **oldest compatible** version.

  * _Example:_ A file saved in 2023 will be opened by InDesign 2023 if you have it. If not, it will try 2024, then 2025.

* **Intelligent Fallback:**

  * If you try to open a 2025 file but only have 2024 installed, the launcher will launch 2024 (your newest version) and let InDesign display its own "cannot open a newer file" error.

* **"Polite" Integration:**

  * On **Windows**, it adds itself to the "Open With..." menu, letting _you_ choose to set it as the default.

  * On **macOS**, it guides you through a simple, one-time manual setup.


- - -

## Installation

You can get the tool in two ways:

### Option 1: Pre-Built Binary (Recommended)

1. Go to the **Releases** page of this project.

2. Download the latest `indesign-launcher.exe` (for Windows) or `indesign-launcher` (for macOS).

3. Place this file in a permanent, convenient location, for example:

   * **Windows:** `C:\Tools\`

   * **macOS:** `/usr/local/bin/` (or your user's `~/bin` folder)

4. On **macOS**, you must make the file executable:

   Bash

   ```
   chmod +x /path/to/indesign-launcher
   ```

### Option 2: Build from Source

You must have the [Go toolchain (v1.20 or later)](https://go.dev/doc/install) installed.

Bash

```
# 1. Clone the repository
git clone https://github.com/your-repo/indesign-launcher.git

# 2. Navigate into the directory
cd indesign-launcher

# 3. Build the binary
go build
```

- - -


# InDesign Launcher

A utility to automatically open Adobe InDesign (`.indd`) files with the correct version of the application.

- - -

## The Problem

Installing multiple versions of Adobe InDesign (e.g., 2023, 2024, and 2025) creates a conflict. Operating systems allow only one default application for a file type.

If InDesign 2025 is your default, opening a file saved in 2023 forces an unnecessary conversion. This alters the file format and triggers compatibility warnings.

## The Solution

**InDesign Launcher** acts as a lightweight dispatcher. When you open a file, it performs three steps:

1. **Reads** the file header to detect the specific version required.

2. **Scans** the system to identify installed InDesign versions.

3. **Launches** the correct version immediately.

It handles the complexity invisibly, ensuring you always work in the correct environment.

- - -

## capabilities

* **Cross-Platform:** Native support for both Windows and macOS.

* **Binary Detection:** Parses the `.indd` binary header to detect the exact creation version (CS6 through CC 2025+).

* **Smart Fallback:** If the exact version is missing, it launches the oldest compatible version to minimize file conversion issues.

* **Clean Integration:**

  * **Windows:** Registers as a valid "Open With" handler.

  * **macOS:** Generates a native `.app` bundle for standard Finder integration.

* **User-Level Operation:** Requires no administrative privileges.

- - -

## Installation

### Option 1: Pre-Built Binary

1. Navigate to the **Releases** page.

2. Download the executable for your system:

   * **Windows:** `indesign-launcher.exe`

   * **macOS:** `indesign-launcher`

3. Move the file to a permanent location (e.g., `C:\Tools\` on Windows or `/usr/local/bin/` on macOS).

### Option 2: Build from Source

Ensure the [Go toolchain (v1.20+)](https://go.dev/doc/install) is installed.

Bash

```
# Clone and build
git clone https://github.com/krommatine/indesign-launcher.git
cd indesign-launcher
go build
```

- - -

## Setup

The launcher must be registered with the operating system to function as a default handler.

### Windows Configuration

1. Open Command Prompt or PowerShell.

2. Navigate to the tool's directory.

3. Run the registration flag:

   PowerShell

   ```
   .\indesign-launcher.exe --register
   ```

4. **Set as Default:**

   * Right-click any `.indd` file.

   * Select **Open with > Choose another app**.

   * Select **InDesign Launcher**.

   * Check **"Always use this app to open .indd files"**.

### macOS Configuration

Previous versions required AppleScript. The tool now generates a native application bundle.

1. Open Terminal and navigate to the directory containing the `indesign-launcher` binary.

2. Run the registration command:

   Bash

   ```
   chmod +x indesign-launcher
   ./indesign-launcher --register
   ```

3. The tool will create a valid **`indesign-launcher.app`** in your current directory (or Desktop).

4. **Install the App:**

   * Move `indesign-launcher.app` to your `/Applications` folder.

5. **Set as Default:**

   * Right-click any `.indd` file in Finder.

   * Select **Get Info**.

   * Under **"Open with:"**, select **InDesign Launcher**.

   * Click **Change All...**.

**Note on Gatekeeper:** Because this tool is open-source and not notarized by Apple, you may need to bypass security checks on the first run.

* **Option A:** Right-click the app and select **Open**, then confirm in the dialog.

* **Option B:** Remove the quarantine attribute via Terminal:

  Bash

  ```
  xattr -d com.apple.quarantine /Applications/indesign-launcher.app
  ```

- - -

## Everyday Use

After you've completed the one-time setup, you're done!

Just **double-click any `.indd` file** on your system. The launcher will run invisibly, find the correct InDesign, and open your file in a fraction of a second.


- - -

## For Developers & Contributors

This project is open-source and contributions are highly welcome! Here are some details to help you get started.

### How It Works (High-Level)

The launcher's logic is split into three phases:

1. **Parse Flags:** The `main()` function first checks for `--register` or `--unregister` flags and routes to the appropriate OS-specific functions.

2. **Read File Header (if opening):** If no flags are present, the `openFile()` function calls `getInDesignVersion()`.

   * This reads the first 37 bytes of the `.indd` file.

   * It validates bytes 0-15 against a "magic number."

   * It checks byte 24 (index 24) to determine the **endianness** (byte order).

   * It reads bytes 29-32 (a 4-byte integer) to get the **raw major version** (e.g., `19`).

3. **Discover & Decide:** `openFile()` then calls `findAllInstalledVersions()`.

   * This OS-specific function (see below) queries the system to find all _installed_ InDesign applications.

   * It returns a map of `majorVersion -> appPath`.

   * The `selectVersionToLaunch()` function compares the file's needed version to the map of installed versions and selects the best one.

4. **Launch:** The chosen `appPath` and the `filePath` are passed to `launchApp()`, which executes the application.

### Code Structure

The project is structured using Go's **build constraints** to keep the logic for each OS separate.

* `main.go`: The main entry point. Handles flag parsing, high-level logic, and file-opening orchestration.

* `versions.go`: A "database" file holding the `versionMap` (e.g., `19` -> `"2024"`), `reverseVersionMap`, and the `ignoreKeywords` list.

* `parse_file.go`: Contains `getInDesignVersion()`, the cross-platform logic for reading and parsing the `.indd` file header.

* `find_app_windows.go`: (`//go:build windows`) Windows-only code. `findAllInstalledVersions()` First searches the default paths for InDesign installations on disk. If a version is not found, it then tries to scan `HKEY_CLASSES_ROOT` for Adobe's `InDesign.Application.XX\CLSID` keys to find `LocalServer32` paths. For old version of InDesign this might fail as the registry paths and setup has changed over time.

* `find_app_darwin.go`: (`//go:build darwin`) macOS-only code. `findAllInstalledVersions()` scans the `/Applications` folder for `Adobe InDesign *` bundles.

* `register_win.go`: (`//go:build windows`) Windows-only code for the `--register` and `--unregister` commands. Modifies the `HKEY_CURRENT_USER` registry, adding a new ProgID and an entry in `OpenWithProgids`.

* `register_mac.go`: (`//go:build darwin`) macOS-only code for the registration flags.



### Building from Source

You must have the Go toolchain installed.

Bash

```
# Build for your current system
go build

# --- Cross-Compiling ---

# Build for Windows (from Mac/Linux)
# 64-bit:
GOOS=windows GOARCH=amd64 go build -ldflags="-H=windowsgui" -o indesign-launcher.exe

# Build for macOS (from Windows/Linux)
# Apple Silicon:
env GOOS=darwin GOARCH=arm64 go build -o indesign-launcher
# Intel:
envGOOS=darwin GOARCH=amd64 go build -o indesign-launcher
```



### Future Goals & To-Do



* **Handle more File Types:** `.indb` and `.indt` files are also Adobe InDesign formats. The launcher could be extended to support these as well.

- - -

## License

This project is open-source and available under the MIT License.

Copyright (c) 2025 Vlad Vladila (Krommatine Systems: [https://krommatine.eu](https://krommatine.eu))

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOTS LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
