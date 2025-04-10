# Google Drive Downloader

A command-line tool for searching and downloading files from Google Drive with support for regex patterns, recursive folder traversal, and sorting by modification date.

## Features

- 🔍 Search files using regex patterns
- 📁 Recursive folder traversal with configurable depth
- 📅 Sort files by modification date
- 🔄 Support for shared drives
- 📊 Limit number of results
- 🏃 Dry-run mode for testing
- 📝 Verbose logging option
- 📂 Maintains folder structure when downloading

## Prerequisites

- Go 1.23 or later
- Google Drive API credentials (`credentials.json`)

## Installation

1. Clone the repository:
```bash
git clone https://github.com/kubenoops-ai/google-drive-downloader.git
cd google-drive-downloader
```

2. Build the binary:
```bash
go build -o google-drive-downloader cmd/main.go
```

## Usage

```bash
./google-drive-downloader [options]
```

### Options

- `-credentials`: Path to Google Drive API credentials file (default: "credentials.json")
- `-folder-id`: Google Drive folder ID to start search from (optional, uses root if not specified)
- `-pattern`: Regex pattern to match files (required)
- `-max-depth`: Maximum depth to search (-1 for unlimited)
- `-max`: Maximum number of files to return (0 for unlimited)
- `-dry-run`: Only list files without downloading
- `-output-dir`: Directory to save downloaded files (default: "output")
- `-verbose`: Enable verbose logging

### Examples

1. Search for TRANSCRIPT files and download the 5 most recently modified:
```bash
./google-drive-downloader -pattern ".*\.TRANSCRIPT$" -max 5
```

2. Search in a specific folder with depth limit and dry run:
```bash
./google-drive-downloader -pattern ".*\.pdf$" -folder-id "your_folder_id" -max-depth 3 -dry-run
```

3. Enable verbose logging and specify custom output directory:
```bash
./google-drive-downloader -pattern ".*\.docx$" -output-dir "downloads" -verbose
```

## Output Structure

Downloaded files maintain their Google Drive folder structure:
```
output/
└── Folder Name/
    └── Subfolder/
        └── file.ext
```

## Authentication

1. Create a Google Cloud project and enable the Google Drive API
2. Create credentials (OAuth 2.0 Client ID) and download as `credentials.json`
3. Place `credentials.json` in the same directory as the binary or specify its path using `-credentials`

## Notes

- Files in trash are automatically skipped
- The tool supports both personal and shared drives
- When using `-max`, files are sorted by modification date (newest first) before limiting
- Use `-dry-run` to preview which files would be downloaded
- The `-verbose` flag provides detailed logging of the search and download process

## License

MIT License 