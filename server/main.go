package main

import (
	"net/http"
	"net/url"

	"github.com/zgiber/proxy"
)

func main() {
	reverseProxy := proxy.New()

	targets := map[string]*url.URL{
		"/google": &url.URL{
			Scheme: "https",
			Host:   "google.com",
		},
		"/bing": &url.URL{
			Scheme: "https",
			Host:   "bing.com",
		},
	}

	reverseProxy.AddDirector(proxy.NewRouterDirector(targets))

	go reverseProxy.ListenAndServeConfigAPI(":9002") // TODO: add some resilience
	http.ListenAndServe(":9001", reverseProxy)
}
