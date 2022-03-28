// +build go1.7

package oss

import (
	"context"
	"io"
	"net/http"
)

// Bucket implements the operations of object.
type Bucket struct {
	Client     Client
	BucketName string
	ctx   context.Context
}

// WithContext support go1.7
func (bucket Bucket)WithContext(ctx context.Context) Bucket  {
	b := bucket // do a copy
	b.ctx = ctx
	return b
}


// Private with ctx support
func (bucket Bucket) do(method, objectName string, params map[string]interface{}, options []Option,
	data io.Reader, listener ProgressListener) (*Response, error) {
	headers := make(map[string]string)
	err := handleOptions(headers, options)
	if err != nil {
		return nil, err
	}

	err = CheckBucketName(bucket.BucketName)
	if len(bucket.BucketName) > 0 && err != nil {
		return nil, err
	}

	conn := bucket.Client.Conn.WithContext(bucket.ctx)
	resp, err := conn.Do(method, bucket.BucketName, objectName,
		params, headers, data, 0, listener)

	// get response header
	respHeader, _ := FindOption(options, responseHeader, nil)
	if respHeader != nil && resp != nil {
		pRespHeader := respHeader.(*http.Header)
		*pRespHeader = resp.Headers
	}

	return resp, err
}