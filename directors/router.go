package directors

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"sync"
)

type routeTree struct {
	sync.RWMutex
	route    string
	director func(*http.Request)
	children map[string]*routeTree
}

func NewRouterDirector(targets map[string]func(*http.Request)) func(req *http.Request) {
	routes := buildRouteTree(targets)
	_ = routes

	return func(req *http.Request) {
		if director, ok := matchRoute(req.URL.Path, targets); ok {
			director(req)
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

func matchRoute(requestPath string, targets map[string]func(*http.Request)) (func(*http.Request), bool) {
	return nil, false
}

func newRouteTree(route string) *routeTree {
	return &routeTree{
		route: route,
		director: func(req *http.Request) {
			// handler on incomplete routes doesn't match by default
			cancelRequestWithError(req, errors.New("location not found")) // TODO: check error handling.. 404?
		},
		children: map[string]*routeTree{},
	}
}

func buildRouteTree(targets map[string]func(*http.Request)) *routeTree {
	rt := &routeTree{
		route:    "/",
		children: map[string]*routeTree{},
	}

	for routeDefinition, target := range targets {
		if !routePathIsValid(routeDefinition) {
			continue
		}

		path := strings.Split(strings.Trim(routeDefinition, "/"), "/")
		root := rt
		var child *routeTree
		for _, pathSegment := range path {
			child, ok := root.children[pathSegment]
			if !ok {
				child = newRouteTree(pathSegment)
			}
			root.children[pathSegment] = child
		}

		// set director on last pathSegment
		child.director = target
	}

	return nil
}

func routePathIsValid(path string) bool {
	return true
}

func cancelRequestWithError(req *http.Request, err error) {
	ctx := context.WithValue(req.Context(), "error", err)
	ctx, cancel := context.WithCancel(ctx)
	cancel()
	*req = *req.WithContext(ctx)
}
