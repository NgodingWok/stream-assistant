package youtube

import "testing"

func TestVideoIDPattern(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		wantID string
		wantOK bool
	}{
		{
			name:   "standard video URL",
			input:  `"url":"watch?v=dQw4w9WgXcQ"`,
			wantID: "dQw4w9WgXcQ",
			wantOK: true,
		},
		{
			name:   "video ID with hyphens and underscores",
			input:  `watch?v=Ab_C-d1E2F3`,
			wantID: "Ab_C-d1E2F3",
			wantOK: true,
		},
		{
			name:   "no video ID present",
			input:  `<html>no results here</html>`,
			wantID: "",
			wantOK: false,
		},
		{
			name:   "short invalid ID",
			input:  `watch?v=short`,
			wantID: "",
			wantOK: false,
		},
		{
			name:   "multiple video IDs returns first",
			input:  `watch?v=AAAAAAAAAAA watch?v=BBBBBBBBBBB`,
			wantID: "AAAAAAAAAAA",
			wantOK: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches := videoIDPattern.FindSubmatch([]byte(tt.input))
			if tt.wantOK {
				if len(matches) < 2 {
					t.Fatalf("expected match, got none")
				}
				got := string(matches[1])
				if got != tt.wantID {
					t.Errorf("expected %q, got %q", tt.wantID, got)
				}
			} else {
				if len(matches) >= 2 {
					t.Errorf("expected no match, got %q", string(matches[1]))
				}
			}
		})
	}
}
