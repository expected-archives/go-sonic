[![GoDoc](https://godoc.org/github.com/expectedsh/go-sonic/sonic?status.svg)](https://godoc.org/github.com/expectedsh/go-sonic/sonic) [![GoLint](https://img.shields.io/badge/golint-ok-green.svg)](https://go-lint.appspot.com/github.com/expectedsh/go-sonic/sonic)

## Go client for the sonic search backend

This package implement all commands to work with sonic. If there is one missing, open an issue ! :)

Sonic: https://github.com/valeriansaliou/sonic

### Install

`go get github.com/expectedsh/go-sonic`

### Example

```go
package main

import (
	"fmt"
	"github.com/expectedsh/go-sonic/sonic"
)

func main() {

	ingester, err := sonic.NewIngester("localhost", 1491, "SecretPassword")
	if err != nil {
		panic(err)
	}

	// I will ignore all errors for demonstration purposes

	_ = ingester.BulkPush("movies", "general", 3, []sonic.IngestBulkRecord{
		{"id:6ab56b4kk3", "Star wars"},
		{"id:5hg67f8dg5", "Spider man"},
		{"id:1m2n3b4vf6", "Batman"},
		{"id:68d96h5h9d0", "This is another movie"},
	})

	search, err := sonic.NewSearch("localhost", 1491, "SecretPassword")
	if err != nil {
		panic(err)
	}

	results, _ := search.Query("movies", "general", "man", 10, 0)

	fmt.Println(results)
}
```

### Benchmark bulk

Method BulkPush and BulkPop use custom connection pool with goroutine dispatch algorithm.
This is the benchmark (file sonic/ingester_test.go):

```
goos: linux
goarch: amd64
pkg: github.com/expectedsh/go-sonic/sonic
BenchmarkIngesterChannel_BulkPushMaxCPUs-8   	       2	 662657959 ns/op
BenchmarkIngesterChannel_BulkPush10-8        	       2	 603779977 ns/op
BenchmarkIngesterChannel_Push-8              	       1	1023322864 ns/op
PASS
```

### Thread Safety

The driver itself isn't thread safe. You could use locks or channels to avoid crashes.

```go
package main

import (
	"fmt"

	"github.com/expectedsh/go-sonic/sonic"
)

func main() {
	events := make(chan []string, 1)

	// simulating a high incoming message load
	tryCrash := func() {
		for {
			events <- []string{"some_text", "some_id"}
		}
	}

	go tryCrash()
	go tryCrash()
	go tryCrash()
	go tryCrash()

	ingester, _ := sonic.NewIngester("localhost", 1491, "SecretPassword")

	for {
		msg := <-events
		// Or use some buffering along with BulkPush
		ingester.Push("collection", "bucket", msg[1], msg[0])
	}
}
```

Bulk push is faster than for loop on Push. 
Hardware detail: Intel(R) Core(TM) i7-8550U CPU @ 1.80GHz
