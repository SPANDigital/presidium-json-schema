package markdown

type Config struct {
	Source      string
	Destination string
	Extension   string
	Recursive   bool
	Local       bool
}

type SchemaHeader struct {
	Id string `json:"$id"`
}
