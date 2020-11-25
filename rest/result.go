package rest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type (
	Result interface {
		Into(obj interface{}) error
	}

	result struct {
		body       []byte
		err        error
		statusCode int
		status     string
	}
)

func (r *result) Into(obj interface{}) error {
	if r.err != nil {
		return r.err
	}

	if r.statusCode != http.StatusOK {
		return fmt.Errorf("http response with status: %s", r.status)
	}

	if obj != nil {
		d := json.NewDecoder(bytes.NewReader(r.body))
		err := d.Decode(obj)
		if err != nil {
			return fmt.Errorf("http response body decode error: %v, raw data: %s", err, string(r.body))
		}
	}

	return nil
}
