package directors

import (
	"net/http"
	"time"

	"github.com/Typeform/ratelimit"
)

func NewRateLimiter(delay, timeout time.Duration, burst int) func(*http.Request) {
	limiter := ratelimit.NewClientLimiter(delay, timeout, burst)

	return func(req *http.Request) {
		limiter.Wait(req.RemoteAddr) // TODO: RemoteAddr won't work properly, it's here just for illustration. A truly unique ID is required.
	}
}
