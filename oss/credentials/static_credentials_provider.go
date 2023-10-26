package credentials

import (
	"context"
)

type StaticCredentialsProvider struct {
	credentials Credentials
}

func NewStaticCredentialsProvider(id, secret, token string) CredentialsProvider {
	return StaticCredentialsProvider{
		credentials: Credentials{
			AccessKeyID:     id,
			AccessKeySecret: secret,
			SessionToken:    token,
		}}
}

func (s StaticCredentialsProvider) GetCredentials(ctx context.Context) (Credentials, error) {
	return s.credentials, nil
}
