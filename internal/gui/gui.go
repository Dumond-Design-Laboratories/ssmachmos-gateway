package gui

import (
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/jukuly/ss_mach_mo/internal/cli/in"
	"github.com/jukuly/ss_mach_mo/internal/cli/out"
	"github.com/jukuly/ss_mach_mo/internal/model"
)

type Console struct {
	output *container.Scroll
	input  *widget.Entry
}

func (c *Console) Write(p []byte) (n int, err error) {
	c.output.Content.(*widget.TextGrid).SetText(c.output.Content.(*widget.TextGrid).Text() + string(p))
	return len(p), nil
}

func Start(sensors *[]model.Sensor, gateway *model.Gateway) {
	myApp := app.New()
	myWindow := myApp.NewWindow("SSMachMos")

	myWindow.Resize(fyne.NewSize(800, 600))

	console := NewConsole(sensors, gateway)

	tabs := container.NewAppTabs(
		container.NewTabItem("Sensors", widget.NewLabel("Hello")),
		container.NewTabItem("Gateway", widget.NewLabel("World!")),
		container.NewTabItem("Console", container.NewBorder(nil, console.input, nil, nil, console.output)),
	)

	tabs.SetTabLocation(container.TabLocationTop)

	myWindow.SetContent(tabs)
	myWindow.ShowAndRun()
}

func NewConsole(sensors *[]model.Sensor, gateway *model.Gateway) *Console {
	console := &Console{}
	textGrid := widget.NewTextGrid()
	console.output = container.NewScroll(textGrid)

	console.input = widget.NewEntry()
	console.input.OnSubmitted = func(s string) {
		in.HandleInput(s, sensors, gateway)
		console.input.SetText("")
		console.output.ScrollToBottom()
	}

	out.SetLogger(log.New(console, "", log.LstdFlags))
	return console
}
