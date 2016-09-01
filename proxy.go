package proxy

import (
	"context"
	"errors"
	"net/http"
	"net/http/httputil"
)

type roundTripper struct {
	rt http.RoundTripper
}

// ReverseProxy is the same as httputil.ReverseProxy
// except that it uses a wrapped Transport, which
// handles errors created by a Director
type ReverseProxy struct {
	*httputil.ReverseProxy
}

func New() *ReverseProxy {
	return &ReverseProxy{
		&httputil.ReverseProxy{
			Transport: newRoundTripper(http.DefaultTransport),
		},
	}
}

func newRoundTripper(t http.RoundTripper) http.RoundTripper {
	return &roundTripper{t}
}

func (rt *roundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if ctx := req.Context(); ctx.Err() != nil {
		return nil, errorFromContext(ctx)
	}
	return rt.rt.RoundTrip(req)
}

func errorFromContext(ctx context.Context) error {
	errVal := ctx.Value("error")
	switch err := errVal.(type) {
	case error:
		return err
	default:
		return errors.New("context expired") // TODO come up with something neater
	}
}
