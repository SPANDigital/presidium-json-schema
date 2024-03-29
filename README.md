# Presidium JSON Schema

A Golang tool for importing your [JSON Schema](http://json-schema.org/) spec into
[Presidium](http://presidium.spandigital.net) documentation.

```text
Usage:
  presidium-json-schema convert [path] [flags]

Flags:
  -d, --destination string   the output directory (default ".")
  -e, --extension string     the schema extension (default "*.schema.json")
  -o, --ordered              preserve the schema order (defaults to alphabetical)
  -w, --walk                 walk through sub-directories
```

To convert a file you simply:

```shell
presidium-json-schema convert <PATH_TO_SCHEMA_DIR> -d <THE_DESTINATION_DIR>
```
