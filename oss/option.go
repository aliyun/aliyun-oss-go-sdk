package oss

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type optionType string

const (
	optionParam optionType = "HTTPParameter" // URL参数
	optionHTTP  optionType = "HTTPHeader"    // HTTP头
	optionArg   optionType = "FuncArgument"  // 函数参数
)

const (
	deleteObjectsQuiet = "delete-objects-quiet"
	routineNum         = "x-routine-num"
	checkpointConfig   = "x-cp-config"
	initCRC64          = "init-crc64"
	progressListener   = "x-progress-listener"
	storageClass       = "storage-class"
)

type (
	optionValue struct {
		Value interface{}
		Type  optionType
	}

	// Option http option
	Option func(map[string]optionValue) error
)

// ACL is an option to set X-Oss-Acl header
func ACL(acl ACLType) Option {
	return setHeader(HTTPHeaderOssACL, string(acl))
}

// ContentType is an option to set Content-Type header
func ContentType(value string) Option {
	return setHeader(HTTPHeaderContentType, value)
}

// ContentLength is an option to set Content-Length header
func ContentLength(length int64) Option {
	return setHeader(HTTPHeaderContentLength, strconv.FormatInt(length, 10))
}

// CacheControl is an option to set Cache-Control header
func CacheControl(value string) Option {
	return setHeader(HTTPHeaderCacheControl, value)
}

// ContentDisposition is an option to set Content-Disposition header
func ContentDisposition(value string) Option {
	return setHeader(HTTPHeaderContentDisposition, value)
}

// ContentEncoding is an option to set Content-Encoding header
func ContentEncoding(value string) Option {
	return setHeader(HTTPHeaderContentEncoding, value)
}

// ContentMD5 is an option to set Content-MD5 header
func ContentMD5(value string) Option {
	return setHeader(HTTPHeaderContentMD5, value)
}

// Expires is an option to set Expires header
func Expires(t time.Time) Option {
	return setHeader(HTTPHeaderExpires, t.Format(http.TimeFormat))
}

// Meta is an option to set Meta header
func Meta(key, value string) Option {
	return setHeader(HTTPHeaderOssMetaPrefix+key, value)
}

// Range is an option to set Range header, [start, end]
func Range(start, end int64) Option {
	return setHeader(HTTPHeaderRange, fmt.Sprintf("bytes=%d-%d", start, end))
}

// NormalizedRange is an option to set Range header, such as 1024-2048 or 1024- or -2048
func NormalizedRange(nr string) Option {
	return setHeader(HTTPHeaderRange, fmt.Sprintf("bytes=%s", strings.TrimSpace(nr)))
}

// AcceptEncoding is an option to set Accept-Encoding header
func AcceptEncoding(value string) Option {
	return setHeader(HTTPHeaderAcceptEncoding, value)
}

// IfModifiedSince is an option to set If-Modified-Since header
func IfModifiedSince(t time.Time) Option {
	return setHeader(HTTPHeaderIfModifiedSince, t.Format(http.TimeFormat))
}

// IfUnmodifiedSince is an option to set If-Unmodified-Since header
func IfUnmodifiedSince(t time.Time) Option {
	return setHeader(HTTPHeaderIfUnmodifiedSince, t.Format(http.TimeFormat))
}

// IfMatch is an option to set If-Match header
func IfMatch(value string) Option {
	return setHeader(HTTPHeaderIfMatch, value)
}

// IfNoneMatch is an option to set IfNoneMatch header
func IfNoneMatch(value string) Option {
	return setHeader(HTTPHeaderIfNoneMatch, value)
}

// CopySource is an option to set X-Oss-Copy-Source header
func CopySource(sourceBucket, sourceObject string) Option {
	return setHeader(HTTPHeaderOssCopySource, "/"+sourceBucket+"/"+sourceObject)
}

// CopySourceRange is an option to set X-Oss-Copy-Source header
func CopySourceRange(startPosition, partSize int64) Option {
	val := "bytes=" + strconv.FormatInt(startPosition, 10) + "-" +
		strconv.FormatInt((startPosition+partSize-1), 10)
	return setHeader(HTTPHeaderOssCopySourceRange, val)
}

// CopySourceIfMatch is an option to set X-Oss-Copy-Source-If-Match header
func CopySourceIfMatch(value string) Option {
	return setHeader(HTTPHeaderOssCopySourceIfMatch, value)
}

// CopySourceIfNoneMatch is an option to set X-Oss-Copy-Source-If-None-Match header
func CopySourceIfNoneMatch(value string) Option {
	return setHeader(HTTPHeaderOssCopySourceIfNoneMatch, value)
}

// CopySourceIfModifiedSince is an option to set X-Oss-CopySource-If-Modified-Since header
func CopySourceIfModifiedSince(t time.Time) Option {
	return setHeader(HTTPHeaderOssCopySourceIfModifiedSince, t.Format(http.TimeFormat))
}

// CopySourceIfUnmodifiedSince is an option to set X-Oss-Copy-Source-If-Unmodified-Since header
func CopySourceIfUnmodifiedSince(t time.Time) Option {
	return setHeader(HTTPHeaderOssCopySourceIfUnmodifiedSince, t.Format(http.TimeFormat))
}

// MetadataDirective is an option to set X-Oss-Metadata-Directive header
func MetadataDirective(directive MetadataDirectiveType) Option {
	return setHeader(HTTPHeaderOssMetadataDirective, string(directive))
}

// ServerSideEncryption is an option to set X-Oss-Server-Side-Encryption header
func ServerSideEncryption(value string) Option {
	return setHeader(HTTPHeaderOssServerSideEncryption, value)
}

// ObjectACL is an option to set X-Oss-Object-Acl header
func ObjectACL(acl ACLType) Option {
	return setHeader(HTTPHeaderOssObjectACL, string(acl))
}

// symlinkTarget is an option to set X-Oss-Symlink-Target
func symlinkTarget(targetObjectKey string) Option {
	return setHeader(HTTPHeaderOssSymlinkTarget, targetObjectKey)
}

// Origin is an option to set Origin header
func Origin(value string) Option {
	return setHeader(HTTPHeaderOrigin, value)
}

// Delimiter is an option to set delimiler parameter
func Delimiter(value string) Option {
	return addParam("delimiter", value)
}

// Marker is an option to set marker parameter
func Marker(value string) Option {
	return addParam("marker", value)
}

// MaxKeys is an option to set maxkeys parameter
func MaxKeys(value int) Option {
	return addParam("max-keys", strconv.Itoa(value))
}

// Prefix is an option to set prefix parameter
func Prefix(value string) Option {
	return addParam("prefix", value)
}

// EncodingType is an option to set encoding-type parameter
func EncodingType(value string) Option {
	return addParam("encoding-type", value)
}

// MaxUploads is an option to set max-uploads parameter
func MaxUploads(value int) Option {
	return addParam("max-uploads", strconv.Itoa(value))
}

// KeyMarker is an option to set key-marker parameter
func KeyMarker(value string) Option {
	return addParam("key-marker", value)
}

// UploadIDMarker is an option to set upload-id-marker parameter
func UploadIDMarker(value string) Option {
	return addParam("upload-id-marker", value)
}

// DeleteObjectsQuiet DeleteObjects详细(verbose)模式或简单(quiet)模式，默认详细模式。
func DeleteObjectsQuiet(isQuiet bool) Option {
	return addArg(deleteObjectsQuiet, isQuiet)
}

// StorageClass bucket的存储方式
func StorageClass(value StorageClassType) Option {
	return addArg(storageClass, value)
}

// 断点续传配置，包括是否启用、cp文件
type cpConfig struct {
	IsEnable bool
	FilePath string
}

// Checkpoint DownloadFile/UploadFile是否开启checkpoint及checkpoint文件路径
func Checkpoint(isEnable bool, filePath string) Option {
	return addArg(checkpointConfig, &cpConfig{isEnable, filePath})
}

// Routines DownloadFile/UploadFile并发数
func Routines(n int) Option {
	return addArg(routineNum, n)
}

// InitCRC AppendObject CRC的校验的初始值
func InitCRC(initCRC uint64) Option {
	return addArg(initCRC64, initCRC)
}

// Progress set progress listener
func Progress(listener ProgressListener) Option {
	return addArg(progressListener, listener)
}

// ResponseContentType is an option to set response-content-type param
func ResponseContentType(value string) Option {
	return addParam("response-content-type", value)
}

// ResponseContentLanguage is an option to set response-content-language param
func ResponseContentLanguage(value string) Option {
	return addParam("response-content-language", value)
}

// ResponseExpires is an option to set response-expires param
func ResponseExpires(value string) Option {
	return addParam("response-expires", value)
}

// ResponseCacheControl is an option to set response-cache-control param
func ResponseCacheControl(value string) Option {
	return addParam("response-cache-control", value)
}

// ResponseContentDisposition is an option to set response-content-disposition param
func ResponseContentDisposition(value string) Option {
	return addParam("response-content-disposition", value)
}

// ResponseContentEncoding is an option to set response-content-encoding param
func ResponseContentEncoding(value string) Option {
	return addParam("response-content-encoding", value)
}

// Process is an option to set X-Oss-Process param
func Process(value string) Option {
	return addParam("X-Oss-Process", value)
}

// UDF相关
// UDFID is an option to set udf Id in ListUDFs
func UDFID(value string) Option {
	return addParam("udfId", value)
}

// UDFImageDesc is an option to set udf Id in ListUDFs
func UDFImageDesc(value string) Option {
	return addParam("udfImageDesc", value)
}

// UDFSince is an option to set since param when get udf app log
func UDFSince(t time.Time) Option {
	return addParam("since", strconv.FormatInt(t.Unix(), 10))
}

// UDFSince is an option to set tail param when get udf app log
func UDFTail(t int64) Option {
	return addParam("tail", strconv.FormatInt(t, 10))
}

func setHeader(key string, value interface{}) Option {
	return func(params map[string]optionValue) error {
		if value == nil {
			return nil
		}
		params[key] = optionValue{value, optionHTTP}
		return nil
	}
}

func addParam(key string, value interface{}) Option {
	return func(params map[string]optionValue) error {
		if value == nil {
			return nil
		}
		params[key] = optionValue{value, optionParam}
		return nil
	}
}

func addArg(key string, value interface{}) Option {
	return func(params map[string]optionValue) error {
		if value == nil {
			return nil
		}
		params[key] = optionValue{value, optionArg}
		return nil
	}
}

func handleOptions(headers map[string]string, options []Option) error {
	params := map[string]optionValue{}
	for _, option := range options {
		if option != nil {
			if err := option(params); err != nil {
				return err
			}
		}
	}

	for k, v := range params {
		if v.Type == optionHTTP {
			headers[k] = v.Value.(string)
		}
	}
	return nil
}

func getRawParams(options []Option) (map[string]interface{}, error) {
	// option
	params := map[string]optionValue{}
	for _, option := range options {
		if option != nil {
			if err := option(params); err != nil {
				return nil, err
			}
		}
	}

	paramsm := map[string]interface{}{}
	// serialize
	for k, v := range params {
		if v.Type == optionParam {
			vs := params[k]
			paramsm[k] = vs.Value.(string)
		}
	}

	return paramsm, nil
}

func findOption(options []Option, param string, defaultVal interface{}) (interface{}, error) {
	params := map[string]optionValue{}
	for _, option := range options {
		if option != nil {
			if err := option(params); err != nil {
				return nil, err
			}
		}
	}

	if val, ok := params[param]; ok {
		return val.Value, nil
	}
	return defaultVal, nil
}

func isOptionSet(options []Option, option string) (bool, interface{}, error) {
	params := map[string]optionValue{}
	for _, option := range options {
		if option != nil {
			if err := option(params); err != nil {
				return false, nil, err
			}
		}
	}

	if val, ok := params[option]; ok {
		return true, val.Value, nil
	}
	return false, nil, nil
}
