package main

import (
	"log"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

func main() {
	// メインウィンドウを作成
	mw := new(walk.MainWindow)

	// メインウィンドウを設定
	if err := (MainWindow{
		AssignTo: &mw,
		Title:    "シンプルなWalk GUIアプリケーション",
		MinSize:  Size{Width: 500, Height: 300},
		Layout:   VBox{},
		Children: []Widget{
			PushButton{
				Text: "クリックしてください！",
				OnClicked: func() {
					walk.MsgBox(mw, "メッセージ", "こんにちは、Walk！", walk.MsgBoxIconInformation)
				},
			},
		},
	}).Create(); err != nil {
		log.Fatal(err)
	}

	// ウィンドウを表示してメッセージループを開始
	mw.Show()
	mw.Run()
}
