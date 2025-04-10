package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kubenoops-ai/google-drive-downloader/pkg/drive"
	"github.com/kubenoops-ai/google-drive-downloader/pkg/transform"
	"github.com/kubenoops-ai/google-drive-downloader/pkg/utils"
)

func main() {
	var (
		credentials string
		folderID    string
		pattern     string
		maxDepth    int
		dryRun      bool
		outputDir   string
		verbose     bool
		maxResults  int
		pathPattern string
		pathFormat  string
	)

	flag.StringVar(&credentials, "credentials", "credentials.json", "Path to credentials file")
	flag.StringVar(&folderID, "folder-id", "", "Folder ID to start search from (optional)")
	flag.StringVar(&pattern, "pattern", "", "Regex pattern to match files")
	flag.IntVar(&maxDepth, "max-depth", -1, "Maximum depth to search (-1 for unlimited)")
	flag.BoolVar(&dryRun, "dry-run", false, "Only list files, don't download")
	flag.StringVar(&outputDir, "output-dir", "output", "Directory to save downloaded files")
	flag.BoolVar(&verbose, "verbose", false, "Enable verbose logging")
	flag.IntVar(&maxResults, "max", 0, "Maximum number of files to return (0 for unlimited)")
	flag.StringVar(&pathPattern, "path-pattern", "", "Regex pattern with named groups to transform output paths (e.g. 'Zoom Recordings/(?P<date>[^/]+)/.*\\.TRANSCRIPT')")
	flag.StringVar(&pathFormat, "path-format", "", "Format string for transformed paths using named groups (e.g. '${date}.TRANSCRIPT')")

	flag.Parse()

	if pattern == "" {
		fmt.Println("Error: pattern is required")
		flag.Usage()
		os.Exit(1)
	}

	// Validate path transformation flags
	if (pathPattern == "") != (pathFormat == "") {
		fmt.Println("Error: both path-pattern and path-format must be provided together")
		flag.Usage()
		os.Exit(1)
	}

	var pathTransformer *transform.PathTransformer
	if pathPattern != "" {
		var err error
		pathTransformer, err = transform.NewPathTransformer(pathPattern, pathFormat)
		if err != nil {
			fmt.Printf("Error creating path transformer: %v\n", err)
			os.Exit(1)
		}
	}

	config := utils.Config{
		Credentials: credentials,
		FolderID:    folderID,
		Pattern:     pattern,
		MaxDepth:    maxDepth,
		DryRun:      dryRun,
		OutputDir:   outputDir,
		Verbose:     verbose,
	}

	driveService, err := drive.NewDriveService(config.Credentials, config.Verbose)
	if err != nil {
		fmt.Printf("Error creating Drive service: %v\n", err)
		os.Exit(1)
	}

	files, err := driveService.ListFiles(config.FolderID, config.Pattern, config.MaxDepth, maxResults)
	if err != nil {
		fmt.Printf("Error listing files: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\nFound %d matching files:\n", len(files))
	for _, file := range files {
		fmt.Printf("- %s (Modified: %s)\n", file.Path, file.ModifiedTime)
	}

	if config.DryRun {
		fmt.Println("\nFound files:")
		for _, file := range files {
			fmt.Printf("- %s (Modified: %s)\n", file.Path, file.ModifiedTime)
		}

		fmt.Println("\nDownload preview:")
		for _, file := range files {
			fmt.Printf("\n📄 Original file: %s\n", file.Path)
			if pathTransformer != nil {
				fmt.Printf("   🔍 Applying pattern: %q\n", pathPattern)
				fmt.Printf("   📝 Using format: %q\n", pathFormat)
				newPath, err := pathTransformer.Transform(file.Path)
				if err != nil {
					fmt.Printf("   ❌ Transformation failed: %v\n", err)
					fmt.Printf("   📁 Will be saved as: %s\n", filepath.Join(config.OutputDir, file.Path))
				} else {
					fmt.Printf("   ✅ Transformed to: %q\n", newPath)
					fmt.Printf("   📁 Will be saved as: %s\n", filepath.Join(config.OutputDir, newPath))
				}
			} else {
				fmt.Printf("   📁 Will be saved as: %s\n", filepath.Join(config.OutputDir, file.Path))
			}
		}
		fmt.Println("\nDry run completed. No files were downloaded.")
		return
	}

	if pathTransformer != nil {
		fmt.Println("\nTransforming file paths before downloading:")
		// Transform file paths before downloading
		for i := range files {
			fmt.Printf("\n🔍 Processing file %d/%d:\n", i+1, len(files))
			fmt.Printf("   Input path: %q\n", files[i].Path)
			fmt.Printf("   Using pattern: %q\n", pathPattern)
			fmt.Printf("   Using format: %q\n", pathFormat)
			newPath, err := pathTransformer.Transform(files[i].Path)
			if err != nil {
				fmt.Printf("   ❌ Warning: Could not transform path: %v\n", err)
				continue
			}
			fmt.Printf("   ✅ Successfully transformed to: %q\n", newPath)
			files[i].Path = newPath
		}
	}

	if err := driveService.DownloadFiles(files, config.OutputDir); err != nil {
		fmt.Printf("Error downloading files: %v\n", err)
		os.Exit(1)
	}
}
