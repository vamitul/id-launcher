//go:build ignore
// +build ignore

// This script generates the Windows resource file (rsrc.syso)
// It reads metadata from versioninfo.json.
//
// How to run:
// go run make_rsrc.go
// or
// go generate

package main

import (
	"encoding/json"
	"io"
	"log"
	"os"

	"github.com/tc-hib/winres"
	"github.com/tc-hib/winres/version"
)

// VersionInfo holds the structure of our versioninfo.json
type VersionInfo struct {
	Version          string `json:"Version"`
	FileVersion      string `json:"FileVersion"`
	ProductVersion   string `json:"ProductVersion"`
	Copyright        string `json:"Copyright"`
	Description      string `json:"Description"`
	InternalName     string `json:"InternalName"`
	OriginalFilename string `json:"OriginalFilename"`
	ProductName      string `json:"ProductName"`
}

func main() {
	// Open and parse versioninfo.json
	file, err := os.Open("versioninfo.json")
	if err != nil {
		log.Fatalf("Could not open versioninfo.json: %v", err)
	}
	defer file.Close()

	bytes, err := io.ReadAll(file)
	if err != nil {
		log.Fatalf("Could not read versioninfo.json: %v", err)
	}

	var v VersionInfo
	if err := json.Unmarshal(bytes, &v); err != nil {
		log.Fatalf("Could not parse versioninfo.json: %v", err)
	}

	rs := winres.ResourceSet{}

	iconFile, err := os.Open("resources/Win/indesign-launcher.ico")
	if err != nil {
		log.Fatalf("Could not open icon file: %v", err)
	}
	defer iconFile.Close()

	icon, err := winres.LoadICO(iconFile)
	if err != nil {
		log.Fatalf("Could not load icon: %v", err)
	}
	rs.SetIcon(winres.Name("APPICON"), icon)

	// Build version.Info using the version package helpers.
	var vi version.Info
	// Set file and product versions (the helper normalizes/pads to 4 parts)
	vi.SetFileVersion(v.FileVersion)
	vi.SetProductVersion(v.ProductVersion)
	// Set the type to application
	vi.Type = version.App
	// Populate string table entries (neutral language)
	vi.Set(version.LangNeutral, version.FileDescription, v.Description)
	vi.Set(version.LangNeutral, version.InternalName, v.InternalName)
	vi.Set(version.LangNeutral, version.OriginalFilename, v.OriginalFilename)
	vi.Set(version.LangNeutral, version.ProductName, v.ProductName)
	vi.Set(version.LangNeutral, version.LegalCopyright, v.Copyright)

	rs.SetVersionInfo(vi)

	// 3. Set the Manifest (read the XML and convert to AppManifest)
	if mf, err := os.ReadFile("resources/Win/indesign-launcher.manifest"); err != nil {
		log.Fatalf("Could not read manifest: %v", err)
	} else {
		manifest, err := winres.AppManifestFromXML(mf)
		if err != nil {
			log.Fatalf("Could not parse manifest XML: %v", err)
		}
		rs.SetManifest(manifest)
	}

	// 4. Populate Version Information from the JSON

	// 5. Compile and save the .syso file
	// This will create 'rsrc.syso' in the current directory
	out, err := os.Create("rsrc.syso")
	if err != nil {
		log.Fatalf("Could not create rsrc.syso: %v", err)
	}
	defer out.Close()
	if err := rs.WriteObject(out, winres.ArchAMD64); err != nil {
		log.Fatalf("Could not write resources: %v", err)
	}

	log.Println("Successfully wrote rsrc.syso")
}
