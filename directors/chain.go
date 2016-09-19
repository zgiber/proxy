package directors

import (
	"errors"
	"net/http"
)

// Chain takes a number of directors and chains them, returning
// a single director.
func Chain(directors ...func(*http.Request)) (func(*http.Request), error) {
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
