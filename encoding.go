package main

import (
	"unicode/utf8"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/encoding/korean"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/encoding/traditionalchinese"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

// Helper function to automatically detect encoding and convert to UTF-8
func autoDetectEncoding(input string) string {
	// Check if the input is already valid UTF-8
	if utf8.ValidString(input) {
		return input
	}

	// List of encodings to try
	encodings := []encoding.Encoding{
		japanese.ShiftJIS,       // Japanese Shift-JIS
		japanese.EUCJP,          // Japanese EUC-JP
		japanese.ISO2022JP,      // Japanese ISO-2022-JP
		korean.EUCKR,            // Korean EUC-KR
		simplifiedchinese.GBK,   // Simplified Chinese GBK
		traditionalchinese.Big5, // Traditional Chinese Big5
		charmap.Windows1252,     // Windows-1252 (Western European)
		unicode.UTF16(unicode.BigEndian, unicode.IgnoreBOM),    // UTF-16BE
		unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM), // UTF-16LE
	}

	// Try each encoding
	for _, enc := range encodings {
		decoder := enc.NewDecoder()
		output, _, err := transform.String(decoder, input)
		if err == nil && utf8.ValidString(output) {
			// Check if the output contains valid characters
			// This helps filter out false positives
			if !containsControlCharacters(output) {
				return output
			}
		}
	}

	// If all encodings fail, try Shift-JIS as a fallback (for backward compatibility)
	transformer := japanese.ShiftJIS.NewDecoder()
	output, _, err := transform.String(transformer, input)
	if err == nil {
		return output
	}

	// If all else fails, return the original string
	return input
}

// Helper function to check if a string contains control characters
// which might indicate an incorrect encoding detection
func containsControlCharacters(s string) bool {
	for _, r := range s {
		// Check for control characters except common whitespace
		if r < 32 && r != '\t' && r != '\n' && r != '\r' {
			return true
		}
	}
	return false
}