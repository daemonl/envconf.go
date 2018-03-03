[![Build Status](https://travis-ci.org/daemonl/envconf.go.svg?branch=master)](https://travis-ci.org/daemonl/envconf.go)
[![GoDoc](https://godoc.org/gopkg.daemonl.com/envconf?status.svg)](https://godoc.org/gopkg.daemonl.com/envconf)
[![codecov](https://codecov.io/gh/daemonl/envconf.go/branch/master/graph/badge.svg)](https://codecov.io/gh/daemonl/envconf.go)



Env Conf
========

Environment Variable config loader for go

The import path "gopkg.daemonl.com/envconf" is equivalent to
"github.com/daemonl/envconf.go", but some tools aren't keen on the
`.go` suffix. (And what if I wanted to write this for another language?)

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
