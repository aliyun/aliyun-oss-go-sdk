package osscrypto

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"hash"
	"hash/crc64"
	"io"
	"net/http"
	"os"
	"strconv"

	kms "github.com/aliyun/alibaba-cloud-sdk-go/services/kms"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

// MasterCipherManager is interface for getting master key with MatDesc(material desc)
// If you may use different master keys for encrypting and decrypting objects,each master
// key must have a unique, non-emtpy, unalterable MatDesc(json string format) and you must provide this interface
// If you always use the same master key for encrypting and decrypting objects, MatDesc
// can be empty and you don't need to provide this interface
//
// matDesc map[string]string:is converted by matDesc json string
// return: []string  the secret key information,such as {"rsa-public-key","rsa-private-key"} or {"non-rsa-key"}
type MasterCipherManager interface {
	GetMasterKey(matDesc map[string]string) ([]string, error)
}

// ExtraCipherBuilder is interface for creating a decrypt ContentCipher with Envelope
// If the objects you need to decrypt are neither encrypted with ContentCipherBuilder
// you provided, nor encrypted with rsa and ali kms master keys, you must provide this interface
//
// ContentCipher  the interface used to decrypt objects
type ExtraCipherBuilder interface {
	GetDecryptCipher(envelope Envelope, cm MasterCipherManager) (ContentCipher, error)
}

// CryptoBucketOption CryptoBucket option such as SetAliKmsClient, SetMasterCipherManager, SetDecryptCipherManager.
type CryptoBucketOption func(*CryptoBucket)

// SetAliKmsClient set field AliKmsClient of CryptoBucket
// If the objects you need to decrypt are encrypted with ali kms master key,but not with ContentCipherBuilder
// you provided, you must provide this interface
func SetAliKmsClient(client *kms.Client) CryptoBucketOption {
	return func(bucket *CryptoBucket) {
		bucket.AliKmsClient = client
	}
}

// SetMasterCipherManager set field MasterCipherManager of CryptoBucket
func SetMasterCipherManager(manager MasterCipherManager) CryptoBucketOption {
	return func(bucket *CryptoBucket) {
		bucket.MasterCipherManager = manager
	}
}

// SetExtraCipherBuilder set field ExtraCipherBuilder of CryptoBucket
func SetExtraCipherBuilder(extraBuilder ExtraCipherBuilder) CryptoBucketOption {
	return func(bucket *CryptoBucket) {
		bucket.ExtraCipherBuilder = extraBuilder
	}
}

// DefaultExtraCipherBuilder is Default implementation of the ExtraCipherBuilder for rsa and kms master keys
type DefaultExtraCipherBuilder struct {
	AliKmsClient *kms.Client
}

// GetDecryptCipher is used to get ContentCipher for decrypt object
func (decb *DefaultExtraCipherBuilder) GetDecryptCipher(envelope Envelope, cm MasterCipherManager) (ContentCipher, error) {
	if cm == nil {
		return nil, fmt.Errorf("DefaultExtraCipherBuilder GetDecryptCipher error,MasterCipherManager is nil")
	}

	if envelope.CEKAlg != AesCtrAlgorithm {
		return nil, fmt.Errorf("DefaultExtraCipherBuilder GetDecryptCipher error,not supported content algorithm %s", envelope.CEKAlg)
	}

	if envelope.WrapAlg != RsaCryptoWrap && envelope.WrapAlg != KmsAliCryptoWrap {
		return nil, fmt.Errorf("DefaultExtraCipherBuilder GetDecryptCipher error,not supported envelope wrap algorithm %s", envelope.WrapAlg)
	}

	matDesc := make(map[string]string)
	if envelope.MatDesc != "" {
		err := json.Unmarshal([]byte(envelope.MatDesc), &matDesc)
		if err != nil {
			return nil, err
		}
	}

	masterKeys, err := cm.GetMasterKey(matDesc)
	if err != nil {
		return nil, err
	}

	var contentCipher ContentCipher
	if envelope.WrapAlg == RsaCryptoWrap {
		// for rsa master key
		if len(masterKeys) != 2 {
			return nil, fmt.Errorf("rsa keys count must be 2,now is %d", len(masterKeys))
		}
		rsaCipher, err := CreateMasterRsa(matDesc, masterKeys[0], masterKeys[1])
		if err != nil {
			return nil, err
		}
		aesCtrBuilder := CreateAesCtrCipher(rsaCipher)
		contentCipher, err = aesCtrBuilder.ContentCipherEnv(envelope)

	} else if envelope.WrapAlg == KmsAliCryptoWrap {
		// for kms master key
		if len(masterKeys) != 1 {
			return nil, fmt.Errorf("non-rsa keys count must be 1,now is %d", len(masterKeys))
		}

		if decb.AliKmsClient == nil {
			return nil, fmt.Errorf("aliyun kms client is nil")
		}

		kmsCipher, err := CreateMasterAliKms(matDesc, masterKeys[0], decb.AliKmsClient)
		if err != nil {
			return nil, err
		}
		aesCtrBuilder := CreateAesCtrCipher(kmsCipher)
		contentCipher, err = aesCtrBuilder.ContentCipherEnv(envelope)
	} else {
		// to do
		// for master keys which are neither rsa nor kms
	}

	return contentCipher, err
}

// CryptoBucket implements the operations for encrypting and decrypting objects
// ContentCipherBuilder is used to encrypt and decrypt objects by default
// when the object's MatDesc which you want to decrypt is emtpy or same to the
// master key's MatDesc you provided in ContentCipherBuilder, sdk try to
// use ContentCipherBuilder to decrypt
type CryptoBucket struct {
	oss.Bucket
	ContentCipherBuilder ContentCipherBuilder
	ExtraCipherBuilder   ExtraCipherBuilder
	MasterCipherManager  MasterCipherManager
	AliKmsClient         *kms.Client
}

// GetCryptoBucket create a client encyrption bucket
func GetCryptoBucket(client *oss.Client, bucketName string, builder ContentCipherBuilder,
	options ...CryptoBucketOption) (*CryptoBucket, error) {
	var cryptoBucket CryptoBucket
	cryptoBucket.Client = *client
	cryptoBucket.BucketName = bucketName
	cryptoBucket.ContentCipherBuilder = builder

	for _, option := range options {
		option(&cryptoBucket)
	}

	if cryptoBucket.ExtraCipherBuilder == nil {
		cryptoBucket.ExtraCipherBuilder = &DefaultExtraCipherBuilder{AliKmsClient: cryptoBucket.AliKmsClient}
	}

	return &cryptoBucket, nil
}

// PutObject creates a new object and encyrpt it on client side when uploading to oss
func (bucket CryptoBucket) PutObject(objectKey string, reader io.Reader, options ...oss.Option) error {
	options = bucket.AddEncryptionUaSuffix(options)
	cc, err := bucket.ContentCipherBuilder.ContentCipher()
	if err != nil {
		return err
	}

	cryptoReader, err := cc.EncryptContent(reader)
	if err != nil {
		return err
	}

	var request *oss.PutObjectRequest
	srcLen, err := oss.GetReaderLen(reader)
	if err != nil {
		request = &oss.PutObjectRequest{
			ObjectKey: objectKey,
			Reader:    cryptoReader,
		}
	} else {
		encryptedLen := cc.GetEncryptedLen(srcLen)
		request = &oss.PutObjectRequest{
			ObjectKey: objectKey,
			Reader:    oss.LimitReadCloser(cryptoReader, encryptedLen),
		}
	}

	opts := addCryptoHeaders(options, cc.GetCipherData())
	resp, err := bucket.DoPutObject(request, opts)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return err
}

// GetObject downloads the object from oss
// If the object is encrypted, sdk decrypt it automaticly
func (bucket CryptoBucket) GetObject(objectKey string, options ...oss.Option) (io.ReadCloser, error) {
	options = bucket.AddEncryptionUaSuffix(options)
	result, err := bucket.DoGetObject(&oss.GetObjectRequest{ObjectKey: objectKey}, options)
	if err != nil {
		return nil, err
	}
	return result.Response, nil
}

// GetObjectToFile downloads the object from oss to local file
// If the object is encrypted, sdk decrypt it automaticly
func (bucket CryptoBucket) GetObjectToFile(objectKey, filePath string, options ...oss.Option) error {
	options = bucket.AddEncryptionUaSuffix(options)
	tempFilePath := filePath + oss.TempFileSuffix

	// Calls the API to actually download the object. Returns the result instance.
	result, err := bucket.DoGetObject(&oss.GetObjectRequest{ObjectKey: objectKey}, options)
	if err != nil {
		return err
	}
	defer result.Response.Close()

	// If the local file does not exist, create a new one. If it exists, overwrite it.
	fd, err := os.OpenFile(tempFilePath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, oss.FilePermMode)
	if err != nil {
		return err
	}

	// Copy the data to the local file path.
	_, err = io.Copy(fd, result.Response.Body)
	fd.Close()
	if err != nil {
		return err
	}

	// Compares the CRC value
	hasRange, _, _ := oss.IsOptionSet(options, oss.HTTPHeaderRange)
	encodeOpt, _ := oss.FindOption(options, oss.HTTPHeaderAcceptEncoding, nil)
	acceptEncoding := ""
	if encodeOpt != nil {
		acceptEncoding = encodeOpt.(string)
	}
	if bucket.GetConfig().IsEnableCRC && !hasRange && acceptEncoding != "gzip" {
		result.Response.ClientCRC = result.ClientCRC.Sum64()
		err = oss.CheckCRC(result.Response, "GetObjectToFile")
		if err != nil {
			os.Remove(tempFilePath)
			return err
		}
	}

	return os.Rename(tempFilePath, filePath)
}

// DoGetObject is the actual API that gets the encrypted or not encrypted object.
// It's the internal function called by other public APIs.
func (bucket CryptoBucket) DoGetObject(request *oss.GetObjectRequest, options []oss.Option) (*oss.GetObjectResult, error) {
	options = bucket.AddEncryptionUaSuffix(options)

	// first,we must head object
	metaInfo, err := bucket.GetObjectDetailedMeta(request.ObjectKey)
	if err != nil {
		return nil, err
	}

	isEncryptedObj := isEncryptedObject(metaInfo)
	if !isEncryptedObj {
		return bucket.Bucket.DoGetObject(request, options)
	}

	envelope, err := getEnvelopeFromHeader(metaInfo)
	if err != nil {
		return nil, err
	}

	if !isValidContentAlg(envelope.CEKAlg) {
		return nil, fmt.Errorf("not supported content algorithm %s,object:%s", envelope.CEKAlg, request.ObjectKey)
	}

	if !envelope.IsValid() {
		return nil, fmt.Errorf("getEnvelopeFromHeader error,object:%s", request.ObjectKey)
	}

	// use ContentCipherBuilder to decrpt object by default
	encryptMatDesc := bucket.ContentCipherBuilder.GetMatDesc()
	var cc ContentCipher
	err = nil
	if envelope.MatDesc == encryptMatDesc {
		cc, err = bucket.ContentCipherBuilder.ContentCipherEnv(envelope)
	} else {
		cc, err = bucket.ExtraCipherBuilder.GetDecryptCipher(envelope, bucket.MasterCipherManager)
	}

	if err != nil {
		return nil, fmt.Errorf("%s,object:%s", err.Error(), request.ObjectKey)
	}

	discardFrontAlignLen := int64(0)
	uRange, err := oss.GetRangeConfig(options)
	if err != nil {
		return nil, err
	}

	if uRange != nil && uRange.HasStart {
		// process range to align key size
		adjustStart := adjustRangeStart(uRange.Start, cc)
		discardFrontAlignLen = uRange.Start - adjustStart
		if discardFrontAlignLen > 0 {
			uRange.Start = adjustStart
			options = oss.DeleteOption(options, oss.HTTPHeaderRange)
			options = append(options, oss.NormalizedRange(oss.GetRangeString(*uRange)))
		}

		// seek iv
		cipherData := cc.GetCipherData().Clone()
		cipherData.SeekIV(uint64(adjustStart))
		cc, _ = cc.Clone(cipherData)
	}

	params, _ := oss.GetRawParams(options)
	resp, err := bucket.Do("GET", request.ObjectKey, params, options, nil, nil)
	if err != nil {
		return nil, err
	}

	result := &oss.GetObjectResult{
		Response: resp,
	}

	// CRC
	var crcCalc hash.Hash64
	hasRange, _, _ := oss.IsOptionSet(options, oss.HTTPHeaderRange)
	if bucket.GetConfig().IsEnableCRC && !hasRange {
		crcCalc = crc64.New(oss.CrcTable())
		result.ServerCRC = resp.ServerCRC
		result.ClientCRC = crcCalc
	}

	// Progress
	listener := oss.GetProgressListener(options)
	contentLen, _ := strconv.ParseInt(resp.Headers.Get(oss.HTTPHeaderContentLength), 10, 64)
	resp.Body = oss.TeeReader(resp.Body, crcCalc, contentLen, listener, nil)
	resp.Body, err = cc.DecryptContent(resp.Body)
	if err == nil && discardFrontAlignLen > 0 {
		resp.Body = &oss.DiscardReadCloser{
			RC:      resp.Body,
			Discard: int(discardFrontAlignLen)}
	}
	return result, err
}

// PutObjectFromFile creates a new object from the local file
// the object will be encrypted automaticly on client side when uploaded to oss
func (bucket CryptoBucket) PutObjectFromFile(objectKey, filePath string, options ...oss.Option) error {
	options = bucket.AddEncryptionUaSuffix(options)
	fd, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer fd.Close()

	opts := oss.AddContentType(options, filePath, objectKey)
	cc, err := bucket.ContentCipherBuilder.ContentCipher()
	if err != nil {
		return err
	}

	cryptoReader, err := cc.EncryptContent(fd)
	if err != nil {
		return err
	}

	var request *oss.PutObjectRequest
	srcLen, err := oss.GetReaderLen(fd)
	if err != nil {
		request = &oss.PutObjectRequest{
			ObjectKey: objectKey,
			Reader:    cryptoReader,
		}
	} else {
		encryptedLen := cc.GetEncryptedLen(srcLen)
		request = &oss.PutObjectRequest{
			ObjectKey: objectKey,
			Reader:    oss.LimitReadCloser(cryptoReader, encryptedLen),
		}
	}

	opts = addCryptoHeaders(opts, cc.GetCipherData())
	resp, err := bucket.DoPutObject(request, opts)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

// AppendObject please refer to Bucket.AppendObject
func (bucket CryptoBucket) AppendObject(objectKey string, reader io.Reader, appendPosition int64, options ...oss.Option) (int64, error) {
	return 0, fmt.Errorf("CryptoBucket doesn't support AppendObject")
}

// DoAppendObject please refer to Bucket.DoAppendObject
func (bucket CryptoBucket) DoAppendObject(request *oss.AppendObjectRequest, options []oss.Option) (*oss.AppendObjectResult, error) {
	return nil, fmt.Errorf("CryptoBucket doesn't support DoAppendObject")
}

// PutObjectWithURL please refer to Bucket.PutObjectWithURL
func (bucket CryptoBucket) PutObjectWithURL(signedURL string, reader io.Reader, options ...oss.Option) error {
	return fmt.Errorf("CryptoBucket doesn't support PutObjectWithURL")
}

// PutObjectFromFileWithURL please refer to Bucket.PutObjectFromFileWithURL
func (bucket CryptoBucket) PutObjectFromFileWithURL(signedURL, filePath string, options ...oss.Option) error {
	return fmt.Errorf("CryptoBucket doesn't support PutObjectFromFileWithURL")
}

// DoPutObjectWithURL please refer to Bucket.DoPutObjectWithURL
func (bucket CryptoBucket) DoPutObjectWithURL(signedURL string, reader io.Reader, options []oss.Option) (*oss.Response, error) {
	return nil, fmt.Errorf("CryptoBucket doesn't support DoPutObjectWithURL")
}

// GetObjectWithURL please refer to Bucket.GetObjectWithURL
func (bucket CryptoBucket) GetObjectWithURL(signedURL string, options ...oss.Option) (io.ReadCloser, error) {
	return nil, fmt.Errorf("CryptoBucket doesn't support GetObjectWithURL")
}

// GetObjectToFileWithURL please refer to Bucket.GetObjectToFileWithURL
func (bucket CryptoBucket) GetObjectToFileWithURL(signedURL, filePath string, options ...oss.Option) error {
	return fmt.Errorf("CryptoBucket doesn't support GetObjectToFileWithURL")
}

// DoGetObjectWithURL please refer to Bucket.DoGetObjectWithURL
func (bucket CryptoBucket) DoGetObjectWithURL(signedURL string, options []oss.Option) (*oss.GetObjectResult, error) {
	return nil, fmt.Errorf("CryptoBucket doesn't support DoGetObjectWithURL")
}

// ProcessObject please refer to Bucket.ProcessObject
func (bucket CryptoBucket) ProcessObject(objectKey string, process string, options ...oss.Option) (oss.ProcessObjectResult, error) {
	var out oss.ProcessObjectResult
	return out, fmt.Errorf("CryptoBucket doesn't support ProcessObject")
}

func (bucket CryptoBucket) AddEncryptionUaSuffix(options []oss.Option) []oss.Option {
	var outOption []oss.Option
	bSet, _, _ := oss.IsOptionSet(options, oss.HTTPHeaderUserAgent)
	if bSet || bucket.Client.Config.UserSetUa {
		outOption = options
		return outOption
	}
	outOption = append(options, oss.UserAgentHeader(bucket.Client.Config.UserAgent+"/"+EncryptionUaSuffix))
	return outOption
}

// isEncryptedObject judge the object is encrypted or not
func isEncryptedObject(headers http.Header) bool {
	encrptedKey := headers.Get(oss.HTTPHeaderOssMetaPrefix + OssClientSideEncryptionKey)
	return len(encrptedKey) > 0
}

// addCryptoHeaders save Envelope information in oss meta
func addCryptoHeaders(options []oss.Option, cd *CipherData) []oss.Option {
	opts := []oss.Option{}

	// convert content-md5
	md5Option, _ := oss.FindOption(options, oss.HTTPHeaderContentMD5, nil)
	if md5Option != nil {
		opts = append(opts, oss.Meta(OssClientSideEncryptionUnencryptedContentMD5, md5Option.(string)))
		options = oss.DeleteOption(options, oss.HTTPHeaderContentMD5)
	}

	// convert content-length
	lenOption, _ := oss.FindOption(options, oss.HTTPHeaderContentLength, nil)
	if lenOption != nil {
		opts = append(opts, oss.Meta(OssClientSideEncryptionUnencryptedContentLength, lenOption.(string)))
		options = oss.DeleteOption(options, oss.HTTPHeaderContentLength)
	}

	opts = append(opts, options...)

	// matDesc
	if cd.MatDesc != "" {
		opts = append(opts, oss.Meta(OssClientSideEncryptionMatDesc, cd.MatDesc))
	}

	// encrypted key
	strEncryptedKey := base64.StdEncoding.EncodeToString(cd.EncryptedKey)
	opts = append(opts, oss.Meta(OssClientSideEncryptionKey, strEncryptedKey))

	// encrypted iv
	strEncryptedIV := base64.StdEncoding.EncodeToString(cd.EncryptedIV)
	opts = append(opts, oss.Meta(OssClientSideEncryptionStart, strEncryptedIV))

	// wrap alg
	opts = append(opts, oss.Meta(OssClientSideEncryptionWrapAlg, cd.WrapAlgorithm))

	// cek alg
	opts = append(opts, oss.Meta(OssClientSideEncryptionCekAlg, cd.CEKAlgorithm))

	return opts
}

func getEnvelopeFromHeader(header http.Header) (Envelope, error) {
	var envelope Envelope
	envelope.IV = header.Get(oss.HTTPHeaderOssMetaPrefix + OssClientSideEncryptionStart)
	decodedIV, err := base64.StdEncoding.DecodeString(envelope.IV)
	if err != nil {
		return envelope, err
	}
	envelope.IV = string(decodedIV)

	envelope.CipherKey = header.Get(oss.HTTPHeaderOssMetaPrefix + OssClientSideEncryptionKey)
	decodedKey, err := base64.StdEncoding.DecodeString(envelope.CipherKey)
	if err != nil {
		return envelope, err
	}
	envelope.CipherKey = string(decodedKey)

	envelope.MatDesc = header.Get(oss.HTTPHeaderOssMetaPrefix + OssClientSideEncryptionMatDesc)
	envelope.WrapAlg = header.Get(oss.HTTPHeaderOssMetaPrefix + OssClientSideEncryptionWrapAlg)
	envelope.CEKAlg = header.Get(oss.HTTPHeaderOssMetaPrefix + OssClientSideEncryptionCekAlg)
	envelope.UnencryptedMD5 = header.Get(oss.HTTPHeaderOssMetaPrefix + OssClientSideEncryptionUnencryptedContentMD5)
	envelope.UnencryptedContentLen = header.Get(oss.HTTPHeaderOssMetaPrefix + OssClientSideEncryptionUnencryptedContentLength)
	return envelope, err
}

func isValidContentAlg(algName string) bool {
	// now content encyrption only support aec/ctr algorithm
	return algName == AesCtrAlgorithm
}

func adjustRangeStart(start int64, cc ContentCipher) int64 {
	alignLen := int64(cc.GetAlignLen())
	return (start / alignLen) * alignLen
}
