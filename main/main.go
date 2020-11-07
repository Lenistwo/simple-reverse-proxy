package main

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"simpleReverseProxy/structs"
	"strings"
)

const (
	host                  = "127.0.0.1"
	port                  = ":8080"
	defaultDestinationUrl = "http://127.0.0.1:8081"
	emptyString           = ""
	slash                 = "/"
	slashIndex            = 1
	replaceFirst          = 1
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

func handleRequest(proxy *httputil.ReverseProxy) func(responseWriter http.ResponseWriter, request *http.Request) {
	return func(responseWriter http.ResponseWriter, request *http.Request) {
		proxy.Director = setDirector(request.URL.Path)
		proxy.ServeHTTP(responseWriter, request)
	}
}

func setDirector(requestUrl string) func(request *http.Request) {
	parse, err := url.Parse(defaultDestinationUrl)

	for _, destination := range destinations {
		if strings.HasPrefix(requestUrl, destination.Invoker) {
			parse, _ = url.Parse(destination.URL)
			break
		}
	}

	if err != nil {
		panic(err)
	}

	return func(request *http.Request) {
		request.URL.Scheme = parse.Scheme
		request.URL.Host = parse.Host
		request.URL.Path, request.URL.RawPath = joinURLPath(parse, prepareURL(request.URL))
	}
}

func joinURLPath(proxyDestinationUrl, requestEndpoint *url.URL) (path, rawPath string) {
	if proxyDestinationUrl.RawPath == emptyString && requestEndpoint.RawPath == emptyString {
		return singleJoiningSlash(proxyDestinationUrl.Path, requestEndpoint.Path), emptyString
	}
	proxyPath := proxyDestinationUrl.EscapedPath()
	endpointPath := requestEndpoint.EscapedPath()

	hasBeforeSlash := strings.HasSuffix(proxyPath, slash)
	hasAfterSlash := strings.HasPrefix(endpointPath, slash)

	switch {
	case hasBeforeSlash && hasAfterSlash:
		return proxyDestinationUrl.Path + requestEndpoint.Path[slashIndex:], proxyPath + endpointPath[slashIndex:]
	case !hasBeforeSlash && !hasAfterSlash:
		return proxyDestinationUrl.Path + slash + requestEndpoint.Path, proxyPath + slash + endpointPath
	}
	return proxyDestinationUrl.Path + requestEndpoint.Path, proxyPath + endpointPath
}

func singleJoiningSlash(proxyDestinationUrl, requestEndpoint string) string {
	aslash := strings.HasSuffix(proxyDestinationUrl, slash)
	bslash := strings.HasPrefix(requestEndpoint, slash)
	switch {
	case aslash && bslash:
		return proxyDestinationUrl + requestEndpoint[slashIndex:]
	case !aslash && !bslash:
		return proxyDestinationUrl + slash + requestEndpoint
	}
	return proxyDestinationUrl + requestEndpoint
}

func prepareURL(url *url.URL) *url.URL {
	for _, destination := range destinations {
		if strings.Contains(url.Path, destination.Invoker) {
			url.Path = strings.Replace(url.Path, destination.Invoker, emptyString, replaceFirst)
			url.RawPath = strings.Replace(url.RawPath, destination.Invoker, emptyString, replaceFirst)
			break
		}
	}
	return url
}
