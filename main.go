package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/k0kubun/pp"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

type options struct {
	EscapeHTML  bool
	Fields      []string
	From        string
	Indent      int
	Input       string
	IsArray     bool
	ListFormats bool
	Scan        bool
	To          string
	WrapInKey   string
}

var opts options
var v *viper.Viper
var dumper func(value interface{}) ([]byte, error)

var rootCmd *cobra.Command = &cobra.Command{
	Use:     "vmap",
	Version: "0.1.1",
	Short:   "convert one data format to another",
	Run: func(cmd *cobra.Command, args []string) {

		// help utilities (not main function)
		if opts.ListFormats {
			list := strings.Join(viper.SupportedExts, ", ")
			fmt.Printf("%s\n", list)
			os.Exit(0)
		}

		opts.From = strings.TrimSpace(strings.ToLower(opts.From))
		opts.To = strings.TrimSpace(strings.ToLower(opts.To))

		// set dumper
		switch opts.To {
		case "json":
			dumper = jsonMarshaller
		case "yaml", "yml":
			dumper = yaml.Marshal
		case "toml":
			dumper = tomlMarshaller
		case "go":
			dumper = prettyGoMarshaller
		default:
			log.Printf("error: could not find marshaller for: %s\n", opts.To)
			log.Fatalf("try one of: %v\n", []string{"go", "json", "yaml", "toml"})
		}

		// configure input format
		var absPath string
		isGuess := opts.From == "" || opts.From == "guess"
		if opts.IsArray {
			opts.From = "json"
		} else if isGuess {
			if opts.Input != "" && !opts.Scan {
				absPath, _ = filepath.Abs(opts.Input)
				ext := filepath.Ext(absPath)
				ext = strings.TrimPrefix(ext, ".")
				opts.From = strings.ToLower(ext)
			} else {
				opts.From = "json"
			}
		}

		if opts.From == "csv" {
			loadData(func(r io.Reader) {
				d := csv.NewReader(r)
				d.TrimLeadingSpace = true
				d.TrailingComma = true
				if records, err := d.ReadAll(); err != nil {
					log.Fatalf("error: %v\n", err)
				} else {
					fromRecords(records)
				}
			})
		} else {
			if !contains(viper.SupportedExts, opts.From) {
				log.Fatalf("error: unsupported input: %s\n", opts.From)
			}
			v.SetConfigType(opts.From)

			loadData(func(r io.Reader) {
				if opts.IsArray {
					data, err := fromArray(r)
					if err != nil {
						log.Fatalf("error: %v\n", err)
					}
					v.Set(opts.WrapInKey, data)
				} else {
					if err := v.ReadConfig(r); err != nil {
						log.Fatalf("error: %v\n", err)
					}
				}
			})
		}

		obj := v.AllSettings()
		result, err := dumper(obj)
		if err != nil {
			log.Fatalf("error: %v\n", err)
		}

		fmt.Println(string(result))
	},
}

func contains(arr []string, elem string) bool {
	for _, val := range arr {
		if val == elem {
			return true
		}
	}
	return false
}

func max(a, b int) int {
	if a < b {
		return b
	}
	return a
}

func tomlMarshaller(value interface{}) ([]byte, error) {
	w := new(bytes.Buffer)
	e := toml.NewEncoder(w)

	err := e.Encode(value)
	return w.Bytes(), err
}

func jsonMarshaller(value interface{}) ([]byte, error) {
	w := new(bytes.Buffer)
	e := json.NewEncoder(w)
	e.SetEscapeHTML(opts.EscapeHTML)
	e.SetIndent("", strings.Repeat(" ", opts.Indent))

	err := e.Encode(value)

	return w.Bytes(), err
}

func prettyGoMarshaller(value interface{}) ([]byte, error) {
	s := pp.Sprint(value)
	return []byte(s), nil
}

func fromArray(r io.Reader) ([]interface{}, error) {
	result := make([]interface{}, 0)
	d := json.NewDecoder(r)
	err := d.Decode(&result)
	return result, err
}

func fromRecords(records [][]string) {
	// create fields
	head := records[0]
	fieldCount := len(opts.Fields)
	elemCount := len(head)
	fields := make([]string, 0)
	if 0 < fieldCount {
		fields = append(fields, opts.Fields...)
	}
	if fieldCount < elemCount {
		for i := max(fieldCount-1, 0); i < elemCount; i++ {
			fields = append(fields, fmt.Sprintf("field%d", i+1))
		}
	}

	data := make([]interface{}, len(records))
	for i, record := range records {
		m := make(map[string]string)
		for j, val := range record {
			m[fields[j]] = val
		}
		data[i] = m
	}

	v.Set(opts.WrapInKey, data)
}

func loadData(callback func(io.Reader)) {

	if opts.Scan {
		// if scan is specified, parse stdin
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) != 0 {
			log.Fatal("error: nothing to read from stdin\n")
		}

		stdin, _ := ioutil.ReadAll(os.Stdin)

		reader := bytes.NewReader(stdin)
		callback(reader)
	} else {
		// else, check for input file
		if opts.Input == "" {
			log.Fatalln("error: input is required when not reading from stdin")
		}

		// normalize source path if any
		absPath, err := filepath.Abs(opts.Input)
		if err != nil {
			log.Fatalf("error: %v\n", err)
		}

		if stat, err := os.Stat(absPath); err != nil {
			log.Fatalf("error: %v\n", err)
		} else if stat.IsDir() {
			log.Fatalf("error: please specify normal file, not directory: %s\n", absPath)
		}

		content, err := ioutil.ReadFile(absPath)
		if err != nil {
			log.Fatalf("error: %v\n", err)
		}
		reader := bytes.NewReader(content)
		callback(reader)
	}
}

func setup() {
	log.SetPrefix("[vmap] ")
	pp.ColoringEnabled = false

	v = viper.New()
	opts = options{}

	flagSet := rootCmd.PersistentFlags()

	flagSet.BoolVar(&opts.ListFormats, "list-formats", false, "list available input format")
	flagSet.BoolVarP(&opts.IsArray, "array", "a", false, "decode as array not map. input forced to json")
	flagSet.BoolVarP(&opts.EscapeHTML, "escape-html", "E", false, "escape html on json output")
	flagSet.BoolVarP(&opts.Scan, "read-stdin", "S", false, "read from stdin")
	flagSet.StringVarP(&opts.From, "from", "f", "guess", "input data format")
	flagSet.StringVarP(&opts.Input, "input", "i", "", "input file path")
	flagSet.StringVarP(&opts.To, "to", "t", "toml", "output data format")
	flagSet.StringVarP(&opts.WrapInKey, "wrap-in", "w", "data", "field name when decoding array")
	flagSet.StringSliceVarP(&opts.Fields, "fields", "c", []string{}, "column names for CSV decode")
	flagSet.IntVarP(&opts.Indent, "indent", "n", 4, "indents for json output")
}

func main() {
	setup()
	if err := rootCmd.Execute(); err != nil {
		// unhandled in Execute()
		os.Exit(2)
	}
}
