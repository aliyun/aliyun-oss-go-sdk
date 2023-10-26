package signer

import (
	"context"
	"net/http"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/v3/oss/credentials"
)

const (
	SubResource = "SubResource"
)

type SigningContext struct {
	//input
	Product string
	Region  string
	Bucket  string
	Key     string
	Request *http.Request

	SubResource []string

	// output
	Credentials   *credentials.Credentials
	Time          time.Time
	SignedHeaders http.Header
	StringToSign  string
}

type Signer interface {
	Sign(ctx context.Context, signingCtx *SigningContext) error
	Presign(ctx context.Context, signingCtx *SigningContext) error
}

type NopSigner struct{}

func (NopSigner) Sign(ctx context.Context, signingCtx *SigningContext) error {
	return nil
}

func (NopSigner) Presign(ctx context.Context, signingCtx *SigningContext) error {
	return nil
}
