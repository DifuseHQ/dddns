package utils

import "strings"

func BoolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func DomainEndsWith(domain, suffix string) bool {
	return strings.HasSuffix(strings.ToLower(domain), strings.ToLower(suffix))
}
