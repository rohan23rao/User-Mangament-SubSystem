// internal/utils/helpers.go
package utils

import "strings"

func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func MaskPassword(url string) string {
	return strings.ReplaceAll(url, "userms_password", "***")
}
