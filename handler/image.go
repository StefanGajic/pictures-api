package handler

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"sort"
	"strings"
)

const (
	uploadDir  = "./uploads/"
	jpgFormat  = "jpg"
	jpegFormat = "jpeg"
	pngFormat  = "png"
	svgFormat  = "svg"
)

type FileInfo struct {
	Name string `json:"name"`
}

// UploadFile hash image name and uploads image in uploads folder.
func UploadFile(w http.ResponseWriter, r *http.Request) *HTTPError {
	err := r.ParseMultipartForm(32 << 20)
	if err != nil {
		return newHTTPError(http.StatusBadRequest, "Error parsing form")
	}

	file, header, err := r.FormFile("image")
	if err != nil {
		return newHTTPError(http.StatusBadRequest, "error getting file")
	}
	defer file.Close()

	splitFileName := strings.Split(header.Filename, ".")
	if len(splitFileName) < 2 {
		return newHTTPError(http.StatusBadRequest, "error spliting file name")
	}

	err = validateImg(splitFileName[1])
	if err != nil {
		return newHTTPError(http.StatusBadRequest, "file is not image type")
	}

	h := sha256.New()
	h.Write([]byte(splitFileName[0]))
	hashImageName := hex.EncodeToString(h.Sum(nil)) + "." + splitFileName[1]

	if _, err := os.Stat(uploadDir + hashImageName); err == nil {
		return newHTTPError(http.StatusBadRequest, "error image already exist")
	}

	err = createDirAndSave(file, uploadDir, hashImageName)
	if err != nil {
		return newHTTPError(http.StatusInternalServerError, "error making directory or saving file")
	}

	w.WriteHeader(http.StatusCreated)

	return nil
}

// ListFiles list all of the images in ascending order.
func ListFiles(w http.ResponseWriter, r *http.Request) *HTTPError {
	w.Header().Set("Content-Type", "application/json")

	dir, err := os.Open("uploads")
	if err != nil {
		return newHTTPError(http.StatusNotFound, err)
	}
	entries, err := dir.Readdir(0)
	if err != nil {
		return newHTTPError(http.StatusNotFound, err)
	}

	list := []FileInfo{}

	for _, entry := range entries {
		f := FileInfo{
			Name: entry.Name(),
		}
		list = append(list, f)
	}

	sort.Slice(list, func(i, j int) bool {
		return list[i].Name < list[j].Name
	})

	output, err := json.Marshal(list)
	if err != nil {
		return newHTTPError(http.StatusInternalServerError, err)
	}

	log.Println(string(output))

	w.Write(output)
	return nil
}

// DownloadFile downloads image selected by name.
func DownloadFile(w http.ResponseWriter, r *http.Request) *HTTPError {
	imageName := r.URL.Query().Get("name")
	destinationPath := "./downloads/"
	filePath := uploadDir + imageName
	downloadPath := destinationPath + imageName

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return newHTTPError(http.StatusNotFound, "error file path does not exist")
	}

	file, err := os.Open(filePath)
	if err != nil {
		return newHTTPError(http.StatusInternalServerError, "error file path")
	}
	defer file.Close()

	if _, err := os.Stat(downloadPath); err == nil {
		return newHTTPError(http.StatusInternalServerError, "error downloaded path")

	}

	err = os.MkdirAll(destinationPath, 0755)
	if err != nil {
		return newHTTPError(http.StatusInternalServerError, "error making directory")
	}

	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		return newHTTPError(http.StatusInternalServerError, "error reading bytes")
	}
	err = ioutil.WriteFile(downloadPath, fileBytes, 0644)
	if err != nil {
		return newHTTPError(http.StatusInternalServerError, "error write file")
	}
	w.WriteHeader(http.StatusOK)
	return nil
}

// DeleteFile deletes selected image by name.
func DeleteFile(w http.ResponseWriter, r *http.Request) *HTTPError {
	fileName := r.URL.Query().Get("name")
	filePath := uploadDir + fileName

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		w.WriteHeader(http.StatusNotFound)
		return newHTTPError(http.StatusNotFound, "error file not found")
	}

	if err := os.Remove(filePath); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return newHTTPError(http.StatusInternalServerError, "error failed to remove file")
	}
	w.WriteHeader(http.StatusNoContent)

	return nil
}

func validateImg(imageExt string) error {
	var ok bool
	switch imageExt {
	case jpgFormat, jpegFormat, pngFormat, svgFormat:
		ok = true
	default:
		ok = false
	}

	if !ok {
		return newHTTPError(http.StatusBadRequest, "invalid file extension")
	}

	return nil
}

func createDirAndSave(file multipart.File, directory, name string) error {
	err := os.MkdirAll(directory, 0755)
	if err != nil {
		return err
	}

	filePath := directory + name
	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(filePath, fileBytes, 0644)
	if err != nil {
		return err
	}

	return nil
}
