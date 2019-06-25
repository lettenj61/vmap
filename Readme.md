# vmap

Your humble data converter

## usage

```
$ vmap -h
convert one data format to another

Usage:
  vmap [flags]

Flags:
  -E, --escape-html    escape html on json output
  -f, --from string    input data format (default "guess")
  -h, --help           help for vmap
  -n, --indent int     indents for json output (default 4)
  -i, --input string   input file path
      --list-formats   list available input format
  -S, --read-stdin     read from stdin
  -t, --to string      output data format (default "toml")
      --version        version for vmap
```

### read from stdin

`vmap -S`

```bat
$ echo tests = ["foo", "bar", "baz"] | vmap -S -f toml -t json
{
    "tests": [
        "foo",
        "bar",
        "baz"
    ]
}
```

### read from file

`vmap -i <FILE>`

```bat
$ type cdnjs.json
{
  "results": [
    {
      "name": "awesomplete",
      "latest": "https://cdnjs.cloudflare.com/ajax/libs/awesomplete/1.1.4/awesomplete.min.js"
    }
  ],
  "total": 1
}

$ vmap -i cdnjs.json -f json -t yaml
results:
- latest: https://cdnjs.cloudflare.com/ajax/libs/awesomplete/1.1.4/awesomplete.min.js
  name: awesomplete
total: 1
```

### dump like pretty-printed Go literals

```bat
$ vmap -i elm.json -t go
map[string]interface {}{
  "dependencies": map[string]interface {}{
    "direct": map[string]interface {}{
      "elm/json":    "1.0.0",
      "elm/browser": "1.0.0",
      "elm/core":    "1.0.0",
      "elm/html":    "1.0.0",
    },
    "indirect": map[string]interface {}{
      "elm/virtual-dom": "1.0.0",
      "elm/time":        "1.0.0",
      "elm/url":         "1.0.0",
    },
  },
  "source-directories": []interface {}{
    "src",
  },
  "elm-version": "0.19.0",
  "type":        "application",
}
```

### list of valid input formats

> It's just a proxy to `viper.SupportedExts`

```bat
vmap --list-formats
json, toml, yaml, yml, properties, props, prop, hcl
```


## tips

* `vmap` decode data with [spf13/viper][viper]. Note that `viper` automatically lower case all the keys it found in input.


## considering enhancements

* Support XML output
* Support CSV output
* Decode non-UTF8 charsets


## packages that power vmap

* [BurntSushi/toml][toml]
* [spf13/cobra][cobra]
* [spf13/viper][viper]
* [k0kubun/pp][pp]
* [go-yaml/yaml][yaml]

Thanks for great creations!


## license

Apache 2.0


[toml]: https://github.com/BurntSushi/toml
[cobra]: https://github.com/spf13/cobra
[viper]: https://github.com/spf13/viper
[pp]: https://github.com/k0kubun/pp
[yaml]: https://github.com/go-yaml/yaml
