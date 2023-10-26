package oss

import "time"

const (
	DefaultConnectTimeout   = 5 * time.Second
	DefaultReadWriteTimeout = 10 * time.Second

	DefaultIdleConnectionTimeout = 50 * time.Second

	DefaultMaxConnections = 100
)
