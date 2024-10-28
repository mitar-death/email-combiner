package records

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"website-copier/cmd/utils"

	"github.com/tealeg/xlsx"
)

type Record struct {
	Name      string
	OrgName   string
	Email     string
	Others    []string
	OthersMap map[string]string // Map headers to values
	FilePath  string
}

// Load records from CSV or XLSX file
func LoadRecords(filename string) ([]Record, []string, error) {

	ext := strings.ToLower(filepath.Ext(filename))
	if ext == ".csv" {
		return loadRecordsFromCSV(filename)
	} else if ext == ".xlsx" {
		return loadRecordsFromXLSX(filename)
	}
	return nil, nil, fmt.Errorf("unsupported file type: %s", ext)
}

func loadRecordsFromCSV(filename string) ([]Record, []string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, nil, err
	}
	defer file.Close()

	var records []Record
	reader := csv.NewReader(file)
	reader.LazyQuotes = true
	reader.FieldsPerRecord = -1
	rows, err := reader.ReadAll()
	if err != nil {
		return nil, nil, err
	}

	if len(rows) == 0 {
		return records, nil, nil
	}

	headers := sanitizeHeaders(rows[0])

	// Find the required column indexes dynamically, using flexible matching
	emailIndex := findFlexibleHeaderIndex(headers, "email")
	nameIndex := findFlexibleHeaderIndex(headers, "name")
	orgNameIndex := findFlexibleHeaderIndex(headers, "organization") // Optional

	// Skip files if required columns are not found
	if emailIndex == -1 || nameIndex == -1 {
		return nil, nil, fmt.Errorf("required columns (Name, Email) not found in CSV file")
	}

	// Process rows, start from 1 to skip header
	for _, row := range rows[1:] {
		if len(row) <= emailIndex || len(row) <= nameIndex {
			// Skip rows that don't have enough columns
			continue
		}

		// Map headers to values
		rowMap := make(map[string]string)
		for i, header := range headers {
			if i < len(row) {
				rowMap[header] = row[i]
			} else {
				rowMap[header] = ""
			}
		}

		email := rowMap[headers[emailIndex]]
		name := rowMap[headers[nameIndex]]
		orgName := ""
		if orgNameIndex != -1 {
			orgName = rowMap[headers[orgNameIndex]]
		}

		// Remove standard fields from OthersMap
		delete(rowMap, headers[nameIndex])
		delete(rowMap, headers[emailIndex])
		if orgNameIndex != -1 {
			delete(rowMap, headers[orgNameIndex])
		}

		// Collect the record
		records = append(records, Record{
			Name:      name,
			OrgName:   orgName,
			Email:     email,
			OthersMap: rowMap,
			FilePath:  filename,
		})
	}

	return records, headers, nil
}

func loadRecordsFromXLSX(filename string) ([]Record, []string, error) {
	file, err := xlsx.OpenFile(filename)
	if err != nil {
		return nil, nil, err
	}

	var records []Record
	var headers []string

	for _, sheet := range file.Sheets {
		if len(sheet.Rows) == 0 {
			continue
		}
		headers = sanitizeHeadersXLSX(sheet.Rows[0].Cells)

		// Find the required column indexes dynamically using flexible matching
		emailIndex := findFlexibleHeaderIndex(headers, "email")
		nameIndex := findFlexibleHeaderIndex(headers, "name")
		orgNameIndex := findFlexibleHeaderIndex(headers, "organization") // Optional

		// Skip files if required columns are not found
		if emailIndex == -1 || nameIndex == -1 {
			return nil, nil, fmt.Errorf("required columns (Name, Email) not found in XLSX file")
		}

		// Process rows, start from 1 to skip header
		for _, row := range sheet.Rows[1:] {
			if len(row.Cells) <= emailIndex || len(row.Cells) <= nameIndex {
				// Skip rows that don't have enough columns
				continue
			}

			// Map headers to values
			rowMap := make(map[string]string)
			for i, header := range headers {
				if i < len(row.Cells) {
					rowMap[header] = row.Cells[i].String()
				} else {
					rowMap[header] = ""
				}
			}

			email := rowMap[headers[emailIndex]]
			name := rowMap[headers[nameIndex]]
			orgName := ""
			if orgNameIndex != -1 {
				orgName = rowMap[headers[orgNameIndex]]
			}

			// Remove standard fields from OthersMap
			delete(rowMap, headers[nameIndex])
			delete(rowMap, headers[emailIndex])
			if orgNameIndex != -1 {
				delete(rowMap, headers[orgNameIndex])
			}

			// Collect the record
			records = append(records, Record{
				Name:      name,
				OrgName:   orgName,
				Email:     email,
				OthersMap: rowMap,
				FilePath:  filename,
			})
		}

	}

	return records, headers, nil
}

func LoadCSV(filename string, recordChan chan<- Record) {
	file, err := os.Open(filename)
	if err != nil {
		utils.LogMessage(fmt.Sprintf("Error opening CSV file: %s - %v", filename, err))
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.LazyQuotes = true    // Allows for malformed CSV fields like bare quotes
	reader.FieldsPerRecord = -1 // Allow variable number of fields per row
	rows, err := reader.ReadAll()
	if err != nil {
		utils.LogMessage(fmt.Sprintf("Error reading CSV file: %s - %v", filename, err))
		return
	}

	if len(rows) == 0 {
		utils.LogMessage(fmt.Sprintf("No data found in CSV file: %s", filename))
		return
	}

	headers := sanitizeHeaders(rows[0])

	// Find the required column indexes dynamically, using flexible matching
	emailIndex := findFlexibleHeaderIndex(headers, "email")
	nameIndex := findFlexibleHeaderIndex(headers, "name")
	orgNameIndex := findFlexibleHeaderIndex(headers, "organization") // Optional

	// Skip files if required columns are not found
	if emailIndex == -1 || nameIndex == -1 {
		utils.LogMessage(fmt.Sprintf("Required columns (Name, Email) not found in CSV file: %s, skipping...", filename))
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

func LoadXLSX(filename string, recordChan chan<- Record) {
	file, err := xlsx.OpenFile(filename)
	if err != nil {
		utils.LogMessage(fmt.Sprintf("Error opening XLSX file: %s - %v", filename, err))
		return
	}

	for _, sheet := range file.Sheets {
		if len(sheet.Rows) == 0 {
			utils.LogMessage(fmt.Sprintf("No data found in XLSX file: %s", filename))
			continue
		}
		headers := sanitizeHeadersXLSX(sheet.Rows[0].Cells)

		// Find the required column indexes dynamically using flexible matching
		emailIndex := findFlexibleHeaderIndex(headers, "email")
		nameIndex := findFlexibleHeaderIndex(headers, "name")
		orgNameIndex := findFlexibleHeaderIndex(headers, "organization") // Optional

		// Skip files if required columns are not found
		if emailIndex == -1 || nameIndex == -1 {
			utils.LogMessage(fmt.Sprintf("Required columns (Name, Email) not found in XLSX file: %s, skipping...", filename))
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

func WriteCSV(filename string, recordsMap map[string]Record) error {
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
	for _, record := range recordsMap {
		row := append([]string{record.Name, record.OrgName, record.Email}, record.Others...)
		err = writer.Write(row)
		if err != nil {
			return err
		}
	}

	return nil
}

func WriteFilteredCSV(filename string, headers []string, records []Record) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	err = writer.Write(headers)
	if err != nil {
		return err
	}

	// Write records
	for _, record := range records {
		// Map field names to values
		recordMap := map[string]string{
			"Name":    record.Name,
			"OrgName": record.OrgName,
			"Email":   record.Email,
		}

		// Merge with OthersMap
		for k, v := range record.OthersMap {
			recordMap[k] = v
		}

		// Prepare the row data based on headers
		var row []string
		for _, header := range headers {
			if val, ok := recordMap[header]; ok {
				row = append(row, val)
			} else {
				// Header not found in record, append empty string
				row = append(row, "")
			}
		}

		err = writer.Write(row)
		if err != nil {
			return err
		}
	}

	return nil
}

// AppendCSV appends records to an existing CSV file
func AppendCSV(filename string, recordsMap map[string]Record) error {
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Append records
	for _, record := range recordsMap {
		row := append([]string{record.Name, record.OrgName, record.Email}, record.Others...)
		err = writer.Write(row)
		if err != nil {
			return err
		}
	}

	return nil
}

func LoadEmailsFromCSV(filename string) (map[string]bool, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	emails := make(map[string]bool)
	reader := csv.NewReader(file)
	reader.LazyQuotes = true
	reader.FieldsPerRecord = -1
	rows, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	if len(rows) == 0 {
		return emails, nil
	}

	headers := sanitizeHeaders(rows[0])
	emailIndex := findFlexibleHeaderIndex(headers, "email")
	if emailIndex == -1 {
		return nil, fmt.Errorf("email column not found in database file")
	}

	for _, row := range rows[1:] {
		if len(row) > emailIndex {
			email := row[emailIndex]
			emails[email] = true
		}
	}

	return emails, nil
}

func ValidateHeaders(headers []string) bool {
	requiredHeaders := []string{"Name", "Email"}
	for _, reqHeader := range requiredHeaders {
		if findFlexibleHeaderIndex(headers, reqHeader) == -1 {
			return false
		}
	}
	return true
}

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

func GetCSVHeaders(filePath string) ([]string, error) {

	file, err := os.Open(filePath)

	if err != nil {

		return nil, err

	}

	defer file.Close()

	reader := csv.NewReader(file)

	headers, err := reader.Read()

	if err != nil {

		return nil, err

	}

	log.Printf("Headers: %v", headers)
	return headers, nil

}

func getRowData(row *xlsx.Row) []string {
	var data []string
	for _, cell := range row.Cells {
		data = append(data, cell.String())
	}
	return data
}

// GetHeaders reads the headers from a CSV or XLSX file
func GetHeaders(filename string) ([]string, error) {
	ext := strings.ToLower(filepath.Ext(filename))
	if ext == ".csv" {
		return getHeadersFromCSV(filename)
	} else if ext == ".xlsx" {
		return getHeadersFromXLSX(filename)
	}
	return nil, fmt.Errorf("unsupported file type: %s", ext)
}

func getHeadersFromCSV(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.LazyQuotes = true
	reader.FieldsPerRecord = -1

	row, err := reader.Read() // Read only the first line
	if err != nil {
		return nil, err
	}

	headers := sanitizeHeaders(row)
	return headers, nil
}

func getHeadersFromXLSX(filename string) ([]string, error) {
	file, err := xlsx.OpenFile(filename)
	if err != nil {
		return nil, err
	}

	for _, sheet := range file.Sheets {
		if len(sheet.Rows) > 0 {
			headers := sanitizeHeadersXLSX(sheet.Rows[0].Cells)
			return headers, nil
		}
	}
	return nil, fmt.Errorf("no headers found in XLSX file")
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
