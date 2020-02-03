package media

import (
	"html/template"
	"net/http"
)

/**
This file will download the media from a URL and save it to disk.
*/

import (
	"bytes"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"io"
	"os"
	"os/exec"
	"sync"
)

const downloadDir = "downloads/"

type ResponseData struct {
	Id string
}

var fetchResponseTmpl = template.Must(template.ParseFiles("templates/media/response.html"))
var fetchIndexTmpl = template.Must(template.ParseFiles("templates/media/index.html"))

func Index(w http.ResponseWriter, _ *http.Request) {
	if err := fetchIndexTmpl.Execute(w, nil); err != nil {
		log.Error().Msgf("Error rendering template: %v", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
	}
}

func FetchMedia(w http.ResponseWriter, r *http.Request) {
	url := r.URL.Query().Get("url")
	if url == "" {
		http.Error(w, "Missing URL", http.StatusBadRequest)
		return
	}

	id, err := fetch(url)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	data := ResponseData{
		Id: id,
	}
	if err := fetchResponseTmpl.Execute(w, data); err != nil {
		log.Error().Msgf("Error rendering template: %v", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
	}
}

// returns the ID of the file
func fetch(url string) (string, error) {
	// This will be the output file name
	id := uuid.New().String()
	// youtube-dl will add the extension as needed
	name := getFilenameWithoutExtensionById(id)

	log.Info().Msgf("Downloading %s to %s", url, id)

	cmd := exec.Command("youtube-dl", "-f", "bestvideo[ext=mp4]+bestaudio[ext=m4a]/mp4/", "-o", name, url)

	var stdoutBuf, stderrBuf bytes.Buffer
	stdoutIn, _ := cmd.StdoutPipe()
	stderrIn, _ := cmd.StderrPipe()

	var errStdout, errStderr error
	stdout := io.MultiWriter(os.Stdout, &stdoutBuf)
	stderr := io.MultiWriter(os.Stderr, &stderrBuf)

	err := cmd.Start()
	if err != nil {
		log.Error().Msgf("Error starting command: %v", err)
		return "", err
	}

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		_, errStdout = io.Copy(stdout, stdoutIn)
		wg.Done()
	}()

	_, errStderr = io.Copy(stderr, stderrIn)
	wg.Wait()
	log.Info().Msgf("Done with %s", id)

	err = cmd.Wait()
	if err != nil {
		log.Error().Msgf("cmd.Run() failed with %s", err)
		return "", err
	} else if errStdout != nil {
		log.Error().Msgf("failed to capture stdout: %v", errStdout)
	} else if errStderr != nil {
		log.Error().Msgf("failed to capture stderr: %v", errStderr)
	}

	return id, nil
}

// Returns the relative filename without the extension. Example:
// downloads/b541cc43-9833-4146-ab19-71334484c0c1/media
// where media can be media.mp4
func getFilenameWithoutExtensionById(id string) string {
	return getMediaDirectory(id) + "media"
}

// Returns the relative directory containing the media file, with a trailing slash
// Id is expected to be pre validated
func getMediaDirectory(id string) string {
	return downloadDir + id + "/"
}
