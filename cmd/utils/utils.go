package utils

import (
	"errors"
	"log"
	"os"
	"strings"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
)

// Global variables to store the last used directories
var LastFileDirectory fyne.ListableURI
var LastFolderDirectory fyne.ListableURI

var Logger *log.Logger
var LogMessages []string
var LogMutex sync.Mutex

func InitializeLogger(logFilePath string) error {
	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	Logger = log.New(logFile, "", log.Ldate|log.Ltime)
	return nil
}

// Helper functions
func ShowError(err error, win fyne.Window) {
	dialog.ShowError(err, win)
}

func ShowInfo(message string, win fyne.Window) {
	dialog.ShowInformation("Info", message, win)
}

func LogMessage(message string) {
	LogMutex.Lock()
	defer LogMutex.Unlock()
	if Logger != nil {
		Logger.Println(message)
	}
	LogMessages = append(LogMessages, message)
}

func ShowFolderSelectionDialog(pathEntry *widget.Entry, win fyne.Window) {
	folderDialog := dialog.NewFolderOpen(func(uri fyne.ListableURI, err error) {
		if err != nil {
			ShowError(err, win)
			return
		}
		if uri != nil {
			pathEntry.SetText(uri.Path())
		}
	}, win)
	folderDialog.SetFilter(storage.NewExtensionFileFilter([]string{}))

	// Reset the location to force refresh
	folderDialog.SetLocation(nil)

	folderDialog.Show()
}

func ShowFileOpenDialog(filePath *string, fileEntry *widget.Entry, win fyne.Window) {
	fileDialog := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil {
			ShowError(err, win)
			return
		}
		if reader != nil {
			ext := strings.ToLower(reader.URI().Extension())
			if ext == ".csv" || ext == ".xlsx" {
				*filePath = reader.URI().Path()
				fileEntry.SetText(*filePath)
				reader.Close()
			} else {
				ShowError(errors.New("Unsupported file type selected"), win)
			}
		}
	}, win)
	fileDialog.SetFilter(storage.NewExtensionFileFilter([]string{".csv", ".xlsx"}))

	// Reset the location to force refresh
	fileDialog.SetLocation(nil)

	fileDialog.Show()
}

func ShowFileSelectionDialog(selectedFiles *[]string, inputPathEntry *widget.Entry, win fyne.Window) {
	fileDialog := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil {
			ShowError(err, win)
			return
		}
		if reader != nil {
			*selectedFiles = append(*selectedFiles, reader.URI().Path())
			inputPathEntry.SetText(strings.Join(*selectedFiles, "\n"))
			reader.Close()
		}
	}, win)
	fileDialog.Show()
}

func TruncateString(s string, length int) string {
	if len(s) > length {
		return s[:length] + "..."
	}
	return s
}

func DisplayHeadersInList(headerDisplay *widget.Entry, headers []string) {
	headerText := strings.Join(headers, ", ")
	headerText = strings.ToTitle(headerText)
	headerDisplay.SetText(headerText)
}
