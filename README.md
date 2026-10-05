# Roll [![Go Report Card](https://goreportcard.com/badge/github.com/darkliquid/roll)](https://goreportcard.com/report/github.com/darkliquid/roll) [![License](https://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/darkliquid/roll/blob/master/LICENSE) [![GoDoc](https://godoc.org/github.com/darkliquid/roll?status.svg)](https://godoc.org/github.com/darkliquid/roll) [![CI](https://github.com/darkliquid/roll/actions/workflows/ci.yml/badge.svg)](https://github.com/darkliquid/roll/actions/workflows/ci.yml)

A simple dice roll compiler and VM that (mostly) supports the [Roll20 Dice Rolling Language Specification][1]

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

Then run `echo '3d6+1' | go run main.go` to get some output like:

`Rolled "3d6+1" and got 6, 6, 4 for a total of 17`

If you want the compiled program directly, use the compile/evaluate API:

```go
program, err := roll.CompileString("3d6+1")
if err != nil {
    panic(err)
}

result, err := roll.EvaluateProgram(program)
if err != nil {
    panic(err)
}

fmt.Println(program.String(), result.Total)
```

## Web (GitHub Pages)

An interactive, fully client-side version runs in the browser via WebAssembly
at <https://darkliquid.github.io/roll/>.

To build the site locally:

```bash
mise run build-wasm
```

This compiles `cmd/wasm` to `docs/roll.wasm` and copies Go's `wasm_exec.js`
into `docs/`. The static site can then be served from `docs/` with any file
server, for example `python3 -m http.server --directory docs`.

## Development

Toolchains (Go, golangci-lint, govulncheck) are pinned in `mise.toml`. Install them
with `mise install`, then run the checks through mise:

```bash
mise run ci        # build, vet, lint, vulncheck and test
mise run build     # compile every package, including js/wasm
mise run lint      # golangci-lint
mise run vulncheck # govulncheck
```

[1]:https://wiki.roll20.net/Dice_Reference
