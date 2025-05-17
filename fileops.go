package main

import (
	"archive/zip"
	"path/filepath"
	"strings"

	"github.com/lxn/walk"
)

// UpdateFileList updates the list of files in the specified directory
func updateFileList(te *walk.TextEdit, zipPath string, dirPath string) error {
	if zipPath == "" {
		return nil
	}

	// Open the ZIP file
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer reader.Close()

	// Clear the text edit
	te.SetText("")

	// Process each file in the ZIP
	var fileList strings.Builder
	for _, file := range reader.File {
		// Skip directories
		if strings.HasSuffix(file.Name, "/") {
			continue
		}

		// Convert the file path to UTF-8
		path := autoDetectEncoding(file.Name)
		dir := filepath.Dir(path)
		name := filepath.Base(path)

		// Normalize directory path for comparison
		dir = strings.TrimSuffix(dir, "/")
		if dir == "." {
			dir = ""
		} else {
			dir += "/"
		}

		// Check if this file is in the current directory
		if dir == dirPath {
			// Add the file to the list
			fileList.WriteString(name)
			fileList.WriteString("\r\n")
		}
	}

	// Update the text edit
	te.SetText(fileList.String())

	return nil
}