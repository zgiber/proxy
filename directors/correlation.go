package directors

import (
	"log"
	"net/http"

	"github.com/zgiber/proxy/auth"
)

func NewCorrelation() func(req *http.Request) {

	return func(req *http.Request) {
		token, err := auth.NewRandomTokenString(16)
		if err != nil {
			log.Println(err)
		}
		req.Header.Set("X-Correlation-ID", string(token))
	}
}
