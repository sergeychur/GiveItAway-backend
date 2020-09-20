package filesystem

import (
	"fmt"
	//_ "github.com/jdeng/goheif"
	_ "golang.org/x/image/webp"
	"image"
	_ "image/gif"
	"image/jpeg"
	_ "image/jpeg"
	_ "image/png"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

func saveFile(img image.Image, handle *multipart.FileHeader, path string, adName string) (string, error) {
	// takes path to directory where to save and adds filename
	filename := getFileName(path, adName, handle.Filename)
	err := CreateDir(path)
	if err != nil {
		return "", err
	}
	f, err := os.Create(filepath.Join(path, filename))
	if err != nil {
		return "", err
	}

	defer func() {
		_ = f.Close()
	}()

	opt := jpeg.Options{
		Quality: 90,
	}

	err = jpeg.Encode(f, img, &opt)
	if err != nil {
		return "", err
	}
	return filename, nil
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
	img, _, err := image.Decode(file)
	if err != nil {
		return "", fmt.Errorf("couldn't detect image format")
	}

	retPath, err := saveFile(img, handle, filepath.Join(basePath, path), path)
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

func getFileName(path string, adName string, initFilename string) string {
	extension := ".jpeg"
	_, errNotFound := os.Stat(filepath.Join(path, adName+extension))
	curFileName := adName + extension
	index := 0
	for !os.IsNotExist(errNotFound) {
		index++
		name := fmt.Sprintf("%s_(%d)", adName, index)
		curFileName = name + extension
		_, errNotFound = os.Stat(filepath.Join(path, curFileName))
	}

	return curFileName
}
