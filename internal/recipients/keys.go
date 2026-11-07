package recipients

// Contains reports whether key is present in keys.
func Contains(keys []string, key string) bool {
	for _, candidate := range keys {
		if candidate == key {
			return true
		}
	}

	return false
}

// Remove returns a copy of keys without key.
func Remove(keys []string, key string) []string {
	result := make([]string, 0, len(keys))
	for _, candidate := range keys {
		if candidate != key {
			result = append(result, candidate)
		}
	}

	return result
}
