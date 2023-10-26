package retry

import (
	"errors"
	"io"
	"net"
	"strings"
)

type HTTPStatusCodeRetryable struct {
}

var retryErrorCodes = []int{
	400, // Bad request
	401, // Unauthorized
	408, // Request Timeout
	429, // Rate exceeded.
}

func (*HTTPStatusCodeRetryable) IsErrorRetryable(err error) bool {
	var v interface{ HttpStatusCode() int }
	if errors.As(err, &v) {
		code := v.HttpStatusCode()
		if code >= 500 {
			return true
		}
		for _, e := range retryErrorCodes {
			if code == e {
				return true
			}
		}
	}
	return false
}

type ConnectionErrorRetryable struct{}

var retriableErrorStrings = []string{
	"connection reset",
	"connection refused",
	"use of closed network connection",
	"unexpected EOF reading trailer",
	"transport connection broken",
	"server closed idle connection",
	"bad record MAC",
	"stream error:",
	"tls: use of closed connection",
}

var retriableErrors = []error{
	io.EOF,
	io.ErrUnexpectedEOF,
}

func (*ConnectionErrorRetryable) IsErrorRetryable(err error) bool {
	if err != nil {
		if r, ok := err.(net.Error); ok && (r.Temporary() || r.Timeout()) {
			return true
		}

		for _, retriableErr := range retriableErrors {
			if err == retriableErr {
				return true
			}
		}

		errString := err.Error()
		for _, phrase := range retriableErrorStrings {
			if strings.Contains(errString, phrase) {
				return true
			}
		}
	}
	return false
}
