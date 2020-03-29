package filesystem

import (
	"fmt"
	"os"
	"path/filepath"
)

func DeleteAdPhotos(basePath string, adId int) error {
	path := filepath.Join(basePath, fmt.Sprintf("ad_%d", adId))
	return os.RemoveAll(path)
}
