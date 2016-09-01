package ratelimit

import (
	"errors"
	"strconv"
	"sync"
	"testing"
	"time"
)

var (
	clientCount     = 10
	clientCallCount = 20
	rps             = 10
	expire          = 1 * time.Second
	burst           = 5
	limiter         = NewClientLimiter(expire, rps, burst)
)

type testClient struct {
	id            string
	limiter       Limiter
	lastCallStart time.Time
	lastCallEnd   time.Time
	wait          time.Duration
	expire        time.Duration
}

func newTestClient(id string, maxRPS int, expire time.Duration) *testClient {
	wait := time.Duration(int64(1.0/float64(maxRPS)*1000)) * time.Millisecond
	return &testClient{
		id:      id,
		limiter: limiter, // test service doing nothing at all
		wait:    wait,
		expire:  expire,
	}
}

func (tc *testClient) call() error {

	switch now := time.Now(); {
	case now.After(tc.lastCallEnd.Add(tc.expire)):
		tc.lastCallStart = now

		// first call since expiration should take less time than tc.wait
		tc.limiter.Limit(tc.id)
		if time.Since(now) > tc.wait {
			return errors.New("First call took too long (longer than limiter's wait duration).")
		}

	case now.After(tc.lastCallStart.Add(tc.wait)):
		tc.lastCallStart = now

		// subsequent call should end after lastCallEnd + tc.wait
		tc.limiter.Limit(tc.id)
		if time.Now().Before(tc.lastCallEnd.Add(tc.wait)) {
			return errors.New("Limited call returned before limiter's wait duration.")
		}
	default:
		tc.lastCallStart = now
		tc.limiter.Limit(tc.id)
	}

	tc.lastCallEnd = time.Now()
	return nil
}

func TestClientLimiter(t *testing.T) {

	var wg sync.WaitGroup
	for i := 0; i < clientCount; i++ {
		client := newTestClient(strconv.Itoa(i), rps, expire)
		wg.Add(1)
		go func(c *testClient) {
			for j := 0; j < clientCallCount; j++ {
				err := client.call()
				if err != nil {
					t.Fatal(err)
				}
				// fmt.Print(c.id)
			}
			time.Sleep(2 * time.Second)
			wg.Done()
		}(client)
	}

	wg.Wait()
}
