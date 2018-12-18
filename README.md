# requests

基于Golang的HTTP请求封装

## 使用 

```go
resp, err := NewHttpRequest("https://github.com", "GET").DoRequest()
b, err := resp.Bytes()
```