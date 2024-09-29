package filter

import (
	"fmt"
	"path/filepath"
	"strings"
	"website-copier/cmd/records"
	"website-copier/cmd/utils"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func CreateFilterScreen(myWindow fyne.Window) fyne.CanvasObject {
	var inputFilePath string
	var databaseFilePath string

	// Input Elements
	inputFileEntry := widget.NewEntry()
	inputFileEntry.SetPlaceHolder("No input file selected")
	inputFileEntry.Disable() // Make it read-only

	selectInputFileBtn := widget.NewButton("Select Input File", func() {
		utils.ShowFileOpenDialog(&inputFilePath, inputFileEntry, myWindow)
	})

	clearInputFileBtn := widget.NewButton("Clear Input File", func() {
		inputFilePath = ""
		inputFileEntry.SetText("")
	})

	// Database File Elements
	databaseFileEntry := widget.NewEntry()
	databaseFileEntry.SetPlaceHolder("No database file selected")
	databaseFileEntry.Disable() // Make it read-only

	selectDatabaseFileBtn := widget.NewButton("Select Database File", func() {
		utils.ShowFileOpenDialog(&databaseFilePath, databaseFileEntry, myWindow)
	})

	clearDatabaseFileBtn := widget.NewButton("Clear Database File", func() {
		databaseFilePath = ""
		databaseFileEntry.SetText("")
	})

	// Output Elements
	outputPathEntry := widget.NewEntry()
	outputPathEntry.SetPlaceHolder("No output folder selected")
	// outputPathEntry.Disable() // Make it read-only

	selectOutputFolderBtn := widget.NewButton("Select Output Folder", func() {
		utils.ShowFolderSelectionDialog(outputPathEntry, myWindow)
	})

	outputFileNameEntry := widget.NewEntry()
	outputFileNameEntry.SetPlaceHolder("Enter output file name (e.g., filtered_output.csv)")

	// Start Button
	startBtn := widget.NewButton("Start Filtering", func() {
		go func() {
			outputPath := outputPathEntry.Text
			outputFileName := outputFileNameEntry.Text

			if inputFilePath == "" {
				utils.ShowError(fmt.Errorf("Please provide an input file"), myWindow)
				return
			}
			if databaseFilePath == "" {
				utils.ShowError(fmt.Errorf("Please provide a database file"), myWindow)
				return
			}
			if outputPath == "" {
				utils.ShowError(fmt.Errorf("Please select an output folder"), myWindow)
				return
			}
			if outputFileName == "" {
				utils.ShowError(fmt.Errorf("Please enter an output file name"), myWindow)
				return
			}

			// Ensure the output file has a .csv extension
			if !strings.HasSuffix(strings.ToLower(outputFileName), ".csv") {
				utils.ShowError(fmt.Errorf("Output file name must have a .csv extension"), myWindow)
				return
			}

			// Perform filtering
			err := filterEmails(inputFilePath, databaseFilePath, outputPath, outputFileName)
			if err != nil {
				utils.ShowError(fmt.Errorf("Error during filtering: %v", err), myWindow)
				return
			}

			utils.ShowInfo("Email filtering completed successfully!", myWindow)
		}()
	})

	// Notes to inform users about dialog refresh
	outputNote := widget.NewLabel("Note: If you create a new file or folder, close and reopen the dialog to refresh.")

	// Layout
	content := container.NewVBox(
		widget.NewLabelWithStyle("Input File", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		inputFileEntry,
		container.NewHBox(selectInputFileBtn, clearInputFileBtn),
		widget.NewLabelWithStyle("Database File", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		databaseFileEntry,
		container.NewHBox(selectDatabaseFileBtn, clearDatabaseFileBtn),
		widget.NewLabelWithStyle("Output Selection", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		outputPathEntry,
		selectOutputFolderBtn,
		outputNote,
		outputFileNameEntry,
		startBtn,
	)

	return container.NewVScroll(content)
}

func filterEmails(inputFilePath, databaseFilePath, outputPath, outputFileName string) error {
	// Load database emails
	dbEmails, err := records.LoadEmailsFromCSV(databaseFilePath)
	if err != nil {
		return fmt.Errorf("failed to load database file: %v", err)
	}

	// Load input records
	inputRecords, headers, err := records.LoadRecords(inputFilePath)
	if err != nil {
		return fmt.Errorf("failed to load input file: %v", err)
	}

	// Filter records
	var filteredRecords []records.Record
	for _, record := range inputRecords {
		if !dbEmails[record.Email] {
			filteredRecords = append(filteredRecords, record)
		}
	}

	// Write output file
	outputFilePath := filepath.Join(outputPath, outputFileName)
	err = records.WriteFilteredCSV(outputFilePath, headers, filteredRecords)
	if err != nil {
		return fmt.Errorf("failed to write output file: %v", err)
	}

	return nil
}
