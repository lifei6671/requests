package requests

import (
	"context"
	"errors"
	"sync"
)

type Callback func(name string, resp *HttpResponse, err error)

type request struct {
	req      *HttpRequest
	callback Callback
}
type response struct {
	resp *HttpResponse
	err  error
}

type Pool struct {
	sync.RWMutex
	req  map[string]*request
	resp map[string]*response
	wait *sync.WaitGroup
}

func NewPool() *Pool {
	return &Pool{
		req:  make(map[string]*request),
		resp: make(map[string]*response),
		wait: &sync.WaitGroup{},
	}
}

func (p *Pool) Add(name string, req *HttpRequest) *Pool {
	p.Lock()
	defer p.Unlock()
	p.req[name] = &request{
		req: req,
	}
	return p
}
func (p *Pool) AddWithCallback(name string, req *HttpRequest, callback Callback) *Pool {
	p.Lock()
	defer p.Unlock()
	p.req[name] = &request{
		req:      req,
		callback: callback,
	}
	return p
}

func (p *Pool) Result(name string) (*HttpResponse, error) {
	p.RLock()
	defer p.RUnlock()
	if resp, ok := p.resp[name]; ok {
		return resp.resp, resp.err
	}
	return nil, errors.New("response does not exist:" + name)
}

func (p *Pool) Wait(ctx context.Context) {
	p.Lock()
	defer p.Unlock()
	for name, req := range p.req {
		p.wait.Add(1)
		go func(name string, req *request) {
			defer p.wait.Done()
			ctx1, cancel := context.WithCancel(ctx)
			defer cancel()
			req.req.WithContext(ctx1)
			resp := &response{}
			resp.resp, resp.err = req.req.GetHttpResponse()
			if req.callback != nil {
				req.callback(name, resp.resp, resp.err)
			} else {
				p.resp[name] = resp
			}
		}(name, req)
	}
	p.wait.Wait()
}
