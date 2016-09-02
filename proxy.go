package proxy

import (
	"context"
	"errors"
	"net/http"
	"net/http/httputil"
	"strings"
)

type roundTripper struct {
	rt http.RoundTripper
}

// ReverseProxy is the same as httputil.ReverseProxy
// except that it uses a wrapped Transport, which
// handles errors created by a Director.
type ReverseProxy struct {
	*httputil.ReverseProxy
	configAPI *http.ServeMux
}

func New() *ReverseProxy {

	return &ReverseProxy{
		&httputil.ReverseProxy{
			Transport: newRoundTripper(http.DefaultTransport),
		},
		http.NewServeMux(),
	}
}

// AddDirector registers a director to be chained after the existing
// proxy director.
func (rp *ReverseProxy) AddDirector(director func(req *http.Request)) error {
	if rp.Director == nil {
		rp.Director = director
		return nil
	}

	d, err := ChainDirectors(rp.Director, director)
	if err != nil {
		rp.Director = d
	}
	return err
}

// AddDynamicDirector registers a director on the reverseproxy and
// registers the given http.Handlers on the configAPI http server.
// This way we can provide a http configuration interface for
// directors to be changed/configured on the fly.
func (rp *ReverseProxy) AddDynamicDirector(
	configPath string,
	configAPIHandler http.Handler,
	director func(req *http.Request),
) error {
	rp.configAPI.Handle(configPath, configAPIHandler)
	d, err := ChainDirectors(rp.Director, director)
	if err != nil {
		rp.Director = d
	}
	return err
}

// ListenAndServeConfigAPI starts the http server for the configuration
// interface on the given addr.
func (rp *ReverseProxy) ListenAndServeConfigAPI(addr string) error {
	return http.ListenAndServe(addr, rp.configAPI)
}

// ListenAndServeConfigAPITLS starts the https server for the configuration
// interface on the given addr.
func (rp *ReverseProxy) ListenAndServeConfigAPITLS(addr, certFile, keyFile string) error {
	return http.ListenAndServeTLS(addr, certFile, keyFile, rp.configAPI)
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

// blatantly coped from standard lib
func singleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	}
	return a + b
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
