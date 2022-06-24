package sonic_test

import (
	"strconv"
	"testing"
	"time"

	"github.com/expectedsh/go-sonic/sonic"
)

func getSearch(tb testing.TB) sonic.Searchable {
	tb.Helper()

	host, port, pass := getSonicConfig(tb)

	srch, err := sonic.NewSearch(host, port, pass)
	if err != nil {
		tb.Fatal(err)
	}

	return srch
}

func TestSearch(t *testing.T) {
	t.Parallel()

	col := t.Name()
	bucket := strconv.FormatInt(time.Now().UnixNano(), 10)

	ing := getIngester(t)
	srch := getSearch(t)

	t.Cleanup(func() {
		_ = ing.FlushBucket(col, bucket)
		_ = ing.Quit()
		_ = srch.Quit()
	})

	err := ing.Push(col, bucket, "obj1", "xxx", sonic.LangAutoDetect)
	if err != nil {
		t.Fatal("Push", err)
	}

	err = ing.Push(col, bucket, "obj1", "yyy", sonic.LangAutoDetect)
	if err != nil {
		t.Fatal("Push", err)
	}

	t.Run("Query_ok", func(t *testing.T) {
		t.Parallel()

		res, err := srch.Query(col, bucket, "xxx", 1, 0, sonic.LangAutoDetect)
		switch {
		case err != nil:
			t.Fatal("Query", err)
		case len(res) != 1:
			t.Fatalf("Actual: %d, expected: %d", len(res), 1)
		}
	})

	t.Run("Query_empty", func(t *testing.T) {
		t.Parallel()

		res, err := srch.Query(col, bucket, "zzz", 1, 0, sonic.LangAutoDetect)
		switch {
		case err != nil:
			t.Fatal("Query", err)
		case len(res) > 1 && res[0] != "":
			t.Fatalf("Actual: %d, expected: %d (%+v)", len(res), 0, res)
		}
	})

	t.Run("Query_outOfOffset", func(t *testing.T) {
		t.Parallel()

		res, err := srch.Query(col, bucket, "xxx", 1, 1, sonic.LangAutoDetect)
		switch {
		case err != nil:
			t.Fatal("Query", err)
		case len(res) > 1 && res[0] != "":
			t.Fatalf("Actual: %d, expected: %d (%+v)", len(res), 0, res)
		}
	})

	t.Run("Suggest", func(t *testing.T) {
		t.Parallel()

		res, err := srch.Suggest(col, bucket, "xx", 1)
		switch {
		case err != nil:
			t.Fatal("Query", err)
		case len(res) != 1:
			t.Fatalf("Actual: %d, expected: %d (%+v)", len(res), 0, res)
		}
	})
}
