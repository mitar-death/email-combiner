# Installation Guide for DataMerge Pro

This guide provides step-by-step instructions to set up the development environment, build, and package the **DataMerge Pro** application for macOS.

## 🖥 Prerequisites

Before you begin, ensure you have the following installed on your macOS system:

- **Go**: Version 1.17 or higher. Download from [golang.org](https://golang.org/dl/).
- **Fyne CLI**: For packaging the application.
- **Xcode Command Line Tools**: Provides necessary build tools.

### Installing Go

1. **Download Go**: Visit the [official Go download page](https://golang.org/dl/) and download the macOS installer.
2. **Install Go**: Run the installer and follow the on-screen instructions.
3. **Verify Installation**:

   ```sh
   go version
   ```

### build

`fyne package -os darwin -icon ./resources/baboon.png -name "DataMerge Pro"`
