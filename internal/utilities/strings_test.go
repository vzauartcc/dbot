package helpers

import (
	"testing"
)

func TestCapitalizeFirst(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "Empty string",
			input: "",
			want:  "",
		},
		{
			name:  "Lowercase ASCII",
			input: "hello",
			want:  "Hello",
		},
		{
			name:  "Already capitalized ASCII",
			input: "World",
			want:  "World",
		},
		{
			name:  "Single character lowercase",
			input: "a",
			want:  "A",
		},
		{
			name:  "Unicode character (accented)",
			input: "étage",
			want:  "Étage",
		},
		{
			name:  "Non-alphabetic start (number)",
			input: "123go",
			want:  "123go",
		},
		{
			name:  "Non-alphabetic start (symbol)",
			input: "!wow",
			want:  "!wow",
		},
		{
			name:  "Multi-byte emoji (no change expected)",
			input: "🚀launch",
			want:  "🚀launch",
		},
		{
			name:  "String with leading whitespace",
			input: " space",
			want:  " space",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CapitalizeFirst(tt.input)
			if got != tt.want {
				t.Errorf("CapitalizeFirst(%q) = %q; want %q", tt.input, got, tt.want)
			}
		})
	}
}
