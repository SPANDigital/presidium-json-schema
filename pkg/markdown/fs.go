package markdown

import (
	"fmt"
	"io/fs"
	"path/filepath"
)

func FindFiles(path string, filter string, recursive bool) ([]string, error) {
	if !recursive {
		return filterFiles(path, filter)
	}

	var files []string
	err := filepath.Walk(path, func(path string, info fs.FileInfo, err error) error {
		if !info.IsDir() {
			return nil
		}

		matches, err := filterFiles(path, filter)
		if err != nil {
			return err
		}

		files = append(files, matches...)
		return nil
	})
	return files, err
}

func filterFiles(path, filter string) ([]string, error) {
	pattern := fmt.Sprintf("%s/%s", path, filter)
	return filepath.Glob(pattern)
}
