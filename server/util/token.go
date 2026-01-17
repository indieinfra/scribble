package util

// PopAccessToken extracts the first string access_token value from a map and removes the key.
// Returns an empty string if not present or not a string.
func PopAccessToken(values map[string]any) string {
	if values == nil {
		return ""
	}

	v, ok := values["access_token"]
	if !ok {
		return ""
	}

	delete(values, "access_token")

	switch t := v.(type) {
	case string:
		return t
	case []any:
		for _, e := range t {
			if s, ok := e.(string); ok && s != "" {
				return s
			}
		}
	}

	return ""
}
