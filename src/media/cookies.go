package media

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func configuredCookieFile(requestArgs map[string]string) (string, error) {
	// Preserve explicit yt-dlp cookie options supplied by API clients.
	for _, option := range []string{"--cookies", "--no-cookies", "--cookies-from-browser"} {
		if _, ok := requestArgs[option]; ok {
			return "", nil
		}
	}

	dir := strings.TrimSpace(os.Getenv("MR_COOKIES_DIR"))
	if dir == "" {
		dir = "cookies"
	}
	filename, err := filepath.Abs(filepath.Join(dir, "cookies.txt"))
	if err != nil {
		return "", err
	}
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	if !info.Mode().IsRegular() {
		return "", fmt.Errorf("cookie file is not a regular file: %s", filename)
	}
	return filename, nil
}
