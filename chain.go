package proxy

import "net/http"

func ChainDirectors(directors ...func(*http.Request)) func(*http.Request) {
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
	}
}
