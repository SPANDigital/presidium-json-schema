package markdown

import (
	"crypto/md5"
	"errors"
	"fmt"
	"github.com/iancoleman/orderedmap"
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

func Hash(s string) string {
	h := md5.New()
	h.Write([]byte(s))
	return fmt.Sprintf("%x", h.Sum(nil))
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

// TrimAnchorPath returns the path before the anchor (#)
// /a/b/c#d/e => /a/b/c
func TrimAnchorPath(path string) string {
	i := strings.Index(path, "#")
	if i < 0 {
		return path
	}
	return path[:i]
}

// AnchorPath returns the path after the anchor (#)
// /a/b/c#d/e => d/e
func AnchorPath(path string) string {
	i := strings.Index(path, "#")
	if i < 0 {
		return ""
	}
	return path[i+1:]
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

func IsRemoteRef(location string) bool {
	var re = regexp.MustCompile(`(?m)^http(s)?:\/\/`)
	return re.MatchString(location)
}

func IsInternalRef(location, ref string) bool {
	return location == TrimAnchorPath(ref)
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

func EscapeRegex(v string) string {
	r := strings.Replace(v, "|", "\\|", -1)
	return fmt.Sprintf("`%s`", r)
}

func LookupRegex(patterns map[string]string) func(regexp.Regexp) string {
	return func(s regexp.Regexp) string {
		regex, ok := patterns[s.String()]
		if !ok {
			return ""
		}
		return EscapeRegex(regex)
	}
}

func Ref(location string) string {
	return Hash(location)[:10]
}

func FileName(title, location string) string {
	fileName := Slugify(title)
	if len(fileName) == 0 {
		return Slugify(FilenameWithoutExt(location))
	}
	return fileName
}

func FilePath(location string) string {
	i := strings.LastIndex(location, "#")
	if i < 0 {
		return ""
	}

	root := Slugify(FilenameWithoutExt(location[:i]))
	if strings.HasPrefix(location[i:], "#/definitions") {
		return filepath.Join(root, "definitions")
	}
	return root
}

func GetPermalink(ref string) func(schema *jsonschema.Schema) string {
	return func(schema *jsonschema.Schema) string {
		alt := Humanize(schema.Location)
		title := FirstNonEmpty(schema.Title, alt)

		fileName := FileName(schema.Title, schema.Location)
		path := FilePath(schema.Location)
		return fmt.Sprintf("[%s]({{<baseurl>}}/%s/%s/#%s)", title, ref, path, fileName)
	}
}

func FindTypeOfs(s *Schema) []*Schema {
	var schemas []*Schema
	var unique = map[string]bool{}
	s.WalkSchema(false, func(s *Schema) error {
		if unique[s.Location] {
			return nil
		}

		if s.AllOf != nil || s.AnyOf != nil || s.OneOf != nil {
			schemas = append(schemas, s)
			unique[s.Location] = true
		}
		return nil
	})
	return schemas
}

func IndexOf(slice []string, val string) int {
	for i, s := range slice {
		if s == val {
			return i + 1
		}
	}
	return -1
}

// GetWeight returns the schema weight based on it's position in the schema file
func GetWeight(schemaOrders map[string]*orderedmap.OrderedMap) func(path, location string) int {
	return func(path, location string) int {
		order, ok := schemaOrders[path]
		if !ok {
			return -1
		}

		anchor := strings.TrimPrefix(AnchorPath(location), "/")
		keys := strings.Split(anchor, "/")
		for i, key := range keys {
			if len(keys) == i+1 {
				return IndexOf(order.Keys(), key)
			}

			val, exist := order.Get(key)
			if !exist {
				return -1
			}

			switch val.(type) {
			case orderedmap.OrderedMap:
				m := val.(orderedmap.OrderedMap)
				order = &m
			case *orderedmap.OrderedMap:
				order = val.(*orderedmap.OrderedMap)
			}
		}
		return -1
	}
}

func FuncMap(ref string, patterns map[string]string, order map[string]*orderedmap.OrderedMap) template.FuncMap {
	return template.FuncMap{
		"slugify":       Slugify,
		"dict":          Dict,
		"join":          Join,
		"ref":           Ref,
		"base":          filepath.Base,
		"firstNonEmpty": FirstNonEmpty,
		"isSlice":       IsSlice,
		"isSchema":      IsSchema,
		"lookupRegex":   LookupRegex(patterns),
		"permalink":     GetPermalink(ref),
		"weight":        GetWeight(order),
		"findTypeOfs":   FindTypeOfs,
		"humanize":      Humanize,
		"slice":         Slice,
		"append":        Append,
		"title":         Title,
	}
}
