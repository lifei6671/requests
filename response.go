package requests

import (
	"compress/gzip"
	"encoding/json"
	"encoding/xml"
	"io"
	"io/ioutil"
	"net/http"
	"os"
)

type HttpResponse struct {
	http.Response
	body []byte
}

func (r *HttpResponse) Bytes() ([]byte, error) {
	if r.body != nil {
		return r.body, nil
	}
	if r.Body == nil {
		return nil, nil
	}
	defer func() {
		_ = r.Body.Close()
	}()
	if r.Header.Get("Content-Encoding") == "gzip" {
		reader, err := gzip.NewReader(r.Body)
		if err != nil {
			return nil, err
		}
		r.body, err = ioutil.ReadAll(reader)
		return r.body, err
	}
	var err error
	r.body, err = ioutil.ReadAll(r.Body)
	return r.body, err
}

func (r *HttpResponse) String() (string, error) {
	if r.body != nil {
		return string(r.body), nil
	}
	b, err := r.Bytes()

	return string(b), err
}

func (r *HttpResponse) SaveFile(filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	if r.Body == nil {
		return nil
	}
	defer r.Body.Close()
	_, err = io.Copy(f, r.Body)
	return err
}

func (r *HttpResponse) ToJson(v interface{}) error {
	data, err := r.Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

func (r *HttpResponse) ToXml(v interface{}) error {
	data, err := r.Bytes()
	if err != nil {
		return err
	}
	return xml.Unmarshal(data, v)
}
