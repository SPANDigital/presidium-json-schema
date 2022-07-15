package markdown

type Config struct {
	Destination  string
	ReferenceUrl string
	Extension    string
	Recursive    bool
	Local        bool
}

type RawSchema map[string]interface{}

func (r RawSchema) Id() string {
	id := r["$id"]
	if _, ok := id.(string); ok {
		return id.(string)
	}
	return ""
}
