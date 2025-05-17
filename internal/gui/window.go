package gui

import (
	"log"
	"path/filepath"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

// CreateMainWindow はメインウィンドウを作成し、表示します
func CreateMainWindow() {
	// メインウィンドウを作成
	mw := new(walk.MainWindow)
	var te *walk.TextEdit

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
					// 左側：ツリービュー（簡略化）
					TreeView{
						StretchFactor:      5, // 左右の比率
						AlwaysConsumeSpace: true,
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

	// ドロップイベントを処理（簡略化）
	mw.DropFiles().Attach(func(files []string) {
		for _, file := range files {
			// ZIPファイルかどうかを確認
			if filepath.Ext(file) == ".zip" {
				// ウィンドウタイトルを更新
				mw.SetTitle("ZIP ファイルビューア - " + filepath.Base(file))
				break
			}
		}
	})

	// ウィンドウを表示してメッセージループを開始
	mw.Show()
	mw.Run()
}