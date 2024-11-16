package utils

import "strings"

func NormalizeUrl(url string) string {
	url = strings.TrimSpace(url)
	parts := strings.Split(url, " ")

	for _, part := range parts {
		// Take the firs string that looks like a URL.
		// TODO: We could try to parse the url, but will save that for later
		if strings.HasPrefix(part, "http") || strings.HasPrefix(part, "www") {
			return part
		}
	}

	return url
}
