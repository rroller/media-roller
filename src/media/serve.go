package media

import (
	"github.com/rs/zerolog/log"
	"net/http"
)

/**
This will serve the fetched files to the client
*/

import (
	"errors"
	"github.com/google/uuid"
	"io"
	"os"
	"strconv"
)

func ServeMedia(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	log.Info().Msgf("Serving file %s", id)
	if id == "" {
		http.Error(w, "Missing file ID", http.StatusBadRequest)
		return
	} else if _, err := uuid.Parse(id); err != nil {
		// Try to parse it just to avoid any type of directory traversal attacks
		http.Error(w, "Invalid file ID", http.StatusBadRequest)
		return
	}

	filename, err := getFileFromId(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	streamFileToClient(w, filename)
}

// id is expected to be validated prior to calling this func
func getFileFromId(id string) (string, error) {
	root := getMediaDirectory(id)
	file, err := os.Open(root)
	if err != nil {
		return "", err
	}
	files, _ := file.Readdirnames(0) // 0 to read all files and folders
	if len(files) == 0 {
		return "", errors.New("ID not found")
	} else if len(files) > 1 {
		// We should only have 1 media file produced
		return "", errors.New("internal error")
	}

	return root + files[0], nil
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
	// Set the following if you want to force the client to download the file
	// writer.Header().Set("Content-Disposition", "attachment; filename="+filepath.Base(filename))
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
	if _, err = io.Copy(writer, Openfile); err != nil {
		log.Error().Msgf("Error copying file %s %v", filename, err)
		http.Error(writer, "Couldn't copy file", 404)
		return
	}
}
