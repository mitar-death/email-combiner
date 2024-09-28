package main

import (
	"encoding/csv"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/tealeg/xlsx"
)

type Record struct {
	Name    string
	OrgName string
	Email   string
	Others  []string
}

var logger *log.Logger

func main() {
	// Open the log file for writing
	logFile, err := os.OpenFile("process_log.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	defer logFile.Close()

	// Set up the logger to write to the log file
	logger = log.New(logFile, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)

	folder := "./your-folder" // Replace with the folder path containing your files
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

	// Walk through the folder and process each file concurrently
	err = filepath.Walk(folder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			logger.Printf("Error accessing file: %s - %v\n", path, err)
			return nil // Continue to the next file
		}
		// Check if the file is either a CSV or XLSX file
		if !info.IsDir() {
			ext := strings.ToLower(filepath.Ext(path))
			if ext == ".csv" || ext == ".xlsx" {
				wg.Add(1)
				go func(filePath string) {
					defer wg.Done()
					logger.Printf("Processing file: %s\n", filePath)
					if ext == ".csv" {
						loadCSV(filePath, recordChan)
					} else if ext == ".xlsx" {
						loadXLSX(filePath, recordChan)
					}
				}(path)
			} else {
				// Log and skip non-CSV and non-XLSX files
				logger.Printf("Skipping unsupported file type: %s\n", path)
			}
		}
		return nil
	})

	if err != nil {
		logger.Printf("Error walking the directory: %v\n", err)
	}

	// Wait for all file processing to complete
	wg.Wait()
	close(recordChan) // Close the channel when all records are processed

	// Write unique records to a new file
	err = writeCSV("combined_output.csv", records)
	if err != nil {
		logger.Printf("Error writing to CSV: %v\n", err)
	}

	logger.Println("Processing completed, duplicates removed!")
}

func loadCSV(filename string, recordChan chan<- Record) {
	file, err := os.Open(filename)
	if err != nil {
		logger.Printf("Error opening CSV file: %s - %v\n", filename, err)
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.LazyQuotes = true    // Allows for malformed CSV fields like bare quotes
	reader.FieldsPerRecord = -1 // Allow variable number of fields per row
	rows, err := reader.ReadAll()
	if err != nil {
		logger.Printf("Error reading CSV file: %s - %v\n", filename, err)
		return
	}

	if len(rows) == 0 {
		logger.Printf("No data found in CSV file: %s\n", filename)
		return
	}

	headers := sanitizeHeaders(rows[0])

	// Find the required column indexes dynamically, using flexible matching
	emailIndex := findFlexibleHeaderIndex(headers, "email")
	nameIndex := findFlexibleHeaderIndex(headers, "name")
	orgNameIndex := findFlexibleHeaderIndex(headers, "organization") // Optional

	// Skip files if required columns are not found
	if emailIndex == -1 || nameIndex == -1 {
		logger.Printf("Required columns (Name, Email) not found in CSV file: %s, skipping...\n", filename)
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
		logger.Printf("Error opening XLSX file: %s - %v\n", filename, err)
		return
	}

	for _, sheet := range file.Sheets {
		if len(sheet.Rows) == 0 {
			logger.Printf("No data found in XLSX file: %s\n", filename)
			return
		}
		headers := sanitizeHeadersXLSX(sheet.Rows[0].Cells)

		// Find the required column indexes dynamically using flexible matching
		emailIndex := findXLSXFlexibleHeaderIndex(headers, "email")
		nameIndex := findXLSXFlexibleHeaderIndex(headers, "name")
		orgNameIndex := findXLSXFlexibleHeaderIndex(headers, "organization") // Optional

		// Skip files if required columns are not found
		if emailIndex == -1 || nameIndex == -1 {
			logger.Printf("Required columns (Name, Email) not found in XLSX file: %s, skipping...\n", filename)
			return
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

// Helper function to find a flexible match for the column header (case insensitive, partial match)
func findFlexibleHeaderIndex(headers []string, keyword string) int {
	for i, h := range headers {
		h = strings.TrimSpace(h)
		if strings.Contains(strings.ToLower(h), strings.ToLower(keyword)) {
			return i
		}
	}
	return -1
}

// Helper function to find a flexible match for the column header in XLSX (case insensitive, partial match)
func findXLSXFlexibleHeaderIndex(headers []string, keyword string) int {
	for i, h := range headers {
		h = strings.TrimSpace(h)
		if strings.Contains(strings.ToLower(h), strings.ToLower(keyword)) {
			return i
		}
	}
	return -1
}

// Sanitize headers by removing any quotes around the header names for CSV
func sanitizeHeaders(headers []string) []string {
	for i := range headers {
		headers[i] = strings.Trim(headers[i], `"`) // Remove surrounding quotes
		headers[i] = strings.TrimSpace(headers[i]) // Remove any extra spaces
	}
	return headers
}

// Sanitize headers by removing any quotes around the header names for XLSX
func sanitizeHeadersXLSX(headers []*xlsx.Cell) []string {
	var sanitizedHeaders []string
	for _, cell := range headers {
		header := strings.Trim(cell.String(), `"`) // Remove surrounding quotes
		header = strings.TrimSpace(header)         // Remove any extra spaces
		sanitizedHeaders = append(sanitizedHeaders, header)
	}
	return sanitizedHeaders
}

// Exclude certain columns from the row (for CSV)
func excludeColumns(row []string, excludeIndexes []int) []string {
	var others []string
	for i, col := range row {
		if !contains(excludeIndexes, i) {
			others = append(others, col)
		}
	}
	return others
}

// Exclude certain columns from the row (for XLSX)
func getRowDataExcluding(row *xlsx.Row, excludeIndexes []int) []string {
	var others []string
	for i, cell := range row.Cells {
		if !contains(excludeIndexes, i) {
			others = append(others, cell.String())
		}
	}
	return others
}

// Helper function to check if a slice contains an element
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
