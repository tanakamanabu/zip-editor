package gui

import (
	"log"
	"os/exec"
	"path/filepath"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"

	"zip-editor/internal/fileops"
	"zip-editor/internal/model"
)

// CreateMainWindow はメインウィンドウを作成し、表示します
func CreateMainWindow() {
	// メインウィンドウを作成
	mw := new(walk.MainWindow)
	var tableView *walk.TableView    // 右側：ディレクトリ内のファイル一覧
	var fileListView *walk.TableView // 左側：ZIPファイル一覧
	var tv *walk.TreeView

	// 現在のZIPファイルパスとモデル
	var currentZipPath string
	var zipModel *model.ZipTreeModel

	// ツリービュー用のコンテキストメニューを作成
	treeContextMenu, err := walk.NewMenu()
	if err != nil {
		log.Fatal(err)
	}

	// 削除メニュー項目を追加
	deleteAction := walk.NewAction()
	deleteAction.SetText("削除")
	deleteAction.Triggered().Attach(func() {
		// 現在選択されているアイテムを取得
		item := tv.CurrentItem()
		if zipItem, ok := item.(*model.ZipTreeItem); ok && zipItem.IsDir() {
			// 再帰的に削除フラグをONに設定
			zipItem.DeleteFlag = true
			fileops.UpdateDeleteFlagRecursively(currentZipPath, zipItem)

			// 現在表示中のファイル一覧を更新
			fileops.UpdateFileList(tableView, zipItem)
		}
	})
	treeContextMenu.Actions().Add(deleteAction)

	// クリアメニュー項目を追加
	clearAction := walk.NewAction()
	clearAction.SetText("クリア")
	clearAction.Triggered().Attach(func() {
		// 現在選択されているアイテムを取得
		item := tv.CurrentItem()
		if zipItem, ok := item.(*model.ZipTreeItem); ok && zipItem.IsDir() {
			// 再帰的に削除フラグをOFFに設定
			zipItem.DeleteFlag = false
			fileops.UpdateDeleteFlagRecursively(currentZipPath, zipItem)

			// 現在表示中のファイル一覧を更新
			fileops.UpdateFileList(tableView, zipItem)
		}
	})
	treeContextMenu.Actions().Add(clearAction)

	// 左ペインのモデル（ZIPファイル一覧）
	fileListModel := model.NewFileListModel()
	// 左ペインの前回選択インデックス
	lastFileListIndex := -1

	// メインウィンドウを設定
	if err := (MainWindow{
		AssignTo: &mw,
		Title:    "ZIP ファイルビューア",
		MinSize:  Size{Width: 700, Height: 400},
		Layout:   VBox{},
		Children: []Widget{
			// 水平分割レイアウト
			HSplitter{
				StretchFactor: 10,
				Children: []Widget{
					Composite{
						StretchFactor: 0,
						Layout:        VBox{MarginsZero: true},
						Children: []Widget{
							// 中身：ZIPファイル一覧（TableView）
							TableView{
								AssignTo:           &fileListView,
								AlwaysConsumeSpace: true,
								Columns: []TableViewColumn{
									{Title: "ファイル"}, // ファイル名のみ表示
								},
							},
						},
					},
					// 中央：ツリービュー
					TreeView{
						AssignTo:           &tv,
						StretchFactor:      4,
						AlwaysConsumeSpace: true,
					},
					// 右側：TableView（選択ディレクトリのファイル一覧表示用）
					TableView{
						AssignTo:           &tableView,
						StretchFactor:      5,
						AlwaysConsumeSpace: true,
						CheckBoxes:         true,
						Columns: []TableViewColumn{
							{Title: "", Width: 20},
							{Title: "ファイル名"},
							{Title: "サイズ"},
							{Title: "日付"},
						},
						OnMouseDown: func(x, y int, button walk.MouseButton) {
							// マウスクリックの位置からアイテムを特定
							index := tableView.IndexAt(x, y)

							// チェックボックスカラムの幅（20ピクセル）内かどうかを確認
							if x <= 20 && index != -1 {
								// 現在選択されているツリーアイテムを取得
								treeItem := tv.CurrentItem()
								if _, ok := treeItem.(*model.ZipTreeItem); ok {

									// モデルから行データを取得
									itemModel := tableView.Model().(*model.FileItemModel)
									if index >= 0 && index < len(itemModel.Items) {
										item := itemModel.Items[index]

										// 削除フラグを更新
										fileops.UpdateDeleteFlagRecursively(currentZipPath, item)

										// テーブルを更新
										err := tableView.SetModel(itemModel)
										if err != nil {
											return
										}
									}
								}
							}
						},
					},
				},
			},
			// ボタンエリア
			Composite{
				Layout: HBox{MarginsZero: true},
				Children: []Widget{
					HSpacer{}, // 右寄せのためのスペーサー
					PushButton{
						Text: "削除",
						OnClicked: func() {
							if currentZipPath == "" {
								return
							}
							// すでに削除中なら実行しない
							if fileListModel.IsDeleting(currentZipPath) {
								walk.MsgBox(mw, "情報", "現在選択中のZIPは削除処理中です。完了までお待ちください。", walk.MsgBoxIconInformation)
								return
							}
							// 確認
							if walk.MsgBox(mw, "確認", "削除フラグが付いたファイルを削除しますか？", walk.MsgBoxIconQuestion|walk.MsgBoxYesNo) != walk.DlgCmdYes {
								return
							}
							// 削除対象のパスをキャプチャ
							targetZip := currentZipPath
							// 左ペインに削除中を表示
							fileListModel.SetDeleting(targetZip, true)
							// 非同期処理開始（並列可）
							go func() {
								err := fileops.DeleteFlaggedFiles(targetZip)
								// UIスレッドで更新
								mw.Synchronize(func() {
									// 状態解除
									fileListModel.SetDeleting(targetZip, false)
									if err != nil {
										walk.MsgBox(mw, "エラー", "ファイルの削除に失敗しました: "+err.Error(), walk.MsgBoxIconError)
										return
									}
									// 成功時はダイアログを表示しない
									// 現在選択中が対象ZIPなら再読み込み
									if currentZipPath == targetZip {
										var loadErr error
										zipModel, loadErr = model.LoadZipFile(targetZip)
										if loadErr != nil {
											walk.MsgBox(mw, "エラー", "ZIPファイルの再読み込みに失敗しました: "+loadErr.Error(), walk.MsgBoxIconError)
											return
										}
										tv.SetModel(zipModel)
									}
								})
							}()
						},
					},
				},
			},
		},
	}).Create(); err != nil {
		log.Fatal(err)
	}

	// ツリービューにコンテキストメニューを設定
	tv.SetContextMenu(treeContextMenu)

	// 左ペインのモデルを設定
	fileListView.SetModel(fileListModel)

	// ドロップイベントを処理（D&DされたZIPを左の一覧に追加）
	mw.DropFiles().Attach(func(files []string) {
		for _, file := range files {
			if filepath.Ext(file) == ".zip" {
				fileListModel.AddPath(file)
			}
		}
	})

	// 左側のZIPファイル一覧の選択変更で読み込み
	fileListView.CurrentIndexChanged().Attach(func() {
		idx := fileListView.CurrentIndex()
		if idx < 0 {
			return
		}
		path := fileListModel.PathAt(idx)
		if path == "" {
			return
		}
		// 削除中は開かない
		if fileListModel.IsDeleting(path) {
			// 元に戻す
			if lastFileListIndex >= 0 && lastFileListIndex < fileListView.Model().(interface{ RowCount() int }).RowCount() {
				fileListView.SetCurrentIndex(lastFileListIndex)
			}
			walk.MsgBox(mw, "情報", "このZIPは削除処理中のため開けません。", walk.MsgBoxIconInformation)
			return
		}
		// 正常読み込み
		var err error
		currentZipPath = path
		zipModel, err = model.LoadZipFile(path)
		if err != nil {
			walk.MsgBox(mw, "エラー", "ZIPファイルの読み込みに失敗しました: "+err.Error(), walk.MsgBoxIconError)
			return
		}
		tv.SetModel(zipModel)
		mw.SetTitle("ZIP ファイルビューア - " + filepath.Base(path))
		lastFileListIndex = idx
	})

	// ツリービューの選択変更イベントを処理
	tv.CurrentItemChanged().Attach(func() {
		if zipModel == nil {
			return
		}

		// 現在選択されているアイテムを取得
		item := tv.CurrentItem()
		if zipItem, ok := item.(*model.ZipTreeItem); ok {
			// ディレクトリしかない
			if zipItem.IsDir() {
				err := fileops.UpdateFileList(tableView, zipItem)
				if err != nil {
					walk.MsgBox(mw, "エラー", "ファイル一覧の更新に失敗しました: "+err.Error(), walk.MsgBoxIconError)
				}
			}
		}
	})

	// アイテムがダブルクリックされたときの処理（ファイルを開くなどの操作を追加できる）
	tableView.ItemActivated().Attach(func() {
		// 現在選択されているツリーアイテム（右側はファイル一覧なのでディレクトリ配下）
		if tv.CurrentItem() == nil || currentZipPath == "" {
			return
		}

		row := tableView.CurrentIndex()
		if row < 0 {
			return
		}
		m, ok := tableView.Model().(*model.FileItemModel)
		if !ok || row < 0 || row >= len(m.Items) {
			return
		}
		fileItem := m.Items[row]

		// 一時フォルダに展開
		extractedPath, err := fileops.ExtractFileToTemp(currentZipPath, fileItem.GetPath())
		if err != nil {
			walk.MsgBox(mw, "エラー", "ファイルの展開に失敗しました: "+err.Error(), walk.MsgBoxIconError)
			return
		}

		// 既定のアプリケーションで開く（Windowsの関連付け）
		// cmd /C start "" <path>
		cmd := exec.Command("cmd", "/C", "start", "", extractedPath)
		if err := cmd.Start(); err != nil {
			walk.MsgBox(mw, "エラー", "ファイルを開けませんでした: "+err.Error(), walk.MsgBoxIconError)
			return
		}
	})

	// ウィンドウを表示してメッセージループを開始
	mw.Show()
	mw.Run()
}
