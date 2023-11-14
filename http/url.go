package http

import (
	"fmt"
	go_http "net/http"
	"strings"
)

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

func splitUrlPath(url string) ([]string, Resolution) {
	var urlParts []string
	if len(url) > 0 && url[0] == '/' {
		urlParts = strings.Split(url, "/")[1:]
	}
	if urlParts == nil || len(urlParts) == 0 {
		return nil, &ErrorResolution{
			Status:  go_http.StatusBadRequest,
			Message: fmt.Sprintf(`http.splitUrlPath: Invalid URL '%s'`, url),
		}
	}
	return urlParts, nil
}
