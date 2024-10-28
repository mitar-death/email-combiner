package lib

import (
	"fmt"
	"path/filepath"
	"strings"
	"website-copier/cmd/utils"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func UpdateHeadersDisplay(headersAccordion *widget.Accordion, fileHeaders map[string][]string) {
	headersAccordion.Items = nil // Clear existing items
	for file, headers := range fileHeaders {
		fileName := filepath.Base(file)
		headerText := strings.Join(headers, ", ")
		headerText = strings.ToTitle(headerText)
		content := widget.NewLabel(headerText)
		item := widget.NewAccordionItem(fileName, content)
		headersAccordion.Append(item)
	}
	headersAccordion.Refresh()
}

func ShowHeaderSelectionModal(win fyne.Window, file string, headers []string, selectedHeaders map[string][]string, headerDisplay *widget.Entry) {
	// Create checkboxes for each header
	var checks []*widget.Check
	headerChecks := make(map[string]*widget.Check)
	for _, header := range headers {
		check := widget.NewCheck(header, nil)
		check.SetChecked(true) // Default to selected
		checks = append(checks, check)
		headerChecks[header] = check
	}

	// Create the content
	content := container.NewVBox()
	content.Add(widget.NewLabelWithStyle(fmt.Sprintf("Select headers to use from %s:", filepath.Base(file)), fyne.TextAlignLeading, fyne.TextStyle{Bold: true}))
	for _, check := range checks {
		content.Add(check)
	}

	var modal *widget.PopUp // Declare modal here to access it in button handlers

	// Create buttons for OK and Cancel
	buttons := container.NewBorder(nil, nil, nil, nil,
		container.NewHBox(
			widget.NewButtonWithIcon("OK", theme.ConfirmIcon(), func() {
				// Collect selected headers
				var selected []string
				for header, check := range headerChecks {
					if check.Checked {
						selected = append(selected, header)
					}
				}
				if len(selected) == 0 {
					dialog.ShowInformation("No Headers Selected", "At least one header must be selected.", win)

					return
				}
				// Store the selected headers
				selectedHeaders[file] = selected
				// Update the header display
				utils.DisplayHeadersInList(headerDisplay, selected)
				// Close the modal
				win.Canvas().Overlays().Remove(modal)
			}),
			widget.NewButtonWithIcon("Cancel", theme.CancelIcon(), func() {
				// Close the modal without saving
				win.Canvas().Overlays().Remove(modal)
			}),
		),
	)
	buttons.Resize(fyne.NewSize(buttons.MinSize().Width, 50)) // Make buttons wider
	buttons.Resize(fyne.NewSize(buttons.MinSize().Width, 50)) // Make buttons wider

	// Create the modal content
	scrollableContent := container.NewVScroll(content)
	scrollableContent.SetMinSize(fyne.NewSize(500, 300)) // Set the minimum size for the scrollable content

	modalContent := container.NewBorder(nil, buttons, nil, nil, scrollableContent)
	modal = widget.NewModalPopUp(modalContent, win.Canvas())
	modal.Show()
}
