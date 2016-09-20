package directors

import (
	"net/http"
	"net/url"
	"strings"
)

// RewriteRequest replaces the url in the request with the provided target
// The followings are considered:
// target.Scheme, target.Host, target.Path
// Empty values on the target are ignored.
func RewriteRequest(target *url.URL, req *http.Request) {

	if target.Scheme != "" {
		req.URL.Scheme = target.Scheme
	}

	if target.Host != "" {
		req.URL.Host = target.Host
	}

	if target.Path != "" {
		RewriteRequestPath(target.Path, req)
	}

	if _, ok := req.Header["User-Agent"]; !ok {
		// explicitly disable User-Agent so it's not set to default value
		req.Header.Set("User-Agent", "")
	}
}

// RewriteRequestPath changes the request's path to the target path.
// The target path may have variables defined e.g.: "/users/:user_id/details"
// In this case the request context is expected to have a value for the
// key 'user_id'.
//
// TODO: If there is no such value.
func RewriteRequestPath(targetPath string, req *http.Request) {
	pathSegments := strings.Split(strings.Trim(targetPath, "/"), "/")

	for i := 0; i < len(pathSegments); i++ {
		segment := pathSegments[i]
		if strings.HasPrefix(segment, ":") {
			value := req.Context().Value(segment[1:])
			if stringValue, ok := value.(string); ok {
				pathSegments[i] = stringValue
			}
		}
		if segment == "+" {
			pathSegments[i] = strings.Trim(req.URL.Path, "/")
		}
	}

	newPath := strings.Join(pathSegments, "/")

	req.URL.Path = newPath
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
