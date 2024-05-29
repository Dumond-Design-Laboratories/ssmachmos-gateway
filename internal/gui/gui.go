package gui

import (
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type Console struct {
}

func (c *Console) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func Start() {
	myApp := app.New()
	myWindow := myApp.NewWindow("SSMachMos")

	tabs := container.NewAppTabs(
		container.NewTabItem("Sensors", widget.NewLabel("Hello")),
		container.NewTabItem("Gateway", widget.NewLabel("World!")),
		container.NewTabItem("Console", widget.NewLabel("Hello World!")),
	)

	tabs.SetTabLocation(container.TabLocationTop)

	myWindow.SetContent(tabs)
	myWindow.ShowAndRun()
}
