package main

import (
	"website-copier/cmd/combine"
	"website-copier/cmd/filter"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func main() {

	// if err := agent.Listen(agent.Options{}); err != nil {
	// 	log.Fatal(err)
	// }

	// Create the GUI application
	myApp := app.New()
	myWindow := myApp.NewWindow("DataMerge Pro")

	// Create a menu to switch between screens
	menu := container.NewHBox()

	// Create content containers for each screen
	combineScreen := combine.CreateCombineScreen(myWindow)
	filterScreen := filter.CreateFilterScreen(myWindow)

	// Create a container to hold the current screen content
	contentContainer := container.NewMax()

	// Function to switch screens
	switchScreen := func(screen fyne.CanvasObject) {
		contentContainer.Objects = []fyne.CanvasObject{screen}
		contentContainer.Refresh()
	}

	// Buttons to switch screens
	combineBtn := widget.NewButton("Combine Files", func() {
		switchScreen(combineScreen)
	})
	filterBtn := widget.NewButton("Filter Emails", func() {
		switchScreen(filterScreen)
	})

	menu.Objects = []fyne.CanvasObject{combineBtn, filterBtn}

	// Initial screen
	switchScreen(combineScreen)

	// Main layout
	mainContainer := container.NewBorder(menu, nil, nil, nil, contentContainer)

	myWindow.SetContent(mainContainer)
	myWindow.Resize(fyne.NewSize(800, 700))
	myWindow.ShowAndRun()

}
