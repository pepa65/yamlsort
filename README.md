# yamlsort 0.1.3

## Sort yaml-file recursively
* Repo: [github.com/pepa65/yamlsort](https://github.com/pepa65/yamlsort)
* After: [github.com/zwopir/yaml-sort](https://github.com/zwopir/yaml-sort)
* Contact: github.com/pepa65
* Install: `wget -qO- gobinaries.com/pepa65/yamlsort |sh`

## Build
```shell
# While in the repo root directory:
go build

# Or anywhere:
go install github.com/pepa65/yamlsort@latest

# Smaller binary:
go build -ldflags="-s -w"

# More extreme shrinking:
upx --best --lzma yamlsort

# Build for various architectures:
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o yamlsort
CGO_ENABLED=0 GOOS=linux GOARCH=arm go build -ldflags="-s -w" -o yamlsort_pi
CGO_ENABLED=0 GOOS=freebsd GOARCH=amd64 go build -ldflags="-s -w" -o yamlsort_bsd
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o yamlsort_osx
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o yamlsort.exe
```

## Usage
```
Usage: yamlsort [<infile>] [flags]

Sort yaml-file recursively

Arguments:
  [<infile>]    Input file [default: stdin]

Flags:
  -h, --help             Show context-sensitive help.
  -o, --outfile=FILE     Output file [default: stdout]
  -i, --in-place=FILE    In-place sort of the provided file
  -V, --version          Display version
```
