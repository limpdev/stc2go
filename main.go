package main

import (
	_ "embed"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
)

//go:embed appicon.png
var appIcon []byte

func main() {
	myApp := app.New()
	myApp.SetIcon(fyne.NewStaticResource("appicon.png", appIcon))
	myWindow := myApp.NewWindow("Fynance")
	myWindow.Resize(fyne.NewSize(500, 400)) // Slightly wider for tabs
	myApp.Settings().SetTheme(newCustomTheme())

	// Create the individual tool interfaces
	stcTab := makeSTCTab(myWindow)
	rsuTab := makeRSUTab(myWindow) // New RSU Tab
	calcTab := makeCalculatorTab()

	// Create the navigation tabs
	tabs := container.NewAppTabs(
		container.NewTabItemWithIcon("EXERCISE", theme.DocumentIcon(), stcTab), // Renamed for clarity
		container.NewTabItemWithIcon("RELEASE", theme.AccountIcon(), rsuTab),   // New Tab
		container.NewTabItemWithIcon("KEYS", theme.ContentAddIcon(), calcTab),
	)

	tabs.SetTabLocation(container.TabLocationTop)

	myWindow.SetContent(tabs)
	myWindow.ShowAndRun()
}
