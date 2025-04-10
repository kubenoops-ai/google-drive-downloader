package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/kubenoops-ai/google-drive-downloader/pkg/drive"
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
	)

	flag.StringVar(&credentials, "credentials", "credentials.json", "Path to credentials file")
	flag.StringVar(&folderID, "folder-id", "", "Folder ID to start search from (optional)")
	flag.StringVar(&pattern, "pattern", "", "Regex pattern to match files")
	flag.IntVar(&maxDepth, "max-depth", -1, "Maximum depth to search (-1 for unlimited)")
	flag.BoolVar(&dryRun, "dry-run", false, "Only list files, don't download")
	flag.StringVar(&outputDir, "output-dir", "output", "Directory to save downloaded files")
	flag.BoolVar(&verbose, "verbose", false, "Enable verbose logging")
	flag.IntVar(&maxResults, "max", 0, "Maximum number of files to return (0 for unlimited)")

	flag.Parse()

	if pattern == "" {
		fmt.Println("Error: pattern is required")
		flag.Usage()
		os.Exit(1)
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
		fmt.Println("\nDry run completed. No files were downloaded.")
		return
	}

	if err := driveService.DownloadFiles(files, config.OutputDir); err != nil {
		fmt.Printf("Error downloading files: %v\n", err)
		os.Exit(1)
	}
}
