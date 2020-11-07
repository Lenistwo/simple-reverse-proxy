package main

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"simpleReverseProxy/structs"
	"strings"
)

const (
	host        = "127.0.0.1"
	port        = ":8080"
	defaultUrl  = "http://127.0.0.1:8081"
	emptyString = ""
	slash       = "/"
)

var (
	destinations = []structs.Destination{{
		Invoker: "/api/auth",
		URL:     "http://127.0.0.1:8410",
	}, {
		Invoker: "/api/customer",
		URL:     "http://127.0.0.1:8411",
	}}
)

func main() {

	reverseProxy := httputil.ReverseProxy{}

	http.HandleFunc(slash, handleRequest(&reverseProxy))
	err := http.ListenAndServe(host+port, nil)

	if err != nil {
		panic(err)
	}
}

func handleRequest(p *httputil.ReverseProxy) func(w http.ResponseWriter, r *http.Request) {

	return func(w http.ResponseWriter, r *http.Request) {
		p.Director = setDirector(r.URL.Path)
		p.ServeHTTP(w, r)
	}
}

func setDirector(requestUrl string) func(r *http.Request) {
	parse, err := url.Parse(defaultUrl)

	for _, destination := range destinations {
		if strings.HasPrefix(requestUrl, destination.Invoker) {
			parse, _ = url.Parse(destination.URL)
			break
		}
	}

	if err != nil {
		panic(err)
	}

	return func(req *http.Request) {
		req.URL.Scheme = parse.Scheme
		req.URL.Host = parse.Host
		req.URL.Path, req.URL.RawPath = joinURLPath(parse, prepareURL(req.URL))
	}
}

func joinURLPath(a, b *url.URL) (path, rawpath string) {
	if a.RawPath == emptyString && b.RawPath == emptyString {
		return singleJoiningSlash(a.Path, b.Path), emptyString
	}
	apath := a.EscapedPath()
	bpath := b.EscapedPath()

	aslash := strings.HasSuffix(apath, slash)
	bslash := strings.HasPrefix(bpath, slash)

	switch {
	case aslash && bslash:
		return a.Path + b.Path[1:], apath + bpath[1:]
	case !aslash && !bslash:
		return a.Path + slash + b.Path, apath + slash + bpath
	}
	return a.Path + b.Path, apath + bpath
}

func singleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, slash)
	bslash := strings.HasPrefix(b, slash)
	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + slash + b
	}
	return a + b
}

func prepareURL(url *url.URL) *url.URL {
	for _, destination := range destinations {
		url.Path = strings.Replace(url.Path, destination.Invoker, emptyString, 1)
		url.RawPath = strings.Replace(url.RawPath, destination.Invoker, emptyString, 1)
	}
	return url
}
