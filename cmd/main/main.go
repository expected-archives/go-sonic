package main

import (
	"fmt"
	"github.com/expectedsh/go-sonic/sonic"
)

func main() {
	connection := &sonic.Connection{
		Host:     "localhost",
		Port:     1491,
		Password: "SecretPassword",
		Channel:  sonic.Ingest,
	}

	e := connection.Connect()
	if e != nil {
		panic(e)
	}

	channel := sonic.IngesterChannel{Connection: connection}
	e = channel.Push("test", "default", "captain", "lol is truth")
	fmt.Println(e)
}
