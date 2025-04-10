package utils

type Config struct {
	FolderID    string
	Pattern     string
	MaxDepth    int
	DryRun      bool
	OutputDir   string
	Credentials string
	TokenPath   string
	Verbose     bool
}

// NewDefaultConfig returns a new Config with default values
func NewDefaultConfig() *Config {
	return &Config{
		MaxDepth:    -1, // -1 means unlimited depth
		DryRun:      true,
		OutputDir:   "downloads",
		Credentials: "credentials.json",
		TokenPath:   "token.json",
		Verbose:     false,
	}
}
