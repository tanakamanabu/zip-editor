package main

import (
	"archive/zip"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"github.com/lxn/win"
)




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
