package hashid

import slug "github.com/goliatone/go-slug"

// Normalizer trims and replaces spaces from the string with the separator:
//  1. Replace unicode chars (by default using charmap.json file)
//  2. remove characters not allowed
//  3. trim leading/trailing spaces
//  4. replaces any redundant whitespaces to single separator chars
//  5. lowercase
func Normalizer(s string) (string, error) {
	return slug.HashNormalize(s)
}

// NormalizerWithSeparator will normalize the string
func NormalizerWithSeparator(s, separator string) (string, error) {
	return slug.HashNormalizeWithSeparator(s, separator)
}

// NormalizerWithCharMap will normalize the string using a custom char map.
func NormalizerWithCharMap(s string, m map[string]string) (string, error) {
	return slug.HashNormalizeWithCharMap(s, m)
}
