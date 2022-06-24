package sonic_test

import (
	"math/rand"
	"runtime"
	"strconv"
	"testing"
	"time"

	"github.com/expectedsh/go-sonic/sonic"
)

var records = make([]sonic.IngestBulkRecord, 0)

func getIngester(tb testing.TB) sonic.Ingestable {
	tb.Helper()

	host, port, pass := getSonicConfig(tb)

	ing, err := sonic.NewIngester(host, port, pass)
	if err != nil {
		tb.Fatal(err)
	}

	return ing
}

func BenchmarkIngesterChannel_BulkPush2XMaxCPUs(b *testing.B) {
	cpus := 2 * runtime.NumCPU()
	ingester := getIngester(b)

	for n := 0; n < b.N; n++ {
		e := ingester.FlushBucket("test", "test2XMaxCpus")
		if e != nil {
			b.Log(e)
			b.Fail()
		}
		be := ingester.BulkPush("test", "test2XMaxCpus", cpus, records, sonic.LangAutoDetect)
		if len(be) > 0 {
			b.Log(be, e)
			b.Fail()
		}
	}
}

func BenchmarkIngesterChannel_BulkPushMaxCPUs(b *testing.B) {
	ingester := getIngester(b)

	cpus := runtime.NumCPU()

	for n := 0; n < b.N; n++ {
		e := ingester.FlushBucket("test", "testMaxCpus")
		if e != nil {
			b.Log(e)
			b.Fail()
		}
		be := ingester.BulkPush("test", "testMaxCpus", cpus, records, sonic.LangAutoDetect)
		if len(be) > 0 {
			b.Log(be, e)
			b.Fail()
		}
	}
}

func BenchmarkIngesterChannel_BulkPush10(b *testing.B) {
	ingester := getIngester(b)

	for n := 0; n < b.N; n++ {
		e := ingester.FlushBucket("test", "test10")
		if e != nil {
			b.Log(e)
			b.Fail()
		}
		be := ingester.BulkPush("test", "test10", 10, records, sonic.LangAutoDetect)
		if len(be) > 0 {
			b.Fail()
		}
	}
}

func BenchmarkIngesterChannel_BulkPop10(b *testing.B) {
	ingester := getIngester(b)

	for n := 0; n < b.N; n++ {
		e := ingester.FlushBucket("test", "popTest10")
		if e != nil {
			b.Log(e)
			b.Fail()
		}
		be := ingester.BulkPop("test", "popTest10", 10, records)
		if len(be) > 0 {
			b.Fail()
		}
	}
}

func BenchmarkIngesterChannel_Push(b *testing.B) {
	ingester := getIngester(b)

	for n := 0; n < b.N; n++ {
		e := ingester.FlushBucket("test", "testBulk")
		if e != nil {
			b.Log(e)
			b.Fail()
		}
		for _, v := range records {
			e := ingester.Push("test", "testBulk", v.Object, v.Text, sonic.LangAutoDetect)
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
		records = append(records, sonic.IngestBulkRecord{randStr(10, charset), randStr(10, charset)})
	}
}

func TestIngester_Push_Count(t *testing.T) {
	t.Parallel()

	host, port, pass := getSonicConfig(t)

	col := t.Name()
	bucket := strconv.FormatInt(time.Now().UnixNano(), 10)

	ing, err := sonic.NewIngester(host, port, pass)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		_ = ing.FlushBucket(col, bucket)
		_ = ing.Quit()
	})

	err = ing.Push(col, bucket, "obj", "test", sonic.LangAutoDetect)
	if err != nil {
		t.Fatal("Push", err)
	}

	count, err := ing.Count(col, bucket, "obj")
	switch {
	case err != nil:
		t.Fatal("Count", err)
	case count != 1:
		t.Fatalf("Actual: %d, expected: %d", count, 1)
	}
}

func TestIngester_BulkPush_Count(t *testing.T) {
	t.Parallel()

	host, port, pass := getSonicConfig(t)

	col := t.Name()
	bucket := strconv.FormatInt(time.Now().UnixNano(), 10)

	ing, err := sonic.NewIngester(host, port, pass)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		_ = ing.FlushBucket(col, bucket)
		_ = ing.Quit()
	})

	errs := ing.BulkPush(col, bucket, 4, []sonic.IngestBulkRecord{{
		Object: "obj1",
		Text:   "test",
	}, {
		Object: "obj2",
		Text:   "test",
	}, {
		Object: "obj3",
		Text:   "test",
	}, {
		Object: "obj4",
		Text:   "test",
	}}, sonic.LangAutoDetect)
	if len(errs) != 0 {
		t.Fatal("BlukPush", errs)
	}

	count, err := ing.Count(col, bucket, "obj4")
	switch {
	case err != nil:
		t.Fatal("Count", err)
	case count != 1:
		t.Fatalf("Actual: %d, expected: %d", count, 1)
	}
}
