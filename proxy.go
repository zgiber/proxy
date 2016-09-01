package proxy

import (
	"context"
	"errors"
	"net/http"
	"net/http/httputil"
	"net/url"
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
func (rp *ReverseProxy) AddDirector(director func(req *http.Request)) {
	rp.Director = ChainDirectors(rp.Director, director)
}

// AddDynamicDirector registers a director on the reverseproxy and
// registers the given http.Handlers on the configAPI http server.
// This way we can provide a http configuration interface for
// directors to be changed/configured on the fly.
func (rp *ReverseProxy) AddDynamicDirector(
	pattern string,
	configAPIHandler http.Handler,
	director func(req *http.Request),
) {
	rp.configAPI.Handle(pattern, configAPIHandler)
	rp.Director = ChainDirectors(rp.Director, director)
}

// ListenAndServeConfigAPI starts the http server for the configuration
// interface on the given addr.
func (rp *ReverseProxy) ListenAndServeConfigAPI(addr string) error {
	return http.ListenAndServe(addr, rp.configAPI)
}

// ListenAndServeConfigAPI starts the https server for the configuration
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

// ChainDirectors takes a number of directors and chains them, returning
// a single director.
func ChainDirectors(directors ...func(*http.Request)) func(*http.Request) {
	return func(req *http.Request) {

		ctx := req.Context()
		if ctx.Err() != nil {
			return
		}

		for _, director := range directors {
			if director == nil {
				continue
			}
			director(req)
			if req.Context().Err() != nil {
				return
				// error is handled by the RoundTripper
			}
		}
	}
}

// NewSingleHostDirector returns a new func(req *http.Request) that routes
// URLs to the scheme, host, and base path provided in target. If the
// target's path is "/base" and the incoming request was for "/dir",
// the target request will be for /base/dir.
// NewSingleHostDirector does not rewrite the Host header.
func NewSingleHostDirector(target *url.URL) func(req *http.Request) {
	return func(req *http.Request) {
		rewriteRequest(target, req)
	}
}

func NewRouterDirector(targets map[string]*url.URL) func(req *http.Request) {
	return func(req *http.Request) {
		if target, ok := targets[req.URL.Path]; ok {
			rewriteRequest(target, req)
		}
	}
}

func rewriteRequest(target *url.URL, req *http.Request) {
	targetQuery := target.RawQuery
	req.URL.Scheme = target.Scheme
	req.URL.Host = target.Host
	req.URL.Path = singleJoiningSlash(target.Path, req.URL.Path)
	if targetQuery == "" || req.URL.RawQuery == "" {
		req.URL.RawQuery = targetQuery + req.URL.RawQuery
	} else {
		req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
	}
	if _, ok := req.Header["User-Agent"]; !ok {
		// explicitly disable User-Agent so it's not set to default value
		req.Header.Set("User-Agent", "")
	}
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
