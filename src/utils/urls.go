package utils

import (
	"bufio"
	"strings"
)

func NormalizeUrl(url string) string {
	url = strings.TrimSpace(url)
	parts := strings.Split(url, " ")

	// Find the first URL. Will split the string by spaces and new lines and return the first thing that looks like a URL
	// TODO: We could try to parse the url, but will save that for later
	for _, part := range parts {
		// Take the firs string that looks like a URL.
		sc := bufio.NewScanner(strings.NewReader(part))
		for sc.Scan() {
			p := sc.Text()
			if strings.HasPrefix(p, "http") || strings.HasPrefix(p, "www") {
				return p
			}
		}
	}

	return url
}
