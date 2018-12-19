package requests

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strings"
	"time"
)

var defaultCookieJar http.CookieJar
var defaultSetting = HttpRequestSetting{
	UserAgent:        "golang-requests-client/1.1",
	ConnectTimeout:   60 * time.Second,
	ReadWriteTimeout: 60 * time.Second,
	Gzip:             true,
}

func createDefaultCookie() {
	defaultCookieJar, _ = cookiejar.New(nil)
}

type Files map[string]string

type HttpRequestSetting struct {
	UserAgent        string
	ConnectTimeout   time.Duration
	ReadWriteTimeout time.Duration
	TLSClientConfig  *tls.Config
	Proxy            func(*http.Request) (*url.URL, error)
	Transport        http.RoundTripper
	EnableCookie     bool
	Gzip             bool
	Retry            int
	CheckRedirect    func(req *http.Request, via []*http.Request) error
}

type HttpRequest struct {
	url     string
	req     *http.Request
	params  url.Values
	files   Files
	setting HttpRequestSetting
	err     []error
	logFunc func(v ...interface{})
	isDebug bool
	cxt     context.Context
}

func NewHttpRequest(method, rawUrl string) *HttpRequest {
	var errs []error
	u, err := url.Parse(rawUrl)
	if err != nil {
		log.Println("Request error:", err)
		errs = append(errs, err)
	}
	req := http.Request{
		URL:        u,
		Method:     method,
		Header:     make(http.Header),
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
	}
	return &HttpRequest{
		url:     rawUrl,
		req:     &req,
		params:  url.Values{},
		files:   Files{},
		setting: defaultSetting,
		err:     errs,
	}
}

func (r *HttpRequest) GetHttpResponse() (*HttpResponse, error) {
	if len(r.err) > 0 {
		return nil, r.Error()
	}
	r.buildURLParams()
	urlParsed, err := url.Parse(r.url)
	if err != nil {
		r.printError("Requests error:", err)
		return nil, err
	}
	r.print(r.url)
	r.req.URL = urlParsed

	trans := r.setting.Transport

	if trans == nil {
		trans = &http.Transport{
			TLSClientConfig: r.setting.TLSClientConfig,
			Proxy:           r.setting.Proxy,
			DialContext: func(ctx context.Context, network, addr string) (conn net.Conn, e error) {
				conn, err := net.DialTimeout(network, addr, r.setting.ConnectTimeout)
				if err != nil {
					r.printError("Requests error:", err)
					return nil, err
				}
				err = conn.SetDeadline(time.Now().Add(r.setting.ReadWriteTimeout))
				if err != nil {
					r.printError("Requests error:", err)
				}
				return conn, err
			},
			MaxIdleConnsPerHost: 100,
		}
	} else {
		if t, ok := trans.(*http.Transport); ok {
			if t.TLSClientConfig == nil {
				t.TLSClientConfig = r.setting.TLSClientConfig
			}
			if t.Proxy == nil {
				t.Proxy = r.setting.Proxy
			}
			if t.DialContext == nil {
				t.DialContext = func(ctx context.Context, network, addr string) (conn net.Conn, e error) {
					conn, err := net.DialTimeout(network, addr, r.setting.ConnectTimeout)
					if err != nil {
						r.printError("Requests error:", err)
						return nil, err
					}
					err = conn.SetDeadline(time.Now().Add(r.setting.ReadWriteTimeout))
					if err != nil {
						r.printError("Requests error:", err)
					}
					return conn, err
				}
			}
			if t.MaxIdleConnsPerHost == 0 {
				t.MaxIdleConnsPerHost = 100
			}
		}
	}

	var jar http.CookieJar
	if r.setting.EnableCookie {
		if defaultCookieJar == nil {
			createDefaultCookie()
		}
		jar = defaultCookieJar
	}

	client := &http.Client{
		Transport: trans,
		Jar:       jar,
	}

	if r.setting.CheckRedirect != nil {
		client.CheckRedirect = r.setting.CheckRedirect
	}
	if r.setting.UserAgent != "" && r.req.Header.Get("User-Agent") == "" {
		r.req.Header.Set("User-Agent", r.setting.UserAgent)
	}

	r.print("Header:", r.req.Header)
	r.print("Params:", r.params)

	//重试指定的次数
	for i := 0; r.setting.Retry == -1 || i <= r.setting.Retry; i++ {
		resp, err := client.Do(r.req)
		if err == nil {
			return &HttpResponse{Response: *resp}, nil
		} else {
			r.printError("Requests error:", err)
		}
	}
	return nil, err
}

func (r *HttpRequest) WithHeader(key string, value string) *HttpRequest {
	r.req.Header.Set(key, value)
	return r
}

func (r *HttpRequest) WithHeaders(headers map[string]string) *HttpRequest {
	for key, value := range headers {
		r.req.Header.Set(key, value)
	}
	return r
}

func (r *HttpRequest) WithPrint(fn func(v ...interface{})) *HttpRequest {
	r.logFunc = fn
	return r
}

func (r *HttpRequest) WithDebug(isDebug bool) *HttpRequest {
	r.isDebug = isDebug
	return r
}

func (r *HttpRequest) WithBody(data interface{}) *HttpRequest {
	switch t := data.(type) {
	case string:
		bf := bytes.NewBufferString(t)
		r.req.Body = ioutil.NopCloser(bf)
		r.req.ContentLength = int64(len(t))
	case []byte:
		bf := bytes.NewBuffer(t)
		r.req.Body = ioutil.NopCloser(bf)
		r.req.ContentLength = int64(len(t))
	case io.Reader:
		r.req.Body = ioutil.NopCloser(t)
	}
	return r
}

func (r *HttpRequest) WithJson(v interface{}) *HttpRequest {
	b, err := json.Marshal(v)

	if err != nil {
		r.err = append(r.err, err)
	}
	r.req.Header.Set("Content-Type", "application/json")
	bf := bytes.NewBuffer(b)
	r.req.Body = ioutil.NopCloser(bf)
	r.req.ContentLength = int64(len(b))

	return r
}

func (r *HttpRequest) WithParam(key string, value string) *HttpRequest {
	r.params.Set(key, value)
	return r
}

func (r *HttpRequest) WithParams(params url.Values) *HttpRequest {
	for key, value := range params {
		if len(value) <= 0 {
			r.params.Set(key, "")
		} else if len(value) == 1 {
			r.params.Set(key, value[0])
		} else {
			for _, v := range value {
				r.params.Set(key, v)
			}
		}
	}
	return r
}

func (r *HttpRequest) WithFile(fileName, filePath string) *HttpRequest {
	r.files[fileName] = filePath
	return r
}

func (r *HttpRequest) WithFiles(files Files) *HttpRequest {
	for name, p := range files {
		r.files[name] = p
	}
	return r
}

func (r *HttpRequest) WithProxy(proxy func(*http.Request) (*url.URL, error)) *HttpRequest {
	r.setting.Proxy = proxy
	return r
}

func (r *HttpRequest) WithTransport(transport http.RoundTripper) *HttpRequest {
	r.setting.Transport = transport
	return r
}

func (r *HttpRequest) WithHost(host string) *HttpRequest {
	r.req.Host = host
	return r
}

func (r *HttpRequest) WithProtocolVersion(vers string) *HttpRequest {
	if len(vers) == 0 {
		vers = "HTTP/1.1"
	}

	major, minor, ok := http.ParseHTTPVersion(vers)
	if ok {
		r.req.Proto = vers
		r.req.ProtoMajor = major
		r.req.ProtoMinor = minor
	}
	return r
}

func (r *HttpRequest) WithTLSConfig(config *tls.Config) *HttpRequest {
	r.setting.TLSClientConfig = config
	return r
}

func (r *HttpRequest) WithConnectTimeout(timeout time.Duration) *HttpRequest {
	r.setting.ConnectTimeout = timeout
	r.setting.ReadWriteTimeout = timeout
	return r
}

func (r *HttpRequest) WithReadWriteTimeout(timeout time.Duration) *HttpRequest {
	r.setting.ReadWriteTimeout = timeout
	return r
}

func (r *HttpRequest) WithUserAgent(ua string) *HttpRequest {
	r.setting.UserAgent = ua
	return r
}

func (r *HttpRequest) WithCookie(cookie *http.Cookie) *HttpRequest {
	r.setting.EnableCookie = true
	r.req.Header.Set("Cookie", cookie.String())
	return r
}

func (r *HttpRequest) WithBasicAuth(username, password string) *HttpRequest {
	r.req.SetBasicAuth(username, password)
	return r
}

func (r *HttpRequest) WithSetting(setting HttpRequestSetting) *HttpRequest {
	r.setting = setting
	return r
}

func (r *HttpRequest) WithReferer(referer string) *HttpRequest {
	r.req.Header.Set("Referer", referer)
	return r
}

func (r *HttpRequest) WidthContentType(contentType string) *HttpRequest {
	if contentType != "" {
		r.req.Header.Set("Content-Type", contentType)
	}
	return r
}

func (r *HttpRequest) WithRedirect(redirect func(req *http.Request, via []*http.Request) error) *HttpRequest {
	r.setting.CheckRedirect = redirect
	return r
}

func (r *HttpRequest) WithContext(ctx context.Context) *HttpRequest {
	r.req = r.req.WithContext(ctx)
	return r
}

func (r *HttpRequest) GetRequest() *http.Request {
	return r.req
}

func (r *HttpRequest) Error() error {
	if i := len(r.err); i > 0 {
		return r.err[i-1]
	}
	return nil
}

func (r *HttpRequest) Errors() []error {
	return r.err
}

func (r *HttpRequest) print(msg ...interface{}) {
	if r.isDebug && len(msg) > 0 {
		if r.logFunc != nil {
			r.logFunc(msg...)
		} else {
			log.Printf("INFO %s \n", fmt.Sprint(msg...))
		}
	}
}

func (r *HttpRequest) printError(v ...interface{}) {
	if len(v) > 0 {
		for _, e := range v {
			switch e.(type) {
			case error:
				r.err = append(r.err, e.(error))
			}
		}
		if r.logFunc != nil {
			r.logFunc(v...)
		} else {
			log.Printf("ERROR %s %s\n", time.Now().Format("2006-01-02 15:04:05"), fmt.Sprint(v...))
		}
	}
}

func (r *HttpRequest) buildURLParams() {

	if r.req.Method == http.MethodGet {
		if strings.Contains(r.url, "?") {
			r.url += "&" + r.params.Encode()
		} else {
			r.url += "?" + r.params.Encode()
		}
		return
	}
	r.print(r.req.Method, r.url)

	if (r.req.Method == http.MethodPost ||
		r.req.Method == http.MethodPut ||
		r.req.Method == http.MethodPatch ||
		r.req.Method == http.MethodDelete) && r.req.Body == nil {
		//如果存在文件
		if len(r.files) > 0 {
			pr, pw := io.Pipe()
			bodyWriter := multipart.NewWriter(pw)
			go func() {
				for formName, filename := range r.files {
					fileWriter, err := bodyWriter.CreateFormFile(formName, filename)
					if err != nil {
						r.printError("Requests error:", err)
						continue
					}
					fh, err := os.Open(filename)
					if err != nil {
						r.printError("Requests error:", err)
						continue
					}
					_, err = io.Copy(fileWriter, fh)
					if err := fh.Close(); err != nil {
						r.printError("Requests error:", err)
					}
				}
				for k, v := range r.params {
					for _, vv := range v {
						if err := bodyWriter.WriteField(k, vv); err != nil {
							r.printError("Requests error:", err)
						}
					}
				}
				if err := bodyWriter.Close(); err != nil {
					r.printError("Requests error:", err)
				}
				if err := pw.Close(); err != nil {
					r.printError("Requests error:", err)
				}
			}()
			r.WithHeader("Content-Type", bodyWriter.FormDataContentType())
			r.req.Body = ioutil.NopCloser(pr)
		} else {
			if r.req.Header.Get("Content-Type") == "" {
				r.WithHeader("Content-Type", "application/x-www-form-urlencoded")
			}
			r.WithBody(r.params.Encode())
		}
	}
}

func Get(rawUrl string) (*HttpResponse, error) {
	req := NewHttpRequest(rawUrl, http.MethodGet)

	if err := req.Error(); err != nil {
		return nil, err
	}
	resp, err := req.GetHttpResponse()
	return resp, err
}

func Post(rawUrl string, body []byte) (*HttpResponse, error) {
	resp, err := NewHttpRequest(rawUrl, http.MethodPost).WithBody(body).GetHttpResponse()
	return resp, err
}

func PostForm(rawUrl string, body url.Values) (*HttpResponse, error) {
	resp, err := NewHttpRequest(rawUrl, http.MethodPost).
		WithBody(body.Encode()).
		WidthContentType("application/x-www-form-urlencoded").
		GetHttpResponse()

	return resp, err
}

func Head(rawUrl string) (*HttpResponse, error) {
	return NewHttpRequest(rawUrl, http.MethodHead).GetHttpResponse()
}

func Delete(rawUrl string) (*HttpResponse, error) {
	return NewHttpRequest(rawUrl, http.MethodDelete).GetHttpResponse()
}
