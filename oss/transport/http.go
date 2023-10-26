package transport

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"time"
)

// Defaults for the Transport
var (
	DefaultTransportConnectTimeout        = 5 * time.Second
	DefaultTransportReadWriteTimeout      = 10 * time.Second
	DefaultTransportIdleConnectionTimeout = 50 * time.Second
	DefaultTransportExpectContinueTimeout = 1 * time.Second
	DefaultDialKeepAliveTimeout           = 30 * time.Second

	DefaultTransportMaxConnections = 100

	// Default to TLS 1.2 for all HTTPS requests.
	DefaultTransportTLSMinVersion uint16 = tls.VersionTLS12
)

// Dialer
type Dialer struct {
	net.Dialer
	timeout time.Duration
}

func newDialer() *Dialer {
	dialer := &Dialer{
		Dialer: net.Dialer{
			Timeout:   DefaultTransportConnectTimeout,
			KeepAlive: DefaultDialKeepAliveTimeout,
		},
		timeout: DefaultTransportReadWriteTimeout,
	}
	return dialer
}

func (d *Dialer) Dial(network, address string) (net.Conn, error) {
	return d.DialContext(context.Background(), network, address)
}

func (d *Dialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	c, err := d.Dialer.DialContext(ctx, network, address)
	if err != nil {
		return c, err
	}

	t := &timeoutConn{
		Conn:    c,
		timeout: d.timeout,
	}
	return t, t.nudgeDeadline()
}

// A net.Conn with Read/Write timeout
type timeoutConn struct {
	net.Conn
	timeout time.Duration
}

func (c *timeoutConn) nudgeDeadline() error {
	if c.timeout > 0 {
		return c.SetDeadline(time.Now().Add(c.timeout))
	}
	return nil
}

func (c *timeoutConn) Read(b []byte) (n int, err error) {
	n, err = c.Conn.Read(b)
	if err == nil && n > 0 && c.timeout > 0 {
		err = c.nudgeDeadline()
	}
	return n, err
}

func (c *timeoutConn) Write(b []byte) (n int, err error) {
	n, err = c.Conn.Write(b)
	if err == nil && n > 0 && c.timeout > 0 {
		err = c.nudgeDeadline()
	}
	return n, err
}

func NewTransportCustom(fns ...func(*http.Transport)) http.RoundTripper {
	tr := &http.Transport{
		DialContext:           newDialer().DialContext,
		TLSHandshakeTimeout:   DefaultTransportConnectTimeout,
		MaxConnsPerHost:       DefaultTransportMaxConnections,
		IdleConnTimeout:       DefaultTransportConnectTimeout,
		ExpectContinueTimeout: DefaultTransportExpectContinueTimeout,
		TLSClientConfig: &tls.Config{
			MinVersion: DefaultTransportTLSMinVersion,
		},
	}

	for _, fn := range fns {
		fn(tr)
	}

	return tr
}
