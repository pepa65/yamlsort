package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"

	"github.com/alecthomas/kong"
	"gopkg.in/yaml.v2"
)

const version = "0.1.3"
const name = "yamlsort"

type VersionStr string

func (v VersionStr) Decode(_ *kong.DecodeContext) error { return nil }
func (v VersionStr) IsBool() bool                       { return true }
func (v VersionStr) BeforeApply(app *kong.Kong) error {
	fmt.Println(name + " " + version)
	app.Exit(0)
	return nil
}

var CLI struct {
	Sort struct {
	} `cmd:"" embed:"" help:"Sort yaml-file recursively"`
	Infile  string     `name:"infile" help:"Input file [default: stdin]" type:"existingfile" arg:"" optional:""`
	Outfile string     `name:"outfile" short:"o" help:"Output file [default: stdout]" type:"path" placeholder:"FILE"`
	InPlace string     `name:"in-place" short:"i" optional:"" help:"In-place sort of the provided file" type:"existingfile" placeholder:"FILE"`
	Version VersionStr `help:"Display version" name:"version" short:"V"`
}

func main() {
	kongCTX := kong.Parse(&CLI, kong.Description("Sort yaml-file recursively"), kong.Vars{"version": version})
	// No arguments shows help
	if strings.EqualFold(CLI.InPlace, "") && strings.EqualFold(CLI.Infile, "") && strings.EqualFold(CLI.Outfile, "") {
		kong.DefaultHelpPrinter(kong.HelpOptions{Compact: false}, kongCTX)
		os.Exit(1)
	}

	if !strings.EqualFold(CLI.InPlace, "") && (!strings.EqualFold(CLI.Infile, "") || !strings.EqualFold(CLI.Outfile, "")) {
		kongCTX.Printf("In-place sorting and Infile/Outfile are mutually exclusive")
		os.Exit(2)
	}
	// Set infile
	infile := os.Stdin
	if !strings.EqualFold(CLI.InPlace, "") {
		var err error
		if infile, err = os.OpenFile(CLI.InPlace, os.O_RDWR, 0644); err != nil {
			kongCTX.Errorf("Failed to open input file %s: %s", err)
			os.Exit(3)
		}
		defer func() {
			_ = infile.Close()
		}()
	}
	if !(strings.EqualFold(CLI.Infile, "-") || strings.EqualFold(CLI.Infile, "")) {
		var err error
		if infile, err = os.Open(CLI.Infile); err != nil {
			kongCTX.Errorf("Failed to open input file %s: %s", err)
			os.Exit(4)
		}
		defer func() {
			_ = infile.Close()
		}()
	}
	// Set outfile
	outfile := os.Stdout
	if !strings.EqualFold(CLI.InPlace, "") {
		// Create a temp file for in-place sorting
		var err error
		dir := filepath.Dir(CLI.InPlace)
		outfile, err = os.CreateTemp(dir, "outfile-*.yaml")
		if err != nil {
			kongCTX.Errorf("Failed to create temp file: %s", err)
			os.Exit(5)
		}
		defer func() {
			_ = outfile.Close()
		}()
		defer func(name string) {
			// It might be ok because of the renaming
			_ = os.Remove(name)
		}(outfile.Name())
	}
	if !(strings.EqualFold(CLI.Outfile, "-") || strings.EqualFold(CLI.Outfile, "")) {
		var err error
		if outfile, err = os.Create(CLI.Outfile); err != nil {
			kongCTX.Errorf("Failed to create output file %s: %s", err)
			os.Exit(6)
		}
		defer func() {
			_ = outfile.Close()
		}()
	}
	var in yaml.MapSlice
	dec := yaml.NewDecoder(infile)
	dec.SetStrict(true)
	if err := dec.Decode(&in); err != nil {
		kongCTX.Errorf("Failed to decode input yaml: %s", err)
		os.Exit(7)
	}
	out := sortYAML(in)
	if err := yaml.NewEncoder(outfile).Encode(out); err != nil {
		kongCTX.Errorf("Failed to encode sorted yaml: %s", err)
		os.Exit(8)
	}
	//_ = outfile.Sync()
	// If in-place, copy content of the outfile (temp file) to infile
	if !strings.EqualFold(CLI.InPlace, "") {
		_, err := infile.Seek(0, io.SeekStart)
		if err != nil {
			kongCTX.Errorf("Failed to rewind infile: %s", err)
			os.Exit(9)
		}
		_, err = outfile.Seek(0, io.SeekStart)
		if err != nil {
			kongCTX.Errorf("Failed to rewind outfile: %s", err)
			os.Exit(10)
		}
		if _, err := io.Copy(infile, outfile); err != nil {
			kongCTX.Errorf("Failed to copy temp file content to in-place file: %s", err)
			os.Exit(11)
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
		panic("Key is not string assertable")
	}
	jStr, ok := s[j].Key.(string)
	if !ok {
		panic("Key is not string assertable")
	}
	return strings.Compare(iStr, jStr) < 0
}

func (s sortedYAML) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func sortYAML(in yaml.MapSlice) sortedYAML {
	sort.Sort(sortedYAML(in))
	for _, v := range in {
		// Can't sort nil
		if v.Value == nil {
			continue
		}
		if obj, ok := v.Value.(yaml.MapSlice); ok {
			sortedObj := sortYAML(obj)
			v.Value = sortedObj
			continue
		}
		// Descend into list of objects and for now preserve the list order
		if obj, ok := v.Value.([]interface{}); ok {
			for idx, elem := range obj {
				if mapSlice, isMapSlice := elem.(yaml.MapSlice); isMapSlice {
					obj[idx] = sortYAML(mapSlice)
				}
			}
			continue
		}
		// By now only basic types should be left over
		t := reflect.TypeOf(v.Value).Kind()
		switch t {
		case reflect.Int, reflect.Float64, reflect.Bool, reflect.String:
			// Those are basic types, nothing to do
			continue
		default:
			fmt.Printf("# XXX %s is %T (kind = %d). Sorting isn't implemented or possible yet\n", v.Key, v.Value, t)
		}
	}
	return sortedYAML(in)
}
