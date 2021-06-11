package osscrypto

import (
	"fmt"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

// CopyFile with multi part mode, temporarily not supported
func (bucket CryptoBucket) CopyFile(srcBucketName, srcObjectKey, destObjectKey string, partSize int64, options ...oss.Option) error {
	return fmt.Errorf("CryptoBucket doesn't support CopyFile")
}
