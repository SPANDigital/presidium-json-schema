package markdown

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestReferenceURL(t *testing.T) {
	testCases := map[string]string{
		"./content/reference":        "reference",
		"not-valid":                  "reference",
		"./content/reference/sample": "reference/sample",
		"content":                    "reference",
	}

	for val, expected := range testCases {
		c := Config{Destination: val}
		actual := c.ReferenceUrl()
		assert.Equal(t, expected, actual)
	}
}
