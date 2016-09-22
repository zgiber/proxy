package directors

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"sync"
)

type errNotFound error

type routeTree struct {
	sync.RWMutex
	route    string
	director func(*http.Request)
	children map[string]*routeTree
}

func NewRouter(targets map[string]func(*http.Request)) func(req *http.Request) {
	rt := buildRouteTree(targets)

	return func(req *http.Request) {
		if d, match := rt.matchRoute(req.URL.Path); match {
			d(req)
		} else {
			cancelRequestWithError(req, errors.New("invalid target"))
		}
	}
}

// matchRoute finds the route for a given path
// variable values defined in the route by ":key" syntax
// are applied to the request context by wrapping the director.
func (rt *routeTree) matchRoute(path string) (func(*http.Request), bool) {
	// TODO: sanitize incoming paths for pattern matching

	var director func(*http.Request)
	pathVariables := map[string]string{}

	var match, wildcardmatch bool
	currentNode := rt
	pathSegments := strings.Split(strings.Trim(path, "/"), "/")
	for _, pathSegment := range pathSegments {

		currentNode.RLock()
		wildcardNode, ok := currentNode.children["*"]
		currentNode.RUnlock()

		if ok {
			wildcardmatch = true
			director = wildcardNode.director
		}

		currentNode.RLock()
		variableNode, ok := currentNode.children[":"]
		currentNode.RUnlock()

		if ok {
			variable := strings.TrimPrefix(variableNode.route, ":")
			pathVariables[variable] = pathSegment
			currentNode = variableNode
			match = true
		}

		currentNode.RLock()
		nextNode, ok := currentNode.children[pathSegment]
		currentNode.RUnlock()

		if ok {
			currentNode = nextNode
			match = true
		} else {
			match = false
		}
	}

	if match {
		director = currentNode.director
	}

	return director, match || wildcardmatch
}

func newRouteTree(route string) *routeTree {
	return &routeTree{
		route: route,
		director: func(req *http.Request) {
			// handler on incomplete routes doesn't match by default
			cancelRequestWithError(req, errNotFound(errors.New("not found")))
			// TODO: check error handling.. 404?
		},
		children: map[string]*routeTree{},
	}
}

func buildRouteTree(targets map[string]func(*http.Request)) *routeTree {
	// TODO: sanitize route definitions
	root := &routeTree{
		children: map[string]*routeTree{},
	}

	root.Lock()
	defer root.Unlock()

	for routeDefinition, target := range targets {
		if !routePathIsValid(routeDefinition) {
			continue
		}

		path := strings.Split(strings.Trim(routeDefinition, "/"), "/")
		currentNode := root

		var child *routeTree
		var ok bool

		for _, pathSegment := range path {

			if strings.HasPrefix(pathSegment, ":") {
				child, ok = currentNode.children[":"]
			} else {
				child, ok = currentNode.children[pathSegment]
			}
			if !ok {
				child = newRouteTree(pathSegment)
			}

			switch {
			case pathSegment == "*":
				currentNode.children["*"] = child
				child.director = target

			case strings.HasPrefix(pathSegment, ":"):
				currentNode.children[":"] = child
				currentNode = child

			default:
				currentNode.children[pathSegment] = child
				currentNode = child
			}
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
