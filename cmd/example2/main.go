package main

import (
	"github.com/expectedsh/go-sonic/sonic/connection"
)

func main() {
	conn := connection.NewConnection("localhost", 1491, "SecretPassword")

	//_ = ingester.BulkPush("movies", "general", 3, []sonic.IngestBulkRecord{
	//	{Object: "id:6ab56b4kk3", Text: "Star wars"},
	//	{Object: "id:5hg67f8dg5", Text: "Spider man"},
	//	{Object: "id:1m2n3b4vf6", Text: "Batman"},
	//	{Object: "id:68d96h5h9d0", Text: "This is another movie"},
	//})
	//
	//search, err := sonic.NewSearch("localhost", 1491, "SecretPassword")
	//if err != nil {
	//	panic(err)
	//}

	if err := conn.Open(); err != nil {
		panic(err)
	}

	//for {
	//	results, err := search.Query("movies", "general", "man", 10, 0)
	//	fmt.Println(results, err)
	//	time.Sleep(time.Second * 5)
	//}
}
