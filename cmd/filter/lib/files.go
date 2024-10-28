package lib

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"website-copier/cmd/records"
	"website-copier/cmd/utils"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"github.com/sqweek/dialog"
)

type FilterConfig struct {
	selectedInputFiles []string
	fileHeaders        map[string][]string
	fileList           *widget.List
	fileListContainer  *fyne.CanvasObject
	inputPathEntry     *widget.Entry
}

func NewFilterConfig(
	selectedInputFiles *[]string,
	fileHeaders map[string][]string,
	fileList *widget.List,
	inputPathEntry *widget.Entry) *FilterConfig {
	return &FilterConfig{
		selectedInputFiles: *selectedInputFiles,
		fileHeaders:        fileHeaders,
		fileList:           fileList,
		inputPathEntry:     inputPathEntry,
	}
}

func ShowFileSelectionDialog(
	selectedInputFiles *[]string,
	fileHeaders map[string][]string,

	inputPathEntry *widget.Entry,
	fileList *widget.List,
	win fyne.Window) *widget.Button {
	selectFilesBtn := widget.NewButton("Select Files", func() {
		files := []string{}
		for {
			file, err := dialog.File().Title("Select Input Files").Filter("CSV and XLSX Files", "csv", "xlsx").Load()
			if err != nil {
				break // User cancelled or an error occurred
			}
			files = append(files, file)

			// Ask user if they want to select another file
			anotherFile := dialog.Message("%s\n\nDo you want to select another file?", file).Title("Select Another File").YesNo()
			if !anotherFile {
				break
			}
		}

		*selectedInputFiles = append(*selectedInputFiles, files...)
		for _, file := range files {
			headers, err := records.GetHeaders(file)
			if err == nil {
				fileHeaders[file] = headers
			}
		}
		inputPathEntry.SetText(strings.Join(*selectedInputFiles, "\n"))
		fileList.Refresh()
	})

	return selectFilesBtn
}

func ShowFolderSelectionDialog(
	selectedInputFiles *[]string,
	fileHeaders map[string][]string,
	inputPathEntry *widget.Entry,
	headerDisplay *widget.Entry,
	fileList *widget.List) *widget.Button {
	selectFolderBtn := widget.NewButton("Select Folder", func() {
		folderPath, err := dialog.Directory().Title("Select Input Folder").Browse()
		if err != nil {
			return // User cancelled or an error occurred
		}

		// Clear previous selections
		selectedInputFiles := ClearPreviousSelection(selectedInputFiles, fileHeaders, inputPathEntry, headerDisplay, fileList)

		// Walk through the folder and collect CSV/XLSX files
		err = filepath.Walk(folderPath, func(p string, info os.FileInfo, err error) error {
			if err != nil {
				return nil // Skip this file and continue
			}
			if !info.IsDir() {
				ext := strings.ToLower(filepath.Ext(p))
				if ext == ".csv" || ext == ".xlsx" {
					*selectedInputFiles = append(*selectedInputFiles, p)
					headers, err := records.GetHeaders(p)
					if err == nil {
						fileHeaders[p] = headers
					}
				}
			}
			return nil
		})
		if err != nil {
			utils.ShowError(fmt.Errorf("Error reading folder: %v", err), nil)
			return
		}
		inputPathEntry.SetText(strings.Join(*selectedInputFiles, "\n"))
		fileList.Refresh()
	})

	return selectFolderBtn
}

func ClearSelectionButton(
	selectedInputFiles *[]string,
	fileHeaders map[string][]string,
	inputPathEntry *widget.Entry,
	headerDisplay *widget.Entry,
	fileList *widget.List) *widget.Button {
	clearSelectionBtn := widget.NewButton("Clear Selection", func() {
		ClearPreviousSelection(selectedInputFiles, fileHeaders, inputPathEntry, headerDisplay, fileList)
	})

	return clearSelectionBtn
}

func ClearPreviousSelection(
	selectedInputFiles *[]string,
	fileHeaders map[string][]string,
	inputPathEntry *widget.Entry,
	headerDisplay *widget.Entry,
	fileList *widget.List) *[]string {

	// Clear previous selections
	*selectedInputFiles = []string{}
	for k := range fileHeaders {
		delete(fileHeaders, k)
	}

	inputPathEntry.SetText("")
	fileList.Refresh()
	headerDisplay.SetText("")

	return selectedInputFiles
}
