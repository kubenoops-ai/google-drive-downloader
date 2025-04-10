package drive

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

type DriveService struct {
	service *drive.Service
	verbose bool
}

type FileInfo struct {
	ID           string
	Name         string
	Path         string
	MimeType     string
	ModifiedTime string
}

func NewDriveService(credentialsFile string, verbose bool) (*DriveService, error) {
	ctx := context.Background()
	srv, err := drive.NewService(ctx, option.WithCredentialsFile(credentialsFile))
	if err != nil {
		return nil, fmt.Errorf("unable to create Drive service: %v", err)
	}

	return &DriveService{service: srv, verbose: verbose}, nil
}

func (d *DriveService) log(format string, args ...interface{}) {
	if d.verbose {
		fmt.Printf(format+"\n", args...)
	}
}

func (d *DriveService) ListFiles(folderID string, pattern string, maxDepth int, maxResults int) ([]FileInfo, error) {
	var files []FileInfo
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid regex pattern: %v", err)
	}

	d.log("Starting search with pattern: %s", pattern)

	// First, get the root folder if no folder ID is provided
	if folderID == "" {
		d.log("No folder ID provided, getting root folder...")
		root, err := d.service.Files.Get("root").Fields("id").Do()
		if err != nil {
			return nil, fmt.Errorf("unable to get root folder: %v", err)
		}
		folderID = root.Id
		d.log("Using root folder ID: %s", folderID)
	}

	err = d.listFilesRecursive(folderID, "", regex, maxDepth, 0, maxResults, &files)
	if err != nil {
		return nil, err
	}

	// Sort files by modification time (newest first)
	sort.Slice(files, func(i, j int) bool {
		return files[i].ModifiedTime > files[j].ModifiedTime
	})

	// Limit results if maxResults is specified
	if maxResults > 0 && len(files) > maxResults {
		files = files[:maxResults]
	}

	d.log("\nSearch completed. Found %d matching files (showing %d).", len(files), len(files))
	return files, nil
}

func (d *DriveService) getFullPath(fileID string, folderNames map[string]string) (string, error) {
	file, err := d.service.Files.Get(fileID).
		Fields("id, name, parents").
		SupportsAllDrives(true).
		Do()
	if err != nil {
		return "", err
	}

	path := file.Name
	if len(file.Parents) > 0 {
		parentPath, err := d.getFullPath(file.Parents[0], folderNames)
		if err != nil {
			return path, nil // Return just the file name if we can't get parent path
		}
		path = filepath.Join(parentPath, path)
	}
	return path, nil
}

func (d *DriveService) cleanPath(path string) string {
	// Remove redundant "Drive/zoom-recordings" prefix if it appears after "Zoom Recordings"
	if strings.Contains(path, "Zoom Recordings/Drive/zoom-recordings/") {
		path = strings.Replace(path, "Zoom Recordings/Drive/zoom-recordings/", "Zoom Recordings/", 1)
	}
	return path
}

func (d *DriveService) listFilesRecursive(folderID, parentPath string, pattern *regexp.Regexp, maxDepth, currentDepth, maxResults int, files *[]FileInfo) error {
	if maxDepth != -1 && currentDepth > maxDepth {
		d.log("Reached max depth (%d) at path: %s", maxDepth, parentPath)
		return nil
	}

	// Early return if we've reached maxResults
	if maxResults > 0 && len(*files) >= maxResults {
		d.log("Reached max results (%d), stopping search", maxResults)
		return nil
	}

	indent := strings.Repeat("  ", currentDepth)
	d.log("%s📂 Entering directory: %s (depth: %d)", indent, parentPath, currentDepth)

	// Try both search methods
	query := fmt.Sprintf("'%s' in parents", folderID)
	d.log("%s🔍 Querying files with: %s", indent, query)

	r, err := d.service.Files.List().
		Q(query).
		Fields("files(id, name, mimeType, trashed, driveId, owners, permissions, parents, modifiedTime)").
		OrderBy("modifiedTime desc").
		IncludeItemsFromAllDrives(true).
		SupportsAllDrives(true).
		PageSize(1000).
		Do()
	if err != nil {
		return fmt.Errorf("unable to list files in folder %s: %v", folderID, err)
	}

	// If no files found, try a broader search
	if len(r.Files) == 0 && currentDepth == 1 { // Only do this for the first level to avoid too many API calls
		d.log("%s📂 Folder appears empty, trying broader search...", indent)
		query = fmt.Sprintf("fullText contains 'TRANSCRIPT' and name contains '.TRANSCRIPT'")
		r, err = d.service.Files.List().
			Q(query).
			Fields("files(id, name, mimeType, trashed, driveId, owners, permissions, parents, modifiedTime)").
			OrderBy("modifiedTime desc").
			IncludeItemsFromAllDrives(true).
			SupportsAllDrives(true).
			PageSize(1000).
			Do()
		if err != nil {
			d.log("%s⚠️ Broader search failed: %v", indent, err)
		}

		// If we found files, get their full paths
		if err == nil && len(r.Files) > 0 {
			// Create a map to store folder names for caching
			folderNames := make(map[string]string)

			// Create a new file list with proper paths
			var newFiles []*drive.File
			for _, f := range r.Files {
				fullPath, err := d.getFullPath(f.Id, folderNames)
				if err != nil {
					d.log("%s⚠️ Error getting full path for %s: %v", indent, f.Name, err)
					continue
				}
				f.Name = d.cleanPath(fullPath)
				newFiles = append(newFiles, f)
			}
			r.Files = newFiles
		}
	}

	d.log("%s📋 Found %d items in current directory", indent, len(r.Files))

	// First list all items to see what we're dealing with
	for _, f := range r.Files {
		if f.MimeType == "application/vnd.google-apps.folder" {
			d.log("%s  📂 Found subfolder: %s (ID: %s, Trashed: %v, DriveId: %s)",
				indent, f.Name, f.Id, f.Trashed, f.DriveId)
		} else {
			d.log("%s  📄 Found file: %s (Type: %s, Trashed: %v, DriveId: %s, Modified: %s)",
				indent, f.Name, f.MimeType, f.Trashed, f.DriveId, f.ModifiedTime)
		}
	}

	// Now process them
	for _, f := range r.Files {
		// Early return if we've reached maxResults
		if maxResults > 0 && len(*files) >= maxResults {
			d.log("%s  🛑 Reached max results (%d), stopping search", indent, maxResults)
			return nil
		}

		// Skip trashed files
		if f.Trashed {
			d.log("%s  ⚠️ Skipping trashed item: %s", indent, f.Name)
			continue
		}

		currentPath := filepath.Join(parentPath, f.Name)
		currentPath = d.cleanPath(currentPath)

		if f.MimeType == "application/vnd.google-apps.folder" {
			d.log("%s  🔍 Exploring subfolder: %s (ID: %s)", indent, f.Name, f.Id)
			err = d.listFilesRecursive(f.Id, currentPath, pattern, maxDepth, currentDepth+1, maxResults, files)
			if err != nil {
				return err
			}
			continue
		}

		if pattern.MatchString(f.Name) {
			d.log("%s  ✅ Found matching file: %s (Modified: %s)", indent, currentPath, f.ModifiedTime)
			*files = append(*files, FileInfo{
				ID:           f.Id,
				Name:         f.Name,
				Path:         currentPath,
				MimeType:     f.MimeType,
				ModifiedTime: f.ModifiedTime,
			})
		}
	}

	d.log("%s📂 Leaving directory: %s", indent, parentPath)
	return nil
}

func (d *DriveService) DownloadFile(fileInfo FileInfo, outputDir string) error {
	d.log("📥 Starting download of: %s", fileInfo.Path)

	outPath := filepath.Join(outputDir, fileInfo.Path)
	d.log("  Creating directory: %s", filepath.Dir(outPath))
	if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
		return fmt.Errorf("unable to create output directory: %v", err)
	}

	d.log("  Downloading file from Drive...")
	resp, err := d.service.Files.Get(fileInfo.ID).Download()
	if err != nil {
		return fmt.Errorf("unable to download file: %v", err)
	}
	defer resp.Body.Close()

	d.log("  Creating output file: %s", outPath)
	outFile, err := os.Create(outPath)
	if err != nil {
		return fmt.Errorf("unable to create output file: %v", err)
	}
	defer outFile.Close()

	d.log("  Copying file contents...")
	_, err = io.Copy(outFile, resp.Body)
	if err != nil {
		return fmt.Errorf("unable to save file: %v", err)
	}

	d.log("✅ Successfully downloaded: %s", fileInfo.Path)
	return nil
}

func (d *DriveService) DownloadFiles(files []FileInfo, outputDir string) error {
	d.log("\n📥 Starting download of %d files...", len(files))
	for _, file := range files {
		fmt.Printf("Downloading: %s\n", file.Path) // Always show this regardless of verbose mode
		if err := d.DownloadFile(file, outputDir); err != nil {
			return fmt.Errorf("error downloading %s: %v", file.Path, err)
		}
	}
	d.log("✅ All files downloaded successfully!")
	return nil
}
