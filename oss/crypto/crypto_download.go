package osscrypto

import (
	"fmt"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

// DownloadFile with multi part mode, temporarily not supported
func (bucket CryptoBucket) DownloadFile(objectKey, filePath string, partSize int64, options ...oss.Option) error {
	return fmt.Errorf("CryptoBucket doesn't support DownloadFile")
}
