package sonic

import (
	"fmt"
	"strconv"
	"strings"
)

type Ingestable interface {
	Push(collection, bucket, object, text string) (err error)
	Pop(collection, bucket, object, text string) (err error)
	Count(collection, bucket, object string) (count int, err error)

	FlushCollection(collection string) (err error)
	FlushBucket(collection, bucket string) (err error)
	FlushObject(collection, bucket, object string) (err error)
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
		Channel:  Ingest,
	}
	err := driver.connect()
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
