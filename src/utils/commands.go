package utils

import (
	"bytes"
	"github.com/rs/zerolog/log"
	"os"
	"os/exec"
	"strings"
)

func RunCommand(name string, args ...string) string {
	cmd := exec.Command(name, args...)

	var s bytes.Buffer
	cmd.Stdout = &s
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		log.Error().Err(err).Msgf("Error running command " + strings.Join(args, " "))
	}

	return strings.TrimSpace(s.String())
}
