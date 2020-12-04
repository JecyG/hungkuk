package rest

import (
	"bytes"
	"encoding/json"
	"fmt"
)

type (
	Result interface {
		StatusCode(statusCode *int) Result
		Into(obj interface{}) error
	}

	result struct {
		body       []byte
		err        error
		statusCode int
	}
)

func (r *result) StatusCode(statusCode *int) Result {
	if statusCode != nil {
		statusCode = &r.statusCode
	}

	return r
}

func (r *result) Into(obj interface{}) error {
	if r.err != nil {
		return r.err
	}

	if obj != nil {
		if len(r.body) == 0 {
			return fmt.Errorf("empty response body with status code: %d", r.statusCode)
		}

		err := json.NewDecoder(bytes.NewReader(r.body)).Decode(obj)
		if err != nil {
			return fmt.Errorf("decode response body error: %v", err)
		}
	}

	return nil
}
