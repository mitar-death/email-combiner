package filter

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"website-copier/cmd/filter/lib"
	"website-copier/cmd/records"
	"website-copier/cmd/utils"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/sqweek/dialog"
)

func CreateFilterScreen(myWindow fyne.Window) fyne.CanvasObject {
	var (
		selectedInputFiles []string
		databaseFilePath   string
		fileHeaders        = make(map[string][]string)
		selectedHeaders    = make(map[string][]string)
	)

	// Initialize UI elements
	inputPathEntry, selectFolderBtn, selectFilesBtn, clearInputSelectionBtn, fileListContainer := createInputElements(&selectedInputFiles, fileHeaders, selectedHeaders, myWindow)
	databaseFileEntry, selectDatabaseFileBtn, clearDatabaseFileBtn := createDatabaseElements(&databaseFilePath)
	outputOptionRadio, outputOptionsContainer := createOutputSelectionElements(&selectedInputFiles)
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

// Helper function to initialize input elements
func createInputElements(selectedInputFiles *[]string, fileHeaders, selectedHeaders map[string][]string, myWindow fyne.Window) (
	*widget.Entry, *widget.Button, *widget.Button, *widget.Button, fyne.CanvasObject) {

	inputPathEntry := widget.NewMultiLineEntry()
	inputPathEntry.SetPlaceHolder("No input files or folders selected")
	inputPathEntry.Disable()

	fileList := widget.NewList(
		func() int { return len(*selectedInputFiles) },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(i widget.ListItemID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(filepath.Base((*selectedInputFiles)[i]))
		},
	)

	headerDisplay := widget.NewMultiLineEntry()
	headerDisplay.SetPlaceHolder("Select a file to view its headers")
	headerDisplay.Disable()

	fileList.OnSelected = func(id widget.ListItemID) {
		file := (*selectedInputFiles)[id]
		headers := fileHeaders[file]
		if len(headers) > 5 {
			lib.ShowHeaderSelectionModal(myWindow, file, headers, selectedHeaders, headerDisplay)
		} else {
			headerText := fmt.Sprintf("Headers for %s:\n%s", filepath.Base(file), strings.Join(headers, ", "))
			headerDisplay.SetText(headerText)
			selectedHeaders[file] = headers
		}
	}

	return inputPathEntry,
		lib.ShowFolderSelectionDialog(selectedInputFiles, fileHeaders, inputPathEntry, headerDisplay, fileList),
		lib.ShowFileSelectionDialog(selectedInputFiles, fileHeaders, inputPathEntry, fileList, myWindow),
		lib.ClearSelectionButton(selectedInputFiles, fileHeaders, inputPathEntry, headerDisplay, fileList),
		container.NewHSplit(container.NewVScroll(fileList), headerDisplay)
}

// Initialize Database Elements
func createDatabaseElements(databaseFilePath *string) (*widget.Entry, *widget.Button, *widget.Button) {
	databaseFileEntry := widget.NewEntry()
	databaseFileEntry.SetPlaceHolder("No database file selected")
	databaseFileEntry.Disable()

	return databaseFileEntry,
		widget.NewButton("Select Database File", func() {
			if filePath, err := dialog.File().Title("Select Database File").Filter("CSV Files", "csv").Load(); err == nil {
				*databaseFilePath = filePath
				databaseFileEntry.SetText(*databaseFilePath)
			}
		}),
		widget.NewButton("Clear Database File", func() {
			*databaseFilePath = ""
			databaseFileEntry.SetText("")
		})
}

// Create Output Selection Elements
func createOutputSelectionElements(selectedInputFiles *[]string) (*widget.RadioGroup, *fyne.Container) {
	outputOptionRadio := widget.NewRadioGroup([]string{"Select Existing CSV File", "Specify Output Folder and Filename", "Generate Output Filename"}, nil)
	outputOptionRadio.SetSelected("Specify Output Folder and Filename")

	outputOptionsContainer := container.NewVBox()
	updateOutputOptions(*selectedInputFiles, outputOptionRadio.Selected, outputOptionsContainer)

	outputOptionRadio.OnChanged = func(selected string) {
		updateOutputOptions(*selectedInputFiles, selected, outputOptionsContainer)
	}

	return outputOptionRadio, outputOptionsContainer
}

// Update Output Options based on selection
func updateOutputOptions(selectedInputFiles []string, selected string, container *fyne.Container) {
	container.Objects = nil
	switch selected {
	case "Select Existing CSV File":
		outputFileEntry := widget.NewEntry()
		outputFileEntry.SetPlaceHolder("No output file selected")
		outputFileEntry.Disable()
		selectOutputFileBtn := widget.NewButton("Select Output File", func() {
			filePath, err := dialog.File().Title("Select Output CSV File").Filter("CSV Files", "csv").Load()
			if err == nil {
				outputFileEntry.SetText(filePath)
			}
		})

		container.Add(outputFileEntry)
		container.Add(selectOutputFileBtn)

	case "Specify Output Folder and Filename":
		outputPathEntry := widget.NewEntry()
		outputPathEntry.SetPlaceHolder("No output folder selected")
		selectOutputFolderBtn := widget.NewButton("Select Output Folder", func() {
			folderPath, err := dialog.Directory().Title("Select Output Folder").Browse()
			if err == nil {
				outputPathEntry.SetText(folderPath)
			}
		})

		outputFileNameEntry := widget.NewEntry()
		outputFileNameEntry.SetPlaceHolder("Enter output file name (e.g., filtered_output.csv)")

		container.Add(outputPathEntry)
		container.Add(selectOutputFolderBtn)
		container.Add(outputFileNameEntry)

	case "Generate Output Filename":
		outputPathEntry := widget.NewEntry()
		outputPathEntry.SetPlaceHolder("No output folder selected")

		selectOutputFolderBtn := widget.NewButton("Select Output Folder", func() {
			folderPath, err := dialog.Directory().Title("Select Output Folder").Browse()
			if err == nil {
				outputPathEntry.SetText(folderPath)
			}
		})

		// Create a disabled entry to display the generated output filename
		outputFileNameEntry := widget.NewEntry()
		outputFileNameEntry.Disable()
		outputFileNameEntry.SetPlaceHolder("Output filename will be generated based on input files")

		// Generate the output filename based on the first selected input file
		if len(selectedInputFiles) > 0 {
			inputFile := selectedInputFiles[0]
			baseFileName := strings.TrimSuffix(filepath.Base(inputFile), filepath.Ext(inputFile))
			outputFileName := fmt.Sprintf("%s_filtered_output.csv", baseFileName)
			outputFileNameEntry.SetText(outputFileName)
		} else {
			outputFileNameEntry.SetText("output_filtered_output.csv")
		}

		container.Add(outputPathEntry)
		container.Add(selectOutputFolderBtn)
		container.Add(outputFileNameEntry)
	}

	container.Refresh()
}

// Create Start Button
func createStartButton(selectedInputFiles *[]string, databaseFilePath *string, outputOptionRadio *widget.RadioGroup, outputOptionsContainer *fyne.Container, myWindow fyne.Window) *widget.Button {
	return widget.NewButton("Start Filtering", func() {
		go func() {
			// Validate inputs
			if err := validateInputs(*selectedInputFiles, *databaseFilePath, outputOptionRadio, outputOptionsContainer, myWindow); err != nil {
				return
			}

			// Determine output file path
			outputFilePath := determineOutputPath(outputOptionRadio, outputOptionsContainer)
			if outputFilePath == "" {
				utils.ShowError(fmt.Errorf("Failed to determine output file path"), myWindow)
				return
			}

			// Open the log file for writing
			logFilePath := filepath.Join(filepath.Dir(""), "process_log.txt")

			if err := utils.InitializeLogger(logFilePath); err != nil {
				utils.ShowError(fmt.Errorf("Failed to initialize logger: %v", err), myWindow)
				return
			}

			// Perform filtering
			if err := filterEmails(*selectedInputFiles, *databaseFilePath, outputFilePath); err != nil {
				utils.ShowError(fmt.Errorf("Error during filtering: %v", err), myWindow)
				return
			}

			utils.ShowInfo("Email filtering completed successfully!", myWindow)
		}()
	})
}

// Validate user inputs
func validateInputs(selectedInputFiles []string, databaseFilePath string, outputOptionRadio *widget.RadioGroup, outputOptionsContainer *fyne.Container, myWindow fyne.Window) error {
	if len(selectedInputFiles) == 0 {
		utils.ShowError(fmt.Errorf("Please select input files or folders"), myWindow)
		return fmt.Errorf("no input files selected")
	}

	if databaseFilePath == "" {
		utils.ShowError(fmt.Errorf("Please select a database file"), myWindow)
		return fmt.Errorf("no database file selected")
	}

	outputOption := outputOptionRadio.Selected
	switch outputOption {
	case "Select Existing CSV File":
		outputFileEntry := outputOptionsContainer.Objects[0].(*widget.Entry)
		if outputFileEntry.Text == "" {
			utils.ShowError(fmt.Errorf("Please select an output file"), myWindow)
			return fmt.Errorf("no output file selected")
		}
	case "Specify Output Folder and Filename":
		outputPathEntry := outputOptionsContainer.Objects[0].(*widget.Entry)
		outputFileNameEntry := outputOptionsContainer.Objects[2].(*widget.Entry)
		if outputPathEntry.Text == "" {
			utils.ShowError(fmt.Errorf("Please select an output folder"), myWindow)
			return fmt.Errorf("no output folder selected")
		}
		if outputFileNameEntry.Text == "" {
			utils.ShowError(fmt.Errorf("Please enter an output file name"), myWindow)
			return fmt.Errorf("no output file name entered")
		}
		if !strings.HasSuffix(strings.ToLower(outputFileNameEntry.Text), ".csv") {
			utils.ShowError(fmt.Errorf("Output file name must have a .csv extension"), myWindow)
			return fmt.Errorf("invalid output file extension")
		}
	case "Generate Output Filename":
		outputPathEntry := outputOptionsContainer.Objects[0].(*widget.Entry)
		if outputPathEntry.Text == "" {
			utils.ShowError(fmt.Errorf("Please select an output folder"), myWindow)
			return fmt.Errorf("no output folder selected")
		}
	default:
		utils.ShowError(fmt.Errorf("Invalid output option selected"), myWindow)
		return fmt.Errorf("invalid output option")
	}

	return nil
}

// Determine the output file path based on user selection
func determineOutputPath(outputOptionRadio *widget.RadioGroup, outputOptionsContainer *fyne.Container) string {
	outputOption := outputOptionRadio.Selected
	var outputFilePath string

	switch outputOption {
	case "Select Existing CSV File":
		outputFileEntry := outputOptionsContainer.Objects[0].(*widget.Entry)
		outputFilePath = outputFileEntry.Text
	case "Specify Output Folder and Filename":
		outputPathEntry := outputOptionsContainer.Objects[0].(*widget.Entry)
		outputFileNameEntry := outputOptionsContainer.Objects[2].(*widget.Entry)
		outputFilePath = filepath.Join(outputPathEntry.Text, outputFileNameEntry.Text)
	case "Generate Output Filename":
		outputPathEntry := outputOptionsContainer.Objects[0].(*widget.Entry)
		outputFileNameEntry := outputOptionsContainer.Objects[2].(*widget.Entry) // The disabled entry showing the filename

		if outputPathEntry.Text == "" {
			return ""
		}

		outputFileName := outputFileNameEntry.Text
		if outputFileName == "" {
			// Fallback in case the output filename is not set
			outputFileName = "output_filtered_output.csv"
		}

		outputFilePath = filepath.Join(outputPathEntry.Text, outputFileName)
	}

	return outputFilePath
}

// Log Viewer for displaying logs
func createLogViewer() *widget.Entry {
	logContent := widget.NewMultiLineEntry()
	logContent.Wrapping = fyne.TextWrapWord
	logContent.Disable()

	// Update the log viewer periodically
	go func() {
		for {
			time.Sleep(1 * time.Second)
			utils.LogMutex.Lock()
			logContent.SetText(strings.Join(utils.LogMessages, "\n"))
			utils.LogMutex.Unlock()
		}
	}()

	return logContent
}

// filterEmails filters emails from input files based on the database file and writes to the output file
func filterEmails(inputPaths []string, databaseFilePath, outputFilePath string) error {
	// Load database emails
	dbEmails, err := records.LoadEmailsFromCSV(databaseFilePath)
	if err != nil {
		return fmt.Errorf("failed to load database file: %v", err)
	}
	utils.LogMessage(fmt.Sprintf("Loaded %d emails from database file", len(dbEmails)))

	var inputRecords []records.Record

	// Process input paths
	for _, path := range inputPaths {
		fileInfo, err := os.Stat(path)
		if err != nil {
			utils.LogMessage(fmt.Sprintf("Failed to access input path: %v", err))
			continue
		}

		if fileInfo.IsDir() {
			utils.LogMessage(fmt.Sprintf("Processing directory: %s", path))
			err = filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
				if err != nil {
					return nil // Skip this file and continue
				}
				if !info.IsDir() {
					ext := strings.ToLower(filepath.Ext(p))
					if ext == ".csv" || ext == ".xlsx" {
						utils.LogMessage(fmt.Sprintf("Loading records from file: %s", p))
						records, _, err := records.LoadRecords(p)
						if err != nil {
							utils.LogMessage(fmt.Sprintf("Error loading records from file %s: %v", p, err))
							return nil
						}
						inputRecords = append(inputRecords, records...)
						utils.LogMessage(fmt.Sprintf("Loaded %d records from file: %s", len(records), p))
					}
				}
				return nil
			})
			if err != nil {
				utils.LogMessage(fmt.Sprintf("Error walking the directory: %v", err))
			}
		} else {
			ext := strings.ToLower(filepath.Ext(path))
			if ext == ".csv" || ext == ".xlsx" {
				utils.LogMessage(fmt.Sprintf("Loading records from file: %s", path))
				records, _, err := records.LoadRecords(path)
				if err != nil {
					utils.LogMessage(fmt.Sprintf("Error loading records from file %s: %v", path, err))
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
	utils.LogMessage("Filtering records based on database emails")
	var filteredRecords []records.Record
	for _, record := range inputRecords {
		if !dbEmails[record.Email] {
			filteredRecords = append(filteredRecords, record)
		}
	}

	utils.LogMessage(fmt.Sprintf("Filtered %d records from %d input records", len(filteredRecords), len(inputRecords)))

	// Write output file
	headers := []string{"Name", "Email", "OrgName"} // Update as per actual headers
	err = records.WriteFilteredCSV(outputFilePath, headers, filteredRecords)
	if err != nil {
		return fmt.Errorf("failed to write output file: %v", err)
	}

	utils.LogMessage(fmt.Sprintf("Wrote %d records to output file: %s", len(filteredRecords), outputFilePath))
	utils.LogMessage("Email filtering completed successfully!")
	return nil
}
