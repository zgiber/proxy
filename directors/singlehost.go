package directors

import (
	"log"
	"net/http"
	"net/url"
)

// NewSingleHost returns a new func(req *http.Request) that routes
// requests to the scheme, host, and path path provided in target.
func NewSingleHost(targetURL string) func(req *http.Request) {
	target, err := url.Parse(targetURL)
	if err != nil {
		log.Fatal(err)
	}

	return func(req *http.Request) {
		rewriteRequest(target, req)
	}
}
