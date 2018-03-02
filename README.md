[![Build Status](https://travis-ci.org/daemonl/envconf.go.svg?branch=master)](https://travis-ci.org/daemonl/envconf.go)
[![GoDoc](https://godoc.org/gopkg.daemonl.com/envconf?status.svg)](https://godoc.org/gopkg.daemonl.com/envconf)
[![codecov](https://codecov.io/gh/daemonl/envconf.go/branch/master/graph/badge.svg)](https://codecov.io/gh/daemonl/envconf.go)



Env Conf
========

Environment Variable config loader for go

## Simple Usage

```

import "gopkg.daemonl.com/envconf"

var config struct {
	Bind string `env:"BIND" default:":8080"`
}

func main() {
	if err := envconf.Parse(&config); err != nil {
		log.Fatal(err.Error())
	}
}
```
