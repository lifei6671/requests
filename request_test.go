package requests

import (
	"net/url"
	"sync"
	"testing"
	"time"
)

func TestGet(t *testing.T) {
	resp, err := Get("https://www.baidu.com")
	if err != nil {
		t.Logf("%+v", err)
	} else {
		body, err := resp.String()

		t.Logf("body=%s;err=%+v", body, err)
	}
}

func TestGetAsync(t *testing.T) {
	GetAsync("https://www.baidu.com").
		Then(func(resp *HttpResponse) {
			body, err := resp.String()
			t.Logf("body=%s;err=%+v", body, err)
		}, func(err *RequestError) {
			t.Logf("%+v", err)
		})
	time.Sleep(time.Second)
}

func TestPost(t *testing.T) {
	resp, err := Post("https://www.baidu.com", []byte("aaaa"))
	if err != nil {
		t.Logf("%+v", err)
	} else {
		body, err := resp.String()

		t.Logf("body=%s;err=%+v", body, err)
	}
}
func TestPostAsync(t *testing.T) {
	PostAsync("https://www.baidu.com", []byte("abc")).
		Then(func(resp *HttpResponse) {
			body, err := resp.String()
			t.Logf("body=%s;err=%+v", body, err)
		}, func(err *RequestError) {
			t.Logf("%+v", err)
		})
	time.Sleep(time.Second)
}

func TestDelete(t *testing.T) {
	resp, err := Delete("https://www.baidu.com")
	if err != nil {
		t.Logf("%+v", err)
	} else {
		body, err := resp.String()

		t.Logf("body=%s;err=%+v", body, err)
	}
}

func TestDeleteAsync(t *testing.T) {
	DeleteAsync("https://www.baidu.com").
		Then(func(resp *HttpResponse) {
			body, err := resp.String()
			t.Logf("body=%s;err=%+v", body, err)
		}, func(err *RequestError) {
			t.Logf("%+v", err)
		})
	time.Sleep(time.Second)
}

func TestHead(t *testing.T) {
	resp, err := Head("https://www.baidu.com")
	if err != nil {
		t.Logf("%+v", err)
	} else {
		body, err := resp.String()

		t.Logf("body=%s;err=%+v", body, err)
	}
}

func TestHeadAsync(t *testing.T) {
	HeadAsync("https://www.baidu.com").
		Then(func(resp *HttpResponse) {
			body, err := resp.String()
			t.Logf("body=%s;err=%+v", body, err)
		}, func(err *RequestError) {
			t.Logf("%+v", err)
		})
	time.Sleep(time.Second)
}

func TestPostForm(t *testing.T) {
	resp, err := PostForm("https://www.baidu.com", url.Values{"key": []string{"abc"}})
	if err != nil {
		t.Logf("%+v", err)
	} else {
		body, err := resp.String()

		t.Logf("body=%s;err=%+v", body, err)
	}
}

func TestPostFormAsync(t *testing.T) {
	w := &sync.WaitGroup{}
	w.Add(1)
	PostFormAsync("https://www.baidu.com", url.Values{"key": []string{"abc"}}).
		Then(func(resp *HttpResponse) {
			defer w.Done()
			body, err := resp.String()
			t.Logf("body=%s;err=%+v", body, err)
		}, func(err *RequestError) {
			defer w.Done()
			t.Logf("%+v", err)
		})
	w.Wait()
}

func TestUploadFile(t *testing.T) {
	filename := "./testdata/test.txt"
	resp, err := UploadFile("https://www.baidu.com", filename)
	if err != nil {
		t.Fatal(err)
	} else {
		body, err := resp.String()
		t.Logf("body=%s;err=%+v", body, err)
	}
}

func TestUploadFileAsync(t *testing.T) {
	w := &sync.WaitGroup{}
	w.Add(1)
	UploadFileAsync("https://www.baidu.com", "./testdata/test.txt").
		Then(func(resp *HttpResponse) {
			defer w.Done()
			body, err := resp.String()
			t.Logf("body=%s;err=%+v", body, err)
		}, func(err *RequestError) {
			defer w.Done()
			t.Logf("%+v", err)
		})
	w.Wait()
}
