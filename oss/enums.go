package oss

// ACLType bucket/object ACL
type ACLType string

const (
	// ACLPrivate definition : private read and write
	ACLPrivate ACLType = "private"

	// ACLPublicRead definition : public read and private write
	ACLPublicRead ACLType = "public-read"

	// ACLPublicReadWrite definition : public read and public write
	ACLPublicReadWrite ACLType = "public-read-write"

	// ACLDefault Object. It's only applicable for object.
	ACLDefault ACLType = "default"
)

// bucket versioning status
type VersioningStatus string

const (
	// Versioning Status definition: Enabled
	VersionEnabled VersioningStatus = "Enabled"

	// Versioning Status definition: Suspended
	VersionSuspended VersioningStatus = "Suspended"
)

// MetadataDirectiveType specifying whether use the metadata of source object when copying object.
type MetadataDirectiveType string

const (
	// MetaCopy the target object's metadata is copied from the source one
	MetaCopy MetadataDirectiveType = "COPY"

	// MetaReplace the target object's metadata is created as part of the copy request (not same as the source one)
	MetaReplace MetadataDirectiveType = "REPLACE"
)

// TaggingDirectiveType specifying whether use the tagging of source object when copying object.
type TaggingDirectiveType string

const (
	// TaggingCopy the target object's tagging is copied from the source one
	TaggingCopy TaggingDirectiveType = "COPY"

	// TaggingReplace the target object's tagging is created as part of the copy request (not same as the source one)
	TaggingReplace TaggingDirectiveType = "REPLACE"
)

// AlgorithmType specifying the server side encryption algorithm name
type AlgorithmType string

const (
	KMSAlgorithm AlgorithmType = "KMS"
	AESAlgorithm AlgorithmType = "AES256"
	SM4Algorithm AlgorithmType = "SM4"
)

// StorageClassType bucket storage type
type StorageClassType string

const (
	// StorageStandard standard
	StorageStandard StorageClassType = "Standard"

	// StorageIA infrequent access
	StorageIA StorageClassType = "IA"

	// StorageArchive archive
	StorageArchive StorageClassType = "Archive"

	// StorageColdArchive cold archive
	StorageColdArchive StorageClassType = "ColdArchive"

	// StorageDeepColdArchive deep cold archive
	StorageDeepColdArchive StorageClassType = "DeepColdArchive"
)

// RedundancyType bucket data Redundancy type
type DataRedundancyType string

const (
	// RedundancyLRS Local redundancy, default value
	RedundancyLRS DataRedundancyType = "LRS"

	// RedundancyZRS Same city redundancy
	RedundancyZRS DataRedundancyType = "ZRS"
)

// PayerType the type of request payer
type PayerType string

const (
	// Requester the requester who send the request
	Requester PayerType = "Requester"

	// BucketOwner the requester who send the request
	BucketOwner PayerType = "BucketOwner"
)

// RestoreMode the restore mode for coldArchive object
type RestoreMode string

const (
	//RestoreExpedited object will be restored in 1 hour
	RestoreExpedited RestoreMode = "Expedited"

	//RestoreStandard object will be restored in 2-5 hours
	RestoreStandard RestoreMode = "Standard"

	//RestoreBulk object will be restored in 5-10 hours
	RestoreBulk RestoreMode = "Bulk"
)

// OSS headers
const (
	HeaderOssPrefix                      string = "X-Oss-"
	HeaderOssMetaPrefix                         = "X-Oss-Meta-"
	HeaderOssACL                                = "X-Oss-Acl"
	HeaderOssObjectACL                          = "X-Oss-Object-Acl"
	HeaderOssSecurityToken                      = "X-Oss-Security-Token"
	HeaderOssServerSideEncryption               = "X-Oss-Server-Side-Encryption"
	HeaderOssServerSideEncryptionKeyID          = "X-Oss-Server-Side-Encryption-Key-Id"
	HeaderOssServerSideDataEncryption           = "X-Oss-Server-Side-Data-Encryption"
	HeaderOssSSECAlgorithm                      = "X-Oss-Server-Side-Encryption-Customer-Algorithm"
	HeaderOssSSECKey                            = "X-Oss-Server-Side-Encryption-Customer-Key"
	HeaderOssSSECKeyMd5                         = "X-Oss-Server-Side-Encryption-Customer-Key-MD5"
	HeaderOssCopySource                         = "X-Oss-Copy-Source"
	HeaderOssCopySourceRange                    = "X-Oss-Copy-Source-Range"
	HeaderOssCopySourceIfMatch                  = "X-Oss-Copy-Source-If-Match"
	HeaderOssCopySourceIfNoneMatch              = "X-Oss-Copy-Source-If-None-Match"
	HeaderOssCopySourceIfModifiedSince          = "X-Oss-Copy-Source-If-Modified-Since"
	HeaderOssCopySourceIfUnmodifiedSince        = "X-Oss-Copy-Source-If-Unmodified-Since"
	HeaderOssMetadataDirective                  = "X-Oss-Metadata-Directive"
	HeaderOssNextAppendPosition                 = "X-Oss-Next-Append-Position"
	HeaderOssRequestID                          = "X-Oss-Request-Id"
	HeaderOssCRC64                              = "X-Oss-Hash-Crc64ecma"
	HeaderOssSymlinkTarget                      = "X-Oss-Symlink-Target"
	HeaderOssStorageClass                       = "X-Oss-Storage-Class"
	HeaderOssCallback                           = "X-Oss-Callback"
	HeaderOssCallbackVar                        = "X-Oss-Callback-Var"
	HeaderOssRequester                          = "X-Oss-Request-Payer"
	HeaderOssTagging                            = "X-Oss-Tagging"
	HeaderOssTaggingDirective                   = "X-Oss-Tagging-Directive"
	HeaderOssTrafficLimit                       = "X-Oss-Traffic-Limit"
	HeaderOssForbidOverWrite                    = "X-Oss-Forbid-Overwrite"
	HeaderOssRangeBehavior                      = "X-Oss-Range-Behavior"
	HeaderOssAllowSameActionOverLap             = "X-Oss-Allow-Same-Action-Overlap"
	HeaderOssDate                               = "X-Oss-Date"
	HeaderOssContentSha256                      = "X-Oss-Content-Sha256"
	HeaderOssEC                                 = "X-Oss-Ec"
	HeaderOssERR                                = "X-Oss-Err"
)

// AuthVersion the version of auth
type AuthVersionType string

const (
	// AuthV1 v1
	AuthV1 AuthVersionType = "v1"
	// AuthV2 v2
	AuthV2 AuthVersionType = "v2"
	// AuthV4 v4
	AuthV4 AuthVersionType = "v4"
)

type UrlRequestStyle int

const (
	VirtualHostedStyle UrlRequestStyle = iota
	PathStyle
	CNameStyle
)

func (f UrlRequestStyle) String() string {
	switch f {
	case VirtualHostedStyle:
		return "virtual-hosted–style"
	case PathStyle:
		return "path-style"
	case CNameStyle:
		return "cname-style"
	default:
		return ""
	}
}
