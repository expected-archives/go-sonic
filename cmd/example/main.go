package main

import (
	"fmt"

	"github.com/expectedsh/go-sonic/sonic"
)

const pswd = "SecretPassword"

func main() {

	ingester, err := sonic.NewIngester("localhost", 1491, pswd)
	if err != nil {
		panic(err)
	}

	// I will ignore all errors for demonstration purposes

	_ = ingester.BulkPush("movies", "general", 3, []sonic.IngestBulkRecord{
		{Object: "id:6ab56b4kk3", Text: "Star wars"},
		{Object: "id:5hg67f8dg5", Text: "Spider man"},
		{Object: "id:1m2n3b4vf6", Text: "Batman"},
		{Object: "id:68d96h5h9d0", Text: "This is another movie"},
	}, sonic.LangAutoDetect)

	search, err := sonic.NewSearch("localhost", 1491, pswd)
	if err != nil {
		panic(err)
	}

	results, _ := search.Query("movies", "general", "man", 10, 0, sonic.LangAutoDetect)

	fmt.Println(results)

	// Search with LANG set to "none" and "eng"

	_ = ingester.FlushCollection("movies")
	_ = ingester.BulkPush("movies", "general", 3, []sonic.IngestBulkRecord{
		{Object: "id:6ab56b4kk3", Text: "Star wars"},
		{Object: "id:5hg67f8dg5", Text: "Spider man"},
		{Object: "id:1m2n3b4vf6", Text: "Batman"},
		{Object: "id:68d96h5h9d0", Text: "This is another movie"},
	}, sonic.LangNone)

	results, _ = search.Query("movies", "general", "this is", 10, 0, sonic.LangNone)
	fmt.Println(results)
	// [id:68d96h5h9d0]

	// English stop words should be encountered by Sonic now
	results, _ = search.Query("movies", "general", "this is", 10, 0, sonic.LangEng)
	fmt.Println(results)
	// []
}
