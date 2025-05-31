package gui

import (
	"log"
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
	var tableView *walk.TableView
	var tv *walk.TreeView

	// 現在のZIPファイルパスとモデル
	var currentZipPath string
	var zipModel *model.ZipTreeModel

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
					// 左側：ツリービュー
					TreeView{
						AssignTo:           &tv,
						StretchFactor:      5, // 左右の比率
						AlwaysConsumeSpace: true,
					},
					// 右側：TableView（ファイル一覧表示用）
					TableView{
						AssignTo:           &tableView,
						StretchFactor:      5, // 左右の比率
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
								if zipItem, ok := treeItem.(*model.ZipTreeItem); ok {
									dirPath := zipItem.GetPath()

									// モデルから行データを取得
									itemModel := tableView.Model().(*model.FileItemModel)
									if index >= 0 && index < len(itemModel.Items) {
										item := itemModel.Items[index]

										// 完全なファイルパスを作成
										fullPath := dirPath + item.GetName()

										// 削除フラグを更新
										fileops.SetDeleteFlag(currentZipPath, fullPath, item.DeleteFlag)

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
			if filepath.Ext(file) == ".zip" {
				// ZIPファイルを読み込む
				var err error
				currentZipPath = file
				zipModel, err = model.LoadZipFile(file)
				if err != nil {
					walk.MsgBox(mw, "エラー", "ZIPファイルの読み込みに失敗しました: "+err.Error(), walk.MsgBoxIconError)
					return
				}

				// ツリービューにモデルを設定
				tv.SetModel(zipModel)

				// ウィンドウタイトルを更新
				mw.SetTitle("ZIP ファイルビューア - " + filepath.Base(file))
				break
			}
		}
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
		// 現在選択されているツリーアイテムを取得
		treeItem := tv.CurrentItem()
		if _, ok := treeItem.(*model.ZipTreeItem); ok {
			// 選択された行のインデックスを取得
			indexes := tableView.SelectedIndexes()
			if len(indexes) > 0 {
				// ここにファイルを開くなどの処理を追加できる
				// 例:
				// row := indexes[0]
				// model := tableView.Model().(*model.FileItemModel)
				// if row >= 0 && row < len(model.Items) {
				//     item := model.Items[row]
				//     walk.MsgBox(mw, "ファイル", item.GetName() + "が選択されました", walk.MsgBoxIconInformation)
				// }
			}
		}
	})

	// ウィンドウを表示してメッセージループを開始
	mw.Show()
	mw.Run()
}
