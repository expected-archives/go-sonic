package sonic

import (
	"math/rand"
	"runtime"
	"testing"
	"time"
)

var records = make([]IngestBulkRecord, 0)
var ingester, err = NewIngester("localhost", 1491, "SecretPassword")

func BenchmarkIngesterChannel_BulkPush2XMaxCPUs(b *testing.B) {
	if err != nil {
		return
	}

	cpus := 2 * runtime.NumCPU()

	for n := 0; n < b.N; n++ {
		e := ingester.FlushBucket("test", "test2XMaxCpus")
		if e != nil {
			b.Log(e)
			b.Fail()
		}
		be := ingester.BulkPush("test", "test2XMaxCpus", cpus, records)
		if len(be) > 0 {
			b.Log(be, e)
			b.Fail()
		}
	}
}

func BenchmarkIngesterChannel_BulkPushMaxCPUs(b *testing.B) {
	if err != nil {
		return
	}

	cpus := runtime.NumCPU()

	for n := 0; n < b.N; n++ {
		e := ingester.FlushBucket("test", "testMaxCpus")
		if e != nil {
			b.Log(e)
			b.Fail()
		}
		be := ingester.BulkPush("test", "testMaxCpus", cpus, records)
		if len(be) > 0 {
			b.Log(be, e)
			b.Fail()
		}
	}
}

func BenchmarkIngesterChannel_BulkPush10(b *testing.B) {
	if err != nil {
		return
	}

	for n := 0; n < b.N; n++ {
		e := ingester.FlushBucket("test", "test10")
		if e != nil {
			b.Log(e)
			b.Fail()
		}
		be := ingester.BulkPush("test", "test10", 10, records)
		if len(be) > 0 {
			b.Log(be, err)
			b.Fail()
		}
	}
}

func BenchmarkIngesterChannel_BulkPop10(b *testing.B) {
	if err != nil {
		return
	}

	for n := 0; n < b.N; n++ {
		e := ingester.FlushBucket("test", "popTest10")
		if e != nil {
			b.Log(e)
			b.Fail()
		}
		be := ingester.BulkPop("test", "popTest10", 10, records)
		if len(be) > 0 {
			b.Log(be, err)
			b.Fail()
		}
	}
}

func BenchmarkIngesterChannel_Push(b *testing.B) {
	if err != nil {
		return
	}

	for n := 0; n < b.N; n++ {
		e := ingester.FlushBucket("test", "testBulk")
		if e != nil {
			b.Log(e)
			b.Fail()
		}
		for _, v := range records {
			e := ingester.Push("test", "testBulk", v.Object, v.Text)
			if e != nil {
				b.Log(e)
				b.Fail()
			}
		}
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
	for n := 0; n < 3000; n++ {
		records = append(records, IngestBulkRecord{randStr(10, charset), randStr(10, charset)})
	}
}
