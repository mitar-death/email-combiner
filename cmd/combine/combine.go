package combine

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"website-copier/cmd/records"
	"website-copier/cmd/utils"

	"github.com/sqweek/dialog"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func CreateCombineScreen(myWindow fyne.Window) fyne.CanvasObject {
	// Variables to store selected files
	var selectedFiles []string

	// Create Input Selection Widgets
	inputPathEntry := createInputPathEntry()
	selectFolderBtn, selectFileBtn, clearFilesBtn := createInputButtons(inputPathEntry, &selectedFiles, myWindow)

	// Create Output Selection Widgets
	outputOptionRadio, outputOptionsContainer := createOutputWidgets()

	// Create Start Button
	startBtn := createStartButton(
		inputPathEntry,
		outputOptionRadio,
		outputOptionsContainer,
		&selectedFiles,
		myWindow,
	)

	// Create Log Viewer
	logContent := createLogViewer()

	// Adjusted Layout
	split := container.NewVSplit(
		container.NewVBox(
			widget.NewLabelWithStyle("Input Selection", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			inputPathEntry,
			container.NewHBox(selectFolderBtn, selectFileBtn, clearFilesBtn),
			widget.NewLabelWithStyle("Output Selection", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			outputOptionRadio,
			outputOptionsContainer,
			startBtn,
		),
		container.NewVScroll(logContent),
	)

	split.Offset = 0.75
	return split
}

// createInputPathEntry creates the input path entry widget
func createInputPathEntry() *widget.Entry {
	inputPathEntry := widget.NewMultiLineEntry()
	inputPathEntry.SetPlaceHolder("No folder selected or files added")
	inputPathEntry.Disable() // Make it read-only
	return inputPathEntry
}

// createInputButtons creates the buttons for selecting folders and files, and clearing the selection
func createInputButtons(inputPathEntry *widget.Entry, selectedFiles *[]string, myWindow fyne.Window) (*widget.Button, *widget.Button, *widget.Button) {
	selectFolderBtn := widget.NewButton("Select Folder", func() {
		folderPath, err := dialog.Directory().Title("Select Input Folder").Browse()
		if err != nil {
			return // User cancelled or an error occurred
		}
		inputPathEntry.SetText(folderPath)
	})

	selectFileBtn := widget.NewButton("Add Files", func() {
		for {
			file, err := dialog.File().Title("Select Files").Filter("CSV and XLSX Files", "csv", "xlsx").Load()
			if err != nil {
				break // User cancelled or an error occurred
			}
			*selectedFiles = append(*selectedFiles, file)
		}
		inputPathEntry.SetText(strings.Join(*selectedFiles, "\n"))
	})

	clearFilesBtn := widget.NewButton("Clear Files", func() {
		*selectedFiles = []string{}
		inputPathEntry.SetText("")
	})

	return selectFolderBtn, selectFileBtn, clearFilesBtn
}

// createOutputWidgets creates the output selection widgets with options
func createOutputWidgets() (*widget.RadioGroup, *fyne.Container) {
	outputOptions := []string{"Select Existing CSV File", "Specify Output Folder and Filename"}
	outputOptionRadio := widget.NewRadioGroup(outputOptions, nil)
	outputOptionRadio.SetSelected("Specify Output Folder and Filename") // Default selection

	outputOptionsContainer := container.NewVBox()
	updateOutputOptions(outputOptionRadio.Selected, outputOptionsContainer)

	outputOptionRadio.OnChanged = func(selected string) {
		updateOutputOptions(selected, outputOptionsContainer)
	}

	return outputOptionRadio, outputOptionsContainer
}

// updateOutputOptions updates the output options container based on the selected option
func updateOutputOptions(selected string, container *fyne.Container) {
	container.Objects = nil

	if selected == "Select Existing CSV File" {
		outputFileEntry := widget.NewEntry()
		outputFileEntry.SetPlaceHolder("No output file selected")
		selectOutputFileBtn := widget.NewButton("Select Output File", func() {
			filePath, err := dialog.File().Title("Select Output CSV File").Filter("CSV Files", "csv").Load()
			if err == nil {
				outputFileEntry.SetText(filePath)
			}
		})

		container.Add(outputFileEntry)
		container.Add(selectOutputFileBtn)

	} else if selected == "Specify Output Folder and Filename" {
		outputPathEntry := widget.NewEntry()
		outputPathEntry.SetPlaceHolder("No output folder selected")
		selectOutputFolderBtn := widget.NewButton("Select Output Folder", func() {
			folderPath, err := dialog.Directory().Title("Select Output Folder").Browse()
			if err == nil {
				outputPathEntry.SetText(folderPath)
			}
		})

		outputFileNameEntry := widget.NewEntry()
		outputFileNameEntry.SetPlaceHolder("Enter output file name (e.g., combined_output.csv)")

		container.Add(outputPathEntry)
		container.Add(selectOutputFolderBtn)
		container.Add(outputFileNameEntry)
	}

	container.Refresh()
}

// createStartButton creates the start button to begin processing
func createStartButton(
	inputPathEntry *widget.Entry,
	outputOptionRadio *widget.RadioGroup,
	outputOptionsContainer *fyne.Container,
	selectedFiles *[]string,
	myWindow fyne.Window,
) *widget.Button {
	return widget.NewButton("Start Processing", func() {
		go func() {
			// Validate inputs
			if err := validateCombineInputs(inputPathEntry, outputOptionRadio, outputOptionsContainer, selectedFiles, myWindow); err != nil {
				return
			}

			// Determine output file path
			outputFilePath, err := determineCombineOutputPath(outputOptionRadio, outputOptionsContainer)
			if err != nil {
				utils.ShowError(err, myWindow)
				return
			}

			// Initialize logger
			if err := utils.InitializeLogger(filepath.Join(filepath.Dir(outputFilePath), "process_log.txt")); err != nil {
				utils.ShowError(fmt.Errorf("Failed to initialize logger: %v", err), myWindow)
				return
			}

			// Start processing
			if err := combineFiles(inputPathEntry.Text, selectedFiles, outputFilePath); err != nil {
				utils.ShowError(err, myWindow)
				return
			}

			utils.ShowInfo("Processing completed successfully!", myWindow)
		}()
	})
}

// validateCombineInputs validates the inputs before processing
func validateCombineInputs(
	inputPathEntry *widget.Entry,
	outputOptionRadio *widget.RadioGroup,
	outputOptionsContainer *fyne.Container,
	selectedFiles *[]string,
	myWindow fyne.Window,
) error {
	inputPath := inputPathEntry.Text
	if inputPath == "" && len(*selectedFiles) == 0 {
		err := fmt.Errorf("Please select an input folder or add files")
		utils.ShowError(err, myWindow)
		return err
	}

	// Validate output options
	outputOption := outputOptionRadio.Selected
	if outputOption == "Select Existing CSV File" {
		outputFileEntry := outputOptionsContainer.Objects[0].(*widget.Entry)
		if outputFileEntry.Text == "" {
			err := fmt.Errorf("Please select an output CSV file")
			utils.ShowError(err, myWindow)
			return err
		}
	} else if outputOption == "Specify Output Folder and Filename" {
		outputPathEntry := outputOptionsContainer.Objects[0].(*widget.Entry)
		outputFileNameEntry := outputOptionsContainer.Objects[2].(*widget.Entry)
		if outputPathEntry.Text == "" {
			err := fmt.Errorf("Please select an output folder")
			utils.ShowError(err, myWindow)
			return err
		}
		if outputFileNameEntry.Text == "" {
			err := fmt.Errorf("Please enter an output file name")
			utils.ShowError(err, myWindow)
			return err
		}
		if !strings.HasSuffix(strings.ToLower(outputFileNameEntry.Text), ".csv") {
			err := fmt.Errorf("Output file name must have a .csv extension")
			utils.ShowError(err, myWindow)
			return err
		}
	} else {
		err := fmt.Errorf("Invalid output option selected")
		utils.ShowError(err, myWindow)
		return err
	}

	return nil
}

// determineCombineOutputPath determines the output file path based on user selection
func determineCombineOutputPath(outputOptionRadio *widget.RadioGroup, outputOptionsContainer *fyne.Container) (string, error) {
	outputOption := outputOptionRadio.Selected

	if outputOption == "Select Existing CSV File" {
		outputFileEntry := outputOptionsContainer.Objects[0].(*widget.Entry)
		return outputFileEntry.Text, nil
	} else if outputOption == "Specify Output Folder and Filename" {
		outputPathEntry := outputOptionsContainer.Objects[0].(*widget.Entry)
		outputFileNameEntry := outputOptionsContainer.Objects[2].(*widget.Entry)
		return filepath.Join(outputPathEntry.Text, outputFileNameEntry.Text), nil
	}

	return "", fmt.Errorf("Invalid output option selected")
}

// combineFiles processes and combines input files into a single output file
func combineFiles(inputPath string, selectedFiles *[]string, outputFilePath string) error {
	recordsMap := make(map[string]records.Record)
	var files []string

	if len(*selectedFiles) > 0 {
		files = *selectedFiles
	} else {
		// Process the inputPath
		fileInfo, err := os.Stat(inputPath)
		if err != nil {
			return fmt.Errorf("Error accessing path: %v", err)
		}

		if fileInfo.IsDir() {
			// Collect all CSV and XLSX files from the directory
			err := filepath.Walk(inputPath, func(path string, info os.FileInfo, err error) error {
				if err == nil && !info.IsDir() {
					ext := strings.ToLower(filepath.Ext(path))
					if ext == ".csv" || ext == ".xlsx" {
						files = append(files, path)
					}
				}
				return nil
			})
			if err != nil {
				utils.LogMessage(fmt.Sprintf("Error walking the directory: %v", err))
			}
		} else {
			// Single file selected
			ext := strings.ToLower(filepath.Ext(inputPath))
			if ext == ".csv" || ext == ".xlsx" {
				files = append(files, inputPath)
			} else {
				return fmt.Errorf("Unsupported file type selected")
			}
		}
	}

	// Update UI with file count
	fileCount := len(files)
	utils.LogMessage(fmt.Sprintf("Total files to process: %d", fileCount))

	if fileCount == 0 {
		return fmt.Errorf("No CSV or XLSX files found in the selected input")
	}

	// Process files
	for _, filePath := range files {
		utils.LogMessage(fmt.Sprintf("Processing file: %s", filePath))
		recordsList, _, err := records.LoadRecords(filePath)
		if err != nil {
			utils.LogMessage(fmt.Sprintf("Error loading records from file %s: %v", filePath, err))
			continue
		}
		for _, record := range recordsList {
			if _, exists := recordsMap[record.Email]; !exists {
				recordsMap[record.Email] = record
			}
		}
	}

	// Write to output file
	headers := []string{"Name", "Email", "OrgName"} // Update as per actual headers
	err := records.WriteFilteredCSV(outputFilePath, headers, mapToSlice(recordsMap))
	if err != nil {
		utils.LogMessage(fmt.Sprintf("Error writing to CSV: %v", err))
		return fmt.Errorf("Error writing to CSV: %v", err)
	}

	utils.LogMessage(fmt.Sprintf("Processing completed, duplicates removed! Output file saved to %s", outputFilePath))
	return nil
}

// mapToSlice converts a map of records to a slice
func mapToSlice(recordsMap map[string]records.Record) []records.Record {
	recordsList := make([]records.Record, 0, len(recordsMap))
	for _, record := range recordsMap {
		recordsList = append(recordsList, record)
	}
	return recordsList
}

// createLogViewer creates the log viewer for displaying log messages
func createLogViewer() *widget.Entry {
	logContent := widget.NewMultiLineEntry()
	logContent.Wrapping = fyne.TextWrapWord
	logContent.Disable()

	// Periodically update the log viewer
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
