## Go client for the sonic search backend

This package implement all commands to work with sonic. If there is one missing, open an issue ! :)

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

	_ = ingester.Push("movies", "general", "id:6ab56b4kk3", "Star wars")
	_ = ingester.Push("movies", "general", "id:5hg67f8dg5", "Spider man")
	_ = ingester.Push("movies", "general", "id:1m2n3b4vf6", "Batman")
	_ = ingester.Push("movies", "general", "id:68d96h5h9d0", "This is another movie")

	search, err := sonic.NewSearch("localhost", 1491, "SecretPassword")
	if err != nil {
		panic(err)
	}

	results, _ := search.Query("movies", "general", "man", 10, 0)

	fmt.Println(results)
}
```