package sonic_test

import (
	"errors"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/expectedsh/go-sonic/sonic"
)

func TestPool_Reconnect(t *testing.T) {
	host, port, pass := getSonicConfig(t)

	proxyLn, proxyDoneCh := runTCPProxy(t,
		fmt.Sprintf("%s:%d", host, port), // Target addr.
		"127.0.0.1:0",                    // Proxy addr.
	)

	proxyHost, proxyPort := mustSplitHostPort(t, proxyLn.Addr().String())

	ing, err := sonic.NewIngester(
		proxyHost,
		proxyPort,
		pass,
		sonic.OptionPoolPingThreshold(time.Nanosecond),
	)
	if err != nil {
		t.Fatal("NewIngester", err)
	}

	// Connection healthy, ping should work.

	err = ing.Ping()
	if err != nil {
		t.Fatal("Ping", err)
	}

	err = ing.Ping()
	if err != nil {
		t.Fatal("Ping", err)
	}

	// Close connection, ping should not work.

	err = proxyLn.Close()
	if err != nil {
		t.Fatal("Close", err)
	}

	select {
	case <-proxyDoneCh:
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout")
	}

	err = ing.Ping()
	if err == nil {
		t.Fatal("Ping", err)
	}

	// Reconnect, ping should work.

	proxyLn, proxyDoneCh = runTCPProxy(t,
		fmt.Sprintf("%s:%d", host, port),           // Target addr.
		fmt.Sprintf("%s:%d", proxyHost, proxyPort), // Proxy addr.
	)

	err = ing.Ping()
	if err != nil {
		t.Fatal("Ping", err)
	}

	err = proxyLn.Close()
	if err != nil {
		t.Fatal("Close", err)
	}

	select {
	case <-proxyDoneCh:
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout")
	}
}

func TestPool_Reconnect_Threshold(t *testing.T) {
	host, port, pass := getSonicConfig(t)

	proxyLn, proxyDoneCh := runTCPProxy(t,
		fmt.Sprintf("%s:%d", host, port), // Target addr.
		"127.0.0.1:0",                    // Proxy addr.
	)

	proxyHost, proxyPort := mustSplitHostPort(t, proxyLn.Addr().String())

	ing, err := sonic.NewIngester(
		proxyHost,
		proxyPort,
		pass,
		sonic.OptionPoolPingThreshold(time.Minute),
	)
	if err != nil {
		t.Fatal("NewIngester", err)
	}

	// Connection healthy, ping should work.

	err = ing.Ping()
	if err != nil {
		t.Fatal("Ping", err)
	}

	err = ing.Ping()
	if err != nil {
		t.Fatal("Ping", err)
	}

	// Close connection, ping should not work.

	err = proxyLn.Close()
	if err != nil {
		t.Fatal("Close", err)
	}

	select {
	case <-proxyDoneCh:
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout")
	}

	err = ing.Ping()
	if err == nil {
		t.Fatal("Ping", err)
	}

	// Reconnect in threshold, ping still should not work.

	proxyLn, proxyDoneCh = runTCPProxy(t,
		fmt.Sprintf("%s:%d", host, port),           // Target addr.
		fmt.Sprintf("%s:%d", proxyHost, proxyPort), // Proxy addr.
	)

	err = ing.Ping()
	if err == nil {
		t.Fatal("Ping", err)
	}

	err = proxyLn.Close()
	if err != nil {
		t.Fatal("Close", err)
	}

	select {
	case <-proxyDoneCh:
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout")
	}
}

func runTCPProxy(
	tb testing.TB,
	targetAddr string,
	proxyAddr string,
) (l net.Listener, doneCh <-chan struct{}) {
	tb.Helper()

	l, err := net.Listen("tcp", proxyAddr)
	if err != nil {
		tb.Fatal("Listen", err)
	}

	tb.Cleanup(func() { _ = l.Close() })

	tb.Log("Proxy:", l.Addr().String(), "To:", l.Addr().String())

	closeCh := make(chan struct{})
	go func() {
		defer close(closeCh)

		for {
			inConn, err := l.Accept()
			if err != nil {
				if !errors.Is(err, net.ErrClosed) {
					tb.Log("Accept", err)
				}

				return
			}

			outConn, err := net.Dial("tcp", targetAddr)
			if err != nil {
				if !errors.Is(err, net.ErrClosed) {
					tb.Error("Dial", err)
				}

				return
			}

			defer func() { _ = inConn.Close() }()
			defer func() { _ = outConn.Close() }()

			go runTCPPipe(tb, inConn, outConn)
			go runTCPPipe(tb, outConn, inConn)
		}
	}()

	return l, closeCh
}

func runTCPPipe(tb testing.TB, inConn, outConn net.Conn) {
	connID := time.Now().UnixNano()
	tb.Log("Acepted new conn", connID)

	buff := make([]byte, 0xFF)
	for {
		_ = inConn.SetReadDeadline(time.Now().Add(5 * time.Second))

		n, err := inConn.Read(buff)
		if err != nil {
			if !errors.Is(err, net.ErrClosed) {
				tb.Error("inConn.ReadAll", connID, err)
			}

			return
		}

		_ = inConn.SetReadDeadline(time.Time{})
		_ = outConn.SetWriteDeadline(time.Now().Add(5 * time.Second))

		tb.Log("read", n, "bytes")

		n, err = outConn.Write(buff[:n])
		if err != nil {
			if !errors.Is(err, net.ErrClosed) {
				tb.Error("outConn.Write", connID, err)
			}

			return
		}

		tb.Log("wrote", n, "bytes")

		_ = outConn.SetWriteDeadline(time.Time{})
	}
}
