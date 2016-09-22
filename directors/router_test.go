package directors

import (
	"net/http"
	"strings"
	"testing"
)

func TestMatchRoutes(t *testing.T) {

	// TODO: exhaustive test with all corner cases
	incomingPaths := []string{
		"/segment1/segment2/segment3",
		"/segment1/segment2/segment3/whatever",
		"/segment2/segment3/",
		"/segment3/user123/resource",
		"/segment3/user123/whatever",
		"/segment4/user123/resource",
		"/nomatch/",
		"/nomatch/",
		"/",
	}

	expectedResults := []string{
		"/segment1/segment2/segment3",
		"/segment1/segment2/segment3/whatever",
		"/segment2/segment3/",
		"/segment3/user123/resource",
		"/segment3/user123/whatever",
		"/segment4/user123/resource",
		"-",
		"-",
		"/",
	}

	results := []string{}
	appendReqestPath := func(req *http.Request) {
		results = append(results, req.URL.Path)
	}

	targets := map[string]func(*http.Request){
		"/segment1/segment2/segment3": appendReqestPath,
		"/segment1/*":                 appendReqestPath,
		"/segment2/segment3":          appendReqestPath,
		"/segment3/:user_id/*":        appendReqestPath,
		"/segment3/:user_id/resource": appendReqestPath,
		"/segment4/:user_id/*":        appendReqestPath,
		"/": appendReqestPath,
	}

	rt := buildRouteTree(targets)
	urlStr := "http://localhost"

	for _, path := range incomingPaths {
		// time.Sleep(50 * time.Millisecond)
		req, _ := http.NewRequest("GET", urlStr+path, nil)
		if d, match := rt.matchRoute(path); match {
			d(req)
		} else {
			results = append(results, "-")
		}
	}

	for i, expected := range expectedResults {
		expected = strings.TrimSpace(expected)
		result := strings.TrimSpace(results[i])
		// fmt.Println(result, expected)
		if expected != result {
			t.Fatalf("Invalid result from director [%v]. Expected:%v Got:%v", i, expected, result)
		}
	}
}
