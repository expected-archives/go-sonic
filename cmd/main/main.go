package main

import (
	"fmt"
	"github.com/expectedsh/go-sonic/sonic"
	"time"
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
	c, e := channel.Count("test", "default", "captain")
	fmt.Println("waiting")
	time.Sleep(time.Second * 10)
	e = channel.Ping()
	//c, e = channel.Count("test", "default", "captain")

	fmt.Println(c, e)
}
