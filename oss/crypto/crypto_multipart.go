package osscrypto

import (
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

// PartCryptoContext save encryption or decryption information
type PartCryptoContext struct {
	ContentCipher ContentCipher
	DataSize      int64
	PartSize      int64
}

// Valid judge PartCryptoContext is valid or not
func (pcc PartCryptoContext) Valid() bool {
	if pcc.ContentCipher == nil || pcc.DataSize == 0 || pcc.PartSize == 0 {
		return false
	}
	return true
}

// InitiateMultipartUpload initializes multipart upload for client encryption
// cryptoContext.PartSize and cryptoContext.DataSize are input parameter
// cryptoContext.PartSize must aligned to the secret iv length
// cryptoContext.ContentCipher is output parameter
// cryptoContext will be used in next API
func (bucket CryptoBucket) InitiateMultipartUpload(objectKey string, cryptoContext *PartCryptoContext, options ...oss.Option) (oss.InitiateMultipartUploadResult, error) {
	options = bucket.AddEncryptionUaSuffix(options)
	var imur oss.InitiateMultipartUploadResult
	if cryptoContext == nil {
		return imur, fmt.Errorf("error,cryptoContext is nil")
	}

	if cryptoContext.PartSize <= 0 {
		return imur, fmt.Errorf("invalid PartCryptoContext's PartSize %d", cryptoContext.PartSize)
	}

	cc, err := bucket.ContentCipherBuilder.ContentCipher()
	if err != nil {
		return imur, err
	}

	if cryptoContext.PartSize%int64(cc.GetAlignLen()) != 0 {
		return imur, fmt.Errorf("PartCryptoContext's PartSize must be aligned to %d", cc.GetAlignLen())
	}

	opts := addCryptoHeaders(options, cc.GetCipherData())
	if cryptoContext.DataSize > 0 {
		opts = append(opts, oss.Meta(OssClientSideEncryptionDataSize, strconv.FormatInt(cryptoContext.DataSize, 10)))
	}
	opts = append(opts, oss.Meta(OssClientSideEncryptionPartSize, strconv.FormatInt(cryptoContext.PartSize, 10)))

	imur, err = bucket.Bucket.InitiateMultipartUpload(objectKey, opts...)
	if err == nil {
		cryptoContext.ContentCipher = cc
	}
	return imur, err
}

// UploadPart uploads parts to oss, the part data are encrypted automaticly on client side
// cryptoContext is the input parameter
func (bucket CryptoBucket) UploadPart(imur oss.InitiateMultipartUploadResult, reader io.Reader,
	partSize int64, partNumber int, cryptoContext PartCryptoContext, options ...oss.Option) (oss.UploadPart, error) {
	options = bucket.AddEncryptionUaSuffix(options)
	var uploadPart oss.UploadPart
	if cryptoContext.ContentCipher == nil {
		return uploadPart, fmt.Errorf("error,cryptoContext is nil or cryptoContext.ContentCipher is nil")
	}

	if partNumber < 1 {
		return uploadPart, fmt.Errorf("partNumber:%d is smaller than 1", partNumber)
	}

	if cryptoContext.PartSize%int64(cryptoContext.ContentCipher.GetAlignLen()) != 0 {
		return uploadPart, fmt.Errorf("PartCryptoContext's PartSize must be aligned to %d", cryptoContext.ContentCipher.GetAlignLen())
	}

	cipherData := cryptoContext.ContentCipher.GetCipherData().Clone()
	// caclulate iv based on part number
	if partNumber > 1 {
		cipherData.SeekIV(uint64(partNumber-1) * uint64(cryptoContext.PartSize))
	}

	// for parallel upload part
	partCC, _ := cryptoContext.ContentCipher.Clone(cipherData)

	cryptoReader, err := partCC.EncryptContent(reader)
	if err != nil {
		return uploadPart, err
	}

	request := &oss.UploadPartRequest{
		InitResult: &imur,
		Reader:     cryptoReader,
		PartSize:   partCC.GetEncryptedLen(partSize),
		PartNumber: partNumber,
	}

	opts := addCryptoHeaders(options, partCC.GetCipherData())
	if cryptoContext.DataSize > 0 {
		opts = append(opts, oss.Meta(OssClientSideEncryptionDataSize, strconv.FormatInt(cryptoContext.DataSize, 10)))
	}
	opts = append(opts, oss.Meta(OssClientSideEncryptionPartSize, strconv.FormatInt(cryptoContext.PartSize, 10)))

	result, err := bucket.Bucket.DoUploadPart(request, opts)
	return result.Part, err
}

// UploadPartFromFile uploads part from the file, the part data are encrypted automaticly on client side
// cryptoContext is the input parameter
func (bucket CryptoBucket) UploadPartFromFile(imur oss.InitiateMultipartUploadResult, filePath string,
	startPosition, partSize int64, partNumber int, cryptoContext PartCryptoContext, options ...oss.Option) (oss.UploadPart, error) {
	options = bucket.AddEncryptionUaSuffix(options)
	var uploadPart = oss.UploadPart{}
	if cryptoContext.ContentCipher == nil {
		return uploadPart, fmt.Errorf("error,cryptoContext is nil or cryptoContext.ContentCipher is nil")
	}

	if cryptoContext.PartSize%int64(cryptoContext.ContentCipher.GetAlignLen()) != 0 {
		return uploadPart, fmt.Errorf("PartCryptoContext's PartSize must be aligned to %d", cryptoContext.ContentCipher.GetAlignLen())
	}

	fd, err := os.Open(filePath)
	if err != nil {
		return uploadPart, err
	}
	defer fd.Close()
	fd.Seek(startPosition, os.SEEK_SET)

	if partNumber < 1 {
		return uploadPart, fmt.Errorf("partNumber:%d is smaller than 1", partNumber)
	}

	cipherData := cryptoContext.ContentCipher.GetCipherData().Clone()
	// calculate iv based on part number
	if partNumber > 1 {
		cipherData.SeekIV(uint64(partNumber-1) * uint64(cryptoContext.PartSize))
	}

	// for parallel upload part
	partCC, _ := cryptoContext.ContentCipher.Clone(cipherData)
	cryptoReader, err := partCC.EncryptContent(fd)
	if err != nil {
		return uploadPart, err
	}

	encryptedLen := partCC.GetEncryptedLen(partSize)
	opts := addCryptoHeaders(options, partCC.GetCipherData())
	if cryptoContext.DataSize > 0 {
		opts = append(opts, oss.Meta(OssClientSideEncryptionDataSize, strconv.FormatInt(cryptoContext.DataSize, 10)))
	}
	opts = append(opts, oss.Meta(OssClientSideEncryptionPartSize, strconv.FormatInt(cryptoContext.PartSize, 10)))

	request := &oss.UploadPartRequest{
		InitResult: &imur,
		Reader:     cryptoReader,
		PartSize:   encryptedLen,
		PartNumber: partNumber,
	}
	result, err := bucket.Bucket.DoUploadPart(request, opts)
	return result.Part, err
}

// UploadPartCopy uploads part copy
func (bucket CryptoBucket) UploadPartCopy(imur oss.InitiateMultipartUploadResult, srcBucketName, srcObjectKey string,
	startPosition, partSize int64, partNumber int, cryptoContext PartCryptoContext, options ...oss.Option) (oss.UploadPart, error) {
	var uploadPart = oss.UploadPart{}
	return uploadPart, fmt.Errorf("CryptoBucket doesn't support UploadPartCopy")
}
