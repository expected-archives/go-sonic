package sonic_test

import (
	"net"
	"net/url"
	"os"
	"strconv"
	"testing"
)

func getSonicConfig(tb testing.TB) (host string, portNum int, password string) {
	tb.Helper()

	const envAddr = "TEST_SONIC_ADDR"

	sonicAddr := os.Getenv(envAddr)
	if sonicAddr == "" {
		tb.Fatal(envAddr + " is not set")
	}

	u, err := url.Parse(sonicAddr)
	if err != nil {
		tb.Fatalf("parsing url: %s", err)
	}

	portNum, err = strconv.Atoi(u.Port())
	if err != nil {
		tb.Fatalf("parsing port: %s", err)
	}

	pass, _ := u.User.Password()

	return u.Hostname(), portNum, pass
}

func mustSplitHostPort(tb testing.TB, hostport string) (host string, port int) {
	tb.Helper()

	host, portStr, err := net.SplitHostPort(hostport)
	if err != nil {
		tb.Fatal("SplitHostPort", err)
	}

	proxyPort, err := strconv.Atoi(portStr)
	if err != nil {
		tb.Fatal("Atoi", err)
	}

	return host, proxyPort
}
