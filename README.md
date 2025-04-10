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
- 🔀 Path transformation using regex capture groups

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
- `-path-pattern`: Regex pattern with named capture groups for path transformation
- `-path-format`: Output format string using captured variables from path-pattern

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

4. Transform paths to include date and room from folder name:
```bash
./google-drive-downloader -pattern ".*\.TRANSCRIPT$" \
  --path-pattern "(?P<date>[^-]+-[^-]+-[^-]+-[^-]+-[^-]+-[^-]+)-(?P<room>[^/]+)/.*\.TRANSCRIPT$" \
  --path-format '${date}-${room}.TRANSCRIPT'
```

5. Extract date components into a formatted filename:
```bash
./google-drive-downloader -pattern ".*\.TRANSCRIPT$" \
  --path-pattern "(?P<month>[^-]+)-(?P<day>[^-]+)-(?P<year>[^-]+)-(?P<hour>[^-]+)-(?P<minute>[^-]+)-(?P<second>[^-]+)-.*" \
  --path-format '${year}-${month}-${day}_${hour}-${minute}.TRANSCRIPT'
```

6. Keep only the room name in the output filename:
```bash
./google-drive-downloader -pattern ".*\.TRANSCRIPT$" \
  --path-pattern ".*-(?P<room>AI_[^/]+)/.*\.TRANSCRIPT$" \
  --path-format '${room}.TRANSCRIPT'
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
2. Create a service account:
   - Go to "IAM & Admin" > "Service Accounts"
   - Click "Create Service Account"
   - Fill in the service account details
   - No need to grant any roles (permissions will be handled in Google Drive)
3. Create and download the service account key:
   - Select your service account
   - Go to "Keys" tab
   - Add Key > Create new key > JSON
   - Save the downloaded JSON file as `credentials.json`
4. Share your Google Drive folders:
   - Copy the service account email (ends with `@...iam.gserviceaccount.com`)
   - Share your Google Drive folders with this email address
   - Grant at least "Viewer" access
5. Place `credentials.json` in the same directory as the binary or specify its path using `-credentials`

## Notes

- Files in trash are automatically skipped
- The tool supports both personal and shared drives
- When using `-max`, files are sorted by modification date (newest first) before limiting
- Use `-dry-run` to preview which files would be downloaded
- The `-verbose` flag provides detailed logging of the search and download process

## Path Transformations

The tool supports transforming output file paths using regex capture groups. This is useful for:
- Extracting date/time components from folder names
- Simplifying complex folder structures
- Organizing files by captured metadata
- Standardizing file naming conventions

### Path Transformation Format

1. Use `-path-pattern` to define a regex with named capture groups:
   - Named groups use the format `(?P<name>pattern)`
   - Example: `(?P<date>[^-]+)-(?P<type>[^/]+)`

2. Use `-path-format` to define the output format:
   - Reference captured groups with `${name}`
   - Example: `${date}_${type}.txt`

### Important Notes

- Quotes in path format strings:
  - Always use single quotes (`'`) around the path format to prevent shell expansion of `${variables}`
  - Using double quotes (`"`) will cause the shell to try to expand the variables before passing to the program
- The path pattern must match the entire path you want to transform
- All capture groups referenced in the format must exist in the pattern
- Use the `-dry-run` flag to test path transformations before downloading

### Example Transformations

Input path: `2025-04-10-17-27-28-AI_TEAM_OFFICE_ROOM-2/audio_transcript.TRANSCRIPT`

1. Basic date extraction:
```
Pattern: (?P<date>[^-]+-[^-]+-[^-]+)-.*\.TRANSCRIPT$
Format: ${date}.TRANSCRIPT
Output: 2025-04-10.TRANSCRIPT
```

2. Date and type combination:
```
Pattern: (?P<date>[^-]+-[^-]+-[^-]+)-.*?/(?P<type>[^_]+)_.*\.TRANSCRIPT$
Format: ${date}_${type}.TRANSCRIPT
Output: 2025-04-10_audio.TRANSCRIPT
```

3. Complex date formatting:
```
Pattern: (?P<year>[^-]+)-(?P<month>[^-]+)-(?P<day>[^-]+)-(?P<time>[^-]+-[^-]+-[^-]+)-.*
Format: ${year}/${month}/${day}/${time}.TRANSCRIPT
Output: 2025/04/10/17-27-28.TRANSCRIPT
```

## License

MIT License 