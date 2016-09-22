package main

import (
	"log"
	"net/http"
	"time"

	"github.com/zgiber/proxy"
	"github.com/zgiber/proxy/directors"
)

func main() {
	reverseProxy := proxy.New()
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Global stuff
	reverseProxy.AddDirector(directors.Chain(
		directors.NewRateLimiter(100*time.Millisecond, 30*time.Second, 10), // delay (between calls), burst timeout, number of burst calls
		directors.NewCorrelation(),
	))

	// routed endpoints
	// TODO: use * wildcard match
	targets := map[string]func(*http.Request){
		"/hello":                 directors.NewSingleHost("http://localhost:8080/mypath"),    // start something on port 8080 first... (python -m SimpleHTTPServer 8080)
		"/api/:user_id/profile":  directors.NewSingleHost("http://localhost:8081"),           // note the lack of '/' in the end.. this will not change paths, just host and scheme
		"/api/:user_id/profile2": directors.NewSingleHost("http://localhost:8081/"),          // just an idea.. disregard for now
		"/api/:user_id/*":        directors.NewSingleHost("http://localhost:8081/whatevers"), // just an idea.. disregard for now
	}

	// add router director
	reverseProxy.AddDirector(directors.NewRouter(targets))

	// start configuration backend
	go reverseProxy.ListenAndServeDirectorConfigAPI(":9002") // TODO: add some resilience to the config backend

	// start proxy
	http.ListenAndServe(":9001", reverseProxy)
}
