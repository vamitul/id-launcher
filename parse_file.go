package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
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