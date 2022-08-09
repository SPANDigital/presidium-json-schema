package markdown

import (
	"fmt"
	"github.com/spf13/afero"
	"gopkg.in/errgo.v2/errors"
	"io/fs"
)

var AppFS = afero.NewOsFs()

func FindFiles(path string, pattern string, recursive bool) ([]string, error) {
	if !recursive {
		return filterFiles(path, pattern)
	}

	var files []string
	err := afero.Walk(AppFS, path, func(path string, info fs.FileInfo, err error) error {
		if !info.IsDir() {
			return nil
		}

		matches, err := filterFiles(path, pattern)
		if err != nil {
			return errors.Wrap(err)
		}

		files = append(files, matches...)
		return nil
	})
	return files, err
}

func filterFiles(path, filter string) ([]string, error) {
	pattern := fmt.Sprintf("%s/%s", path, filter)
	return afero.Glob(AppFS, pattern)
}
