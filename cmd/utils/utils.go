package utils

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// Logger and Log Management
var (
	Logger      *log.Logger
	LogMessages []string
	LogMutex    sync.Mutex
)

// InitializeLogger sets up the logger to write to a specified file.
func InitializeLogger(logFilePath string) error {
	logDir := filepath.Dir(logFilePath)
	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		err = os.MkdirAll(logDir, os.ModePerm)
		if err != nil {
			return fmt.Errorf("failed to create log directory: %v", err)
		}
	}

	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		return fmt.Errorf("failed to open log file: %v", err)
	}
	Logger = log.New(logFile, "", log.Ldate|log.Ltime)
	return nil
}

// ShowError displays an error dialog.
func ShowError(err error, win fyne.Window) {
	if err != nil && win != nil {
		dialog.ShowError(err, win)
	}
}

// ShowInfo displays an information dialog.
func ShowInfo(message string, win fyne.Window) {
	if win != nil {
		dialog.ShowInformation("Info", message, win)
	}
}

// LogMessage logs a message to both the logger and the in-memory log.
func LogMessage(message string) {
	LogMutex.Lock()
	defer LogMutex.Unlock()
	if Logger != nil {
		Logger.Println(message)
	}
	LogMessages = append(LogMessages, message)
}

// ShowFolderSelectionDialog displays a folder selection dialog and updates the path entry.
func ShowFolderSelectionDialog(pathEntry *widget.Entry, win fyne.Window) {
	if pathEntry == nil || win == nil {
		return
	}

	dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
		if err != nil {
			ShowError(err, win)
			return
		}
		if uri != nil {
			pathEntry.SetText(uri.Path())
		}
	}, win)
}

// ShowFileOpenDialog displays a file open dialog for selecting a single file.
func ShowFileOpenDialog(filePath *string, fileEntry *widget.Entry, win fyne.Window) {
	if filePath == nil || fileEntry == nil || win == nil {
		return
	}

	dialog.ShowFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil {
			ShowError(err, win)
			return
		}
		if reader != nil {
			defer reader.Close()
			ext := strings.ToLower(reader.URI().Extension())
			if ext == ".csv" || ext == ".xlsx" {
				*filePath = reader.URI().Path()
				fileEntry.SetText(*filePath)
			} else {
				ShowError(errors.New("unsupported file type selected"), win)
			}
		}
	}, win)
}

// ShowFileSelectionDialog displays a file selection dialog for selecting multiple files.
func ShowFileSelectionDialog(selectedFiles *[]string, inputPathEntry *widget.Entry, win fyne.Window) {
	if selectedFiles == nil || inputPathEntry == nil || win == nil {
		return
	}

	dialog.ShowFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil {
			ShowError(err, win)
			return
		}
		if reader != nil {
			defer reader.Close()
			*selectedFiles = append(*selectedFiles, reader.URI().Path())
			inputPathEntry.SetText(strings.Join(*selectedFiles, "\n"))
		}
	}, win)
}

// TruncateString truncates a string to a specified length, adding "..." if it exceeds that length.
func TruncateString(s string, length int) string {
	if len(s) > length {
		return s[:length] + "..."
	}
	return s
}

// DisplayHeadersInList displays headers in a text entry widget.
func DisplayHeadersInList(headerDisplay *widget.Entry, headers []string) {
	if headerDisplay != nil {
		headerText := strings.Join(headers, ", ")
		headerText = strings.Title(headerText)
		headerDisplay.SetText(headerText)
	}
}
