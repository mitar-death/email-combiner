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

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func CreateCombineScreen(myWindow fyne.Window) fyne.CanvasObject {
	// Variable to store selected files
	var selectedFiles []string

	// Input Elements
	inputPathEntry := widget.NewMultiLineEntry()
	inputPathEntry.SetPlaceHolder("No folder selected or files added")
	inputPathEntry.Disable() // Make it read-only

	selectFolderBtn := widget.NewButton("Select Folder", func() {
		utils.ShowFolderSelectionDialog(inputPathEntry, myWindow)
	})
	selectFileBtn := widget.NewButton("Add File", func() {
		utils.ShowFileSelectionDialog(&selectedFiles, inputPathEntry, myWindow)
	})
	clearFilesBtn := widget.NewButton("Clear Files", func() {
		selectedFiles = []string{}
		inputPathEntry.SetText("")
	})

	// Output Elements
	outputPathEntry := widget.NewEntry()
	outputPathEntry.SetPlaceHolder("No output folder selected")
	// outputPathEntry.Disable() // Make it read-only

	selectOutputFolderBtn := widget.NewButton("Select Output Folder", func() {
		utils.ShowFolderSelectionDialog(outputPathEntry, myWindow)
	})

	outputFileNameEntry := widget.NewEntry()
	outputFileNameEntry.SetPlaceHolder("Enter output file name (e.g., combined_output.csv)")

	// Start Button
	startBtn := widget.NewButton("Start Processing", func() {
		go func() {
			inputPath := inputPathEntry.Text
			outputPath := outputPathEntry.Text
			outputFileName := outputFileNameEntry.Text

			if inputPath == "" && len(selectedFiles) == 0 {
				utils.ShowError(fmt.Errorf("Please select an input folder or add files"), myWindow)
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

			// Open the log file for writing
			logFilePath := filepath.Join(outputPath, "process_log.txt")
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

			if len(selectedFiles) > 0 {
				files = selectedFiles
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

			// Write unique records to the output file
			outputFilePath := filepath.Join(outputPath, outputFileName)
			err = records.WriteCSV(outputFilePath, recordsMap)
			if err != nil {
				utils.LogMessage(fmt.Sprintf("Error writing to CSV: %v", err))
			} else {
				utils.LogMessage(fmt.Sprintf("Processing completed, duplicates removed! Output file saved to %s", outputFilePath))
				utils.ShowInfo("Processing completed successfully!", myWindow)
			}
		}()
	})

	// Log viewer
	logContent := widget.NewMultiLineEntry()
	logContent.Wrapping = fyne.TextWrapWord
	// logContent.Disable() // Make it read-only

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

	// Notes to inform users about dialog refresh
	inputNote := widget.NewLabel("Note: If you add a new folder, please close and reopen the dialog to see it.")
	outputNote := widget.NewLabel("Note: If you create a new folder, close and reopen the dialog to refresh.")

	// Adjusted Layout: Use a VSplit to control how much space is allocated to the logs and controls
	split := container.NewVSplit(
		container.NewVBox(
			widget.NewLabelWithStyle("Input Selection", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			inputPathEntry,
			container.NewHBox(selectFolderBtn, selectFileBtn, clearFilesBtn),
			inputNote,
			widget.NewLabelWithStyle("Output Selection", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			outputPathEntry,
			selectOutputFolderBtn,
			outputNote,
			outputFileNameEntry,
			startBtn,
		),
		container.NewVScroll(logContent),
	)

	// Adjust split ratio to ensure the log section gets sufficient space
	split.Offset = 0.75 // 75% for controls, 25% for logs (you can adjust this as needed)

	return split
}
