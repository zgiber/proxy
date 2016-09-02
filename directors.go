package proxy

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sync"
)

type Router interface {
	GetURL(path string) *url.URL
	SetURL(path string, u *url.URL)
	Delete(path string)
}

type mappedRouter struct {
	sync.RWMutex
	s []*segment
}

type segment struct {
	m map[string]*segment
	u *url.URL
}

// func (r *mappedRouter) GetURL(path string) *url.URL {
// 	slicedPath := strings.Split(strings.Trim(path, "/"), "/")
// 	r.RLock()
// 	u := r.s.getURL(slicedPath)
// 	r.RUnlock()
// 	return u
// }

// func (s *segment) getURL(path []string) *url.URL {

// 	for segment, v := range s.m {

// 	}

// }

// ChainDirectors takes a number of directors and chains them, returning
// a single director.
func ChainDirectors(directors ...func(*http.Request)) (func(*http.Request), error) {
	for _, director := range directors {
		if director == nil {
			return nil, errors.New("director can not be nil")
		}

	}
	return func(req *http.Request) {

		ctx := req.Context()
		if ctx.Err() != nil {
			return
		}

		for _, director := range directors {
			director(req)
			if req.Context().Err() != nil {
				return
				// error is handled by the RoundTripper
			}
		}
	}, nil
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
		} else {
			// raise error
			err := errors.New("invalid target")
			ctx := context.WithValue(req.Context(), "error", err)
			ctx, cancel := context.WithCancel(ctx)
			cancel()
			*req = *req.WithContext(ctx)
		}
	}
}

func rewriteRequest(target *url.URL, req *http.Request) {
	fmt.Println("DEBUG - rewriteRequest")
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
	fmt.Println("DEBUG - rewriteRequest DONE")
	fmt.Println(req.URL.Scheme, req.URL.Host, req.URL.RawPath, req.URL.RawQuery)
}
