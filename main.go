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
	"github.com/lxn/win"
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

func main() {
	// メインウィンドウを作成
	mw := new(walk.MainWindow)
	var tv *walk.TreeView
	var te *walk.TextEdit
	var model *ZipTreeModel
	var currentZipPath string
	var currentSelectedPath string

	// メインウィンドウを設定
	if err := (MainWindow{
		AssignTo: &mw,
		Title:    "ZIP ファイルビューア",
		MinSize:  Size{Width: 700, Height: 400},
		Layout:   VBox{},
		Children: []Widget{
			// 水平分割レイアウト
			Composite{
				Layout:         HBox{MarginsZero: true},
				StretchFactor:  10,
				Children: []Widget{
					// 左側：ツリービュー
					TreeView{
						AssignTo:           &tv,
						StretchFactor:      5, // 左右の比率
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

									// Toggle the deleteFlag and propagate to children
									newFlag := !zipItem.deleteFlag
									zipItem.SetDeleteFlagRecursively(newFlag, model)
								},
							},
						},
						OnMouseDown: func(x, y int, button walk.MouseButton) {
							// Check if left mouse button was clicked
							if button != walk.LeftButton {
								return
							}

							// Check if Ctrl key is pressed
							if win.GetKeyState(win.VK_CONTROL) >= 0 { // Not pressed if high bit is not set
								return
							}

							// Get the item at the clicked position
							item := tv.ItemAt(x, y)
							if item == nil {
								return
							}

							// Get the ZipTreeItem
							zipItem, ok := item.(*ZipTreeItem)
							if !ok {
								return
							}

							// Toggle the deleteFlag and propagate to children
							newFlag := !zipItem.deleteFlag
							zipItem.SetDeleteFlagRecursively(newFlag, model)
						},
						OnCurrentItemChanged: func() {
							// Get the selected item
							item := tv.CurrentItem()
							// Get the ZipTreeItem
							zipItem, ok := item.(*ZipTreeItem)
							if !ok || !zipItem.isDir {
								return
							}
							
							// Save the current selected path
							currentSelectedPath = zipItem.path
							
							// Update the list view with files in this directory
							updateFileList(te, currentZipPath, currentSelectedPath)
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
					// 右側：テキストエディット（ファイル一覧表示用）
					TextEdit{
						AssignTo:           &te,
						StretchFactor:      5, // 左右の比率
						AlwaysConsumeSpace: true,
						ReadOnly:           true,
						VScroll:            true,
					},
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