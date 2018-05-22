package oss

import (
	"time"
)

// HTTPTimeout defines http timeout.
type HTTPTimeout struct {
	ConnectTimeout   time.Duration
	ReadWriteTimeout time.Duration
	HeaderTimeout    time.Duration
	LongTimeout      time.Duration
	IdleConnTimeout  time.Duration
}

// Config defines oss configuration
type Config struct {
	Endpoint        string      // oss endpoint
	AccessKeyID     string      // accessId
	AccessKeySecret string      // accessKey
	RetryTimes      uint        // retry count by default it's 5.
	UserAgent       string      // SDK name/version/system information
	IsDebug         bool        // enable debug mode. Default is false.
	Timeout         uint        // timeout in seconds. By default it's 60.
	SecurityToken   string      // STS Token
	IsCname         bool        // if cname is in the endpoint.
	HTTPTimeout     HTTPTimeout // HTTP timeout
	IsUseProxy      bool        // flag of using proxy.
	ProxyHost       string      // flag of using proxy host.
	IsAuthProxy     bool        // flag of needs authentication
	ProxyUser       string      // proxy user
	ProxyPassword   string      // proxy password
	IsEnableMD5     bool        // flag of enabling MD5 for upload
	MD5Threshold    int64       // Memory footprint threshold for each MD5 computation (16MB is the default), in byte. When the data is more than that, temp file is used.
	IsEnableCRC     bool        // flag of enabling CRC for upload.
}

// getDefaultOssConfig gets the default config.
func getDefaultOssConfig() *Config {
	config := Config{}

	config.Endpoint = ""
	config.AccessKeyID = ""
	config.AccessKeySecret = ""
	config.RetryTimes = 5
	config.IsDebug = false
	config.UserAgent = userAgent
	config.Timeout = 60 // seconds
	config.SecurityToken = ""
	config.IsCname = false

	config.HTTPTimeout.ConnectTimeout = time.Second * 30   // 30s
	config.HTTPTimeout.ReadWriteTimeout = time.Second * 60 // 60s
	config.HTTPTimeout.HeaderTimeout = time.Second * 60    // 60s
	config.HTTPTimeout.LongTimeout = time.Second * 300     // 300s
	config.HTTPTimeout.IdleConnTimeout = time.Second * 50  // 50s

	config.IsUseProxy = false
	config.ProxyHost = ""
	config.IsAuthProxy = false
	config.ProxyUser = ""
	config.ProxyPassword = ""

	config.MD5Threshold = 16 * 1024 * 1024 // 16MB
	config.IsEnableMD5 = false
	config.IsEnableCRC = true

	return &config
}
