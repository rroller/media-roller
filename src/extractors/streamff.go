package extractors

import (
	"regexp"
)

// https://streamff.com/v/e70b90d8
var streamffRe = regexp.MustCompile(`^(?:https?://)?(?:www)?\.?streamff\.com/v/([A-Za-z0-9]+)/?`)

func GetUrl(url string) string {
	if matches := streamffRe.FindStringSubmatch(url); len(matches) == 2 {
		return "https://ffedge.streamff.com/uploads/" + matches[1] + ".mp4"
	}
	return ""
}
