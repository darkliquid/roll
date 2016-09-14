# Roll

A simple dice roll parsing engine

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
