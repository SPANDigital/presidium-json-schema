package markdown

import (
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestFindFiles(t *testing.T) {
	AppFS = afero.NewMemMapFs()
	create(t, "a.test")
	create(t, "a/b.test")
	create(t, "c.test")
	create(t, "d.json")
	create(t, "e.test.json")

	files, err := FindFiles(".", "*.test", false)
	assert.Nil(t, err)
	assert.ElementsMatch(t, []string{"a.test", "c.test"}, files)
}

func TestFindFilesRecursive(t *testing.T) {
	AppFS = afero.NewMemMapFs()
	create(t, "a.test")
	create(t, "a/b.test")
	create(t, "a/b/c.test")
	create(t, "d.test")
	create(t, "e.test.json")

	files, err := FindFiles(".", "*.test", true)
	assert.Nil(t, err)
	assert.ElementsMatch(t, []string{"a.test", "a/b.test", "a/b/c.test", "d.test"}, files)
}

func create(t *testing.T, path string) {
	err := afero.WriteFile(AppFS, path, []byte{}, os.ModePerm)
	assert.Nil(t, err)
}
