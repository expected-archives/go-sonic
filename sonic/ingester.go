package sonic

import (
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"
)

// IngestBulkRecord is the struct to be used as a list in bulk operation.
type IngestBulkRecord struct {
	Object, Text string
}

// IngestBulkError represent an error for a given object in a bulk operation.
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
	BulkPush(collection, bucket string, parallelRoutines int, records []IngestBulkRecord) []IngestBulkError

	// Pop search data from the index.
	// Command syntax POP <collection> <bucket> <object> "<text>".
	Pop(collection, bucket, object, text string) (err error)

	// BulkPop will execute N (parallelRoutines) goroutines at the same time to
	// dispatch the records at best.
	// If parallelRoutines <= 0; parallelRoutines will be equal to 1.
	// If parallelRoutines > len(records); parallelRoutines will be equal to len(records).
	BulkPop(collection, bucket string, parallelRoutines int, records []IngestBulkRecord) []IngestBulkError

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

	// Ping refer to the Base interface
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

type ingesterChannel struct {
	*driver
}

// NewIngester create a new driver instance with a ingesterChannel instance.
// Only way to get a Ingestable implementation.
func NewIngester(host string, port int, password string) (Ingestable, error) {
	driver := &driver{
		Host:     host,
		Port:     port,
		Password: password,
		channel:  Ingest,
	}
	err := driver.Connect()
	if err != nil {
		return nil, err
	}
	return ingesterChannel{
		driver: driver,
	}, nil
}

func (i ingesterChannel) Push(collection, bucket, object, text string) (err error) {
	//
	patterns := []struct {
		Pattern     string
		Replacement string
	}{{"\\", "\\\\"},
		{"\n", "\\n"},
		{"\"", "\\\""}}
	for _, v := range patterns {
		text = strings.Replace(text, v.Pattern, v.Replacement, -1)
	}

	chunks := splitText(text, i.cmdMaxBytes/2)
	// split chunks with partial success will yield single error
	for _, chunk := range chunks {
		err = i.write(fmt.Sprintf("%s %s %s %s \"%s\"", push, collection, bucket, object, chunk))

		if err != nil {
			return err
		}

		// sonic should sent OK
		_, err = i.read()
		if err != nil {
			return err
		}
	}

	return nil
}

// Ensure splitting on a valid leading byte
// Slicing the string directly is more efficient than converting to []byte and back because
// since a string is immutable and a []byte isn't,
// the data must be copied to new memory upon conversion,
// taking O(n) time (both ways!),
// whereas slicing a string simply returns a new string header backed by the same array as the original
// (taking constant time).
func splitText(longString string, maxLen int) []string {
	splits := []string{}

	var l, r int
	for l, r = 0, maxLen; r < len(longString); l, r = r, r+maxLen {
		for !utf8.RuneStart(longString[r]) {
			r--
		}
		splits = append(splits, longString[l:r])
	}
	splits = append(splits, longString[l:])
	return splits
}

func (i ingesterChannel) BulkPush(collection, bucket string, parallelRoutines int, records []IngestBulkRecord) (errs []IngestBulkError) {
	if parallelRoutines <= 0 {
		parallelRoutines = 1
	}

	// chunk array into N (parallelRoutines) parts
	divided := divideIngestBulkRecords(records, parallelRoutines)

	bulkWriteChan := make(chan []IngestBulkRecord, len(divided))
	bulkErrorChan := make(chan []IngestBulkError)

	defer close(bulkWriteChan)
	defer close(bulkErrorChan)

	for _, r := range divided {
		bulkWriteChan <- r
		go bulkPushRoutine(i.driver, collection, bucket, bulkWriteChan, bulkErrorChan)
	}

	errs = make([]IngestBulkError, 0)
	for range divided {
		for _, e := range <-bulkErrorChan {
			errs = append(errs, e)
		}
	}

	return errs
}

// bulkPushRoutine creates a new ingester instance using the provided driver
// and pushes all the records received from bulkWriteChan
// any errors are sent to bulkErrorChan
func bulkPushRoutine(driver *driver, collection, bucket string, bulkWriteChan <-chan []IngestBulkRecord, bulkErrorChan chan<- []IngestBulkError) {
	recs := <-bulkWriteChan
	errs := make([]IngestBulkError, 0)
	newIngester, _ := NewIngester(driver.Host, driver.Port, driver.Password)

	for _, rec := range recs {
		if newIngester == nil {
			addBulkError(&errs, rec, ErrClosed)
			continue
		}
		err := newIngester.Push(collection, bucket, rec.Object, rec.Text)
		if err != nil {
			addBulkError(&errs, rec, err)
		}
	}

	if newIngester != nil {
		_ = newIngester.Quit()
	}
	bulkErrorChan <- errs
}

func (i ingesterChannel) Pop(collection, bucket, object, text string) (err error) {
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

func (i ingesterChannel) BulkPop(collection, bucket string, parallelRoutines int, records []IngestBulkRecord) (errs []IngestBulkError) {
	if parallelRoutines <= 0 {
		parallelRoutines = 1
	}

	// chunk array into N (parallelRoutines) parts
	divided := divideIngestBulkRecords(records, parallelRoutines)

	bulkPopChan := make(chan []IngestBulkRecord, len(divided))
	bulkErrorChan := make(chan []IngestBulkError)

	defer close(bulkPopChan)
	defer close(bulkErrorChan)

	for _, r := range divided {
		bulkPopChan <- r
		go bulkPopRoutine(i.driver, collection, bucket, bulkPopChan, bulkErrorChan)
	}

	errs = make([]IngestBulkError, 0)
	for range divided {
		for _, e := range <-bulkErrorChan {
			errs = append(errs, e)
		}
	}

	close(bulkPopChan)
	close(bulkErrorChan)
	return errs
}

// bulkPopRoutine creates a new ingester instance using the provided driver
// and pops all the records received from bulkPopChan
// any errors are sent to bulkErrorChan
func bulkPopRoutine(driver *driver, collection, bucket string, bulkPopChan <-chan []IngestBulkRecord, bulkErrorChan chan<- []IngestBulkError) {
	recs := <-bulkPopChan
	errs := make([]IngestBulkError, 0)
	newIngester, _ := NewIngester(driver.Host, driver.Port, driver.Password)

	for _, rec := range recs {
		if newIngester == nil {
			addBulkError(&errs, rec, ErrClosed)
			continue
		}

		err := newIngester.Pop(collection, bucket, rec.Object, rec.Text)

		if err != nil {
			addBulkError(&errs, rec, err)
		}
	}

	if newIngester != nil {
		_ = newIngester.Quit()
	}

	bulkErrorChan <- errs
}

func (i ingesterChannel) Count(collection, bucket, object string) (cnt int, err error) {
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

func (i ingesterChannel) FlushCollection(collection string) (err error) {
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

func (i ingesterChannel) FlushBucket(collection, bucket string) (err error) {
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

func (i ingesterChannel) FlushObject(collection, bucket, object string) (err error) {
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

func divideIngestBulkRecords(records []IngestBulkRecord, parallelRoutines int) [][]IngestBulkRecord {
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

func addBulkError(e *[]IngestBulkError, record IngestBulkRecord, err error) {
	*e = append(*e, IngestBulkError{record.Object, err})
}
