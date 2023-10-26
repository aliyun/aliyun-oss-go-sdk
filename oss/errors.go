package oss

import (
	"encoding/xml"
	"fmt"
	"time"
)

type ServiceError struct {
	XMLName   xml.Name `xml:"Error"`
	Code      string   `xml:"Code"`
	Message   string   `xml:"Message"`
	RequestID string   `xml:"RequestId"`
	EC        string   `xml:"EC"`

	StatusCode    int
	Snapshot      []byte
	Timestamp     time.Time
	RequestTarget string
}

func (e *ServiceError) Error() string {
	return fmt.Sprintf(`Error returned by Service. Http Status Code: %d. Error Code: %s. Request Id: %s. Message: %s.
	EC: %s.
	Timestamp: %s.
	Request Endpoint: %s.`,
		e.StatusCode, e.Code, e.RequestID, e.Message, e.EC, e.Timestamp, e.RequestTarget)
}

func (e *ServiceError) HttpStatusCode() int {
	return e.StatusCode
}

type ClientError struct {
	Code    string
	Message string
	Err     error
}

func (e *ClientError) Unwrap() error { return e.Err }

func (e *ClientError) Error() string {
	return fmt.Sprintf("client error: %v", e.Err)
}

type OperationError struct {
	OperationName string
	Err           error
}

func (e *OperationError) Operation() string { return e.OperationName }

func (e *OperationError) Unwrap() error { return e.Err }

func (e *OperationError) Error() string {
	return fmt.Sprintf("operation error %s: %v", e.OperationName, e.Err)
}

type DeserializationError struct {
	Err      error
	Snapshot []byte
}

func (e *DeserializationError) Error() string {
	const msg = "deserialization failed"
	if e.Err == nil {
		return msg
	}
	return fmt.Sprintf("%s, %v", msg, e.Err)
}

func (e *DeserializationError) Unwrap() error { return e.Err }

type SerializationError struct {
	Err error
}

func (e *SerializationError) Error() string {
	const msg = "serialization failed"
	if e.Err == nil {
		return msg
	}
	return fmt.Sprintf("%s: %v", msg, e.Err)
}

func (e *SerializationError) Unwrap() error { return e.Err }

type CanceledError struct {
	Err error
}

func (*CanceledError) CanceledError() bool { return true }

func (e *CanceledError) Unwrap() error {
	return e.Err
}

func (e *CanceledError) Error() string {
	return fmt.Sprintf("canceled, %v", e.Err)
}

/*
func IsServiceError(err error) (failure ServiceError, ok bool) {
	failure, ok = err.(ServiceError)
	return
}
*/
