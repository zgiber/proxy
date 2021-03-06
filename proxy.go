package proxy

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"

	"github.com/zgiber/proxy/directors"
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
	if director == nil {
		log.Fatal("director must be non nil")
	}

	if rp.Director == nil {
		rp.Director = director
		return
	}

	rp.Director = directors.Chain(rp.Director, director)
}

// AddDynamicDirector registers a director on the reverseproxy and
// registers the given http.Handlers on the configAPI http server.
// This way we can provide a http configuration interface for
// directors to be changed/configured on the fly.
func (rp *ReverseProxy) AddDynamicDirector(
	directorConfigPath string,
	directorConfigHandler http.Handler,
	director func(req *http.Request),
) {
	rp.configAPI.Handle(directorConfigPath, directorConfigHandler)
	rp.Director = directors.Chain(rp.Director, director)
}

// ListenAndServeDirectorConfig starts the http server for the configuration
// interface on the given addr.
func (rp *ReverseProxy) ListenAndServeDirectorConfig(addr string) error {
	return http.ListenAndServe(addr, rp.configAPI)
}

// ListenAndServeDirectorConfigTLS starts the https server for the configuration
// interface on the given addr.
func (rp *ReverseProxy) ListenAndServeDirectorConfigTLS(addr, certFile, keyFile string) error {
	return http.ListenAndServeTLS(addr, certFile, keyFile, rp.configAPI)
}

func newRoundTripper(t http.RoundTripper) http.RoundTripper {
	return &roundTripper{t}
}

func (rt *roundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	fmt.Printf("% v", req)
	if ctx := req.Context(); ctx.Err() != nil {
		return nil, errorFromContext(ctx)
		// TODO: return appropriate responses for certain errors
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
