package oss

import (
	"io"
	"net/http"
)

type OperationMetadata struct {
	values map[interface{}]interface{}
}

func (m OperationMetadata) Get(key interface{}) interface{} {
	return m.values[key]
}

func (m OperationMetadata) Clone() OperationMetadata {
	vs := make(map[interface{}]interface{}, len(m.values))
	for k, v := range m.values {
		vs[k] = v
	}

	return OperationMetadata{
		values: vs,
	}
}

func (m *OperationMetadata) Set(key, value interface{}) {
	if m.values == nil {
		m.values = map[interface{}]interface{}{}
	}
	m.values[key] = value
}

func (m OperationMetadata) Has(key interface{}) bool {
	if m.values == nil {
		return false
	}
	_, ok := m.values[key]
	return ok
}

type RequestCommon struct {
	Headers    map[string]string
	Parameters map[string]string
}

type ResultCommon struct {
	Status     string
	StatusCode int
	Headers    http.Header
	Metadata   OperationMetadata
}

type OperationInput struct {
	OperationName string

	Bucket string
	Key    string

	Method     string
	Headers    map[string]string
	Parameters map[string]string
	Body       io.Reader

	Metadata OperationMetadata
}

type OperationOutput struct {
	Input *OperationInput

	Status     string
	StatusCode int
	Headers    http.Header
	Body       io.ReadCloser

	Metadata OperationMetadata
}
