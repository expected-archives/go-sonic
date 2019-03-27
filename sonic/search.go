package sonic

import (
	"fmt"
	"strings"
)

// Searchable is used for querying the search index.
type Searchable interface {

	// Query the database, return a list of object, represented as a string.
	// Sonic default limit is 10.
	// Command syntax QUERY <collection> <bucket> "<terms>" [LIMIT(<count>)]? [OFFSET(<count>)]?.
	Query(collection, bucket, terms string, limit, offset int) (results []string, err error)

	// Suggest auto-completes word, return a list of words as a string.
	// Command syntax SUGGEST <collection> <bucket> "<word>" [LIMIT(<count>)]?.
	Suggest(collection, bucket, word string, limit int) (results []string, err error)

	// Quit refer to the Base interface
	Quit() (err error)

	// Ping refer to the Base interface
	Ping() (err error)
}

type searchCommands string

const (
	query   searchCommands = "QUERY"
	suggest searchCommands = "SUGGEST"
)

type searchChannel struct {
	*driver
}

// NewIngester create a new driver instance with a searchChannel instance.
// Only way to get a Searchable implementation.
func NewSearch(host string, port int, password string) (Searchable, error) {
	driver := &driver{
		Host:     host,
		Port:     port,
		Password: password,
		channel:  Search,
	}
	err := driver.Connect()
	if err != nil {
		return nil, err
	}
	return searchChannel{
		driver: driver,
	}, nil
}

func (s searchChannel) Query(collection, bucket, term string, limit, offset int) (results []string, err error) {
	err = s.write(fmt.Sprintf("%s %s %s \"%s\" LIMIT(%d) OFFSET(%d)", query, collection, bucket, term, limit, offset))
	if err != nil {
		return nil, err
	}

	// pending, should be PENDING ID_EVENT
	_, err = s.read()
	if err != nil {
		return nil, err
	}

	// event query, should be EVENT QUERY ID_EVENT RESULT1 RESULT2 ...
	read, err := s.read()
	if err != nil {
		return nil, err
	}
	return getSearchResults(read, string(query)), nil
}

func (s searchChannel) Suggest(collection, bucket, word string, limit int) (results []string, err error) {
	err = s.write(fmt.Sprintf("%s %s %s \"%s\" LIMIT(%d)", suggest, collection, bucket, word, limit))
	if err != nil {
		return nil, err
	}

	// pending, should be PENDING ID_EVENT
	_, err = s.read()
	if err != nil {
		return nil, err
	}

	// event query, should be EVENT SUGGEST ID_EVENT RESULT1 RESULT2 ...
	read, err := s.read()
	if err != nil {
		return nil, err
	}
	return getSearchResults(read, string(suggest)), nil
}

func getSearchResults(line string, eventType string) []string {
	if strings.HasPrefix(line, "EVENT "+eventType) {
		return strings.Split(line, " ")[3:]
	}
	return []string{}
}
