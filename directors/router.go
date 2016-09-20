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

func NewRouter(targets map[string]func(*http.Request)) func(req *http.Request) {
	rt := buildRouteTree(targets)

	return func(req *http.Request) {
		if director, ok := rt.matchRoute(req.URL.Path); ok {
			director(req)
		} else {
			cancelRequestWithError(req, errors.New("invalid target"))
		}
	}
}

// matchRoute finds the route for a given path
// variable values defined in the route by ":key" syntax
// are applied to the request context by wrapping the director.
func (rt *routeTree) matchRoute(path string) (func(*http.Request), bool) {
	pathSegments := strings.Split(strings.Trim(path, "/"), "/")
	var match bool
	variables := map[string]string{}
	currentNode := rt

	for _, pathSegment := range pathSegments {
		currentNode, match = currentNode.matchSegment(pathSegment)
		if !match {
			return nil, false
		}

		if strings.HasPrefix(currentNode.route, ":") {
			key := currentNode.route[1:]
			variables[key] = pathSegment
		}
	}

	director := directorWithVariables(currentNode.director, variables)
	return director, match
}

func (rt *routeTree) matchSegment(segment string) (*routeTree, bool) {
	next, ok := rt.children[segment]
	if !ok {
		next, ok = rt.children["*"]
	}

	return next, ok
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
	root := &routeTree{
		children: map[string]*routeTree{},
	}

	for routeDefinition, target := range targets {
		if !routePathIsValid(routeDefinition) {
			continue
		}

		path := strings.Split(strings.Trim(routeDefinition, "/"), "/")
		currentNode := root

		var child *routeTree
		var ok bool
		for _, pathSegment := range path {

			child, ok = currentNode.children[pathSegment]
			if !ok {
				child = newRouteTree(pathSegment)
			}
			if strings.HasPrefix(pathSegment, ":") {
				currentNode.children["*"] = child
			} else {
				currentNode.children[pathSegment] = child
			}
			currentNode = child
		}

		// set director on the final pathSegment (full match)
		child.director = target
	}

	return root
}

func directorWithVariables(director func(*http.Request), variables map[string]string) func(*http.Request) {
	return func(req *http.Request) {
		for key, value := range variables {
			setRequestVariable(req, key, value)
		}
		director(req)
	}
}

func setRequestVariable(req *http.Request, key, value string) {
	*req = *req.WithContext(context.WithValue(req.Context(), key, value))
}

func routePathIsValid(path string) bool {
	// validate path here
	// If it contains * it must be the last character
	// TODO: better use of wildcards (allow surrounded wildcard etc.)
	return true
}

func cancelRequestWithError(req *http.Request, err error) {
	ctx := context.WithValue(req.Context(), "error", err)
	ctx, cancel := context.WithCancel(ctx)
	cancel()
	*req = *req.WithContext(ctx)
}
