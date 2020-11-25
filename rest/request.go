package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"syscall"
	"time"
)

const (
	maxRetries    = 3                     // 最大重试次数
	retryInterval = 20 * time.Millisecond // 重试时间间隔
)

type (
	Request interface {
		WithParam(name, value string) Request
		WithParams(params map[string]string) Request
		WithHeader(header http.Header) Request
		WithContext(ctx context.Context) Request
		WithTimeout(d time.Duration) Request
		SubResourcef(subPath string, args ...interface{}) Request
		Body(body interface{}) Request
		Do() Result
	}

	request struct {
		client      *http.Client
		method      string
		params      url.Values
		header      http.Header
		body        []byte
		ctx         context.Context
		baseURL     string
		subPath     string
		subPathArgs []interface{}
		timeout     time.Duration
		username    string
		password    string
		err         error
	}
)

func (r *request) WithParam(name, value string) Request {
	if r.params == nil {
		r.params = make(url.Values)
	}

	r.params[name] = append(r.params[name], value)

	return r
}

func (r *request) WithParams(params map[string]string) Request {
	if r.params == nil {
		r.params = make(url.Values)
	}

	for name, value := range params {
		r.params[name] = append(r.params[name], value)
	}

	return r
}

func (r *request) WithHeader(header http.Header) Request {
	if r.header == nil {
		r.header = header
		return r
	}

	for key, values := range header {
		for _, v := range values {
			r.header.Add(key, v)
		}
	}

	return r
}

func (r *request) WithContext(ctx context.Context) Request {
	r.ctx = ctx
	return r
}

func (r *request) WithTimeout(d time.Duration) Request {
	r.timeout = d
	return r
}

func (r *request) SubResourcef(subPath string, args ...interface{}) Request {
	r.subPathArgs = args
	return r.subResource(subPath)
}

func (r *request) subResource(subPath string) Request {
	subPath = strings.TrimLeft(subPath, "/")
	r.subPath = subPath
	return r
}

func (r *request) Body(body interface{}) Request {
	if body == nil {
		r.body = []byte("")
		return r
	}

	data, err := json.Marshal(body)
	if err != nil {
		r.err = err
		r.body = []byte("")
		return r
	}

	r.body = data

	return r
}

func (r *request) wrapURL() *url.URL {
	finalUrl := &url.URL{}
	if len(r.baseURL) != 0 {
		u, err := url.Parse(r.baseURL)
		if err != nil {
			r.err = err
			return new(url.URL)
		}
		*finalUrl = *u
	}

	if len(r.subPathArgs) > 0 {
		finalUrl.Path = finalUrl.Path + fmt.Sprintf(r.subPath, r.subPathArgs...)
	} else {
		finalUrl.Path = finalUrl.Path + r.subPath
	}

	query := url.Values{}
	for key, values := range r.params {
		for _, value := range values {
			query.Add(key, value)
		}
	}

	if r.timeout != 0 {
		query.Set("timeout", r.timeout.String())
	}

	finalUrl.RawQuery = query.Encode()

	return finalUrl
}

func (r *request) Do() Result {
	rt := &result{}

	if r.err != nil {
		rt.err = r.err
		return rt
	}

	client := r.client
	if client == nil {
		client = http.DefaultClient
	}

	for try := 0; try < maxRetries; try++ {
		u := r.wrapURL().String()
		req, err := http.NewRequest(r.method, u, bytes.NewReader(r.body))
		if err != nil {
			rt.err = err
			return rt
		}

		if r.ctx != nil {
			req.WithContext(r.ctx)
		}

		req.Header = r.header.Clone()
		if len(req.Header) == 0 {
			req.Header = make(http.Header)
		}

		req.Header.Del("Accept-Encoding")
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Accept-Charset", "utf-8")

		req.SetBasicAuth(r.username, r.password)

		resp, err := client.Do(req)
		if err != nil {
			if !isConnectionReset(err) || r.method != http.MethodGet {
				rt.err = err
				return rt
			}

			time.Sleep(retryInterval)
			continue
		}

		var body []byte
		if resp.Body != nil {
			data, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				if err == io.ErrUnexpectedEOF {
					time.Sleep(retryInterval)
					continue
				}

				rt.err = err
				return rt
			}

			body = data
		}

		rt.body = body
		rt.statusCode = resp.StatusCode
		rt.status = resp.Status

		return rt
	}

	rt.err = errors.New("unexpected error")

	return rt
}

func isConnectionReset(err error) bool {
	if urlErr, ok := err.(*url.Error); ok {
		err = urlErr.Err
	}

	if opErr, ok := err.(*net.OpError); ok {
		err = opErr.Err
	}

	if osErr, ok := err.(*os.SyscallError); ok {
		err = osErr.Err
	}

	if errno, ok := err.(syscall.Errno); ok && errno == syscall.ECONNRESET {
		return true
	}

	return false
}
