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
	targets := map[string]func(*http.Request){
		"/hello": directors.NewSingleHost("http://localhost:8080/"), // start something on port 8080 first... (python -m SimpleHTTPServer 8080)
	}

	// add router director
	reverseProxy.AddDirector(directors.NewRouter(targets))

	// start configuration backend
	go reverseProxy.ListenAndServeDirectorConfigAPI(":9002") // TODO: add some resilience to the config backend

	// start proxy
	http.ListenAndServe(":9001", reverseProxy)
}
