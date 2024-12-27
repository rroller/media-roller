package media

import (
	"crypto/md5"
	"errors"
	"fmt"
	"github.com/dustin/go-humanize"
	"golang.org/x/sync/errgroup"
	"html/template"
	"media-roller/src/utils"
	"net/http"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

/**
This file will download the media from a URL and save it to disk.
*/

import (
	"bytes"
	"github.com/rs/zerolog/log"
	"io"
	"os"
	"os/exec"
)

type Media struct {
	Id          string
	Name        string
	SizeInBytes int64
	HumanSize   string
}

var fetchIndexTmpl = template.Must(template.ParseFiles("templates/media/index.html"))

// Where the media files are saved. Always has a trailing slash
var downloadDir = getDownloadDir()
var idCharSet = regexp.MustCompile(`^[a-zA-Z0-9]+$`).MatchString

func Index(w http.ResponseWriter, _ *http.Request) {
	data := map[string]string{
		"ytDlpVersion": CachedYtDlpVersion,
	}
	if err := fetchIndexTmpl.Execute(w, data); err != nil {
		log.Error().Msgf("Error rendering template: %v", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
	}
}

func FetchMedia(w http.ResponseWriter, r *http.Request) {
	url, args := getUrl(r)

	media, ytdlpErrorMessage, err := getMediaResults(url, args)
	data := map[string]interface{}{
		"url":          url,
		"media":        media,
		"error":        ytdlpErrorMessage,
		"ytDlpVersion": CachedYtDlpVersion,
	}
	if err != nil {
		_ = fetchIndexTmpl.Execute(w, data)
		return
	}

	if err = fetchIndexTmpl.Execute(w, data); err != nil {
		log.Error().Msgf("Error rendering template: %v", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
	}
}

func FetchMediaApi(w http.ResponseWriter, r *http.Request) {
	url, args := getUrl(r)
	medias, _, err := getMediaResults(url, args)
	if err != nil {
		log.Error().Msgf("error getting media results: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if len(medias) == 0 {
		log.Error().Msgf("not media found")
		http.Error(w, "Media not found", http.StatusBadRequest)
		return
	}

	// just take the first one
	streamFileToClientById(w, r, medias[0].Id)
}

func getUrl(r *http.Request) (string, map[string]string) {
	u := strings.TrimSpace(r.URL.Query().Get("url"))

	// Support yt-dlp arguments passed in via the url. We'll assume anything starting with a dash - is an argument
	args := make(map[string]string)
	for k, v := range r.URL.Query() {
		if strings.HasPrefix(k, "-") {
			if len(v) > 0 {
				args[k] = v[0]
			} else {
				args[k] = ""
			}
		}
	}

	return u, args
}

func getMediaResults(inputUrl string, args map[string]string) ([]Media, string, error) {
	if inputUrl == "" {
		return nil, "", errors.New("missing URL")
	}

	url := utils.NormalizeUrl(inputUrl)
	log.Info().Msgf("Got input '%s' and extracted '%s' with args %v", inputUrl, url, args)

	// NOTE: This system is for a simple use case, meant to run at home. This is not a great design for a robust system.
	// We are hashing the URL here and writing files to disk to a consistent directory based on the ID. You can imagine
	// concurrent users would break this for the same URL. That's fine given this is for a simple home system.
	// Future work can make this more sophisticated.
	id := GetMD5Hash(url, args)
	// Look to see if we already have the media on disk
	medias, err := getAllFilesForId(id)
	if err != nil {
		return nil, "", err
	}
	if len(medias) == 0 {
		// We don't, so go fetch it
		errMessage := ""
		id, errMessage, err = downloadMedia(url, args)
		if err != nil {
			return nil, errMessage, err
		}
		medias, err = getAllFilesForId(id)
		if err != nil {
			return nil, "", err
		}
	}

	return medias, "", nil
}

// returns the ID of the file, and error message, and an error
func downloadMedia(url string, requestArgs map[string]string) (string, string, error) {
	// The id will be used as the name of the parent directory of the output files
	id := GetMD5Hash(url, requestArgs)
	name := getMediaDirectory(id) + "%(id)s.%(ext)s"

	log.Info().Msgf("Downloading %s to %s", url, name)

	defaultArgs := map[string]string{
		"--format":              "bestvideo[ext=mp4]+bestaudio[ext=m4a]/best[ext=mp4]/best",
		"--merge-output-format": "mp4",
		"--trim-filenames":      "100",
		"--recode-video":        "mp4",
		"--restrict-filenames":  "",
		"--write-info-json":     "",
		"--verbose":             "",
		"--output":              name,
	}

	args := make([]string, 0)

	// First add all default arguments that were not supplied as request level arguments
	for arg, value := range defaultArgs {
		if _, has := requestArgs[arg]; !has {
			args = append(args, arg)
			if value != "" {
				args = append(args, value)
			}
		}
	}

	// Now add all request level arguments
	for arg, value := range requestArgs {
		args = append(args, arg)
		if value != "" {
			args = append(args, value)
		}
	}

	// And finally add any environment level arguments not supplied as request level args
	for arg, value := range getEnvVars() {
		if _, has := requestArgs[arg]; !has {
			args = append(args, arg)
			if value != "" {
				args = append(args, value)
			}
		}
	}

	args = append(args, url)

	cmd := exec.Command("yt-dlp", args...)

	var stdoutBuf, stderrBuf bytes.Buffer
	stdoutIn, _ := cmd.StdoutPipe()
	stderrIn, _ := cmd.StderrPipe()

	var errStdout, errStderr error
	stdout := io.MultiWriter(os.Stdout, &stdoutBuf)
	stderr := io.MultiWriter(os.Stderr, &stderrBuf)

	err := cmd.Start()
	if err != nil {
		log.Error().Msgf("Error starting command: %v", err)
		return "", err.Error(), err
	}

	eg := errgroup.Group{}

	eg.Go(func() error {
		_, errStdout = io.Copy(stdout, stdoutIn)
		return nil
	})

	_, errStderr = io.Copy(stderr, stderrIn)
	_ = eg.Wait()
	log.Info().Msgf("Done with %s", id)

	err = cmd.Wait()
	if err != nil {
		log.Error().Err(err).Msgf("cmd.Run() failed with %s", err)
		return "", strings.TrimSpace(stderrBuf.String()), err
	} else if errStdout != nil {
		log.Error().Msgf("failed to capture stdout: %v", errStdout)
	} else if errStderr != nil {
		log.Error().Msgf("failed to capture stderr: %v", errStderr)
	}

	return id, "", nil
}

// Returns the relative directory containing the media file, with a trailing slash.
// Id is expected to be pre validated
func getMediaDirectory(id string) string {
	return downloadDir + id + "/"
}

// id is expected to be validated prior to calling this func
func getAllFilesForId(id string) ([]Media, error) {
	root := getMediaDirectory(id)
	file, err := os.Open(root)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	files, _ := file.Readdirnames(0) // 0 to read all files and folders
	if len(files) == 0 {
		return nil, errors.New("ID not found: " + id)
	}

	var medias []Media

	// We expect two files to be produced for each video, a json manifest and an mp4.
	for _, f := range files {
		if !strings.HasSuffix(f, ".json") {
			fi, err2 := os.Stat(root + f)
			var size int64 = 0
			if err2 == nil {
				size = fi.Size()
			}

			media := Media{
				Id:          id,
				Name:        filepath.Base(f),
				SizeInBytes: size,
				HumanSize:   humanize.Bytes(uint64(size)),
			}
			medias = append(medias, media)
		}
	}

	return medias, nil
}

// id is expected to be validated prior to calling this func
// TODO: This needs to handle multiple files in the directory
func getFileFromId(id string) (string, error) {
	root := getMediaDirectory(id)
	file, err := os.Open(root)
	if err != nil {
		return "", err
	}
	files, _ := file.Readdirnames(0) // 0 to read all files and folders
	if len(files) == 0 {
		return "", errors.New("ID not found")
	}

	// We expect two files to be produced, a json manifest and an mp4. We want to return the mp4
	// Sometimes the video file might not have an mp4 extension, so filter out the json file
	for _, f := range files {
		if !strings.HasSuffix(f, ".json") {
			// TODO: This is just returning the first file found. We need to handle multiple
			return root + f, nil
		}
	}

	return "", errors.New("unable to find file")
}

func GetMD5Hash(url string, args map[string]string) string {
	id := url
	if len(args) > 0 {
		tmp := make([]string, 0)
		for k, v := range args {
			tmp = append(tmp, k, v)
		}
		sort.Strings(tmp)
		id += ":" + strings.Join(tmp, ",")
	}
	return fmt.Sprintf("%x", md5.Sum([]byte(id)))
}

func isValidId(id string) bool {
	return idCharSet(id)
}

func getDownloadDir() string {
	dir := os.Getenv("MR_DOWNLOAD_DIR")
	if dir != "" {
		if !strings.HasSuffix(dir, "/") {
			return dir + "/"
		}
		return dir
	}
	return "downloads/"
}

func getEnvVars() map[string]string {
	vars := make(map[string]string)
	if ev := strings.TrimSpace(os.Getenv("MR_PROXY")); ev != "" {
		vars["--proxy"] = ev
	}
	return vars
}
