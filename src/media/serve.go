package media

import (
	"github.com/rs/zerolog/log"
	"net/http"
	"os"
	"path/filepath"
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

	streamFileToClientById(w, r, id)
}

func streamFileToClientById(w http.ResponseWriter, r *http.Request, id string) {
	filename, err := getFileFromId(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	streamFileToClient(w, r, filename)
}

func streamFileToClient(w http.ResponseWriter, r *http.Request, filename string) {
	// Check if file exists and open
	openfile, err := os.Open(filename)
	if err != nil {
		//File not found, send 404
		http.Error(w, "File not found.", 404)
		return
	}
	defer openfile.Close()

	// Get the Content-Type of the file
	// Create a buffer to store the header of the file in
	fileHeader := make([]byte, 100)
	//Copy the headers into the FileHeader buffer
	if _, err = openfile.Read(fileHeader); err != nil {
		log.Error().Msgf("File not found, couldn't open for reading at %s %v", filename, err)
		http.Error(w, "File not found", 404)
		return
	}

	// Get content type of file
	fileContentType := http.DetectContentType(fileHeader)

	// Send the headers
	w.Header().Set("Content-Disposition", "filename="+filepath.Base(filename))
	w.Header().Set("Content-Type", fileContentType)

	// Send the file
	// We read n bytes from the file already, so we reset the offset back to 0
	if _, err = openfile.Seek(0, 0); err != nil {
		log.Error().Msgf("Error seeking into file %s %v", filename, err)
		http.Error(w, "File not found", 404)
		return
	}
	http.ServeFile(w, r, filename)
}
