package markdown

import (
	"errors"
	"fmt"
	"github.com/iancoleman/strcase"
	"github.com/santhosh-tekuri/jsonschema/v5"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"net/url"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"text/template"
)

func Slugify(s string) string {
	s = strcase.ToKebab(s)
	var nonWordRe = regexp.MustCompile(`(?m)(\W|_)+`)
	slug := nonWordRe.ReplaceAllString(s, "-")
	return strings.Trim(slug, "-")
}

func Dict(values ...interface{}) (map[string]interface{}, error) {
	if len(values)%2 != 0 {
		return nil, errors.New("invalid dict call")
	}
	dict := make(map[string]interface{}, len(values)/2)
	for i := 0; i < len(values); i += 2 {
		key, ok := values[i].(string)
		if !ok {
			return nil, errors.New("dict keys must be strings")
		}
		dict[key] = values[i+1]
	}
	return dict, nil
}

func FilenameWithoutExt(path string) string {
	return strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
}

func Join(list []string, sep string) string {
	return strings.Join(list, sep)
}

func Title(val string) string {
	return cases.Title(language.English).String(val)
}

func TrimAnchor(path string) string {
	i := strings.Index(path, "#")
	if i < 0 {
		return path
	}
	return path[:i]
}

func Anchor(path string) string {
	i := strings.Index(path, "#")
	if i < 0 || i == len(path)-1 {
		return FilenameWithoutExt(path)
	}
	return path[i+1:]
}

func Anchorize(path string) string {
	return Slugify(Anchor(path))
}

func Humanize(path string) string {
	base := FilenameWithoutExt(path)
	base = strings.TrimSuffix(base, "#")
	p, err := url.PathUnescape(base)
	if err != nil {
		return base
	}
	return p
}

func SchemaFileName(title, location string) string {
	fileName := Slugify(title)
	if len(fileName) == 0 {
		return Slugify(FilenameWithoutExt(location))
	}
	return fileName
}

func SchemaFilePath(location string) string {
	anchor := Anchor(location)
	if len(anchor) == 0 {
		return "/"
	}

	dir, _ := filepath.Split(anchor)
	return dir
}

func IsRemoteRef(location string) bool {
	var re = regexp.MustCompile(`(?m)^http(s)?:\/\/`)
	return re.MatchString(location)
}

func IsInternalRef(location, ref string) bool {
	return location == TrimAnchor(ref)
}

func FirstNonEmpty(values ...string) string {
	for _, value := range values {
		if len(value) > 0 {
			return value
		}
	}
	return ""
}

func IsSlice(v interface{}) bool {
	return reflect.TypeOf(v).Kind() == reflect.Slice
}

func IsSchema(v interface{}) bool {
	switch v.(type) {
	case *jsonschema.Schema:
		return true
	case jsonschema.Schema:
		return true
	default:
		return false
	}
}

func Slice(items ...string) []string {
	var s []string
	s = append(s, items...)
	return s
}

func Append(slice []string, value ...string) []string {
	return append(slice, value...)
}

func EscapeRegex(v *regexp.Regexp) string {
	r := strings.Replace(v.String(), "|", "\\|", -1)
	return fmt.Sprintf("`%s`", r)
}

func FuncMap() template.FuncMap {
	return template.FuncMap{
		"slugify":       Slugify,
		"dict":          Dict,
		"join":          Join,
		"base":          filepath.Base,
		"firstNonEmpty": FirstNonEmpty,
		"isSlice":       IsSlice,
		"isSchema":      IsSchema,
		"escapeRegex":   EscapeRegex,
		"anchorize":     Anchorize,
		"humanize":      Humanize,
		"slice":         Slice,
		"append":        Append,
		"title":         Title,
	}
}
