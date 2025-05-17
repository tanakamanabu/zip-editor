package main

import (
	"archive/zip"
	"path/filepath"
	"strings"

	"github.com/lxn/walk"
)

// ZipTreeItem represents an item in the ZIP file tree
type ZipTreeItem struct {
	name       string
	path       string
	children   []*ZipTreeItem
	parent     *ZipTreeItem
	isDir      bool
	deleteFlag bool
}

// SetDeleteFlagRecursively sets the delete flag for this item and all its children
func (item *ZipTreeItem) SetDeleteFlagRecursively(flag bool, model *ZipTreeModel) {
	item.deleteFlag = flag
	model.PublishItemChanged(item)

	// Recursively set the flag for all children
	for _, child := range item.children {
		child.SetDeleteFlagRecursively(flag, model)
	}
}

// Text returns the display text for the item
func (item *ZipTreeItem) Text() string {
	if item.deleteFlag {
		return item.name + " [削除予定]"
	}
	return item.name
}

// Parent returns the parent item
func (item *ZipTreeItem) Parent() walk.TreeItem {
	if item.parent == nil {
		return nil
	}
	return item.parent
}

// ChildCount returns the number of children
func (item *ZipTreeItem) ChildCount() int {
	return len(item.children)
}

// ChildAt returns the child at the specified index
func (item *ZipTreeItem) ChildAt(index int) walk.TreeItem {
	return item.children[index]
}

// Image returns the image index for the item
func (item *ZipTreeItem) Image() interface{} {
	if item.isDir {
		return 0 // Folder icon
	}
	return 1 // File icon
}

// ZipTreeModel represents the tree model for the ZIP file
type ZipTreeModel struct {
	walk.TreeModelBase
	rootItem *ZipTreeItem
}

// LazyPopulation returns false as we populate the tree immediately
func (m *ZipTreeModel) LazyPopulation() bool {
	return false
}

// RootCount returns the number of root items
func (m *ZipTreeModel) RootCount() int {
	return 1
}

// RootAt returns the root item at the specified index
func (m *ZipTreeModel) RootAt(index int) walk.TreeItem {
	return m.rootItem
}

// LoadZipFile loads a ZIP file and populates the tree model
func LoadZipFile(filePath string) (*ZipTreeModel, error) {
	reader, err := zip.OpenReader(filePath)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	// Create root item with the ZIP file name
	rootItem := &ZipTreeItem{
		name:  filepath.Base(filePath),
		path:  "",
		isDir: true,
	}

	// Map to store directory items for quick lookup
	dirMap := make(map[string]*ZipTreeItem)
	dirMap[""] = rootItem

	// Process each file in the ZIP
	for _, file := range reader.File {
		// Skip directories (they are created as needed)
		if strings.HasSuffix(file.Name, "/") {
			continue
		}

		// Split the path into components and automatically detect encoding
		// This will try various encodings (Shift-JIS, EUC-JP, UTF-8, etc.) and convert to UTF-8
		path := autoDetectEncoding(file.Name)
		dir := filepath.Dir(path)
		dir = strings.TrimSuffix(dir, "/")

		// Ensure all parent directories exist
		parentPath := ""
		parentItem := rootItem
		for _, part := range strings.Split(dir, "/") {
			if part == "" {
				continue
			}

			currentPath := parentPath + part + "/"
			if item, exists := dirMap[currentPath]; exists {
				parentItem = item
			} else {
				// Create new directory item
				newDir := &ZipTreeItem{
					name:   part,
					path:   currentPath,
					parent: parentItem,
					isDir:  true,
				}
				parentItem.children = append(parentItem.children, newDir)
				dirMap[currentPath] = newDir
				parentItem = newDir
			}
			parentPath = currentPath
		}

		// Skip adding file items to the tree to show only folders
		// File items are not added to parentItem.children
	}

	return &ZipTreeModel{rootItem: rootItem}, nil
}