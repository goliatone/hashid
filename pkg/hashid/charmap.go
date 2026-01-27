package hashid

import slug "github.com/goliatone/go-slug"

func GetCharMap() (map[string]string, error) {
	return slug.GetCharMap()
}

func SetCharMap(mapping map[string]string) {
	slug.SetCharMap(mapping)
}

func ResetCharMap() error {
	return slug.ResetCharMap()
}
