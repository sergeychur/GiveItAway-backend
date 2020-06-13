package filesystem

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	clientPrefix = "upload"
)

func DeleteAdPhotos(basePath string, adId int) error {
	path := filepath.Join(basePath, fmt.Sprintf("ad_%d", adId))
	return os.RemoveAll(path)
}

func DeleteAdPhoto(basePath string, photUrl string) error {
	path := strings.Trim(photUrl, clientPrefix)
	return os.Remove(basePath + path)
}
