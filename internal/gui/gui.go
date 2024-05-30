package gui

import (
	"fmt"
	"log"
	"strconv"

	"fyne.io/fyne/theme"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
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

	sensorTab := NewSensorTab(sensors)
	gatewayTab := NewGatewayTab(gateway)
	consoleTab := NewConsoleTab(sensors, gateway)

	tabs := container.NewAppTabs(
		container.NewTabItem("Sensors", sensorTab),
		container.NewTabItem("Gateway", gatewayTab),
		container.NewTabItem("Console", consoleTab),
	)

	// Disable form submit button initially
	// Coulnd't do it in NewGatewayTab because the form is not yet created
	gatewayTab.Objects[0].(*widget.Form).Disable()

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

func NewSensorTab(sensors *[]model.Sensor) *fyne.Container {
	sensorNames := make([]string, len(*sensors))
	for i, s := range *sensors {
		sensorNames[i] = s.Name
	}
	sensorSelect := widget.NewSelect(sensorNames, nil)
	sensorSelect.PlaceHolder = "View sensor details"

	leftContainer := container.NewVBox(
		sensorSelect,
		NewSensorDetails(nil, sensors),
	)
	sensorSelect.OnChanged = func(s string) {
		for _, sensor := range *sensors {
			if sensor.Name == s {
				leftContainer.Objects[1] = NewSensorDetails(&sensor, sensors)
			}
		}
	}

	//rightContainer := container.NewVBox(
	//	widget.
	//)

	return container.NewBorder(nil, nil, leftContainer, nil, widget.NewSeparator())
}

func NewSensorDetails(sensor *model.Sensor, sensors *[]model.Sensor) *fyne.Container {
	if sensor == nil {
		return container.NewCenter(widget.NewLabel("No sensor selected"))
	}

	nameEntry := widget.NewEntry()
	nameEntry.SetText(sensor.Name)

	sensorTypes := ""
	for i, t := range sensor.Types {
		if i == len(sensor.Types)-1 {
			sensorTypes += t
		} else {
			sensorTypes += t + ", "
		}
	}

	wakeUpIntervalEntry := widget.NewEntry()
	wakeUpIntervalUnitsSelect := widget.NewSelect([]string{"seconds", "minutes", "hours"}, nil)
	wakeUpIntervalEntry.SetText(strconv.Itoa(sensor.WakeUpInterval))
	wakeUpIntervalUnitsSelect.Selected = "seconds"

	batteryLevelLabel := widget.NewLabel("Unknown")
	if sensor.BatteryLevel != -1 {
		batteryLevelLabel.SetText(fmt.Sprintf("%d mV", sensor.BatteryLevel))
	}

	details := container.NewGridWithColumns(
		2,
		widget.NewLabel("Name:"), nameEntry,
		widget.NewLabel("MAC Address:"), widget.NewLabel(model.MacToString(sensor.Mac)),
		widget.NewLabel("Sensor Types:"), widget.NewLabel(sensorTypes),
		widget.NewLabel("Wake-up Interval:"), container.NewGridWithColumns(2, wakeUpIntervalEntry, wakeUpIntervalUnitsSelect),
		widget.NewLabel("Battery Level:"), batteryLevelLabel,
	)

	settingsWidget := widget.NewAccordion()
	for setting, value := range sensor.Settings {
		settingWidget := container.NewVBox()
		for k, v := range value {
			if k == "active" {
				settingCheckbox := widget.NewCheck("", nil)
				settingCheckbox.SetChecked(v == "true")
				settingWidget.Add(container.NewGridWithColumns(2, widget.NewLabel(k+":"), settingCheckbox))
			} else {
				settingEntry := widget.NewEntry()
				settingEntry.SetText(v)
				settingWidget.Add(container.NewGridWithColumns(2, widget.NewLabel(k+":"), container.NewGridWithColumns(2, settingEntry, widget.NewLabel("Hz"))))
			}
		}
		settingsWidget.Append(widget.NewAccordionItem(setting, settingWidget))
	}

	buttons := container.NewHBox(
		layout.NewSpacer(),
		&widget.Button{
			Text:       "Save",
			OnTapped:   nil,
			Importance: widget.HighImportance,
			Icon:       theme.ConfirmIcon(),
		},
		&widget.Button{
			Text: "Delete",
			OnTapped: func() {
				model.RemoveSensor(sensor.Mac, sensors)
			},
			Importance: widget.DangerImportance,
			Icon:       theme.DeleteIcon(),
		},
	)

	return container.NewVBox(details, container.NewGridWithColumns(2, widget.NewLabel("Settings:"), settingsWidget), buttons)
}
