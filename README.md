# InDesign Launcher

A smart, cross-platform utility to automatically open `.indd` files with the correct version of Adobe InDesign.

- - -

## 🎯 The Problem

When you have multiple versions of Adobe InDesign installed (e.g., 2023, 2024, and 2025), Windows and macOS will only set **one** of them as the default application for `.indd` files.

This means if InDesign 2025 is your default, double-clicking a file saved in 2023 will try to open it in 2025. This forces an unnecessary conversion, may update file formats, and can cause compatibility warnings.

This tool fixes that. It's a tiny, lightning-fast "dispatcher" that you set as the default app. When you open a file, it:

1. Instantly reads the file's header to see what version it _needs_.

2. Scans your system to see what versions you _have_.

3. Launches the **correct** version of InDesign.

- - -

## ✨ Features

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

* **No Admin Rights:** All operations, including registration, are done at the user level.

- - -

## ⚙️ Installation

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

## 🚀 Usage (One-Time Setup)

To use the launcher, you must register it with your operating system.

### On Windows

The launcher adds itself to the "Open With..." menu. You just need to run this command once.

1. Open a Command Prompt (cmd) or PowerShell.

2. Navigate to the directory where you saved the tool (e.g., `cd C:\Tools`).

3. Run the registration command:

   Shell

   ```
   .\indesign-launcher.exe --register
   ```

4. The tool will confirm it has been added and give you the next steps.

**To set it as the default:**

1. Right-click any `.indd` file.

2. Select **Open with > Choose another app**.

3. Select **InDesign Launcher** from the list.

4. **Important:** Check the box **"Always use this app to open .indd files"**.

To undo this, you can run `.\indesign-launcher.exe --unregister`.

- - -

### On macOS

macOS will not allow a command-line tool to be set as a file handler. Our tool will guide you to create a simple "wrapper" app using AppleScript.

1. Open a Terminal.

2. Navigate to the directory where you saved the `indesign-launcher` binary.

3. Run the registration command:

   Shell

   ```
   ./indesign-launcher --register
   ```

4. This will create a new file named **`indesign_launcher.applescript`** in the same directory.

5. Follow the instructions printed in your terminal (copied here for reference):

   1. Open 'Script Editor' (it's in your Utilities folder).

   2. Drag the 'indesign\_launcher.applescript' file into it.

   3. Go to 'File' > 'Export...'.

   4. Set 'File Format' to 'Application'.

   5. Save it as 'InDesignLauncher.app' in your Applications folder.

**To set it as the default:**

1. Right-click any `.indd` file and select **Get Info**.

2. Find the **"Open with:"** section.

3. Click the dropdown menu and select your new **InDesignLauncher.app**.

4. Click the **"Change All..."** button to make this the default for all `.indd` files.

To undo this, simply repeat the "Get Info" steps and set a normal Adobe InDesign version as the default.

- - -

## ☀️ Everyday Use

After you've completed the one-time setup, you're done!

Just **double-click any `.indd` file** on your system. The launcher will run invisibly, find the correct InDesign, and open your file in a fraction of a second.

- - -

- - -

## 🧑‍💻 For Developers & Contributors

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

* `find_app_windows.go`: (`//go:build windows`) Windows-only code. `findAllInstalledVersions()` scans `HKEY_CLASSES_ROOT` for Adobe's `InDesign.Application.XX\CLSID` keys to find `LocalServer32` paths.

* `find_app_darwin.go`: (`//go:build darwin`) macOS-only code. `findAllInstalledVersions()` scans the `/Applications` folder for `Adobe InDesign *` bundles.

* `register_win.go`: (`//go:build windows`) Windows-only code for the `--register` and `--unregister` commands. Modifies the `HKEY_CURRENT_USER` registry, adding a new ProgID and an entry in `OpenWithProgids`.

* `register_mac.go`: (`//go:build darwin`) macOS-only code for the registration flags. Prints guidance and creates the helper `applescript` file.

* _(Not present, but implied by the others)_ `find_app_unsupported.go` & `registry_stub.go`: These files would provide empty stubs for other OSs (like Linux) to allow the code to compile.

### Building from Source

You must have the Go toolchain installed.

Bash

```
# Build for your current system
go build

# --- Cross-Compiling ---

# Build for Windows (from Mac/Linux)
# 64-bit:
GOOS=windows GOARCH=amd64 go build -o indesign-launcher.exe

# Build for macOS (from Windows/Linux)
# Apple Silicon:
GOOS=darwin GOARCH=arm64 go build -o indesign-launcher
# Intel:
GOOS=darwin GOARCH=amd64 go build -o indesign-launcher
```

**⭐ Important Windows Build Note:**

When you run the final `indesign-launcher.exe` on Windows (by double-clicking a file), a console window will flash open. To create a "headless" production build that hides this window, build with the following `-ldflags`:

Bash

```
# Build for Windows production (hides console)
go build -ldflags="-H=windowsgui" -o indesign-launcher.exe
```

### 🚀 Future Goals & To-Do



* **Handle more File Types:** `.indb` and `.indt` files are also Adobe InDesign formats. The launcher could be extended to support these as well.

* **Add an Application Icon:**

  * **Windows:** The `-ldflags="-H=windowsgui"` build would benefit from embedding an icon resource (`.syso` file).

  * **macOS:** The AppleScript wrapper should be part of a build script that also assigns a custom icon.

* **Automate macOS Bundling:** The current "guide the user" step for macOS is a good start, but a proper build script (`create_bundle.sh`?) could automate the creation of the `InDesignLauncher.app` bundle, including the `Info.plist` and AppleScript stub.

- - -

## 📜 License

This project is open-source and available under the MIT License.

Copyright (c) 2025 Vlad Vladila (Krommatine Systems: [https://krommatine.eu](https://krommatine.eu))

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOTS LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.