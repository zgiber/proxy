package main

import (
	"log"
	"net/http"
	"net/url"

	"github.com/zgiber/proxy"
)

func main() {
	reverseProxy := proxy.New()
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	targets := map[string]*url.URL{
		"/google": &url.URL{
			Scheme: "http",
			Host:   "google.com",
		},
		"/bing": &url.URL{
			Scheme: "http",
			Host:   "bing.com",
		},
	}

	err := reverseProxy.AddDirector(directors.NewRouterDirector(targets))
	if err != nil {
		log.Println(err)
	}

	go reverseProxy.ListenAndServeDirectorConfigAPI(":9002") // TODO: add some resilience
	http.ListenAndServe(":9001", reverseProxy)
}
