package rest

import (
	"net/http"
)

type (
	RESTClient interface {
		SetBasicAuth(username, password string) RESTClient
		Post() Request
		Put() Request
		Get() Request
		Delete() Request
		Patch() Request
	}

	restClient struct {
		client   *http.Client
		baseURL  string
		username string
		password string
	}
)

func NewClient(client *http.Client, baseURL string) RESTClient {
	return &restClient{
		client:  client,
		baseURL: baseURL,
	}
}

func (rc *restClient) Post() Request {
	return rc.method(http.MethodPost)
}

func (rc *restClient) Put() Request {
	return rc.method(http.MethodPut)
}

func (rc *restClient) Get() Request {
	return rc.method(http.MethodGet)
}

func (rc *restClient) Delete() Request {
	return rc.method(http.MethodDelete)
}

func (rc *restClient) Patch() Request {
	return rc.method(http.MethodPatch)
}

func (rc *restClient) SetBasicAuth(username, password string) RESTClient {
	rc.username = username
	rc.password = password
	return rc
}

func (rc *restClient) method(method string) Request {
	return &request{
		method:   method,
		baseURL:  rc.baseURL,
		client:   rc.client,
		username: rc.username,
		password: rc.password,
	}
}
