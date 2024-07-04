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
  -p, --orderedfilepath      preserve the schema order (defaults to alphabetical) by appending a digit to the filename prefix
  -c, --clean                removes the output directory before generating output files, negative by default
  -w, --walk                 walk through sub-directories
```

To convert a file you simply:

```shell
presidium-json-schema convert <PATH_TO_SCHEMA_DIR> -d <THE_DESTINATION_DIR>
```

### Releasing a new version

This project uses [GoReleaser](https://goreleaser.com/) to automate the release process. When you push a new tag to the repository, GoReleaser will create a new release with the artifacts for the supported platforms and publish it to the [Span Homebrew tap](https://github.com/SPANDigital/homebrew-tap).

To release a new version, you need to create a new tag and push it to the repository. The version number should follow the [Semantic Versioning](https://semver.org/) specification.

```shell
git tag -a v1.0.0 -m "Release version 1.0.0"
git push origin v1.0.0
```

Once a release is published, you can install the new version or upgrade an existing installation using Homebrew.

To install the latest version:

```shell
brew tap SPANDigital/homebrew-tap https://github.com/SPANDigital/homebrew-tap.git
brew install presidium-json-schema
```


To upgrade an existing installation:

```shell
brew upgrade presidium-json-schema
```

