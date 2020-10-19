package requests

import (
	"context"
	"testing"
	"time"
)

func TestPool_Result(t *testing.T) {
	pool := NewPool()

	pool.Add("baidu.com", NewHttpRequest("GET", "https://www.baidu.com")).
		AddWithCallback("zhihu.com", NewHttpRequest("GET", "https://www.zhihu.com"), func(name string, resp *HttpResponse, err error) {
			t.Log(name)
		}).
		Add("oschina.net", NewHttpRequest("GET", "https://www.oschina.net/"))

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	pool.Wait(ctx)

	resp, err := pool.Result("baidu.com")
	if err != nil {
		t.Fatalf("name=baidu.com,%+v", err)
	}
	body, _ := resp.String()
	t.Logf("%s,%s", resp.Status, body)
}
