package oss

import (
	"net/http"

	"github.com/aliyun/aliyun-oss-go-sdk/v3/oss/credentials"
	"github.com/aliyun/aliyun-oss-go-sdk/v3/oss/retry"
)

type Config struct {
	// The region to send requests to.
	Region string

	Endpoint string

	// RetryMaxAttempts specifies the maximum number attempts an API client will call
	// an operation that fails with a retryable error.
	RetryMaxAttempts int

	// Retryer guides how HTTP requests should be retried in case of recoverable failures.
	Retryer retry.Retryer

	// Allows you to enable the client to use path-style addressing, i.e., https://oss-cn-hangzhou.aliyuncs.com/bucket/key.
	// By default, the oss client will use virtual hosted addressing i.e., https://bucket.oss-cn-hangzhou.aliyuncs.com/key.
	UsePathStyle bool

	// The HTTP client to invoke API calls with. Defaults to client's default HTTP
	// implementation if nil.
	HTTPClient *http.Client

	CredentialsProvider credentials.CredentialsProvider
}

func NewConfig() *Config {
	return &Config{}
}

func (c Config) Copy() Config {
	cp := c
	return cp
}

func LoadDefaultConfig() *Config {
	config := &Config{
		Region:           "cn-hangzhou",
		RetryMaxAttempts: 3,
		UsePathStyle:     false,
	}

	return config
}

func (c *Config) WithRegion(region string) *Config {
	c.Region = region
	return c
}

func (c *Config) WithEndpoint(endpoint string) *Config {
	c.Endpoint = endpoint
	return c
}

func (c *Config) WithRetryMaxAttempts(value int) *Config {
	c.RetryMaxAttempts = value
	return c
}

func (c *Config) WithRetryer(retryer retry.Retryer) *Config {
	c.Retryer = retryer
	return c
}

func (c *Config) WithUsePathStyle(enable bool) *Config {
	c.UsePathStyle = enable
	return c
}

func (c *Config) WithHTTPClient(client *http.Client) *Config {
	c.HTTPClient = client
	return c
}

func (c *Config) WithCredentialsProvider(provider credentials.CredentialsProvider) *Config {
	c.CredentialsProvider = provider
	return c
}
