package oss

import (
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/v3/oss/credentials"
	"github.com/aliyun/aliyun-oss-go-sdk/v3/oss/readers"
	"github.com/aliyun/aliyun-oss-go-sdk/v3/oss/retry"
	"github.com/aliyun/aliyun-oss-go-sdk/v3/oss/signer"
	"github.com/aliyun/aliyun-oss-go-sdk/v3/oss/transport"
)

type Options struct {
	Region string

	Endpoint *url.URL

	RetryMaxAttempts int

	Retryer retry.Retryer

	Signer signer.Signer

	CredentialsProvider credentials.CredentialsProvider

	HttpClient *http.Client

	ResponseHandlers []func(*http.Response) error
}

func (c Options) Copy() Options {
	to := c
	to.ResponseHandlers = make([]func(*http.Response) error, len(c.ResponseHandlers))
	copy(to.ResponseHandlers, c.ResponseHandlers)
	return to
}

type Client struct {
	options Options
}

func New(cfg *Config, optFns ...func(*Options)) *Client {
	options := Options{
		Region:              cfg.Region,
		RetryMaxAttempts:    cfg.RetryMaxAttempts,
		Retryer:             cfg.Retryer,
		CredentialsProvider: cfg.CredentialsProvider,
		HttpClient:          cfg.HTTPClient,
	}
	resolveEndpoint(cfg, &options)
	resolveRetryer(cfg, &options)
	resolveHTTPClient(cfg, &options)
	resolveSigner(cfg, &options)

	for _, fn := range optFns {
		fn(&options)
	}

	client := &Client{
		options: options,
	}

	return client
}

func resolveEndpoint(cfg *Config, o *Options) {
	var scheme string
	var endpoint string
	if strings.HasPrefix(cfg.Endpoint, "http://") {
		scheme = "http"
		endpoint = cfg.Endpoint[len("http://"):]
	} else if strings.HasPrefix(endpoint, "https://") {
		scheme = "https"
		endpoint = cfg.Endpoint[len("https://"):]
	} else {
		scheme = "http"
		endpoint = cfg.Endpoint
	}

	strUrl := fmt.Sprintf("%s://%s", scheme, endpoint)
	o.Endpoint, _ = url.Parse(strUrl)
}

func resolveRetryer(cfg *Config, o *Options) {
	if o.Retryer != nil {
		return
	}

	o.Retryer = retry.NewStandard()
}

func resolveHTTPClient(cfg *Config, o *Options) {
	if o.HttpClient != nil {
		return
	}

	//TODO timeouts from config

	o.HttpClient = &http.Client{
		Transport: transport.NewTransportCustom(),
	}
}

func resolveSigner(cfg *Config, o *Options) {
	if o.Signer != nil {
		return
	}

	o.Signer = signer.SignerV1{}
}

func (c *Client) invokeOperation(ctx context.Context, input *OperationInput, optFns []func(*Options)) (output *OperationOutput, err error) {
	options := c.options.Copy()
	opOpt := Options{}

	for _, fn := range optFns {
		fn(&opOpt)
	}

	applyOperationOpt(&options, &opOpt)

	output, err = c.sendRequest(ctx, input, &options)

	if err != nil {
		return output, &OperationError{
			OperationName: input.OperationName,
			Err:           err}
	}

	return output, err
}

func (c *Client) sendRequest(ctx context.Context, input *OperationInput, opts *Options) (output *OperationOutput, err error) {
	// covert input into httpRequest
	if !isValidEndpoint(opts.Endpoint) {
		return output, NewErrParamInvalid("Endpoint")
	}

	// host & path
	host, path := buildURL(input, opts)
	strUrl := fmt.Sprintf("%s://%s%s", opts.Endpoint.Scheme, host, path)

	// querys
	if len(input.Parameters) > 0 {
		var buf bytes.Buffer
		for k, v := range input.Parameters {
			if buf.Len() > 0 {
				buf.WriteByte('&')
			}
			buf.WriteString(url.QueryEscape(k))
			if len(v) > 0 {
				buf.WriteString("=" + strings.Replace(url.QueryEscape(v), "+", "%20", -1))
			}
		}
		strUrl += "?" + buf.String()
	}

	request, err := http.NewRequestWithContext(ctx, input.Method, strUrl, nil)
	if err != nil {
		return output, err
	}

	// headers
	for k, v := range input.Headers {
		if len(k) > 0 && len(v) > 0 {
			request.Header.Add(k, v)
		}
	}
	request.Header.Set("User-Agent", defaultUserAgent())

	// body
	var body readers.ReadSeekerNopClose
	if input.Body == nil {
		body = readers.ReadSeekNopCloser(strings.NewReader(""))
	} else {
		body = readers.ReadSeekNopCloser(input.Body)
	}
	len, _ := body.GetLen()
	if len >= 0 && request.Header.Get("Content-Length") == "" {
		request.ContentLength = len
	}
	request.Body = body

	//signing context
	subResource, _ := input.Metadata.Get(signer.SubResource).([]string)
	signingCtx := &signer.SigningContext{
		Product:     "oss",
		Region:      opts.Region,
		Bucket:      input.Bucket,
		Key:         input.Key,
		Request:     request,
		SubResource: subResource,
	}

	// send http request
	response, err := c.sendHttpRequest(ctx, signingCtx, opts)

	if err != nil {
		return output, err
	}

	// covert http response into output context
	output = &OperationOutput{
		Input:      input,
		Status:     response.Status,
		StatusCode: response.StatusCode,
		Body:       response.Body,
		Headers:    response.Header,
	}

	// save other info by Metadata filed, ex. retry detail info
	//output.Metadata.Set()

	return output, err
}

func (c *Client) sendHttpRequest(ctx context.Context, signingCtx *signer.SigningContext, opts *Options) (response *http.Response, err error) {
	request := signingCtx.Request
	retryer := opts.Retryer
	body, _ := request.Body.(readers.ReadSeekerNopClose)
	bodyStart, _ := body.Seek(0, io.SeekCurrent)
	for tries := 1; tries <= retryer.MaxAttempts(); tries++ {
		if tries > 1 {
			delay, err := retryer.RetryDelay(tries, err)
			if err != nil {
				break
			}
			if err = sleepWithContext(ctx, delay); err != nil {
				err = &CanceledError{Err: err}
				break
			}

			if _, err = body.Seek(bodyStart, io.SeekStart); err != nil {
				break
			}
		}

		if response, err = c.sendHttpRequestOnce(ctx, signingCtx, opts); err == nil {
			break
		}

		if isContextError(ctx, &err) {
			err = &CanceledError{Err: err}
			break
		}

		if !readers.IsReaderSeekable(request.Body) {
			break
		}

		if !retryer.IsErrorRetryable(err) {
			break
		}
	}
	return response, err
}

func (c *Client) sendHttpRequestOnce(ctx context.Context, signingCtx *signer.SigningContext, opts *Options) (
	response *http.Response, err error,
) {
	cred, err := opts.CredentialsProvider.GetCredentials(ctx)
	if err != nil {
		return response, err
	}

	signingCtx.Credentials = &cred
	if err = c.options.Signer.Sign(ctx, signingCtx); err != nil {
		return response, err
	}

	if response, err = c.options.HttpClient.Do(signingCtx.Request); err != nil {
		return response, err
	}

	for _, fn := range opts.ResponseHandlers {
		if err = fn(response); err != nil {
			return response, err
		}
	}

	return response, err
}

func buildURL(input *OperationInput, opts *Options) (string, string) {
	var host = ""
	var path = ""

	if input == nil || opts == nil || opts.Endpoint == nil {
		return host, path
	}

	bucket := input.Bucket
	object := escapePath(input.Key, false)

	if bucket == "" {
		host = opts.Endpoint.Host
		path = "/"
	} else {
		host = bucket + "." + opts.Endpoint.Host
		path = "/" + object
	}

	return host, path
}

func serviceErrorResponseHandler(response *http.Response) error {
	if response.StatusCode/100 == 2 {
		return nil
	}

	timestamp, err := time.Parse(http.TimeFormat, response.Header.Get("Date"))
	if err != nil {
		timestamp = time.Now()
	}

	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)

	se := &ServiceError{
		StatusCode:    response.StatusCode,
		Code:          "BadErrorResponse",
		RequestID:     response.Header.Get("x-oss-request-id"),
		Timestamp:     timestamp,
		RequestTarget: fmt.Sprintf("%s %s", response.Request.Method, response.Request.URL),
		Snapshot:      body,
	}

	if err != nil {
		se.Message = fmt.Sprintf("The body of the response was not readable, due to :%s", err.Error())
		return se
	}

	err = xml.Unmarshal(body, &se)
	if err != nil {
		len := len(body)
		if len > 256 {
			len = 256
		}
		se.Message = fmt.Sprintf("Failed to parse xml from response body due to: %s. With part response body %s.", err.Error(), string(body[:len]))
		return se
	}
	return se
}

func applyOperationOpt(c *Options, op *Options) {
	if c == nil || op == nil {
		return
	}

	if op.Endpoint != nil {
		c.Endpoint = op.Endpoint
	}

	if op.RetryMaxAttempts > 0 {
		c.RetryMaxAttempts = op.RetryMaxAttempts
	}

	if op.Retryer != nil {
		c.Retryer = op.Retryer
	}

	if c.Retryer == nil {
		c.Retryer = retry.NopRetryer{}
	}

	//response handler
	handlers := []func(*http.Response) error{
		serviceErrorResponseHandler,
	}
	handlers = append(handlers, c.ResponseHandlers...)
	handlers = append(handlers, op.ResponseHandlers...)
	c.ResponseHandlers = handlers
}
