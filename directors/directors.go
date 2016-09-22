package directors

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// rewriteRequest replaces the url in the request with the provided target
// The followings are considered:
// target.Scheme, target.Host, target.Path
// Empty values on the target are ignored.
func rewriteRequest(target *url.URL, req *http.Request) {
	if target.Scheme != "" {
		// fmt.Println(target.Scheme)
		req.URL.Scheme = target.Scheme
	}

	if target.Host != "" {
		req.URL.Host = target.Host
	}

	if target.Path != "" {
		rewriteRequestPath(target.Path, req)
	}

	//req.URL.Path = SingleJoiningSlash(target.Path, req.URL.Path)
	if _, ok := req.Header["User-Agent"]; !ok {
		// explicitly disable User-Agent so it's not set to default value
		req.Header.Set("User-Agent", "")
	}

	fmt.Println("target Scheme", target.Scheme)
	fmt.Println("request Scheme", req.URL.Scheme)
}

// rewriteRequestPath changes the request's path to the target path.
// The target path may have variables defined e.g.: "/users/:user_id/details"
// In this case the request context is expected to have a value for the
// key 'user_id'.
//
// TODO: If there is no such value.
func rewriteRequestPath(targetPath string, req *http.Request) {
	pathSegments := strings.Split(strings.Trim(targetPath, "/"), "/")

	for i := 0; i < len(pathSegments); i++ {
		segment := pathSegments[i]

		if strings.HasPrefix(segment, ":") {
			value := req.Context().Value(segment[1:])
			if stringValue, ok := value.(string); ok {
				pathSegments[i] = stringValue
			}
		}

		if segment == "*" {
			if requestPathIgnored, ok := req.Context().Value("request.path.ignored").(string); ok {
				pathSegments = append(pathSegments[:i], requestPathIgnored)
				break
			}
		}
	}

	newPath := strings.Join(pathSegments, "/")

	req.URL.Path = "/" + newPath
	req.URL.RawPath = req.URL.EscapedPath()

}

// singleJoiningSlash properly joins two paths
// handling trailing and leading '/' characters.
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
