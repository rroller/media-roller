package media

import (
	"bytes"
	"github.com/rs/zerolog/log"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
)

func UpdateYtDlp() (string, error) {
	log.Info().Msgf("Updateing yt-dlp")

	cmd := exec.Command("yt-dlp",
		"--update",
		"--update-to", "nightly",
	)

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

	err = cmd.Wait()
	if err != nil {
		log.Error().Msgf("cmd.Run() failed with %s", err)
		return "", err
	} else if errStdout != nil {
		log.Error().Msgf("failed to capture stdout: %v", errStdout)
	} else if errStderr != nil {
		log.Error().Msgf("failed to capture stderr: %v", errStderr)
	}
	log.Info().Msgf("Done updating yt-dlp")

	return "", nil
}

func GetInstalledVersion() string {
	cmd := exec.Command("yt-dlp", "--version")

	var s bytes.Buffer
	cmd.Stdout = &s
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		log.Error().Err(err).Msgf("Error getting installed version")
	}

	version := strings.TrimSpace(string(s.Bytes()))
	if version != "" {
		return version
	}
	return "unknown"
}
