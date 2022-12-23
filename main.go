package main

import (
	"fmt"
	"github.com/alecthomas/kong"
	"gopkg.in/yaml.v2"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
)

var CLI struct {
	Sort struct {
	} `cmd:"" embed:"" help:"Sort YAML"`
	Infile  string `name:"infile" help:"input file. Defaults to stdin"`
	Outfile string `name:"outfile" help:"output file. Defaults to stdout"`
	InPlace string `name:"in-place" short:"i" optional:"" help:"sort the provided file in-place"`
}

func main() {
	kongCTX := kong.Parse(&CLI)
	// in-place sorting and Infile/Outfile are mutually exclusive
	if strings.EqualFold(CLI.InPlace, "") && (!strings.EqualFold(CLI.Infile, "") || !strings.EqualFold(CLI.Outfile, "")) {
		kongCTX.Printf("in-place sorting and Infile/Outfile are mutually exclusive")
		os.Exit(1)
	}
	// set infile
	infile := os.Stdin
	if !strings.EqualFold(CLI.InPlace, "") {
		var err error
		if infile, err = os.Open(CLI.InPlace); err != nil {
			kongCTX.Errorf("failed to open input file %s: %s", err)
			os.Exit(1)
		}
		defer func() {
			_ = infile.Close()
		}()
	}
	if !(strings.EqualFold(CLI.Infile, "-") || strings.EqualFold(CLI.Infile, "")) {
		var err error
		if infile, err = os.Open(CLI.Infile); err != nil {
			kongCTX.Errorf("failed to open input file %s: %s", err)
			os.Exit(1)
		}
		defer func() {
			_ = infile.Close()
		}()
	}
	// set outfile
	outfile := os.Stdout
	if !strings.EqualFold(CLI.InPlace, "") {
		// create a temp file for in-place sorting and
		var err error
		dir := filepath.Dir(CLI.InPlace)
		outfile, err = os.CreateTemp(dir, "outfile-*.yaml")
		if err != nil {
			kongCTX.Errorf("failed to create temp file: %s", err)
			os.Exit(1)
		}
		defer func() {
			_ = outfile.Close()
		}()
		defer func(name string) {
			// it might be ok because of the renaming
			_ = os.Remove(name)
		}(outfile.Name())
	}
	if !(strings.EqualFold(CLI.Outfile, "-") || strings.EqualFold(CLI.Outfile, "")) {
		var err error
		if outfile, err = os.Create(CLI.Outfile); err != nil {
			kongCTX.Errorf("failed to create output file %s: %s", err)
			os.Exit(1)
		}
		defer func() {
			_ = outfile.Close()
		}()
	}
	var in yaml.MapSlice
	dec := yaml.NewDecoder(infile)
	dec.SetStrict(true)
	if err := dec.Decode(&in); err != nil {
		kongCTX.Errorf("failed to decode input yaml: %s", err)
		os.Exit(1)
	}
	out := sortYAML(in)
	if err := yaml.NewEncoder(outfile).Encode(out); err != nil {
		kongCTX.Errorf("failed to encode sorted yaml: %s", err)
		os.Exit(1)
	}
	// if in-place, copy content of the outfile (temp file) to infile
	if !strings.EqualFold(CLI.InPlace, "") {
		_, err := infile.Seek(0, 0)
		if err != nil {
			kongCTX.Errorf("failed to rewind infile: %s", err)
			os.Exit(1)
		}
		if _, err := io.Copy(infile, outfile); err != nil {
			kongCTX.Errorf("failed to copy temp file content to in-place file: %s", err)
			os.Exit(1)
		}
	}
}

type sortedYAML []yaml.MapItem

func (s sortedYAML) Len() int {
	return len(s)
}

func (s sortedYAML) Less(i, j int) bool {
	iStr, ok := s[i].Key.(string)
	if !ok {
		panic("key is not string assertable")
	}
	jStr, ok := s[j].Key.(string)
	if !ok {
		panic("key is not string assertable")
	}
	return strings.Compare(iStr, jStr) < 0
}

func (s sortedYAML) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func sortYAML(in yaml.MapSlice) sortedYAML {
	sort.Sort(sortedYAML(in))
	for _, v := range in {
		// can't sort nil
		if v.Value == nil {
			continue
		}
		if obj, ok := v.Value.(yaml.MapSlice); ok {
			sortedObj := sortYAML(obj)
			v.Value = sortedObj
			continue
		}
		// descend into list of objects and for now preserve the list order
		if obj, ok := v.Value.([]interface{}); ok {
			for idx, elem := range obj {
				if mapSlice, isMapSlice := elem.(yaml.MapSlice); isMapSlice {
					obj[idx] = sortYAML(mapSlice)
				}
			}
			continue
		}
		// by now only basic types should be left over
		t := reflect.TypeOf(v.Value).Kind()
		switch t {
		case reflect.Int, reflect.Float64, reflect.Bool, reflect.String:
			// those are basic types, nothing to do
			continue
		default:
			fmt.Printf("# XXX %s is %T (kind = %d). Sorting isn't implemented or possible yet\n", v.Key, v.Value, t)
		}
	}
	return sortedYAML(in)
}
