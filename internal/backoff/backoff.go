package backoff

import "time"

const (
	DefaultMaxRetry = 5
	DefaultInterval = 500 * time.Millisecond
)

type Backoff struct {
	description string
	maxAttempt  int
	attempt     int
	function    func() error
	interval    time.Duration
}

// New return a new Backoff
// entry can be null, by default logs will appear by default only if you define
// an entry
func New(description string, run func() error) *Backoff {
	return &Backoff{
		description: description,
		maxAttempt:  DefaultMaxRetry,
		attempt:     1,
		function:    run,
		interval:    DefaultInterval,
	}
}

func (s *Backoff) Clone() *Backoff {
	return &Backoff{
		maxAttempt: s.maxAttempt,
		attempt:    s.attempt,
		interval:   s.interval,
		function:   s.function,
	}
}

func (s *Backoff) WithMaxAttempt(maxRetry int) *Backoff {
	s.maxAttempt = maxRetry
	return s
}

func (s *Backoff) WithInterval(interval time.Duration) *Backoff {
	s.interval = interval
	return s
}

func (s *Backoff) Run() error {
	for {
		if err := s.function(); err != nil {
			time.Sleep(time.Duration(fibonacci(s.attempt)) * s.interval)
			if s.attempt == s.maxAttempt {
				return err
			}
			s.attempt++
			continue
		}
		return nil
	}
}

func fibonacci(n int) int {
	f := make([]int, n+1, n+2)
	if n < 2 {
		f = f[0:2]
	}
	f[0] = 0
	f[1] = 1
	for i := 2; i <= n; i++ {
		f[i] = f[i-1] + f[i-2]
	}
	return f[n]
}
