package directors

import (
	"net/http"
	"net/url"
)

// NewSingleHostDirector returns a new func(req *http.Request) that routes
// URLs to the scheme, host, and base path provided in target. If the
// target's path is "/base" and the incoming request was for "/dir",
// the target request will be for /base/dir.
// NewSingleHostDirector does not rewrite the Host header.
func NewSingleHostDirector(target *url.URL) func(req *http.Request) {
	return func(req *http.Request) {
		RewriteRequest(target, req)
	}
}
