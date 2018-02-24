[![Build Status](https://travis-ci.org/daemonl/envconf.go?branch=master)](https://travis-ci.org/daemonl/envconf.go)
[![GoDoc](https://godoc.org/github.com/daemonl/envconf.go?status.svg)](https://godoc.org/github.com/daemonl/envconf.go)

Env Conf
========

Environment Variable config loader for go

Simple Usage

```
var config struct {
	Bind string `env:"BIND" default:":8080"`
}

func main() {
	if err := envconf.Parse(&config); err != nil {
		log.Fatal(err.Error())
	}
}
```


