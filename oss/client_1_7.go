// +build go1.7

package oss

import (
	"context"
	"net/http"
	"io"
)

// Client OSS client
type Client struct {
	Config     *Config      // OSS client configuration
	Conn       *Conn        // Send HTTP request
	HTTPClient *http.Client //http.Client to use - if nil will make its own
	ctx context.Context
}

// WithContext support go1.7 context
func (client Client)WithContext(ctx context.Context) Client  {
	c := client // do a copy
	c.ctx = ctx
	return c
}

// Private
func (client Client) do(method, bucketName string, params map[string]interface{},
	headers map[string]string, data io.Reader, options ...Option) (*Response, error) {
	err := CheckBucketName(bucketName)
	if len(bucketName) > 0 && err != nil {
		return nil, err
	}

	// option headers
	addHeaders := make(map[string]string)
	err = handleOptions(addHeaders, options)
	if err != nil {
		return nil, err
	}

	// merge header
	if headers == nil {
		headers = make(map[string]string)
	}

	for k, v := range addHeaders {
		if _, ok := headers[k]; !ok {
			headers[k] = v
		}
	}

	conn := client.Conn.WithContext(client.ctx)
	resp, err := conn.Do(method, bucketName, "", params, headers, data, 0, nil)

	// get response header
	respHeader, _ := FindOption(options, responseHeader, nil)
	if respHeader != nil {
		pRespHeader := respHeader.(*http.Header)
		*pRespHeader = resp.Headers
	}

	return resp, err
}

