package filter

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"website-copier/cmd/filter/lib"
	"website-copier/cmd/records"
	"website-copier/cmd/utils"

	"github.com/sqweek/dialog"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func CreateFilterScreen(myWindow fyne.Window) fyne.CanvasObject {
	// Variables to store selected files
	var selectedInputFiles []string
	var databaseFilePath string
	fileHeaders := make(map[string][]string)
	selectedHeaders := make(map[string][]string)

	// Input Elements
	inputPathEntry, selectFolderBtn, selectFilesBtn, clearInputSelectionBtn, fileListContainer := createInputElements(&selectedInputFiles, fileHeaders, selectedHeaders, myWindow)
	databaseFileEntry, selectDatabaseFileBtn, clearDatabaseFileBtn := createDatabaseElements(&databaseFilePath)

	// Output Elements
	outputOptionRadio, outputOptionsContainer := createOutputSelectionElements(selectedInputFiles)
	startBtn := createStartButton(&selectedInputFiles, &databaseFilePath, outputOptionRadio, outputOptionsContainer, myWindow)

	// Log Viewer
	logViewer := createLogViewer()

	// Layout
	content := container.NewVSplit(
		container.NewVBox(
			widget.NewLabelWithStyle("Input Selection", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			inputPathEntry,
			widget.NewLabelWithStyle("Selected Files", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			container.NewHBox(selectFolderBtn, selectFilesBtn, clearInputSelectionBtn),
			widget.NewLabelWithStyle("File Headers", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			fileListContainer,
			widget.NewLabelWithStyle("Database File", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			databaseFileEntry,
			container.NewHBox(selectDatabaseFileBtn, clearDatabaseFileBtn),
			widget.NewLabelWithStyle("Output Selection", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			outputOptionRadio,
			outputOptionsContainer,
			startBtn,
		),
		container.NewVScroll(logViewer),
	)

	content.Offset = 0.75
	return content
}

// createInputElements initializes the input selection elements
func createInputElements(selectedInputFiles *[]string,
	fileHeaders map[string][]string,
	selectedHeaders map[string][]string, myWindow fyne.Window) (*widget.Entry, *widget.Button, *widget.Button, *widget.Button, fyne.CanvasObject) {
	inputPathEntry := widget.NewMultiLineEntry()
	inputPathEntry.SetPlaceHolder("No input files or folders selected")
	inputPathEntry.Disable() // Make it read-only

	// File list and header display
	fileList := widget.NewList(
		func() int { return len(*selectedInputFiles) },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(i widget.ListItemID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(filepath.Base((*selectedInputFiles)[i]))
		},
	)

	headerDisplay := widget.NewMultiLineEntry()
	headerDisplay.SetPlaceHolder("Select a file to view its headers")
	headerDisplay.Disable() // Read-only

	fileList.OnSelected = func(id widget.ListItemID) {
		file := (*selectedInputFiles)[id]
		headers := fileHeaders[file]
		if len(headers) > 5 {
			// Show modal to select headers
			lib.ShowHeaderSelectionModal(myWindow, file, headers, selectedHeaders, headerDisplay)
		} else {
			// Display headers in the headerDisplay area
			headerText := fmt.Sprintf("Headers for %s:\n%s", filepath.Base(file), strings.Join(headers, ", "))
			headerDisplay.SetText(headerText)
			// Store the selected headers
			selectedHeaders[file] = headers
		}

	}

	// Show the folder selection dialog
	selectFolderBtn := lib.ShowFolderSelectionDialog(selectedInputFiles, fileHeaders, inputPathEntry, headerDisplay, fileList)

	//Show the file selection dialog
	selectFilesBtn := lib.ShowFileSelectionDialog(selectedInputFiles, fileHeaders, inputPathEntry, fileList, myWindow)

	//Clear selection button
	clearInputSelectionBtn := lib.ClearSelectionButton(selectedInputFiles, fileHeaders, inputPathEntry, headerDisplay, fileList)

	// Container for file list and header display
	fileListContainer := container.NewHSplit(
		container.NewVScroll(fileList),
		headerDisplay,
	)
	fileListContainer.Offset = 0.3 // Adjust the split ratio as needed

	return inputPathEntry, selectFolderBtn, selectFilesBtn, clearInputSelectionBtn, fileListContainer
}

// createDatabaseElements initializes the database selection elements
func createDatabaseElements(databaseFilePath *string) (*widget.Entry, *widget.Button, *widget.Button) {
	databaseFileEntry := widget.NewEntry()
	databaseFileEntry.SetPlaceHolder("No database file selected")
	databaseFileEntry.Disable() // Make it read-only

	selectDatabaseFileBtn := widget.NewButton("Select Database File", func() {
		filePath, err := dialog.File().Title("Select Database File").Filter("CSV Files", "csv").Load()
		if err != nil {
			return // User cancelled or an error occurred
		}
		*databaseFilePath = filePath
		databaseFileEntry.SetText(*databaseFilePath)
	})

	clearDatabaseFileBtn := widget.NewButton("Clear Database File", func() {
		*databaseFilePath = ""
		databaseFileEntry.SetText("")
	})

	return databaseFileEntry, selectDatabaseFileBtn, clearDatabaseFileBtn
}

// createOutputSelectionElements initializes the output selection elements
func createOutputSelectionElements(selectedInputFiles []string) (*widget.RadioGroup, *fyne.Container) {
	outputOptionRadio := widget.NewRadioGroup([]string{"Select Existing CSV File", "Specify Output Folder and Filename", "Generate Output Filename"}, nil)
	outputOptionRadio.SetSelected("Specify Output Folder and Filename") // Default option

	outputOptionsContainer := container.NewVBox()
	updateOutputOptions(selectedInputFiles, outputOptionRadio.Selected, outputOptionsContainer)

	outputOptionRadio.OnChanged = func(selected string) {
		updateOutputOptions(selectedInputFiles, selected, outputOptionsContainer)
	}

	return outputOptionRadio, outputOptionsContainer
}

// updateOutputOptions updates the output options container based on selected option
func updateOutputOptions(selectedInputFiles []string, selected string, container *fyne.Container) {
	container.Objects = nil
	if selected == "Select Existing CSV File" {
		outputFileEntry := widget.NewEntry()
		outputFileEntry.SetPlaceHolder("No output file selected")
		outputFileEntry.Disable()
		selectOutputFileBtn := widget.NewButton("Select Output File", func() {
			filePath, err := dialog.File().Title("Select Output CSV File").Filter("CSV Files", "csv").Load()
			if err != nil {
				return // User cancelled or an error occurred
			}
			outputFileEntry.SetText(filePath)
		})

		container.Add(outputFileEntry)
		container.Add(selectOutputFileBtn)
	} else if selected == "Specify Output Folder and Filename" {
		outputPathEntry := widget.NewEntry()
		outputPathEntry.SetPlaceHolder("No output folder selected")
		selectOutputFolderBtn := widget.NewButton("Select Output Folder", func() {
			folderPath, err := dialog.Directory().Title("Select Output Folder").Browse()
			if err != nil {
				return // User cancelled or an error occurred
			}
			outputPathEntry.SetText(folderPath)
		})

		outputFileNameEntry := widget.NewEntry()
		outputFileNameEntry.SetPlaceHolder("Enter output file name (e.g., filtered_output.csv)")

		container.Add(outputPathEntry)
		container.Add(selectOutputFolderBtn)
		container.Add(outputFileNameEntry)

	} else if selected == "Generate Output Filename" {
		outputPathEntry := widget.NewEntry()
		outputPathEntry.SetPlaceHolder("No output folder selected")
		selectOutputFolderBtn := widget.NewButton("Select Output Folder", func() {
			folderPath, err := dialog.Directory().Title("Select Output Folder").Browse()
			if err != nil {
				return // User cancelled or an error occurred
			}
			outputPathEntry.SetText(folderPath)
			if len(selectedInputFiles) > 0 {
				// Automatically generate output filename from the first input file
				inputFile := selectedInputFiles[0]
				fileName := strings.TrimSuffix(filepath.Base(inputFile), filepath.Ext(inputFile))
				outputFileName := fmt.Sprintf("%s_filtered_output.csv", fileName)
				outputPathEntry.SetText(filepath.Join(folderPath, outputFileName))
			}
		})

		container.Add(outputPathEntry)
		container.Add(selectOutputFolderBtn)
	}

	container.Refresh()
}

// createStartButton initializes the start button for filtering
func createStartButton(selectedInputFiles *[]string, databaseFilePath *string, outputOptionRadio *widget.RadioGroup, outputOptionsContainer *fyne.Container, myWindow fyne.Window) *widget.Button {
	return widget.NewButton("Start Filtering", func() {
		go func() {
			// Input validation
			if len(*selectedInputFiles) == 0 {
				utils.ShowError(fmt.Errorf("Please select input files or folders"), myWindow)
				return
			}

			if *databaseFilePath == "" {
				utils.ShowError(fmt.Errorf("Please select a database file"), myWindow)
				return
			}

			var outputFilePath string
			outputOption := outputOptionRadio.Selected
			if outputOption == "Select Existing CSV File" {
				outputFileEntry := outputOptionsContainer.Objects[0].(*widget.Entry)
				outputFilePath = outputFileEntry.Text
				if outputFilePath == "" {
					utils.ShowError(fmt.Errorf("Please select an output file"), myWindow)
					return
				}
			} else if outputOption == "Specify Output Folder and Filename" {
				outputPathEntry := outputOptionsContainer.Objects[0].(*widget.Entry)
				outputFileNameEntry := outputOptionsContainer.Objects[2].(*widget.Entry)
				outputFolder := outputPathEntry.Text
				outputFileName := outputFileNameEntry.Text

				if outputFolder == "" {
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
				outputFilePath = filepath.Join(outputFolder, outputFileName)
			} else if outputOption == "Generate Output Filename" {
				outputPathEntry := outputOptionsContainer.Objects[0].(*widget.Entry)
				outputFolder := outputPathEntry.Text

				if outputFolder == "" {
					utils.ShowError(fmt.Errorf("Please select an output folder"), myWindow)
					return
				}

				if len(*selectedInputFiles) > 0 {
					// Automatically generate output filename from the first input file
					inputFile := (*selectedInputFiles)[0]
					fileName := strings.TrimSuffix(filepath.Base(inputFile), filepath.Ext(inputFile))
					outputFilePath = filepath.Join(outputFolder, fmt.Sprintf("%s_filtered_output.csv", fileName))
				}
			}
			// Open the log file for writing
			logFilePath := filepath.Join(filepath.Dir(outputFilePath), "process_log.txt")
			logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
			if err != nil {
				utils.ShowError(fmt.Errorf("Failed to open log file: %v", err), myWindow)
				return
			}
			defer logFile.Close()

			// Set up the logger to write to the log file
			utils.Logger = log.New(logFile, "", log.Ldate|log.Ltime)

			// Perform filtering
			err = filterEmails(*selectedInputFiles, *databaseFilePath, outputFilePath)
			if err != nil {
				utils.ShowError(fmt.Errorf("Error during filtering: %v", err), myWindow)
				return
			}

			utils.ShowInfo("Email filtering completed successfully!", myWindow)
		}()
	})
}

func createLogViewer() *widget.Entry {
	logContent := widget.NewMultiLineEntry()
	logContent.Wrapping = fyne.TextWrapWord

	// Periodically update the log viewer
	go func() {
		for {
			utils.LogMutex.Lock()
			logText := strings.Join(utils.LogMessages, "\n")
			logContent.SetText(logText)
			utils.LogMutex.Unlock()
			time.Sleep(500 * time.Millisecond)
		}
	}()

	return logContent
}

// filterEmails filters emails from input files based on the database file and writes to the output file
func filterEmails(inputPaths []string, databaseFilePath, outputFilePath string) error {
	// Load database emails
	dbEmails, err := records.LoadEmailsFromCSV(databaseFilePath)
	utils.LogMessage(fmt.Sprintf("Loaded %d emails from database file", len(dbEmails)))

	if err != nil {
		return fmt.Errorf("failed to load database file: %v", err)
	}

	// Load input records from all selected files or folders
	var inputRecords []records.Record
	for _, path := range inputPaths {
		fileInfo, err := os.Stat(path)
		if err != nil {
			return fmt.Errorf("failed to access input path: %v", err)
		}
		if fileInfo.IsDir() {
			utils.LogMessage(fmt.Sprintf("Processing directory: %s", path))
			// If it's a directory, walk through it and load records
			err = filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
				if err != nil {
					return nil // Skip this file and continue
				}

				utils.LogMessage(fmt.Sprintf("Processing file: %s", p))
				if !info.IsDir() {
					ext := strings.ToLower(filepath.Ext(p))
					if ext == ".csv" || ext == ".xlsx" {
						utils.LogMessage(fmt.Sprintf("Loading records from file: %s", p))
						records, _, err := records.LoadRecords(p)
						if err != nil {
							// Log the error and continue
							return nil
						}
						inputRecords = append(inputRecords, records...)
						utils.LogMessage(fmt.Sprintf("Loaded %d records from file: %s", len(records), p))
					}
				}
				return nil
			})
			if err != nil {
				return fmt.Errorf("error walking the directory: %v", err)
			}
			utils.LogMessage(fmt.Sprintf("Processed directory: %s", path))
		} else {
			// It's a file
			ext := strings.ToLower(filepath.Ext(path))
			if ext == ".csv" || ext == ".xlsx" {
				utils.LogMessage(fmt.Sprintf("Loading records from file: %s", path))
				records, _, err := records.LoadRecords(path)
				if err != nil {
					// Log the error and continue
					continue
				}
				inputRecords = append(inputRecords, records...)
				utils.LogMessage(fmt.Sprintf("Loaded %d records from file: %s", len(records), path))
			}
		}

	}

	if len(inputRecords) == 0 {
		return fmt.Errorf("no valid input records found")
	}

	// Filter records
	var filteredRecords []records.Record
	utils.LogMessage(fmt.Sprintf("Filtering records based on database file: %s", databaseFilePath))
	for _, record := range inputRecords {
		utils.LogMessage(fmt.Sprintf("Processing record: %s", record.Email))
		if !dbEmails[record.Email] {
			filteredRecords = append(filteredRecords, record)
		}
	}

	utils.LogMessage(fmt.Sprintf("Filtered %d records based on database file", len(filteredRecords)))
	// Write output file
	headers := []string{"Name", "Email", "OrgName"} // Replace with actual headers if different
	err = records.WriteFilteredCSV(outputFilePath, headers, filteredRecords)
	utils.LogMessage(fmt.Sprintf("Wrote %d records to output file: %s", len(filteredRecords), outputFilePath))
	if err != nil {
		return fmt.Errorf("failed to write output file: %v", err)
	}

	utils.LogMessage(fmt.Sprint("Email filtering completed successfully!"))

	return nil
}
