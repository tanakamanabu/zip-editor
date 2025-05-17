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
	var te *walk.TextEdit
	var tv *walk.TreeView

	// 現在のZIPファイルパスとモデル
	var currentZipPath string
	var zipModel = model.CreateEmptyZipTreeModel()

	// メインウィンドウを設定
	if err := (MainWindow{
		AssignTo: &mw,
		Title:    "ZIP ファイルビューア",
		MinSize:  Size{Width: 700, Height: 400},
		Layout:   VBox{},
		Children: []Widget{
			// 水平分割レイアウト
			Composite{
				Layout:        HBox{MarginsZero: true},
				StretchFactor: 10,
				Children: []Widget{
					// 左側：ツリービュー
					TreeView{
						AssignTo:           &tv,
						StretchFactor:      5, // 左右の比率
						AlwaysConsumeSpace: true,
						Model:              zipModel, // 初期状態ではnilになります
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
			// 選択されたディレクトリ内のファイル一覧を表示
			err := fileops.UpdateFileList(te, currentZipPath, zipItem.GetPath())
			if err != nil {
				walk.MsgBox(mw, "エラー", "ファイル一覧の更新に失敗しました: "+err.Error(), walk.MsgBoxIconError)
			}
		}
	})

	// ウィンドウを表示してメッセージループを開始
	mw.Show()
	mw.Run()
}
