package main

import (
	"log"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

func main() {
	// Create a new main window
	mw := new(walk.MainWindow)

	// Configure the main window
	if err := (MainWindow{
		AssignTo: &mw,
		Title:    "Simple Walk GUI Application",
		MinSize:  Size{Width: 500, Height: 300},
		Layout:   VBox{},
		Children: []Widget{
			PushButton{
				Text: "Click Me!",
				OnClicked: func() {
					walk.MsgBox(mw, "Message", "Hello, Walk!", walk.MsgBoxIconInformation)
				},
			},
		},
	}).Create(); err != nil {
		log.Fatal(err)
	}

	// Show the window and start the message loop
	mw.Show()
	mw.Run()
}