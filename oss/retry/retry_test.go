package retry

import (
	"errors"
	"io"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const (
	RetriesAttemptedCelling = 64
)

var noretryHttpStatus = []int{
	403, 300,
}

var noretryErrorPattern = []string{
	"123",
	"abc refused",
}

func TestEqualJitterBackoff(t *testing.T) {
	baseDelay := 1 * time.Second
	maxDelay := 20 * time.Second
	r := NewEqualJJitterBackoff(baseDelay, maxDelay)
	assert.NotNil(t, r)

	for i := 0; i < RetriesAttemptedCelling*2; i++ {
		delay, _ := r.BackoffDelay(i, nil)
		assert.Greater(t, delay, 0*time.Second)
		assert.Less(t, delay, maxDelay+1*time.Second)
	}
}

func TestFullJitterBackoff(t *testing.T) {
	baseDelay := 1 * time.Second
	maxDelay := 20 * time.Second
	r := NewFullJitterBackoff(baseDelay, maxDelay)
	assert.NotNil(t, r)

	for i := 0; i < RetriesAttemptedCelling*2; i++ {
		delay, _ := r.BackoffDelay(i, nil)
		assert.Greater(t, delay, 0*time.Second)
		assert.Less(t, delay, maxDelay+1*time.Second)
	}
}

func TestNewFixedDelayBackoff(t *testing.T) {
	maxDelay := 20 * time.Second
	r := NewFixedDelayBackoff(maxDelay)
	assert.NotNil(t, r)

	for i := 0; i < RetriesAttemptedCelling*2; i++ {
		delay, _ := r.BackoffDelay(i, nil)
		assert.Equal(t, maxDelay, delay)
	}
}

type statusCodeError struct {
	StatusCode int
}

func (e *statusCodeError) Error() string {
	return "error"
}

func (e *statusCodeError) HttpStatusCode() int {
	return e.StatusCode
}

func TestHTTPStatusCodeRetryable(t *testing.T) {
	r := &HTTPStatusCodeRetryable{}
	assert.NotNil(t, r)

	assert.False(t, r.IsErrorRetryable(nil))

	for _, code := range retryErrorCodes {
		assert.True(t, r.IsErrorRetryable(&statusCodeError{StatusCode: code}))
	}

	assert.True(t, r.IsErrorRetryable(&statusCodeError{StatusCode: 500}))
	assert.True(t, r.IsErrorRetryable(&statusCodeError{StatusCode: 502}))

	for _, code := range noretryHttpStatus {
		assert.False(t, r.IsErrorRetryable(&statusCodeError{StatusCode: code}))
	}
}

func TestConnectionErrorRetryable(t *testing.T) {
	var err error
	r := &ConnectionErrorRetryable{}
	assert.NotNil(t, r)

	assert.False(t, r.IsErrorRetryable(nil))

	err = &net.DNSError{Err: "error test", Name: "name", Server: "server", IsTimeout: true}
	assert.True(t, r.IsErrorRetryable(err))

	//support error
	for _, err := range retriableErrors {
		assert.True(t, r.IsErrorRetryable(err))
	}

	// not support error
	assert.False(t, r.IsErrorRetryable(io.ErrNoProgress))

	//support error pattern
	for _, pattern := range retriableErrorStrings {
		assert.True(t, r.IsErrorRetryable(errors.New(pattern)))
	}

	//not support error pattern
	for _, pattern := range noretryErrorPattern {
		assert.False(t, r.IsErrorRetryable(errors.New(pattern)))
	}
}

func TestNopRetryer(t *testing.T) {
	r := NopRetryer{}
	assert.NotNil(t, r)
	assert.False(t, r.IsErrorRetryable(nil))
	assert.Equal(t, 1, r.MaxAttempts())
	_, err := r.RetryDelay(1, nil)
	assert.NotNil(t, err)
}

func TestStandard(t *testing.T) {
	r := NewStandard()
	assert.NotNil(t, r)

	assert.Equal(t, defaultMaxAttempts, r.MaxAttempts())
	for i := 0; i < RetriesAttemptedCelling*2; i++ {
		delay, _ := r.RetryDelay(i, nil)
		assert.Less(t, delay, defaultMaxBackoff+1*time.Second)
	}

	assert.False(t, r.IsErrorRetryable(nil))

	for _, code := range retryErrorCodes {
		assert.True(t, r.IsErrorRetryable(&statusCodeError{StatusCode: code}))
	}

	assert.True(t, r.IsErrorRetryable(&statusCodeError{StatusCode: 500}))
	assert.True(t, r.IsErrorRetryable(&statusCodeError{StatusCode: 502}))

	for _, code := range noretryHttpStatus {
		assert.False(t, r.IsErrorRetryable(&statusCodeError{StatusCode: code}))
	}

	assert.True(t, r.IsErrorRetryable(&net.DNSError{Err: "error test", Name: "name", Server: "server", IsTimeout: true}))

	//support error
	for _, err := range retriableErrors {
		assert.True(t, r.IsErrorRetryable(err))
	}

	// not support error
	assert.False(t, r.IsErrorRetryable(io.ErrNoProgress))

	//support error pattern
	for _, pattern := range retriableErrorStrings {
		assert.True(t, r.IsErrorRetryable(errors.New(pattern)))
	}

	//not support error pattern
	for _, pattern := range noretryErrorPattern {
		assert.False(t, r.IsErrorRetryable(errors.New(pattern)))
	}

	//overwrite default
	r = NewStandard(
		func(ro *RetryOptions) { ro.MaxAttempts = 4 },
		func(ro *RetryOptions) { ro.MaxBackoff = 10 * time.Second },
		func(ro *RetryOptions) { ro.Backoff = NewFixedDelayBackoff(ro.MaxBackoff) },
		func(ro *RetryOptions) { ro.ErrorRetryables = []ErrorRetryable{} },
	)
	assert.Equal(t, 4, r.MaxAttempts())
	for i := 0; i < RetriesAttemptedCelling*2; i++ {
		delay, _ := r.RetryDelay(i, nil)
		assert.Equal(t, 10*time.Second, delay)
	}
	for _, code := range retryErrorCodes {
		assert.False(t, r.IsErrorRetryable(&statusCodeError{StatusCode: code}))
	}
	for _, pattern := range retriableErrorStrings {
		assert.False(t, r.IsErrorRetryable(errors.New(pattern)))
	}
}
