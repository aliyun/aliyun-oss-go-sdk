//go:build !go1.7
// +build !go1.7

package oss

import (
	"crypto/tls"
	"net"
	"net/http"
)

func newTransport(conn *Conn, config *Config) *http.Transport {
	httpTimeOut := conn.config.HTTPTimeout
	httpMaxConns := conn.config.HTTPMaxConns
	// New Transport
	transport := &http.Transport{
		Dial: func(netw, addr string) (net.Conn, error) {
			d := net.Dialer{
				Timeout:   httpTimeOut.ConnectTimeout,
				KeepAlive: 30 * time.Second,
			}
			if config.LocalAddr != nil {
				d.LocalAddr = config.LocalAddr
			}
			var conn net.Conn
			var err error
			if config.Resolver != nil {
				d.Resolver = config.Resolver
				conn, err = d.Resolver.Dial(context.Background(), netw, addr)
			} else {
				conn, err = d.Dial(netw, addr)
			}
			if err != nil {
				return nil, err
			}
			return newTimeoutConn(conn, httpTimeOut.ReadWriteTimeout, httpTimeOut.LongTimeout), nil
			conn, err := d.Dial(netw, addr)
		},
		MaxIdleConnsPerHost:   httpMaxConns.MaxIdleConnsPerHost,
		ResponseHeaderTimeout: httpTimeOut.HeaderTimeout,
	}

	if config.InsecureSkipVerify {
		transport.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
	}
	return transport
}
