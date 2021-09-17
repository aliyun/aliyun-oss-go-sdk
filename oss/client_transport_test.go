package oss

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"testing"
)

func Test_ClientTransPort_CreateBucket(t *testing.T) {
	t.Logf("ok")

	rt := &tripper{}
	rt.AddResponse(200)

	client, err := New("cn-beijing", "abc", "dfe")
	if err != nil {
		t.Fatalf("New Error %v", err)
	}

	if client.HTTPClient == nil {
		client.HTTPClient = &http.Client{
			Transport: rt,
		}
	} else {
		client.HTTPClient.Transport = rt
	}
	client.SetTransport(rt)

	// Create Bucket
	err = client.CreateBucket("abc")
	if err != nil {
		t.Fatalf("CreateBucket Error %v", err)
	} else {
		t.Logf("CreateBucket Success")
	}

}

func Test_ClientTransPort_DeleteBucket(t *testing.T) {
	t.Logf("ok")

	rt := &tripper{}
	rt.AddResponse(204)

	client, err := New("cn-beijing", "abc", "dfe")
	if err != nil {
		t.Fatalf("New Error %v", err)
	}

	if client.HTTPClient == nil {
		client.HTTPClient = &http.Client{
			Transport: rt,
		}
	} else {
		client.HTTPClient.Transport = rt
	}
	client.SetTransport(rt)

	// Delete Bucket
	err = client.DeleteBucket("abc")
	if err != nil {
		t.Fatalf("DeleteBucket Error %v", err)
	} else {
		t.Logf("DeleteBucket Success")
	}
}

type tripper struct {
	req           int
	reqBodies     [][]byte
	responseCodes []int
}

func (r *tripper) RoundTrip(req *http.Request) (*http.Response, error) {
	defer func() {
		r.req++
	}()

	if req.Body != nil {
		dt, err := ioutil.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
		r.reqBodies = append(r.reqBodies, dt)
	}

	code := http.StatusOK
	if r.req < len(r.responseCodes) {
		code = r.responseCodes[r.req]
	}

	return &http.Response{
		StatusCode: code,
		Body:       ioutil.NopCloser(bytes.NewBufferString("{}")),
	}, nil
}

func (r *tripper) AddResponse(code int) {
	r.responseCodes = append(r.responseCodes, code)
}
