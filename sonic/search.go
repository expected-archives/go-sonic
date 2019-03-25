package sonic

import (
	"fmt"
	"strings"
)

type Searchable interface {
	Query(collection, bucket, term string, limit, offset int) (results []string, err error)
	Suggest(collection, bucket, word string, limit int) (results []string, err error)
}

type searchCommands string

const (
	query   searchCommands = "QUERY"
	suggest searchCommands = "SUGGEST"
)

type SearchChannel struct {
	*Driver
}

func NewSearch(host string, port int, password string) (Searchable, error) {
	driver := &Driver{
		Host:     host,
		Port:     port,
		Password: password,
		Channel:  Search,
	}
	err := driver.connect()
	if err != nil {
		return nil, err
	}
	return SearchChannel{
		Driver: driver,
	}, nil
}

func (s SearchChannel) Query(collection, bucket, term string, limit, offset int) (results []string, err error) {
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

func (s SearchChannel) Suggest(collection, bucket, word string, limit int) (results []string, err error) {
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
