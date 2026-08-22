package opener

import "testing"

func TestMatchPattern(t *testing.T) {
	tests := []struct {
		name     string
		pattern  string
		filename string
		want     bool
	}{
		{"glob matches", "*.pdf", "document.pdf", true},
		{"glob case-insensitive", "*.pdf", "document.PDF", true},
		{"glob no match", "*.pdf", "document.png", false},
		{"bare extension treated as glob", ".pdf", "document.pdf", true},
		{"bare extension no match", ".pdf", "document.png", false},
		{"literal filename match", "document.pdf", "document.pdf", true},
		{"literal filename no match", "document.pdf", "other.pdf", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := matchPattern(tt.pattern, tt.filename)
			if err != nil {
				t.Fatalf("matchPattern() error = %v, want nil", err)
			}
			if got != tt.want {
				t.Errorf("matchPattern(%q, %q) = %v, want %v", tt.pattern, tt.filename, got, tt.want)
			}
		})
	}
}

func TestSplitCommand(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    []string
		wantErr bool
	}{
		{"single word", "nvim", []string{"nvim"}, false},
		{"multiple words", "open -a Safari", []string{"open", "-a", "Safari"}, false},
		{"single-quoted argument with space", `open -a 'Google Chrome'`, []string{"open", "-a", "Google Chrome"}, false},
		{"double-quoted argument with space", `open -a "Google Chrome"`, []string{"open", "-a", "Google Chrome"}, false},
		{"unterminated quote is an error", `open -a 'Google Chrome`, nil, true},
		{"empty string", "", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := splitCommand(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatal("splitCommand() error = nil, want error")
				}
				return
			}
			if err != nil {
				t.Fatalf("splitCommand() error = %v, want nil", err)
			}
			if !slicesEqual(got, tt.want) {
				t.Errorf("splitCommand(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}
