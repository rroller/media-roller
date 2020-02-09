package media

import (
	"github.com/rs/zerolog/log"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

/**
This will serve the fetched files to the client
*/

func ServeMedia(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	log.Info().Msgf("Serving file %s", id)
	if id == "" {
		http.Error(w, "Missing file ID", http.StatusBadRequest)
		return
	} else if !isValidId(id) {
		// Try to parse it just to avoid any type of directory traversal attacks
		http.Error(w, "Invalid file ID", http.StatusBadRequest)
		return
	}

	streamFileToClientById(w, id)
}

func streamFileToClientById(w http.ResponseWriter, id string) {
	filename, err := getFileFromId(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	streamFileToClient(w, filename)
}

func streamFileToClient(writer http.ResponseWriter, filename string) {
	// Check if file exists and open
	Openfile, err := os.Open(filename)
	defer Openfile.Close() //Close after function return
	if err != nil {
		//File not found, send 404
		http.Error(writer, "File not found.", 404)
		return
	}

	// Get the Content-Type of the file
	// Create a buffer to store the header of the file in
	fileHeader := make([]byte, 100)
	//Copy the headers into the FileHeader buffer
	if _, err = Openfile.Read(fileHeader); err != nil {
		log.Error().Msgf("File not found, couldn't open for reading at %s %v", filename, err)
		http.Error(writer, "File not found", 404)
		return
	}

	// Get content type of file
	fileContentType := http.DetectContentType(fileHeader)

	// Get the file size as a string
	fileStat, _ := Openfile.Stat()
	fileSize := strconv.FormatInt(fileStat.Size(), 10)

	// Send the headers
	writer.Header().Set("Content-Disposition", "filename="+filepath.Base(filename))
	writer.Header().Set("Content-Type", fileContentType)
	writer.Header().Set("Content-Length", fileSize)

	// Send the file
	// We read n bytes from the file already, so we reset the offset back to 0
	if _, err = Openfile.Seek(0, 0); err != nil {
		log.Error().Msgf("Error seeking into file %s %v", filename, err)
		http.Error(writer, "File not found", 404)
		return
	}

	// Copy the file to the client
	_, _ = io.Copy(writer, Openfile)
}
