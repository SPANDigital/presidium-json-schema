package markdown

import "strings"

type Config struct {
	Destination     string
	Extension       string
	Recursive       bool
	Ordered         bool
	OrderedFilePath bool
	Local           bool
	Clean           bool
}

func (c Config) ReferenceUrl() string {
	prefix := "content/"
	i := strings.Index(c.Destination, prefix)
	offset := i + len(prefix)
	if i < 0 || offset > len(c.Destination) {
		return "reference"
	}
	return c.Destination[offset:]
}
