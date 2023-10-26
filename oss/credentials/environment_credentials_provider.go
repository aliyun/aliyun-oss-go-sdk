package credentials

import (
	"context"
	"fmt"
	"os"
)

type EnvironmentVariableCredentialsProvider struct {
	credentials *Credentials
}

func (s *EnvironmentVariableCredentialsProvider) GetCredentials(ctx context.Context) (Credentials, error) {
	var err error
	if s.credentials == nil {
		id := os.Getenv("OSS_ACCESS_KEY_ID")
		secret := os.Getenv("OSS_ACCESS_KEY_SECRET")
		token := os.Getenv("OSS_SESSION_TOKEN")
		if len(id) == 0 || len(secret) == 0 {
			err = fmt.Errorf("access key id or access key secret is empty!")
		} else {
			s.credentials = &Credentials{
				AccessKeyID:     id,
				AccessKeySecret: secret,
				SessionToken:    token,
			}
		}
	}
	return *s.credentials, err
}

func NewEnvironmentVariableCredentialsProvider() (CredentialsProvider, error) {
	provider := &EnvironmentVariableCredentialsProvider{}
	_, err := provider.GetCredentials(context.TODO())
	return provider, err
}
