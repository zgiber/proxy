package directors

import (
	"net/http"
	"testing"
)

func TestMatchRoutes(t *testing.T) {

	incomingPaths := []string{
		"/segment1/segment2/segment3", // simple route
		"/segment2/segment3/",         // ending in '/'
		"/segment3/user123/resource",  // expected to match wildcard
		"/nomatch/",                   // no match
		"/nomatch",                    // no match
		"/",                           // root
	}

	results := []interface{}{}
	expected := []interface{}{1, 2, "user123", -1, -1, 3}

	targets := map[string]func(*http.Request){
		"/segment1/segment2/segment3": func(req *http.Request) { results = append(results, 1) },
		"/segment2/segment3":          func(req *http.Request) { results = append(results, 2) },
		"/segment3/:user_id/resource": func(req *http.Request) { results = append(results, req.Context().Value("user_id")) },
		"/": func(req *http.Request) { results = append(results, 3) },
	}

	rt := buildRouteTree(targets)
	req := &http.Request{}

	for _, path := range incomingPaths {
		d, ok := rt.matchRoute(path)
		if !ok {
			results = append(results, -1)
			continue
		}

		d(req)
	}

	for i, result := range results {
		if result != expected[i] {
			t.Fatal("Invalid result from director")
		}
	}
}
