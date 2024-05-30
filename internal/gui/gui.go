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
	c.output.ScrollToBottom()
	return len(p), nil
}

func Start(sensors *[]model.Sensor, gateway *model.Gateway) {
	myApp := app.New()
	myWindow := myApp.NewWindow("SSMachMos")

	myWindow.Resize(fyne.NewSize(800, 600))

	gatewayTab := NewGatewayTab(gateway)
	consoleTab := NewConsoleTab(sensors, gateway)

	tabs := container.NewAppTabs(
		container.NewTabItem("Sensors", widget.NewLabel("Hello")),
		container.NewTabItem("Gateway", gatewayTab),
		container.NewTabItem("Console", consoleTab),
	)

	tabs.SetTabLocation(container.TabLocationTop)

	myWindow.SetContent(tabs)
	myWindow.ShowAndRun()
}

func NewConsoleTab(sensors *[]model.Sensor, gateway *model.Gateway) *fyne.Container {
	console := &Console{}
	textGrid := widget.NewTextGrid()
	console.output = container.NewScroll(textGrid)

	console.input = widget.NewEntry()
	console.input.OnSubmitted = func(s string) {
		in.HandleInput(s, sensors, gateway)
		console.input.SetText("")
	}

	out.SetLogger(log.New(console, "", log.LstdFlags))
	return container.NewBorder(nil, console.input, nil, nil, console.output)
}

func NewGatewayTab(gateway *model.Gateway) *fyne.Container {
	idEntry := widget.NewEntry()
	passwordEntry := widget.NewPasswordEntry()

	idEntry.SetText(gateway.Id)
	passwordEntry.SetText(gateway.Password)

	form := widget.NewForm(
		widget.NewFormItem("Gateway ID:", idEntry),
		widget.NewFormItem("Gateway Password:", passwordEntry),
		widget.NewFormItem("Data Characteristic UUID:", widget.NewLabel(model.UuidToString(gateway.DataCharUUID))),
	)
	form.SubmitText = "Save"
	form.Disable()
	form.OnSubmit = func() {
		model.SetGatewayId(gateway, idEntry.Text)
		model.SetGatewayPassword(gateway, passwordEntry.Text)
		form.Disable()
	}
	idEntry.OnChanged = func(s string) {
		form.Enable()
	}
	passwordEntry.OnChanged = func(s string) {
		form.Enable()
	}

	return container.NewCenter(form)
}
