package main

import (
	"net/http/httputil"
	"net/url"

	"github.com/zgiber/proxy"
)

func main() {
	reverseProxy := proxy.New()
	reverseProxy.ReverseProxy = httputil.NewSingleHostReverseProxy(
		&url.URL{
			Scheme: "http",
			Host:   "localhost",
		})
}
