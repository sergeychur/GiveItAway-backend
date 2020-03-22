package filesystem

import (
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

func saveFile(file multipart.File, handle *multipart.FileHeader, path string) (string, error) {
	// takes path to directory where to save and adds filename
	data, err := ioutil.ReadAll(file)
	if err != nil {
		return "", err
	}
	err = CreateDir(path)
	if err != nil {
		return "", err
	}
	err = ioutil.WriteFile(filepath.Join(path, handle.Filename), data, 0644)
	if err != nil {
		return "", err
	}
	return handle.Filename, nil
}

func UploadFile(w http.ResponseWriter, r *http.Request, callback func(header multipart.FileHeader) error,
	basePath string, path string) (string, error) {
		// takes base path to all uploads and directory to where save the current file
	file, handle, err := r.FormFile("file")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return "", err
	}
	err = callback(*handle)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return "", err
	}

	defer func() {
		_ = file.Close()
	}()

	retPath, err := saveFile(file, handle, filepath.Join(basePath, path))
	if err != nil {
		return "", err
	}
	return filepath.Join("upload/", path, retPath), nil
}

func CreateDir(path string) error {
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		return nil
	}
	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}