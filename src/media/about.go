package media

import (
	"github.com/matishsiao/goInfo"
	"github.com/rs/zerolog/log"
	"html/template"
	"media-roller/src/utils"
	"net/http"
	"regexp"
	"strings"
)

var aboutIndexTmpl = template.Must(template.ParseFiles("templates/media/about.html"))

var newlineRegex = regexp.MustCompile("\r?\n")

func AboutIndex(w http.ResponseWriter, _ *http.Request) {
	pythonVersion := utils.RunCommand("python3", "--version")
	if pythonVersion == "" {
		pythonVersion = utils.RunCommand("python", "--version")
	}

	gi, _ := goInfo.GetInfo()

	data := map[string]interface{}{
		"ytDlpVersion":  CachedYtDlpVersion,
		"goVersion":     strings.TrimPrefix(utils.RunCommand("go", "version"), "go version "),
		"pythonVersion": strings.TrimPrefix(pythonVersion, "Python "),
		"ffmpegVersion": newlineRegex.Split(utils.RunCommand("ffmpeg", "-version"), -1),
		"os":            gi.OS,
		"kernel":        gi.Kernel,
		"core":          gi.Core,
		"platform":      gi.Platform,
		"hostname":      gi.Hostname,
		"cpus":          gi.CPUs,
	}

	if err := aboutIndexTmpl.Execute(w, data); err != nil {
		log.Error().Msgf("Error rendering template: %v", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
	}
}
