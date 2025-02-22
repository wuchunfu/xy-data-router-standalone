package es

import (
	"crypto/tls"
	"crypto/x509"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/valyala/fasthttp"

	"github.com/fufuok/xy-data-router/conf"
)

var (
	fasthttpClient       *fasthttp.Client
	maxIdleConnDuration  = 10 * time.Minute
	dialConcurrency      = 4096
	dialDNSCacheDuration = 1 * time.Hour
)

// transport implements the elastictransport interface with
// the github.com/valyala/fasthttp HTTP client.
type transport struct{}

func initFasthttpClient() {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: conf.Config.DataConf.ESInsecureSkipVerify,
	}
	if len(conf.Config.DataConf.ESRootCA) > 0 {
		tlsConfig.RootCAs = x509.NewCertPool()
		tlsConfig.RootCAs.AppendCertsFromPEM(conf.Config.DataConf.ESRootCA)
	}
	dialer := &fasthttp.TCPDialer{
		Concurrency:      dialConcurrency,
		DNSCacheDuration: dialDNSCacheDuration,
	}
	fasthttpClient = &fasthttp.Client{
		TLSConfig:                     tlsConfig,
		ReadTimeout:                   conf.Config.WebConf.ESAPITimeout,
		WriteTimeout:                  conf.Config.WebConf.ESAPITimeout,
		MaxIdleConnDuration:           maxIdleConnDuration,
		MaxIdemponentCallAttempts:     0,
		NoDefaultUserAgentHeader:      true,
		DisableHeaderNamesNormalizing: true,
		DisablePathNormalizing:        true,
		Dial:                          dialer.Dial,
	}
}

// RoundTrip performs the request and returns a response or error
func (t *transport) RoundTrip(req *http.Request) (*http.Response, error) {
	freq := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(freq)

	fres := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(fres)

	t.copyRequest(freq, req)

	err := fasthttpClient.Do(freq, fres)
	if err != nil {
		return nil, err
	}

	res := &http.Response{Header: make(http.Header)}
	t.copyResponse(res, fres)

	return res, nil
}

// copyRequest converts a http.Request to fasthttp.Request
func (t *transport) copyRequest(dst *fasthttp.Request, src *http.Request) *fasthttp.Request {
	if src.Method == "GET" && src.Body != nil {
		src.Method = "POST"
	}

	dst.SetHost(src.Host)
	dst.SetRequestURI(src.URL.String())

	dst.Header.SetRequestURI(src.URL.String())
	dst.Header.SetMethod(src.Method)

	for k, vv := range src.Header {
		for _, v := range vv {
			dst.Header.Set(k, v)
		}
	}

	if src.Body != nil {
		dst.SetBodyStream(src.Body, -1)
	}

	return dst
}

// copyResponse converts a http.Response to fasthttp.Response
func (t *transport) copyResponse(dst *http.Response, src *fasthttp.Response) *http.Response {
	dst.StatusCode = src.StatusCode()

	src.Header.VisitAll(func(k, v []byte) {
		dst.Header.Set(string(k), string(v))
	})

	// Cast to a string to make a copy seeing as src.Body() won't
	// be valid after the response is released back to the pool (fasthttp.ReleaseResponse).
	dst.Body = io.NopCloser(strings.NewReader(string(src.Body())))

	return dst
}
