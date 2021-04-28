package httphelper

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

const HttpClientErrCode = 0

const reqTimeout time.Duration = 10 // seconds

// HTTP METHOD WRAPPER
type ClientRequest struct {
	Url     string
	Query   map[string][]string
	Headers map[string]string
	Body    []byte
	Timeout int

	V interface{} // response body unmarshal struct

	method string
}

type ClientResponse struct {
	Code int   // http status code and default err code(0)
	Err  error // when program err occurred
	Body []byte
	Raw  *http.Response // be careful, response.Body can be read exactly once
}

func Get(req *ClientRequest) *ClientResponse {
	req.method = http.MethodGet
	resp := httpRequest(req)
	return resp
}

func Post(req *ClientRequest) *ClientResponse {
	req.method = http.MethodPost
	resp := httpRequest(req)
	return resp
}

func Put(req *ClientRequest) *ClientResponse {
	req.method = http.MethodPut
	resp := httpRequest(req)
	return resp
}

func Delete(req *ClientRequest) *ClientResponse {
	req.method = http.MethodDelete
	resp := httpRequest(req)
	return resp
}

func API(req *ClientRequest, method string) *ClientResponse {
	req.method = method
	resp := httpRequest(req)
	return resp
}

func httpRequest(req *ClientRequest) *ClientResponse {
	var err error
	resp := &ClientResponse{
		Code: HttpClientErrCode,
		Err:  err,
	}

	// generate req
	var newReq *http.Request
	if req.Body != nil {
		newReq, err = http.NewRequest(req.method, req.Url, bytes.NewBuffer(req.Body))
	} else {
		newReq, err = http.NewRequest(req.method, req.Url, nil)
	}
	if err != nil {
		resp.Err = err
		return resp
	}

	// process url query string
	if req.Query != nil {
		urlV := url.Values{}
		for k, vs := range req.Query {
			if len(vs) == 1 {
				urlV.Set(k, vs[0])
			} else if len(vs) > 1 {
				for _, v := range vs {
					urlV.Add(k, v)
				}
			}
		}
		newReq.URL.RawQuery = urlV.Encode()
	}

	// process headers
	for k, v := range req.Headers {
		newReq.Header.Add(k, v)
	}

	// timeout
	timeout := reqTimeout
	if req.Timeout > 0 {
		timeout = time.Duration(req.Timeout)
	}
	client := &http.Client{
		Timeout: timeout * time.Second,
	}

	doResp, err := client.Do(newReq)
	if err != nil {
		resp.Err = err
		return resp
	}
	resp.Raw = doResp
	resp.Code = doResp.StatusCode

	respBody, err := ioutil.ReadAll(doResp.Body)
	if err != nil {
		resp.Err = err
		return resp
	}
	defer doResp.Body.Close()
	resp.Body = respBody

	// check if need unmarshal response body
	if req.V != nil {
		err = json.Unmarshal(respBody, &req.V)
		if err != nil {
			resp.Err = err
			return resp
		}
	}

	return resp
}
