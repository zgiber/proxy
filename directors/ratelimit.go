package directors

import (
	"net/http"
	"time"

	"github.com/Typeform/ratelimit"
)

func NewRateLimiter() func(*http.Request) {
	limiter := ratelimit.NewClientLimiter(100*time.Millisecond, 30*time.Second, 10)

	return func(req *http.Request) {
		limiter.Wait(req.RemoteAddr) // RemoteAddr won't work properly, it's here just for illustration. A truly unique ID is required.
	}
}
