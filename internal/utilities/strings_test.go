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

func TestTitleCase(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Capitalize first word even if it's a minor word",
			input:    "of mice and men",
			expected: "Of Mice and Men",
		},
		{
			name:     "Lowercase minor words in the middle",
			input:    "The Lord OF the Rings",
			expected: "The Lord of the Rings",
		},
		{
			name:     "Normal sentence",
			input:    "the quick brown fox jumps",
			expected: "The Quick Brown Fox Jumps",
		},
		{
			name:     "Handle multiple internal spaces",
			input:    "the  wide   road",
			expected: "The Wide Road",
		},
		{
			name:     "Handle leading and trailing spaces",
			input:    "  trim me  ",
			expected: "Trim Me",
		},
		{
			name:     "Handle tabs and newlines",
			input:    "text\twith\nwhitespace",
			expected: "Text with Whitespace",
		},
		{
			name:     "Preserve mid-word capitalization (McAfee)",
			input:    "the story of McAfee",
			expected: "The Story of McAfee",
		},
		{
			name:     "Preserve abbreviations (FBI)",
			input:    "a report for the FBI",
			expected: "A Report for the FBI",
		},
		{
			name:     "Handle already correct CamelCase at start",
			input:    "iPhone and iPad",
			expected: "IPhone and IPad",
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "Single minor word",
			input:    "and",
			expected: "And",
		},
		{
			name:     "Single character words",
			input:    "a b c",
			expected: "A B C",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := TitleCase(tt.input)
			if actual != tt.expected {
				t.Errorf("TitleCase(%q) = %q; want %q", tt.input, actual, tt.expected)
			}
		})
	}
}
