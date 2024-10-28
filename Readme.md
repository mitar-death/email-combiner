# DataMerge Pro

![DataMerge Pro Logo](./resources/baboon.png)

**DataMerge Pro** is a powerful and user-friendly tool designed to help you efficiently combine and filter your CSV and XLSX files. Whether you're managing large datasets or performing complex data manipulations, DataMerge Pro provides an intuitive graphical interface built with [Fyne](https://fyne.io/) to streamline your workflow.

## 🚀 Features

- **Combine Files**: Merge multiple CSV and XLSX files into a single consolidated file.
- **Filter Emails**: Remove duplicate or unwanted email entries based on your criteria.
- **User-Friendly Interface**: Easy-to-use GUI built with Fyne, offering seamless navigation between features.
- **Logging**: Track processing steps and errors with detailed logs.
- **Custom Icon**: Personalized application icon for a professional appearance on macOS Dock and Finder.

## 📦 Installation

Follow the instructions in the [INSTALL.md](./INSTALL.md) file to set up the development environment, build, and package DataMerge Pro for macOS.

## 🛠 Usage

### Combining Files

1. **Select Input Folder or Add Files**: Choose a folder containing your CSV/XLSX files or add individual files manually.
2. **Select Output Folder**: Specify where the combined file will be saved.
3. **Enter Output File Name**: Provide a name for the combined output file (e.g., `combined_output.csv`).
4. **Start Processing**: Click the "Start Processing" button to begin merging files. Monitor progress and logs in the log viewer.

### Filtering Emails

1. **Select Input File**: Choose the CSV/XLSX file containing the emails you want to filter.
2. **Select Database File**: Choose the CSV/XLSX file containing the database of emails to filter against.
3. **Select Output Folder**: Specify where the filtered file will be saved.
4. **Enter Output File Name**: Provide a name for the filtered output file (e.g., `filtered_output.csv`).
5. **Start Filtering**: Click the "Start Filtering" button to begin the filtering process. Monitor progress and logs in the log viewer.

## 📂 Project Structure

email-combiner/ ├── combine/ │ └── combine.go ├── filter/ │ └── filter.go ├── droparea/ │ └── droparea.go ├── records/ │ └── records.go ├── utils/ │ └── utils.go ├── resources/ │ ├── baboon.icns │ └── baboon.png ├── fyne.yaml ├── main.go ├── go.mod ├── go.sum ├── README.md └── INSTALL.md

## 🧰 Dependencies

- [Go](https://golang.org/) (version 1.17 or higher)
- [Fyne](https://fyne.io/) GUI toolkit
- [tealeg/xlsx](https://github.com/tealeg/xlsx) for handling XLSX files

## 📄 License

[MIT](./LICENSE)
