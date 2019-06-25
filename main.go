package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

type options struct {
	EscapeHTML  bool
	Input       string
	ListFormats bool
	From        string
	To          string
	Scan        bool
	Indent      int
}

var opts options
var v *viper.Viper
var dumper func(value interface{}) ([]byte, error)

var rootCmd *cobra.Command = &cobra.Command{
	Use:     "vmap",
	Version: "0.1.0",
	Short:   "convert one data format to another",
	Long:    "",
	Run: func(cmd *cobra.Command, args []string) {

		// help utilities (not main function)
		if opts.ListFormats {
			list := strings.Join(viper.SupportedExts, ", ")
			fmt.Printf("%s\n", list)
			os.Exit(0)
		}

		var absPath string

		opts.From = strings.TrimSpace(strings.ToLower(opts.From))
		opts.To = strings.TrimSpace(strings.ToLower(opts.To))

		isGuess := opts.From == "" || opts.From == "guess"
		if isGuess && opts.Input != "" && !opts.Scan {
			absPath, _ = filepath.Abs(opts.Input)
			ext := filepath.Ext(absPath)
			ext = strings.TrimPrefix(ext, ".")
			opts.From = strings.ToLower(ext)
		} else if isGuess && opts.Scan {
			opts.From = "json"
		}

		// initialize from options
		if !contains(viper.SupportedExts, opts.From) {
			log.Fatalf("error: unsupported input: %s\n", opts.From)
		}
		if !contains(viper.SupportedExts, opts.To) {
			log.Fatalf("error: unsupported output: %s\n", opts.To)
		}
		v.SetConfigType(opts.From)

		// set dumper
		switch opts.To {
		case "json":
			dumper = jsonMarshaller
		case "yaml":
			dumper = yaml.Marshal
		case "toml":
			dumper = tomlMarshaller
		default:
			log.Fatalf("error: datix could not find marshaller for: %s\n", opts.To)
		}

		if opts.Scan {
			// if scan is specified, parse stdin
			stat, _ := os.Stdin.Stat()
			if (stat.Mode() & os.ModeCharDevice) != 0 {
				log.Fatal("error: nothing to read from stdin\n")
			}

			stdin, _ := ioutil.ReadAll(os.Stdin)

			reader := bytes.NewReader(stdin)
			if err := v.ReadConfig(reader); err != nil {
				log.Fatalf("error: %v\n", err)
			}
		} else {
			// else, check for input file
			if opts.Input == "" {
				log.Fatal("error: input is required when not reading from stdin\n")
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
			v.ReadConfig(bytes.NewReader(content))
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

func setup() {
	log.SetPrefix("[vmap] ")

	v = viper.New()
	opts = options{}

	rootCmd.PersistentFlags().BoolVar(&opts.ListFormats, "list-formats", false, "list available input format")
	rootCmd.PersistentFlags().BoolVarP(&opts.EscapeHTML, "escape-html", "E", false, "escape html on json output")
	rootCmd.PersistentFlags().StringVarP(&opts.Input, "input", "i", "", "input file path")
	rootCmd.PersistentFlags().StringVarP(&opts.From, "from", "f", "guess", "input data format")
	rootCmd.PersistentFlags().StringVarP(&opts.To, "to", "t", "toml", "output data format")
	rootCmd.PersistentFlags().BoolVarP(&opts.Scan, "read-stdin", "S", false, "read from stdin")
	rootCmd.PersistentFlags().IntVarP(&opts.Indent, "indent", "n", 4, "indents for json output")
}

func main() {
	setup()
	if err := rootCmd.Execute(); err != nil {
		// unhandled in Execute()
		os.Exit(2)
	}
}
