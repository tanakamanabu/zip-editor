# Simple Walk GUI Application

A simple Windows GUI application created with [Walk](https://github.com/lxn/walk), a Windows GUI toolkit for Go.

## Features

- Displays a window with a button
- Shows a message box when the button is clicked

## Requirements

- Go 1.24 or later
- Windows operating system (Walk is Windows-only)
- GCC compiler (for CGo)

## Installation

1. Clone this repository:
   ```
   git clone https://github.com/yourusername/zip-editor.git
   cd zip-editor
   ```

2. Install dependencies:
   ```
   go mod tidy
   ```

## Building and Running

### Regular Build

```
go build
```

Then run the executable:

```
zip-editor.exe
```

### Building with manifest for better DPI support

For better DPI scaling support, you can build with the rsrc tool:

1. Install rsrc:
   ```
   go install github.com/akavel/rsrc@latest
   ```

2. Create a manifest file (already included in this repo)

3. Build with rsrc:
   ```
   rsrc -manifest zip-editor.manifest -o rsrc.syso
   go build
   ```

## License

[MIT License](LICENSE)