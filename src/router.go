package main

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync"
)

var challengeProxy *proxy
var proxies []*proxy

const maxProxies = 100

func init() {
	remote, _ := url.Parse("http://localhost:8080")
	challengeProxy =  NewProxy(remote)

	remote, _ = url.Parse("http://localhost:82")
	proxies = make([]*proxy, maxProxies)
	for i := range proxies {
		proxies[i] = NewProxy(remote)
	}
}

type router struct {
	sync.Mutex
	iProxy int
}

func (r *router) getProxy() *proxy {
	r.Lock()
	defer r.Unlock()

	if r.iProxy >= maxProxies {
		r.iProxy = 0
	}
	rp := proxies[r.iProxy]
	r.iProxy++
	return rp
}

type proxy struct {
	*httputil.ReverseProxy
	Host string
}

func NewProxy(target *url.URL) *proxy {
	rp := httputil.NewSingleHostReverseProxy(target)
	return &proxy{
		rp,
		target.Host,
	}
}

func (this *router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	var rp *proxy
	if strings.Contains(path, "/.well-known/acme-challenge") {
		rp = challengeProxy
	} else {
		rp = this.getProxy()
	}

	r.Host = rp.Host
	rp.ServeHTTP(w, r)
}
