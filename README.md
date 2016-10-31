# Roll [![Go Report Card](https://goreportcard.com/badge/github.com/darkliquid/roll)](https://goreportcard.com/report/github.com/darkliquid/roll) [![License](https://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/darkliquid/roll/blob/master/LICENSE) [![GoDoc](https://godoc.org/github.com/darkliquid/roll?status.svg)](https://godoc.org/github.com/darkliquid/roll) [![Build Status](https://travis-ci.org/darkliquid/roll.svg?branch=master)](https://travis-ci.org/darkliquid/roll)

A simple dice roll parsing engine that (mostly) supports the [Roll20 Dice Rolling Language Specification][1]

## Usage

```go
package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/darkliquid/roll"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	out, err := roll.Parse(os.Stdin)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	fmt.Println(out)
}
```

then run `echo '3d6+1' | go run main.go` to get some output like:

`Rolled "3d6+1" and got 6, 6, 4 for a total of 17`

[1]:https://wiki.roll20.net/Dice_Reference
