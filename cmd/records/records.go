package records

import (
	"encoding/csv"
	"fmt"
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
	OthersMap map[string]string
	FilePath  string
}

// Load records from CSV or XLSX file
func LoadRecords(filename string) ([]Record, []string, error) {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".csv":
		return loadRecordsFromCSV(filename)
	case ".xlsx":
		return loadRecordsFromXLSX(filename)
	default:
		return nil, nil, fmt.Errorf("unsupported file type: %s", ext)
	}
}

// loadRecordsFromCSV loads records from a CSV file.
func loadRecordsFromCSV(filename string) ([]Record, []string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, nil, fmt.Errorf("error opening CSV file %s: %v", filename, err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.LazyQuotes = true
	reader.FieldsPerRecord = -1

	rows, err := reader.ReadAll()
	if err != nil {
		return nil, nil, fmt.Errorf("error reading CSV file %s: %v", filename, err)
	}

	if len(rows) == 0 {
		return nil, nil, fmt.Errorf("no data found in CSV file: %s", filename)
	}

	headers := sanitizeHeaders(rows[0])

	// Find required columns
	emailIndex := findFlexibleHeaderIndex(headers, "email")
	nameIndex := findFlexibleHeaderIndex(headers, "name")
	orgNameIndex := findFlexibleHeaderIndex(headers, "organization")

	if emailIndex == -1 || nameIndex == -1 {
		return nil, nil, fmt.Errorf("required columns (Name, Email) not found in CSV file: %s", filename)
	}

	records := make([]Record, 0, len(rows)-1)

	for _, row := range rows[1:] {
		record, err := createRecordFromRow(row, headers, nameIndex, emailIndex, orgNameIndex)
		if err != nil {
			// Skip invalid rows but log the error
			utils.LogMessage(fmt.Sprintf("Error processing row in file %s: %v", filename, err))
			continue
		}
		record.FilePath = filename
		records = append(records, record)
	}

	return records, headers, nil
}

// loadRecordsFromXLSX loads records from an XLSX file.
func loadRecordsFromXLSX(filename string) ([]Record, []string, error) {
	file, err := xlsx.OpenFile(filename)
	if err != nil {
		return nil, nil, fmt.Errorf("error opening XLSX file %s: %v", filename, err)
	}

	var records []Record
	var headers []string
	for _, sheet := range file.Sheets {
		if len(sheet.Rows) == 0 {
			continue
		}
		headers = sanitizeHeadersXLSX(sheet.Rows[0].Cells)

		// Find required columns
		emailIndex := findFlexibleHeaderIndex(headers, "email")
		nameIndex := findFlexibleHeaderIndex(headers, "name")
		orgNameIndex := findFlexibleHeaderIndex(headers, "organization")

		if emailIndex == -1 || nameIndex == -1 {
			return nil, nil, fmt.Errorf("required columns (Name, Email) not found in XLSX file: %s", filename)
		}

		for _, row := range sheet.Rows[1:] {
			record, err := createRecordFromXLSXRow(row, headers, nameIndex, emailIndex, orgNameIndex)
			if err != nil {
				utils.LogMessage(fmt.Sprintf("Error processing row in file %s: %v", filename, err))
				continue
			}
			record.FilePath = filename
			records = append(records, record)
		}
	}

	if len(records) == 0 {
		return nil, nil, fmt.Errorf("no valid records found in XLSX file: %s", filename)
	}

	return records, headers, nil
}

// createRecordFromRow creates a Record from a CSV row.
func createRecordFromRow(row []string, headers []string, nameIndex, emailIndex, orgNameIndex int) (Record, error) {
	if len(row) <= emailIndex || len(row) <= nameIndex {
		return Record{}, fmt.Errorf("row does not have required columns")
	}

	rowMap := mapRowData(row, headers)
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

	return Record{
		Name:      name,
		OrgName:   orgName,
		Email:     email,
		OthersMap: rowMap,
	}, nil
}

// createRecordFromXLSXRow creates a Record from an XLSX row.
func createRecordFromXLSXRow(row *xlsx.Row, headers []string, nameIndex, emailIndex, orgNameIndex int) (Record, error) {
	if len(row.Cells) <= emailIndex || len(row.Cells) <= nameIndex {
		return Record{}, fmt.Errorf("row does not have required columns")
	}

	rowMap := mapXLSXRowData(row, headers)
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

	return Record{
		Name:      name,
		OrgName:   orgName,
		Email:     email,
		OthersMap: rowMap,
	}, nil
}

// mapXLSXRowData maps XLSX row data to a map using headers.
func mapXLSXRowData(row *xlsx.Row, headers []string) map[string]string {
	rowMap := make(map[string]string)
	for i, header := range headers {
		if i < len(row.Cells) {
			rowMap[header] = row.Cells[i].String()
		} else {
			rowMap[header] = ""
		}
	}
	return rowMap
}

// mapRowData maps CSV row data to a map using headers.
func mapRowData(row []string, headers []string) map[string]string {
	rowMap := make(map[string]string)
	for i, header := range headers {
		if i < len(row) {
			rowMap[header] = row[i]
		} else {
			rowMap[header] = ""
		}
	}
	return rowMap
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

// WriteFilteredCSV writes filtered records to a CSV file.
func WriteFilteredCSV(filename string, headers []string, records []Record) error {
	// **Ensure the output directory exists**

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("error creating output CSV file %s: %v", filename, err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write headers
	if err := writer.Write(headers); err != nil {
		return fmt.Errorf("error writing headers to CSV file %s: %v", filename, err)
	}

	for _, record := range records {
		row := make([]string, len(headers))
		for i, header := range headers {
			value := record.OthersMap[header]
			switch header {
			case "Name":
				value = record.Name
			case "OrgName":
				value = record.OrgName
			case "Email":
				value = record.Email
			}
			row[i] = value
		}
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("error writing record to CSV file %s: %v", filename, err)
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

// LoadEmailsFromCSV loads emails from a CSV file into a map for quick lookup.
func LoadEmailsFromCSV(filename string) (map[string]bool, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("error opening CSV file %s: %v", filename, err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.LazyQuotes = true
	reader.FieldsPerRecord = -1

	rows, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("error reading CSV file %s: %v", filename, err)
	}

	if len(rows) == 0 {
		return nil, fmt.Errorf("no data found in CSV file: %s", filename)
	}

	headers := sanitizeHeaders(rows[0])
	emailIndex := findFlexibleHeaderIndex(headers, "email")
	if emailIndex == -1 {
		return nil, fmt.Errorf("email column not found in database file: %s", filename)
	}

	emails := make(map[string]bool, len(rows)-1)
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

// findFlexibleHeaderIndex finds the index of a header containing the keyword.
func findFlexibleHeaderIndex(headers []string, keyword string) int {
	lowerKeyword := strings.ToLower(keyword)
	for i, header := range headers {
		h := strings.ToLower(header)
		if strings.Contains(h, lowerKeyword) {
			return i
		}
	}
	return -1
}

// sanitizeHeaders trims headers and removes surrounding quotes.
func sanitizeHeaders(headers []string) []string {
	for i, header := range headers {
		header = strings.TrimSpace(header)
		header = strings.Trim(header, `"`)
		headers[i] = header
	}
	return headers
}

// sanitizeHeadersXLSX trims XLSX headers and removes surrounding quotes.
func sanitizeHeadersXLSX(cells []*xlsx.Cell) []string {
	headers := make([]string, len(cells))
	for i, cell := range cells {
		header := strings.TrimSpace(cell.String())
		header = strings.Trim(header, `"`)
		headers[i] = header
	}
	return headers
}

// GetHeaders reads the headers from a CSV or XLSX file.
func GetHeaders(filename string) ([]string, error) {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".csv":
		return getHeadersFromCSV(filename)
	case ".xlsx":
		return getHeadersFromXLSX(filename)
	default:
		return nil, fmt.Errorf("unsupported file type: %s", ext)
	}
}

func getHeadersFromCSV(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("error opening CSV file %s: %v", filename, err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.LazyQuotes = true
	reader.FieldsPerRecord = -1

	row, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("error reading headers from CSV file %s: %v", filename, err)
	}

	headers := sanitizeHeaders(row)
	return headers, nil
}

func getHeadersFromXLSX(filename string) ([]string, error) {
	file, err := xlsx.OpenFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error opening XLSX file %s: %v", filename, err)
	}

	for _, sheet := range file.Sheets {
		if len(sheet.Rows) > 0 {
			headers := sanitizeHeadersXLSX(sheet.Rows[0].Cells)
			return headers, nil
		}
	}
	return nil, fmt.Errorf("no headers found in XLSX file: %s", filename)
}
