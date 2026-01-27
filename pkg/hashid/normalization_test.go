package hashid

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNormalizer2(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Basic normalization",
			input:    "Hello World!",
			expected: "hello-world",
		},
		{
			name:     "Unicode replacement",
			input:    "Hellö Wørld!",
			expected: "hello-world",
		},
		{
			name:     "Special characters",
			input:    "special@#-$-%^-&*-chars",
			expected: "special-dollar-percent-and-chars",
		},
		{
			name:     "Multiple spaces",
			input:    "Multiple   Spaces",
			expected: "multiple-spaces",
		},
		{
			name:     "Leading and trailing spaces",
			input:    "  Trim Spaces  ",
			expected: "trim-spaces",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Normalizer(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNormalizerWithSeparator2(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		separator string
		expected  string
	}{
		{
			name:      "Custom separator",
			input:     "Custom Separator",
			separator: "_",
			expected:  "custom_separator",
		},
		{
			name:      "Empty separator",
			input:     "Empty Separator",
			separator: "",
			expected:  "empty-separator",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := NormalizerWithSeparator(tt.input, tt.separator)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}
