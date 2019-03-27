package sonic

import (
	"math/rand"
	"runtime"
	"testing"
	"time"
)

var records = make([]IngestBulkRecord, 0)
var ingester, err = NewIngester("localhost", 1491, "SecretPassword")

func BenchmarkIngesterChannel_BulkPushMaxCPUs(b *testing.B) {
	if err != nil {
		return
	}

	cpus := runtime.NumCPU()

	for n := 0; n < b.N; n++ {
		_ = ingester.FlushBucket("test", "testMaxCpus")
		ingester.BulkPush("test", "testMaxCpus", cpus, records)
	}
}

func BenchmarkIngesterChannel_Push(b *testing.B) {
	if err != nil {
		return
	}

	for n := 0; n < b.N; n++ {
		_ = ingester.FlushBucket("test", "testMaxCpus")
		for _, v := range records {
			_ = ingester.Push("test", "testMaxCpus", v.Object, v.Text)
		}
	}
}

func BenchmarkIngesterChannel_BulkPush10(b *testing.B) {
	if err != nil {
		return
	}

	for n := 0; n < b.N; n++ {
		_ = ingester.FlushBucket("test", "testMaxCpus")
		ingester.BulkPush("test", "testMaxCpus", 10, records)
	}
}

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_"

var seededRand = rand.New(rand.NewSource(time.Now().UnixNano()))

func randStr(length int, charset string) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func init() {
	for n := 0; n < 1000; n++ {
		records = append(records, IngestBulkRecord{randStr(10, charset), randStr(10, charset)})
	}
}
