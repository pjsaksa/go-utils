package http

import "strings"

func UrlPartsMatch(url []string, match ...string) bool {
	if len(url) != len(match) {
		return false
	}

	for i := 0; i < len(url); i++ {
		if match[i] == "*" {
			continue
		}
		if url[i] != match[i] {
			return false
		}
	}
	return true
}

// ------------------------------------------------------------

func splitUrlPath(url string) []string {
	if len(url) < 1 || url[0] != '/' {
		return nil
	}
	return strings.Split(url, "/")[1:]
}
