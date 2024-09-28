package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/tealeg/xlsx"
)

type Record struct {
	Name    string
	OrgName string
	Email   string
	Others  []string
}

var logger *log.Logger
var logMessages []string
var logMutex sync.Mutex

func main() {
	// Create the GUI application
	myApp := app.New()
	myWindow := myApp.NewWindow("CSV/XLSX Combiner")

	// Variable to store selected files
	var selectedFiles []string

	// Input Elements
	inputPathEntry := widget.NewMultiLineEntry()
	inputPathEntry.SetPlaceHolder("No folder selected or files added")
	inputPathEntry.Disable() // Make it read-only

	selectFolderBtn := widget.NewButton("Select Folder", func() {
		dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
			if err != nil {
				showError(err, myWindow)
				return
			}
			if uri != nil {
				inputPathEntry.SetText(uri.Path())
			}
		}, myWindow)
	})

	selectFileBtn := widget.NewButton("Add File", func() {
		dialog.ShowFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err != nil {
				showError(err, myWindow)
				return
			}
			if reader != nil {
				selectedFiles = append(selectedFiles, reader.URI().Path())
				inputPathEntry.SetText(strings.Join(selectedFiles, "\n"))
				reader.Close()
			}
		}, myWindow)
	})

	clearFilesBtn := widget.NewButton("Clear Files", func() {
		selectedFiles = []string{}
		inputPathEntry.SetText("")
	})

	// Output Elements
	outputPathEntry := widget.NewEntry()
	outputPathEntry.SetPlaceHolder("No output folder selected")
	outputPathEntry.Disable() // Make it read-only

	selectOutputFolderBtn := widget.NewButton("Select Output Folder", func() {
		dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
			if err != nil {
				showError(err, myWindow)
				return
			}
			if uri != nil {
				outputPathEntry.SetText(uri.Path())
			}
		}, myWindow)
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
				showError(fmt.Errorf("Please select an input folder or add files"), myWindow)
				return
			}
			if outputPath == "" {
				showError(fmt.Errorf("Please select an output folder"), myWindow)
				return
			}
			if outputFileName == "" {
				showError(fmt.Errorf("Please enter an output file name"), myWindow)
				return
			}

			// Ensure the output file has a .csv extension
			if !strings.HasSuffix(strings.ToLower(outputFileName), ".csv") {
				showError(fmt.Errorf("Output file name must have a .csv extension"), myWindow)
				return
			}

			// Open the log file for writing
			logFilePath := filepath.Join(outputPath, "process_log.txt")
			logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
			if err != nil {
				showError(fmt.Errorf("Failed to open log file: %v", err), myWindow)
				return
			}
			defer logFile.Close()

			// Set up the logger to write to the log file
			logger = log.New(logFile, "", log.Ldate|log.Ltime)

			records := make(map[string]Record)

			var wg sync.WaitGroup
			recordChan := make(chan Record)

			// Start a goroutine to collect all records into the map
			go func() {
				for record := range recordChan {
					if _, exists := records[record.Email]; !exists {
						records[record.Email] = record
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
					showError(fmt.Errorf("Error accessing path: %v", err), myWindow)
					return
				}

				if fileInfo.IsDir() {
					// Walk through the folder and process each file
					err := filepath.Walk(inputPath, func(path string, info os.FileInfo, err error) error {
						if err != nil {
							logMessage(fmt.Sprintf("Error accessing file: %s - %v", path, err))
							return nil // Continue to the next file
						}
						// Check if the file is either a CSV or XLSX file
						if !info.IsDir() {
							ext := strings.ToLower(filepath.Ext(path))
							if ext == ".csv" || ext == ".xlsx" {
								files = append(files, path)
							} else {
								// Log and skip non-CSV and non-XLSX files
								logMessage(fmt.Sprintf("Skipping unsupported file type: %s", path))
							}
						}
						return nil
					})
					if err != nil {
						logMessage(fmt.Sprintf("Error walking the directory: %v", err))
					}
				} else {
					// Single file selected
					ext := strings.ToLower(filepath.Ext(inputPath))
					if ext == ".csv" || ext == ".xlsx" {
						files = append(files, inputPath)
					} else {
						showError(fmt.Errorf("Unsupported file type selected"), myWindow)
						return
					}
				}
			}

			// Update UI with file count
			fileCount := len(files)
			logMessage(fmt.Sprintf("Total files to process: %d", fileCount))

			if fileCount == 0 {
				showError(fmt.Errorf("No CSV or XLSX files found in the selected input"), myWindow)
				return
			}

			// Process files concurrently
			for _, filePath := range files {
				wg.Add(1)
				go func(filePath string) {
					defer wg.Done()
					ext := strings.ToLower(filepath.Ext(filePath))
					logMessage(fmt.Sprintf("Processing file: %s", filePath))
					if ext == ".csv" {
						loadCSV(filePath, recordChan)
					} else if ext == ".xlsx" {
						loadXLSX(filePath, recordChan)
					}
				}(filePath)
			}

			// Wait for all file processing to complete
			wg.Wait()
			close(recordChan) // Close the channel when all records are processed

			// Write unique records to the output file
			outputFilePath := filepath.Join(outputPath, outputFileName)
			err = writeCSV(outputFilePath, records)
			if err != nil {
				logMessage(fmt.Sprintf("Error writing to CSV: %v", err))
			} else {
				logMessage(fmt.Sprintf("Processing completed, duplicates removed! Output file saved to %s", outputFilePath))
				showInfo("Processing completed successfully!", myWindow)
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
			logMutex.Lock()
			logText := strings.Join(logMessages, "\n")
			logContent.SetText(logText)
			logMutex.Unlock()
			time.Sleep(500 * time.Millisecond)
		}
	}()

	// Adjusted Layout: Use a VSplit to control how much space is allocated to the logs and controls
	split := container.NewVSplit(
		container.NewVBox(
			widget.NewLabelWithStyle("Input Selection", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			inputPathEntry,
			container.NewHBox(selectFolderBtn, selectFileBtn, clearFilesBtn),
			widget.NewLabelWithStyle("Output Selection", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			outputPathEntry,
			selectOutputFolderBtn,
			outputFileNameEntry,
			startBtn,
		),
		container.NewVScroll(logContent),
	)

	// Adjust split ratio to ensure the log section gets sufficient space
	split.Offset = 0.75 // 75% for controls, 25% for logs (you can adjust this as needed)

	myWindow.SetContent(split)
	myWindow.Resize(fyne.NewSize(600, 700))
	myWindow.ShowAndRun()
}

func showError(err error, win fyne.Window) {
	dialog.ShowError(err, win)
}

func showInfo(message string, win fyne.Window) {
	dialog.ShowInformation("Info", message, win)
}

func logMessage(message string) {
	logMutex.Lock()
	defer logMutex.Unlock()
	logger.Println(message)
	logMessages = append(logMessages, message)
	fmt.Println(message)
}

func loadCSV(filename string, recordChan chan<- Record) {
	file, err := os.Open(filename)
	if err != nil {
		logMessage(fmt.Sprintf("Error opening CSV file: %s - %v", filename, err))
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.LazyQuotes = true    // Allows for malformed CSV fields like bare quotes
	reader.FieldsPerRecord = -1 // Allow variable number of fields per row
	rows, err := reader.ReadAll()
	if err != nil {
		logMessage(fmt.Sprintf("Error reading CSV file: %s - %v", filename, err))
		return
	}

	if len(rows) == 0 {
		logMessage(fmt.Sprintf("No data found in CSV file: %s", filename))
		return
	}

	headers := sanitizeHeaders(rows[0])

	// Find the required column indexes dynamically, using flexible matching
	emailIndex := findFlexibleHeaderIndex(headers, "email")
	nameIndex := findFlexibleHeaderIndex(headers, "name")
	orgNameIndex := findFlexibleHeaderIndex(headers, "organization") // Optional

	// Skip files if required columns are not found
	if emailIndex == -1 || nameIndex == -1 {
		logMessage(fmt.Sprintf("Required columns (Name, Email) not found in CSV file: %s, skipping...", filename))
		return
	}

	// Process rows, start from 1 to skip header
	for _, row := range rows[1:] {
		if len(row) <= emailIndex || len(row) <= nameIndex {
			// Skip rows that don't have enough columns
			continue
		}
		email := row[emailIndex]
		name := row[nameIndex]
		orgName := ""
		if orgNameIndex != -1 && len(row) > orgNameIndex {
			orgName = row[orgNameIndex]
		}

		// Send the record to the channel
		recordChan <- Record{
			Name:    name,
			OrgName: orgName,
			Email:   email,
			Others:  excludeColumns(row, []int{nameIndex, orgNameIndex, emailIndex}),
		}
	}
}

func loadXLSX(filename string, recordChan chan<- Record) {
	file, err := xlsx.OpenFile(filename)
	if err != nil {
		logMessage(fmt.Sprintf("Error opening XLSX file: %s - %v", filename, err))
		return
	}

	for _, sheet := range file.Sheets {
		if len(sheet.Rows) == 0 {
			logMessage(fmt.Sprintf("No data found in XLSX file: %s", filename))
			continue
		}
		headers := sanitizeHeadersXLSX(sheet.Rows[0].Cells)

		// Find the required column indexes dynamically using flexible matching
		emailIndex := findFlexibleHeaderIndex(headers, "email")
		nameIndex := findFlexibleHeaderIndex(headers, "name")
		orgNameIndex := findFlexibleHeaderIndex(headers, "organization") // Optional

		// Skip files if required columns are not found
		if emailIndex == -1 || nameIndex == -1 {
			logMessage(fmt.Sprintf("Required columns (Name, Email) not found in XLSX file: %s, skipping...", filename))
			continue
		}

		// Process rows, start from 1 to skip header
		for _, row := range sheet.Rows[1:] {
			if len(row.Cells) <= emailIndex || len(row.Cells) <= nameIndex {
				// Skip rows that don't have enough columns
				continue
			}
			email := row.Cells[emailIndex].String()
			name := row.Cells[nameIndex].String()
			orgName := ""
			if orgNameIndex != -1 && len(row.Cells) > orgNameIndex {
				orgName = row.Cells[orgNameIndex].String()
			}

			// Send the record to the channel
			recordChan <- Record{
				Name:    name,
				OrgName: orgName,
				Email:   email,
				Others:  getRowDataExcluding(row, []int{nameIndex, orgNameIndex, emailIndex}),
			}
		}
	}
}

// Helper functions for header management and data extraction
func findFlexibleHeaderIndex(headers []string, keyword string) int {
	for i, h := range headers {
		h = strings.TrimSpace(h)
		h = strings.Trim(h, `"`) // Remove surrounding quotes
		if strings.Contains(strings.ToLower(h), strings.ToLower(keyword)) {
			return i
		}
	}
	return -1
}

func sanitizeHeaders(headers []string) []string {
	for i := range headers {
		headers[i] = strings.Trim(headers[i], `"`) // Remove surrounding quotes
		headers[i] = strings.TrimSpace(headers[i]) // Remove any extra spaces
	}
	return headers
}

func sanitizeHeadersXLSX(headers []*xlsx.Cell) []string {
	var sanitizedHeaders []string
	for _, cell := range headers {
		header := strings.Trim(cell.String(), `"`) // Remove surrounding quotes
		header = strings.TrimSpace(header)         // Remove any extra spaces
		sanitizedHeaders = append(sanitizedHeaders, header)
	}
	return sanitizedHeaders
}

func excludeColumns(row []string, excludeIndexes []int) []string {
	var others []string
	for i, col := range row {
		if !contains(excludeIndexes, i) {
			others = append(others, col)
		}
	}
	return others
}

func getRowDataExcluding(row *xlsx.Row, excludeIndexes []int) []string {
	var others []string
	for i, cell := range row.Cells {
		if !contains(excludeIndexes, i) {
			others = append(others, cell.String())
		}
	}
	return others
}

func contains(slice []int, elem int) bool {
	for _, e := range slice {
		if e == elem {
			return true
		}
	}
	return false
}

func writeCSV(filename string, records map[string]Record) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	err = writer.Write([]string{"Name", "OrgName", "Email", "Others"})
	if err != nil {
		return err
	}

	// Write records
	for _, record := range records {
		row := append([]string{record.Name, record.OrgName, record.Email}, record.Others...)
		err = writer.Write(row)
		if err != nil {
			return err
		}
	}

	return nil
}
