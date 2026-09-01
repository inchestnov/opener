package opener

import (
	"slices"
	"testing"
)

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
			if !slices.Equal(got, tt.want) {
				t.Errorf("splitCommand(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}
