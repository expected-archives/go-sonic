package sonic

import (
	"fmt"
	"strconv"
	"strings"
)

type IngestBulkRecord struct {
	Object, Text string
}

type IngestBulkError struct {
	Object string
	Error  error
}

// Ingestable is used for altering the search index (push, pop and flush).
type Ingestable interface {
	// Push search data in the index.
	// Command syntax PUSH <collection> <bucket> <object> "<text>"
	Push(collection, bucket, object, text string) (err error)

	// BulkPush will execute N (parallelRoutines) goroutines at the same time to
	// dispatch the records at best.
	// If parallelRoutines <= 0; parallelRoutines will be equal to 1.
	// If parallelRoutines > len(records); parallelRoutines will be equal to len(records).
	BulkPush(collection, bucket string, records []IngestBulkRecord) []IngestBulkError

	// Pop search data from the index.
	// Command syntax POP <collection> <bucket> <object> "<text>".
	Pop(collection, bucket, object, text string) (err error)

	// BulkPop will execute N (parallelRoutines) goroutines at the same time to
	// dispatch the records at best.
	// If parallelRoutines <= 0; parallelRoutines will be equal to 1.
	// If parallelRoutines > len(records); parallelRoutines will be equal to len(records).
	BulkPop(collection, bucket string, records []IngestBulkRecord) []IngestBulkError

	// Count indexed search data.
	// bucket and object are optionals, empty string ignore it.
	// Command syntax COUNT <collection> [<bucket> [<object>]?]?.
	Count(collection, bucket, object string) (count int, err error)

	// FlushCollection Flush all indexed data from a collection.
	// Command syntax FLUSHC <collection>.
	FlushCollection(collection string) (err error)

	// Flush all indexed data from a bucket in a collection.
	// Command syntax FLUSHB <collection> <bucket>.
	FlushBucket(collection, bucket string) (err error)

	// Flush all indexed data from an object in a bucket in collection.
	// Command syntax FLUSHO <collection> <bucket> <object>.
	FlushObject(collection, bucket, object string) (err error)

	// Quit refer to the Base interface
	Quit() (err error)

	// Quit refer to the Base interface
	Ping() (err error)
}

type ingesterCommands string

const (
	push   ingesterCommands = "PUSH"
	pop    ingesterCommands = "POP"
	count  ingesterCommands = "COUNT"
	flushb ingesterCommands = "FLUSHB"
	flushc ingesterCommands = "FLUSHC"
	flusho ingesterCommands = "FLUSHO"
)

type IngesterChannel struct {
	*Driver
}

func NewIngester(host string, port int, password string) (Ingestable, error) {
	driver := &Driver{
		Host:     host,
		Port:     port,
		Password: password,
		channel:  Ingest,
	}
	err := driver.Connect()
	if err != nil {
		return nil, err
	}
	return IngesterChannel{
		Driver: driver,
	}, nil
}

func (i IngesterChannel) Push(collection, bucket, object, text string) (err error) {
	err = i.write(fmt.Sprintf("%s %s %s %s \"%s\"", push, collection, bucket, object, text))
	if err != nil {
		return err
	}

	// sonic should sent OK
	_, err = i.read()
	if err != nil {
		return err
	}
	return nil
}

func (i IngesterChannel) BulkPush(collection, bucket string, records []IngestBulkRecord) []IngestBulkError {
	errs := make([]IngestBulkError, 0)

	for _, v := range records {
		if err := i.Push(collection, bucket, v.Object, v.Text); err != nil {
			errs = append(errs, IngestBulkError{v.Object, err})
		}
	}

	return errs
}

func (i IngesterChannel) Pop(collection, bucket, object, text string) (err error) {
	err = i.write(fmt.Sprintf("%s %s %s %s \"%s\"", pop, collection, bucket, object, text))
	if err != nil {
		return err
	}

	// sonic should sent OK
	_, err = i.read()
	if err != nil {
		return err
	}
	return nil
}

func (i IngesterChannel) BulkPop(collection, bucket string, records []IngestBulkRecord) []IngestBulkError {
	errs := make([]IngestBulkError, 0)

	for _, v := range records {
		if err := i.Push(collection, bucket, v.Object, v.Text); err != nil {
			errs = append(errs, IngestBulkError{v.Object, err})
		}
	}

	return errs
}

func (i IngesterChannel) Count(collection, bucket, object string) (cnt int, err error) {
	err = i.write(fmt.Sprintf("%s %s %s", count, collection, buildCountQuery(bucket, object)))
	if err != nil {
		return 0, err
	}

	// RESULT NUMBER
	r, err := i.read()
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(r[7:])
}

func buildCountQuery(bucket, object string) string {
	builder := strings.Builder{}
	if bucket != "" {
		builder.WriteString(bucket)
		if object != "" {
			builder.WriteString(" " + object)
		}
	}
	return builder.String()
}

func (i IngesterChannel) FlushCollection(collection string) (err error) {
	err = i.write(fmt.Sprintf("%s %s", flushc, collection))
	if err != nil {
		return err
	}

	// sonic should sent OK
	_, err = i.read()
	if err != nil {
		return err
	}
	return nil
}

func (i IngesterChannel) FlushBucket(collection, bucket string) (err error) {
	err = i.write(fmt.Sprintf("%s %s %s", flushb, collection, bucket))
	if err != nil {
		return err
	}

	// sonic should sent OK
	_, err = i.read()
	if err != nil {
		return err
	}
	return nil
}

func (i IngesterChannel) FlushObject(collection, bucket, object string) (err error) {
	err = i.write(fmt.Sprintf("%s %s %s %s", flusho, collection, bucket, object))
	if err != nil {
		return err
	}

	// sonic should sent OK
	_, err = i.read()
	if err != nil {
		return err
	}
	return nil
}

func (i IngesterChannel) divideIngestBulkRecords(records []IngestBulkRecord, parallelRoutines int) [][]IngestBulkRecord {
	var divided [][]IngestBulkRecord
	chunkSize := (len(records) + parallelRoutines - 1) / parallelRoutines
	for i := 0; i < len(records); i += chunkSize {
		end := i + chunkSize
		if end > len(records) {
			end = len(records)
		}
		divided = append(divided, records[i:end])
	}
	return divided
}
