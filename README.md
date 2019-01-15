# requests

基于Golang的HTTP请求封装

## 使用 

```go
go get github.com/lifei6671/requests
```

### GET请求

```go
resp, err := requests.NewHttpRequest("GET", "https://www.baidu.com/s").
	WithParam("wd", "golang").
	WithHeader("Referer", "https://www.baidu.com").
	WithHeader("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8").
	WithHeader("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8,zh-TW;q=0.7").
	WithHeader("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/71.0.3578.80 Safari/537.36").
	WithDebug(true).
	GetHttpResponse()
if err != nil {
	fmt.Println(err)
} else {
	body,err := resp.Bytes()
}
```

### POST 请求

```go
body := []byte("www.baidu.com")
resp, err := requests.NewHttpRequest( "POST", "https://www.baidu.com/").WithBody(body).GetHttpResponse()
if err != nil {
	fmt.Println(err)
} else {
	body,err := resp.String()
}
```