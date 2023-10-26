package credentials

import (
	"context"
	"time"
)

type Credentials struct {
	AccessKeyID     string     // Access key ID
	AccessKeySecret string     // Access Key Secret
	SessionToken    string     // Session Token
	Expires         *time.Time // The time the credentials will expire at.
}

func (v Credentials) Expired() bool {
	if v.Expires != nil {
		return !v.Expires.After(time.Now().Round(0))
	}
	return false
}

func (v Credentials) HasKeys() bool {
	return len(v.AccessKeyID) > 0 && len(v.AccessKeySecret) > 0
}

type CredentialsProvider interface {
	GetCredentials(ctx context.Context) (Credentials, error)
}

type AnonymousCredentialsProvider struct{}

func (AnonymousCredentialsProvider) GetCredentials(ctx context.Context) (Credentials, error) {
	return Credentials{AccessKeyID: "", AccessKeySecret: ""}, nil
}
