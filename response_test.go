package requests

import (
	"fmt"
	"testing"
)

func TestHttpResponse_Bytes(t *testing.T) {
	resp, err := NewHttpRequest("https://www.baidu.com/s", "GET").
		WithParam("wd", "golang").
		WithHeader("Referer", "https://www.baidu.com").
		WithHeader("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8").
		WithHeader("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8,zh-TW;q=0.7").
		WithHeader("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/71.0.3578.80 Safari/537.36").
		WithDebug(true).
		GetHttpResponse()

	if err != nil {
		t.Fatal(err)
	} else {
		b, err := resp.Bytes()

		t.Log(resp.StatusCode, err, string(b))
	}
}

func TestHttpResponse_SaveFile(t *testing.T) {
	resp, err := NewHttpRequest("http://f.hiphotos.baidu.com/image/pic/item/f2deb48f8c5494eee5a348a020f5e0fe98257e81.jpg", "GET").
		WithHeader("Referer", "https://www.baidu.com").
		WithHeader("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8").
		WithHeader("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8,zh-TW;q=0.7").
		WithHeader("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/71.0.3578.80 Safari/537.36").
		WithDebug(true).
		GetHttpResponse()

	if err != nil {
		t.Fatal(err)
	} else {
		err := resp.SaveFile("./f2deb48f8c5494eee5a348a020f5e0fe98257e81.jpg")

		t.Log(resp.StatusCode, err)
	}
}

func TestHttpResponse_ToJson(t *testing.T) {
	resp, err := NewHttpRequest("https://api.apiopen.top/searchPoetry", "GET").
		WithParam("name", "古风二首 二").
		WithReferer("https://www.baidu.com").
		WithHeader("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8").
		WithHeader("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8,zh-TW;q=0.7").
		WithUserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/71.0.3578.80 Safari/537.36").
		WithDebug(true).
		GetHttpResponse()

	if err != nil {
		t.Fatal(err)
	} else {
		type Data struct {
			Title   string `json:"title"`
			Content string `json:"content"`
			Authors string `json:"authors"`
		}
		type Result struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
			Result  []*Data
		}

		var result Result
		err := resp.ToJson(&result)

		t.Log(resp.StatusCode, err, fmt.Sprintf("%v", result))
	}
}

func TestHttpRequest_Error(t *testing.T) {
	req := NewHttpRequest("://www.baidu.com/s", "GET").
		WithParam("wd", "golang").
		WithHeader("Referer", "https://www.baidu.com").
		WithHeader("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8").
		WithHeader("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8,zh-TW;q=0.7").
		WithHeader("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/71.0.3578.80 Safari/537.36").
		WithDebug(true)
	if err := req.Error(); err != nil {
		t.Log(err)
	}
}

func TestHttpRequest_WithFiles(t *testing.T) {
	req := NewHttpRequest("https://www.baidu.com", "POST").
		WithFile("requests.go", "./requests.go")

	if err := req.Error(); err != nil {
		t.Log(err)
	} else {
		resp, err := req.GetHttpResponse()

		if err != nil {
			t.Log(err)
		} else {
			t.Log(resp.String())
		}
	}
}
