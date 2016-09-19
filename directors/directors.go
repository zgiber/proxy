package directors

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// RewriteRequest replaces target url in the request with the provided one
func RewriteRequest(target *url.URL, req *http.Request) {
	fmt.Println("DEBUG - rewriteRequest")
	targetQuery := target.RawQuery
	req.URL.Scheme = target.Scheme
	req.URL.Host = target.Host
	req.URL.Path = SingleJoiningSlash(target.Path, req.URL.Path)
	if targetQuery == "" || req.URL.RawQuery == "" {
		req.URL.RawQuery = targetQuery + req.URL.RawQuery
	} else {
		req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
	}
	if _, ok := req.Header["User-Agent"]; !ok {
		// explicitly disable User-Agent so it's not set to default value
		req.Header.Set("User-Agent", "")
	}
	fmt.Println("DEBUG - rewriteRequest DONE")
	fmt.Println(req.URL.Scheme, req.URL.Host, req.URL.RawPath, req.URL.RawQuery)
}

// SingleJoiningSlash properly joins two paths
// handling trailing and leading '/' characters.
func SingleJoiningSlash(a, b string) string {
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
