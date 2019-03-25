package sonic

import "fmt"

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
	*Connection
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
	panic("implement me")
}

func (i IngesterChannel) Count(collection, bucket, object string) (count int, err error) {
	panic("implement me")
}

func (i IngesterChannel) FlushCollection(collection string) (err error) {
	panic("implement me")
}

func (i IngesterChannel) FlushBucket(collection, bucket string) (err error) {
	panic("implement me")
}

func (i IngesterChannel) FlushObject(collection, bucket, object string) (err error) {
	panic("implement me")
}
