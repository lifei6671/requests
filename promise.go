package requests

import "context"

type Fulfilled func(httpResponse *HttpResponse)
type Rejected func(err *RequestError)

type RequestError struct {
	error
	req *HttpRequest
}

func (e *RequestError) Error() string {
	if e.error != nil {
		return e.error.Error()
	}
	return ""
}
func (e *RequestError) Request() *HttpRequest {
	return e.req
}

type Promise struct {
	reqErr *RequestError
	cancel context.CancelFunc
}

func NewPromise(req *HttpRequest) *Promise {
	return &Promise{reqErr: &RequestError{req: req}}
}

func (p *Promise) Then(onFulfilled Fulfilled, onRejected Rejected) {
	if err := p.reqErr.req.Error(); err != nil && onRejected != nil {
		p.reqErr.error = err
		onRejected(p.reqErr)
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	p.cancel = cancel

	go func() {
		defer func() {
			if p.cancel != nil {
				p.cancel()
				p.cancel = nil
			}
		}()
		p.reqErr.req.WithContext(ctx)

		resp, err := p.reqErr.req.GetHttpResponse()
		if err != nil {
			if onRejected != nil {
				p.reqErr.error = err
				onRejected(p.reqErr)
			}
			return
		}
		if onFulfilled != nil {
			onFulfilled(resp)
			return
		}
		_ = resp.Body.Close()
	}()
}

func (p *Promise) Otherwise(onRejected Rejected) {
	p.Then(nil, onRejected)
}
func (p *Promise) Cancel() {
	if p.cancel != nil {
		p.cancel()
		p.cancel = nil
	}
}
