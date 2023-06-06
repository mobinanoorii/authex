package helpers

import "strings"

// IsEmpty check if a string is empty
func IsEmpty(s string) bool {
	return strings.TrimSpace(s) == ""
}
