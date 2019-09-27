package osscrypto

// for client sider encryption oss meta
const (
	OssClientSideEncryptionKey                      string = "client-side-encryption-key"
	OssClientSideEncryptionStart                           = "client-side-encryption-start"
	OssClientSideEncryptionCekAlg                          = "client-side-encryption-cek-alg"
	OssClientSideEncryptionWrapAlg                         = "client-side-encryption-wrap-alg"
	OssClientSideEncryptionMatDesc                         = "client-side-encryption-matdesc"
	OssClientSideEncryptionUnencryptedContentLength        = "client-side-encryption-unencrypted-content-length"
	OssClientSideEncryptionUnencryptedContentMD5           = "client-side-encryption-unencrypted-content-md5"
	OssClientSideEncryptionDataSize                        = "client-side-encryption-data-size"
	OssClientSideEncryptionPartSize                        = "client-side-encryption-part-size"
)

// encryption Algorithm
const (
	RsaCryptoWrap    string = "RSA/NONE/PKCS1Padding"
	KmsAliCryptoWrap string = "KMS/ALICLOUD"
	AesCtrAlgorithm  string = "AES/CTR/NoPadding"
)

// user agent tag for client encryption
const (
	EncryptionUaSuffix string = "OssEncryptionClient"
)
