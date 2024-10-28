package combine

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"website-copier/cmd/records"
	"website-copier/cmd/utils"

	"github.com/sqweek/dialog"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func CreateCombineScreen(myWindow fyne.Window) fyne.CanvasObject {
	// Variable to store selected files
	var selectedFiles []string

	// Create Input Selection Widgets
	inputPathEntry := createInputPathEntry()
	selectFolderBtn, selectFileBtn, clearFilesBtn := createInputButtons(inputPathEntry, &selectedFiles)

	// Create Output Selection Widgets
	outputPathEntry, _, outputFileNameEntry, outputFileEntry, outputOptionRadio, outputOptionsContainer := createOutputWidgets()

	// Create Start Button
	startBtn := createStartButton(
		inputPathEntry,
		outputPathEntry,
		outputFileNameEntry,
		outputFileEntry,
		outputOptionRadio,
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
func createInputButtons(inputPathEntry *widget.Entry, selectedFiles *[]string) (*widget.Button, *widget.Button, *widget.Button) {
	selectFolderBtn := widget.NewButton("Select Folder", func() {
		folderPath, err := dialog.Directory().Title("Select Input Folder").Browse()
		if err != nil {
			return // User cancelled or an error occurred
		}
		inputPathEntry.SetText(folderPath)
	})

	selectFileBtn := widget.NewButton("Add File", func() {
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

// createOutputWidgets creates the output selection widgets with a toggle between existing CSV file and folder path with filename
func createOutputWidgets() (*widget.Entry, *widget.Button, *widget.Entry, *widget.Entry, *widget.RadioGroup, *fyne.Container) {
	// Output Option RadioGroup
	outputOptions := []string{"Select Existing CSV File", "Specify Output Folder and Filename"}
	outputOptionRadio := widget.NewRadioGroup(outputOptions, nil)
	outputOptionRadio.SetSelected("Specify Output Folder and Filename") // Default selection

	// Widgets for "Select Existing CSV File" option
	outputFileEntry := widget.NewEntry()
	outputFileEntry.SetPlaceHolder("No output file selected")
	selectOutputFileBtn := widget.NewButton("Select Output File", func() {
		filePath, err := dialog.File().Title("Select Output CSV File").Filter("CSV Files", "csv").Load()
		if err != nil {
			return // User cancelled or an error occurred
		}
		outputFileEntry.SetText(filePath)
	})

	// Widgets for "Specify Output Folder and Filename" option
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
	outputFileNameEntry.SetPlaceHolder("Enter output file name (e.g., combined_output.csv)")

	// Container to hold the widgets that will change based on selection
	outputOptionsContainer := container.NewVBox()

	// Function to update the output options container
	updateOutputOptions := func(selected string) {
		outputOptionsContainer.Objects = nil
		if selected == "Select Existing CSV File" {
			outputOptionsContainer.Add(outputFileEntry)
			outputOptionsContainer.Add(selectOutputFileBtn)
		} else if selected == "Specify Output Folder and Filename" {
			outputOptionsContainer.Add(outputPathEntry)
			outputOptionsContainer.Add(selectOutputFolderBtn)
			outputOptionsContainer.Add(outputFileNameEntry)
		}
		outputOptionsContainer.Refresh()
	}

	// Set the initial state
	updateOutputOptions(outputOptionRadio.Selected)

	// Set the handler for when the selection changes
	outputOptionRadio.OnChanged = updateOutputOptions

	return outputPathEntry, selectOutputFolderBtn, outputFileNameEntry, outputFileEntry, outputOptionRadio, outputOptionsContainer
}

// createStartButton creates the start button to begin processing
func createStartButton(
	inputPathEntry *widget.Entry,
	outputPathEntry *widget.Entry,
	outputFileNameEntry *widget.Entry,
	outputFileEntry *widget.Entry,
	outputOptionRadio *widget.RadioGroup,
	selectedFiles *[]string,
	myWindow fyne.Window,
) *widget.Button {
	return widget.NewButton("Start Processing", func() {
		go func() {
			inputPath := inputPathEntry.Text
			outputPath := outputPathEntry.Text
			outputFileName := outputFileNameEntry.Text
			outputFile := outputFileEntry.Text
			outputOption := outputOptionRadio.Selected

			if inputPath == "" && len(*selectedFiles) == 0 {
				utils.ShowError(fmt.Errorf("Please select an input folder or add files"), myWindow)
				return
			}

			// Output validation
			var outputFilePath string
			if outputOption == "Select Existing CSV File" {
				if outputFile == "" {
					utils.ShowError(fmt.Errorf("Please select an output CSV file"), myWindow)
					return
				}
				outputFilePath = outputFile
			} else if outputOption == "Specify Output Folder and Filename" {
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
				outputFilePath = filepath.Join(outputPath, outputFileName)
			} else {
				utils.ShowError(fmt.Errorf("Invalid output option selected"), myWindow)
				return
			}

			// Check if the output file exists
			var existingHeaders []string
			if _, err := os.Stat(outputFilePath); err == nil {
				// File exists, load headers
				existingHeaders, err = records.GetCSVHeaders(outputFilePath)
				if err != nil {
					utils.ShowError(fmt.Errorf("Error reading existing file headers: %v", err), myWindow)
					return
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

			recordsMap := make(map[string]records.Record)

			var wg sync.WaitGroup
			recordChan := make(chan records.Record)

			// Start a goroutine to collect all records into the map
			go func() {
				for record := range recordChan {
					if _, exists := recordsMap[record.Email]; !exists {
						recordsMap[record.Email] = record
					}
				}
			}()

			var files []string

			if len(*selectedFiles) > 0 {
				files = *selectedFiles
			} else {
				// Process the inputPath
				fileInfo, err := os.Stat(inputPath)
				if err != nil {
					utils.ShowError(fmt.Errorf("Error accessing path: %v", err), myWindow)
					return
				}

				if fileInfo.IsDir() {
					// Walk through the folder and process each file
					err := filepath.Walk(inputPath, func(path string, info os.FileInfo, err error) error {
						if err != nil {
							utils.LogMessage(fmt.Sprintf("Error accessing file: %s - %v", path, err))
							return nil // Continue to the next file
						}
						// Check if the file is either a CSV or XLSX file
						if !info.IsDir() {
							ext := strings.ToLower(filepath.Ext(path))
							if ext == ".csv" || ext == ".xlsx" {
								files = append(files, path)
							} else {
								// Log and skip non-CSV and non-XLSX files
								utils.LogMessage(fmt.Sprintf("Skipping unsupported file type: %s", path))
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
						utils.ShowError(fmt.Errorf("Unsupported file type selected"), myWindow)
						return
					}
				}
			}

			// Update UI with file count
			fileCount := len(files)
			utils.LogMessage(fmt.Sprintf("Total files to process: %d", fileCount))

			if fileCount == 0 {
				utils.ShowError(fmt.Errorf("No CSV or XLSX files found in the selected input"), myWindow)
				return
			}

			// Process files concurrently
			for _, filePath := range files {
				wg.Add(1)
				go func(filePath string) {
					defer wg.Done()
					ext := strings.ToLower(filepath.Ext(filePath))
					utils.LogMessage(fmt.Sprintf("Processing file: %s", filePath))
					if ext == ".csv" {
						records.LoadCSV(filePath, recordChan)
					} else if ext == ".xlsx" {
						records.LoadXLSX(filePath, recordChan)
					}
				}(filePath)
			}

			// Wait for all file processing to complete
			wg.Wait()
			close(recordChan) // Close the channel when all records are processed

			// Check if the headers match if the file already exists
			if len(existingHeaders) > 0 {
				if !records.ValidateHeaders(existingHeaders) {
					utils.LogMessage("Existing file headers do not match requirements, creating a new file.")
					// Create a new file if headers do not match
					err := records.WriteCSV(outputFilePath, recordsMap)
					if err != nil {
						utils.LogMessage(fmt.Sprintf("Error writing to CSV: %v", err))
					} else {
						utils.LogMessage(fmt.Sprintf("Processing completed, duplicates removed! Output file saved to %s", outputFilePath))
						utils.ShowInfo("Processing completed successfully!", myWindow)
					}
				} else {
					// Append to the existing file if headers match
					err := records.AppendCSV(outputFilePath, recordsMap)
					if err != nil {
						utils.LogMessage(fmt.Sprintf("Error appending to CSV: %v", err))
					} else {
						utils.LogMessage(fmt.Sprintf("Records appended to existing file: %s", outputFilePath))
						utils.ShowInfo("Records appended successfully!", myWindow)
					}
				}
			} else {
				// Create a new file if it does not exist
				err := records.WriteCSV(outputFilePath, recordsMap)
				if err != nil {
					utils.LogMessage(fmt.Sprintf("Error writing to CSV: %v", err))
				} else {
					utils.LogMessage(fmt.Sprintf("Processing completed, duplicates removed! Output file saved to %s", outputFilePath))
					utils.ShowInfo("Processing completed successfully!", myWindow)
				}
			}
		}()
	})
}

// createLogViewer creates the log viewer for displaying log messages
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
