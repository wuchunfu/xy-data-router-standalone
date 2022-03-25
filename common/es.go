package common

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/fufuok/utils"
	"github.com/fufuok/utils/pools/bufferpool"
	"github.com/tidwall/gjson"
	"github.com/valyala/fasthttp"

	"github.com/fufuok/xy-data-router/conf"
)

var (
	ES *elasticsearch.Client

	// ESVersionServer ESVersionClient 版本信息
	ESVersionServer string
	ESVersionClient string

	// ESVersionMain 大版本号: 6 / 7 / 8
	ESVersionMain int

	// ESLessThan7 大版本号小于 7
	ESLessThan7 bool
)

func initES() {
	// 首次初始化 ES 连接, PING 失败时允许启动程序
	es, cfgErr, esErr := newES()
	if cfgErr != nil || esErr != nil {
		log.Fatalln("Failed to initialize ES:", cfgErr, esErr, "\nbye.")
	}

	ES = es
}

// InitES 重新初始化 ES 连接, PING 成功则更新连接
func InitES() error {
	es, cfgErr, esErr := newES()
	if cfgErr != nil || esErr != nil {
		return fmt.Errorf("%s%s", cfgErr, esErr)
	}

	ES = es

	return nil
}

func newES() (es *elasticsearch.Client, cfgErr error, esErr error) {
	Log.Info().Strs("hosts", conf.Config.SYSConf.ESAddress).Msg("Initialize ES connection")
	es, cfgErr = elasticsearch.NewClient(elasticsearch.Config{
		Addresses:     conf.Config.SYSConf.ESAddress,
		RetryOnStatus: conf.Config.SYSConf.ESRetryOnStatus,
		MaxRetries:    conf.Config.SYSConf.ESMaxRetries,
		DisableRetry:  conf.Config.SYSConf.ESDisableRetry,
		Transport:     &transport{},
	})
	if cfgErr != nil {
		return nil, cfgErr, nil
	}

	// 数据转发时不涉及 ES
	if conf.ForwardTunnel != "" {
		return
	}

	res, err := es.Info()
	if err != nil {
		return nil, nil, err
	}
	if res.IsError() {
		err = fmt.Errorf("ES info error, status: %s", res.Status())
		Log.Error().Err(err).Msg("es.Info")
		return nil, nil, err
	}

	buf := bufferpool.Get()
	defer bufferpool.Put(buf)
	n, _ := buf.ReadFrom(res.Body)
	if n == 0 {
		err = fmt.Errorf("ES info error: nil")
		Log.Error().Err(err).Msg("es.Info")
		return nil, nil, err
	}
	ESVersionServer = gjson.GetBytes(buf.Bytes(), "version.number").String()
	ESVersionClient = elasticsearch.Version
	ESVersionMain = utils.MustInt(strings.SplitN(ESVersionServer, ".", 2)[0])
	ESLessThan7 = ESVersionMain < 7
	Log.Info().Str("server_version", ESVersionServer).Str("client_version", ESVersionClient).Msg("ES info")

	return
}

// transport implements the elastictransport interface with
// the github.com/valyala/fasthttp HTTP client.
type transport struct{}

// RoundTrip performs the request and returns a response or error
func (t *transport) RoundTrip(req *http.Request) (*http.Response, error) {
	freq := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(freq)

	fres := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(fres)

	t.copyRequest(freq, req)

	err := fasthttp.Do(freq, fres)
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
