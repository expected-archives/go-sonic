package sonic_test

import (
	"testing"

	"github.com/expectedsh/go-sonic/sonic"
)

func TestController(t *testing.T) {
	t.Parallel()

	var ctrl sonic.Base = getIngester(t)

	err := ctrl.Ping()
	if err != nil {
		t.Fatal("Ping", err)
	}

	err = ctrl.Quit()
	if err != nil {
		t.Fatal("Quit", err)
	}

	err = ctrl.Ping()
	if err == nil {
		t.Fatal("Ping", err)
	}
}
