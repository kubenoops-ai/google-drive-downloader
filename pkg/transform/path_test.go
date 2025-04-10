package transform

import (
	"strings"
	"testing"
)

func TestNewPathTransformer(t *testing.T) {
	tests := []struct {
		name        string
		pattern     string
		format      string
		wantErr     bool
		errContains string
	}{
		{
			name:        "empty pattern",
			pattern:     "",
			format:      "${date}.txt",
			wantErr:     true,
			errContains: "both pattern and format must be non-empty",
		},
		{
			name:        "empty format",
			pattern:     "(?P<date>\\d{4}-\\d{2}-\\d{2})",
			format:      "",
			wantErr:     true,
			errContains: "both pattern and format must be non-empty",
		},
		{
			name:        "invalid regex pattern",
			pattern:     "(?P<date>[",
			format:      "${date}.txt",
			wantErr:     true,
			errContains: "invalid regex pattern",
		},
		{
			name:    "valid pattern and format",
			pattern: "(?P<date>\\d{4}-\\d{2}-\\d{2})",
			format:  "${date}.txt",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transformer, err := NewPathTransformer(tt.pattern, tt.format)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				} else if tt.errContains != "" && !contains(err.Error(), tt.errContains) {
					t.Errorf("error = %v, want error containing %v", err, tt.errContains)
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if transformer == nil {
				t.Error("expected non-nil transformer")
			}
		})
	}
}

func TestPathTransformer_Transform(t *testing.T) {
	tests := []struct {
		name        string
		pattern     string
		format      string
		input       string
		want        string
		wantErr     bool
		errContains string
	}{
		{
			name:    "simple date extraction",
			pattern: "Zoom Recordings/(?P<date>[^/]+(?:-[^/]+){4})/.*\\.TRANSCRIPT$",
			format:  "${date}.TRANSCRIPT",
			input:   "Zoom Recordings/apr-10-2025-17-27-28-AI_TEAM_OFFICE_ROOM-2/audio_transcript.TRANSCRIPT",
			want:    "apr-10-2025-17-27-28-AI_TEAM_OFFICE_ROOM-2.TRANSCRIPT",
		},
		{
			name:    "multiple capture groups",
			pattern: "Zoom Recordings/(?P<date>[^/]+(?:-[^/]+){4})/(?P<type>[^/]+)_transcript\\.TRANSCRIPT$",
			format:  "${date}_${type}.TRANSCRIPT",
			input:   "Zoom Recordings/apr-10-2025-17-27-28-AI_TEAM_OFFICE_ROOM-2/audio_transcript.TRANSCRIPT",
			want:    "apr-10-2025-17-27-28-AI_TEAM_OFFICE_ROOM-2_audio.TRANSCRIPT",
		},
		{
			name:        "no match",
			pattern:     "Zoom Recordings/(?P<date>[^/]+(?:-[^/]+){4})/.*\\.TRANSCRIPT$",
			format:      "${date}.TRANSCRIPT",
			input:       "invalid/path/file.txt",
			wantErr:     true,
			errContains: "path does not match pattern",
		},
		{
			name:        "missing placeholder replacement",
			pattern:     "Zoom Recordings/(?P<date>[^/]+(?:-[^/]+){4})/.*\\.TRANSCRIPT$",
			format:      "${date}_${type}.TRANSCRIPT",
			input:       "Zoom Recordings/apr-10-2025-17-27-28-AI_TEAM_OFFICE_ROOM-2/audio_transcript.TRANSCRIPT",
			wantErr:     true,
			errContains: "some placeholders in format string were not replaced",
		},
		// User's example and variations
		{
			name:    "extract date from path",
			pattern: "(?P<date>[^-]+-[^-]+-[^-]+-[^-]+-[^-]+-[^-]+)-.*\\.TRANSCRIPT$",
			format:  "${date}.TRANSCRIPT",
			input:   "apr-10-2025-17-27-28-AI_TEAM_OFFICE_ROOM-2/audio_transcript.TRANSCRIPT",
			want:    "apr-10-2025-17-27-28.TRANSCRIPT",
		},
		{
			name:    "extract date and room",
			pattern: "(?P<date>[^-]+-[^-]+-[^-]+-[^-]+-[^-]+-[^-]+)-(?P<room>[^/]+)/.*\\.TRANSCRIPT$",
			format:  "${date}-${room}.TRANSCRIPT",
			input:   "apr-10-2025-17-27-28-AI_TEAM_OFFICE_ROOM-2/audio_transcript.TRANSCRIPT",
			want:    "apr-10-2025-17-27-28-AI_TEAM_OFFICE_ROOM-2.TRANSCRIPT",
		},
		{
			name:    "extract date components",
			pattern: "(?P<month>[^-]+)-(?P<day>[^-]+)-(?P<year>[^-]+)-(?P<hour>[^-]+)-(?P<minute>[^-]+)-(?P<second>[^-]+)-(?P<room>[^/]+)/.*\\.TRANSCRIPT$",
			format:  "${year}-${month}-${day}_${hour}-${minute}.TRANSCRIPT",
			input:   "apr-10-2025-17-27-28-AI_TEAM_OFFICE_ROOM-2/audio_transcript.TRANSCRIPT",
			want:    "2025-apr-10_17-27.TRANSCRIPT",
		},
		{
			name:    "extract room name only",
			pattern: ".*-(?P<room>AI_[^/]+)/.*\\.TRANSCRIPT$",
			format:  "${room}.TRANSCRIPT",
			input:   "apr-10-2025-17-27-28-AI_TEAM_OFFICE_ROOM-2/audio_transcript.TRANSCRIPT",
			want:    "AI_TEAM_OFFICE_ROOM-2.TRANSCRIPT",
		},
		{
			name:        "format string missing captured variables",
			pattern:     "(?P<date>[^-]+-[^-]+-[^-]+-[^-]+-[^-]+-[^-]+)-(?P<room>[^/]+)/.*\\.TRANSCRIPT$",
			format:      "-.TRANSCRIPT",
			input:       "apr-10-2025-17-27-28-AI_TEAM_OFFICE_ROOM-2/audio_transcript.TRANSCRIPT",
			wantErr:     true,
			errContains: "format string does not use any captured variables",
		},
		{
			name:    "extract with different output format",
			pattern: "(?P<date>[^-]+-[^-]+-[^-]+-[^-]+-[^-]+-[^-]+)-[^/]+/(?P<type>[^_]+)_.*\\.TRANSCRIPT$",
			format:  "${type}_${date}.TRANSCRIPT",
			input:   "apr-10-2025-17-27-28-AI_TEAM_OFFICE_ROOM-2/audio_transcript.TRANSCRIPT",
			want:    "audio_apr-10-2025-17-27-28.TRANSCRIPT",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transformer, err := NewPathTransformer(tt.pattern, tt.format)
			if err != nil {
				t.Fatalf("failed to create transformer: %v", err)
			}

			got, err := transformer.Transform(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				} else if tt.errContains != "" && !contains(err.Error(), tt.errContains) {
					t.Errorf("error = %v, want error containing %v", err, tt.errContains)
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("Transform() = %v, want %v", got, tt.want)
			}
		})
	}
}

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
