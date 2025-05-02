package main

import (
	"archive/zip"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
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

// ZipTreeItem represents an item in the ZIP file tree
type ZipTreeItem struct {
	name       string
	path       string
	children   []*ZipTreeItem
	parent     *ZipTreeItem
	isDir      bool
	deleteFlag bool
}

// Text returns the display text for the item
func (item *ZipTreeItem) Text() string {
	if item.deleteFlag {
		return "☑ " + item.name
	}
	return "◻ " + item.name
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

// ZipTreeModel represents the model for the ZIP file tree
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
		dir, name := filepath.Split(path)
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

		// Create file item
		fileItem := &ZipTreeItem{
			name:   name,
			path:   path,
			parent: parentItem,
			isDir:  false,
		}
		parentItem.children = append(parentItem.children, fileItem)
	}

	return &ZipTreeModel{rootItem: rootItem}, nil
}

func main() {
	// メインウィンドウを作成
	mw := new(walk.MainWindow)
	var tv *walk.TreeView
	var model *ZipTreeModel
	var currentZipPath string

	// メインウィンドウを設定
	if err := (MainWindow{
		AssignTo: &mw,
		Title:    "ZIP ファイルビューア",
		MinSize:  Size{Width: 500, Height: 300},
		Layout:   VBox{},
		Children: []Widget{
			// ツリービュー
			TreeView{
				AssignTo:           &tv,
				StretchFactor:      10, // ウィンドウサイズに合わせて拡大
				AlwaysConsumeSpace: true,
				ContextMenuItems: []MenuItem{
					Action{
						Text: "Toggle Delete Flag",
						OnTriggered: func() {
							// Get the selected item
							item := tv.CurrentItem()
							// Get the ZipTreeItem
							zipItem, ok := item.(*ZipTreeItem)
							if !ok {
								return
							}

							// Toggle the deleteFlag
							zipItem.deleteFlag = !zipItem.deleteFlag

							// Update the tree view
							model.PublishItemChanged(zipItem)
						},
					},
				},
				OnItemActivated: func() {
					// Get the selected item
					item := tv.CurrentItem()
					// Get the ZipTreeItem
					zipItem, ok := item.(*ZipTreeItem)
					if !ok || zipItem.isDir {
						// Ignore if it's not a ZipTreeItem or if it's a directory
						return
					}

					// Extract the file to a temporary location
					reader, err := zip.OpenReader(currentZipPath)
					if err != nil {
						walk.MsgBox(mw, "エラー", "ZIPファイルを開けませんでした: "+err.Error(), walk.MsgBoxIconError)
						return
					}
					defer reader.Close()

					// Find the file in the ZIP
					var zipFile *zip.File
					for _, f := range reader.File {
						// Automatically detect encoding and convert to UTF-8 for comparison
						// This will try various encodings (Shift-JIS, EUC-JP, UTF-8, etc.) and convert to UTF-8
						if autoDetectEncoding(f.Name) == zipItem.path {
							zipFile = f
							break
						}
					}

					if zipFile == nil {
						walk.MsgBox(mw, "エラー", "ファイルが見つかりませんでした: "+zipItem.path, walk.MsgBoxIconError)
						return
					}

					// Create a temporary directory
					tempDir, err := os.MkdirTemp("", "zip-editor-")
					if err != nil {
						walk.MsgBox(mw, "エラー", "一時ディレクトリを作成できませんでした: "+err.Error(), walk.MsgBoxIconError)
						return
					}

					// Create the full path for the extracted file
					// zipItem.name is already in UTF-8 from our earlier conversion
					tempFilePath := filepath.Join(tempDir, zipItem.name)

					// Extract the file
					srcFile, err := zipFile.Open()
					if err != nil {
						walk.MsgBox(mw, "エラー", "ファイルを開けませんでした: "+err.Error(), walk.MsgBoxIconError)
						return
					}
					defer srcFile.Close()

					destFile, err := os.Create(tempFilePath)
					if err != nil {
						walk.MsgBox(mw, "エラー", "一時ファイルを作成できませんでした: "+err.Error(), walk.MsgBoxIconError)
						return
					}
					defer destFile.Close()

					_, err = io.Copy(destFile, srcFile)
					if err != nil {
						walk.MsgBox(mw, "エラー", "ファイルを抽出できませんでした: "+err.Error(), walk.MsgBoxIconError)
						return
					}

					// Close the file before opening it
					destFile.Close()

					// Open the file with the default application
					cmd := exec.Command("cmd", "/c", "start", "", tempFilePath)
					err = cmd.Start()
					if err != nil {
						walk.MsgBox(mw, "エラー", "ファイルを開けませんでした: "+err.Error(), walk.MsgBoxIconError)
						return
					}
				},
			},
			// ボタンエリア
			Composite{
				Layout: HBox{MarginsZero: true},
				Children: []Widget{
					HSpacer{}, // 右寄せのためのスペーサー
					PushButton{
						Text: "OK",
						OnClicked: func() {
							walk.MsgBox(mw, "確認", "OKボタンがクリックされました", walk.MsgBoxIconInformation)
						},
					},
				},
			},
		},
	}).Create(); err != nil {
		log.Fatal(err)
	}

	// ドロップイベントを処理
	mw.DropFiles().Attach(func(files []string) {
		for _, file := range files {
			// ZIPファイルかどうかを確認
			if strings.ToLower(filepath.Ext(file)) == ".zip" {
				// ZIPファイルを読み込む
				newModel, err := LoadZipFile(file)
				if err != nil {
					walk.MsgBox(mw, "エラー", "ZIPファイルを開けませんでした: "+err.Error(), walk.MsgBoxIconError)
					continue
				}

				// 現在のZIPファイルパスを保存
				currentZipPath = file

				model = newModel
				tv.SetModel(model)

				// ウィンドウタイトルを更新
				mw.SetTitle("ZIP ファイルビューア - " + filepath.Base(file))

				// 最初のZIPファイルだけ処理
				break
			}
		}
	})

	// ウィンドウを表示してメッセージループを開始
	mw.Show()
	mw.Run()
}
