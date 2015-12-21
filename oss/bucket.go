package oss

import (
	"bytes"
	"encoding/xml"
	"io"
	"net/http"
	"os"
	"strconv"
)

// Bucket implements the operations of object.
type Bucket struct {
	Client     Client
	BucketName string
}

//
// PutObject 新建Object，如果Object已存在，覆盖原有Object。
//
// objectKey  上传对象的名称，使用UTF-8编码、长度必须在1-1023字节之间、不能以“/”或者“\”字符开头。
// reader     io.Reader读取object的数据。
// options    上传对象时可以指定对象的属性，可用选项有CacheControl、ContentDisposition、ContentEncoding、
// Expires、ServerSideEncryption、ObjectACL、Meta，具体含义请参看
// https://help.aliyun.com/document_detail/oss/api-reference/object/PutObject.html
//
// error  操作无错误为nil，非nil为错误信息。
//
func (bucket Bucket) PutObject(objectKey string, reader io.Reader, options ...Option) error {
	opts := addContentType(options, objectKey)
	resp, err := bucket.do("PUT", objectKey, "", "", opts, reader)
	if err != nil {
		return err
	}

	defer resp.body.Close()
	return checkRespCode(resp.statusCode, []int{http.StatusOK})
}

//
// PutObjectFromFile 新建Object，内容从本地文件中读取。
//
// objectKey 上传对象的名称。
// filePath  本地文件，上传对象的值为该文件内容。
// options   上传对象时可以指定对象的属性。详见PutObject的options。
//
// error  操作无错误为nil，非nil为错误信息。
//
func (bucket Bucket) PutObjectFromFile(objectKey, filePath string, options ...Option) error {
	fd, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer fd.Close()

	opts := addContentType(options, filePath, objectKey)

	resp, err := bucket.do("PUT", objectKey, "", "", opts, fd)
	if err != nil {
		return err
	}
	defer resp.body.Close()
	return checkRespCode(resp.statusCode, []int{http.StatusOK})
}

//
// GetObject 下载文件。
//
// objectKey 下载的文件名称。
// options   对象的属性限制项，可选值有Range、IfModifiedSince、IfUnmodifiedSince、IfMatch、
// IfNoneMatch、AcceptEncoding，详细请参考
// https://help.aliyun.com/document_detail/oss/api-reference/object/GetObject.html
//
// io.ReadCloser  reader，读取数据后需要close。error为nil时有效。
// error  操作无错误为nil，非nil为错误信息。
//
func (bucket Bucket) GetObject(objectKey string, options ...Option) (io.ReadCloser, error) {
	resp, err := bucket.do("GET", objectKey, "", "", options, nil)
	if err != nil {
		return nil, err
	}
	return resp.body, nil
}

//
// GetObjectToFile 下载文件。
//
// objectKey  下载的文件名称。
// filePath   下载对象的内容写到该本地文件。
// options    对象的属性限制项。详见GetObject的options。
//
// error  操作无错误时返回error为nil，非nil为错误说明。
//
func (bucket Bucket) GetObjectToFile(objectKey, filePath string, options ...Option) error {
	resp, err := bucket.do("GET", objectKey, "", "", options, nil)
	if err != nil {
		return err
	}
	defer resp.body.Close()

	fd, err := os.OpenFile(filePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0660)
	if err != nil {
		return err
	}
	defer fd.Close()

	_, err = io.Copy(fd, resp.body)
	if err != nil {
		return err
	}

	return nil
}

//
// CopyObject 同一个bucket内拷贝Object。
//
// srcObjectKey  Copy的源对象。
// destObjectKey Copy的目标对象。
// options  Copy对象时，您可以指定源对象的限制条件，满足限制条件时copy，不满足时返回错误，您可以选择如下选项CopySourceIfMatch、
// CopySourceIfNoneMatch、CopySourceIfModifiedSince、CopySourceIfUnmodifiedSince、MetadataDirective。
// Copy对象时，您可以指定目标对象的属性，如CacheControl、ContentDisposition、ContentEncoding、Expires、
// ServerSideEncryption、ObjectACL、Meta，选项的含义请参看
// https://help.aliyun.com/document_detail/oss/api-reference/object/CopyObject.html
//
// error 操作无错误为nil，非nil为错误信息。
//
func (bucket Bucket) CopyObject(srcObjectKey, destObjectKey string, options ...Option) (CopyObjectResult, error) {
	var out CopyObjectResult
	options = append(options, CopySource(bucket.BucketName, srcObjectKey))
	resp, err := bucket.do("PUT", destObjectKey, "", "", options, nil)
	if err != nil {
		return out, err
	}
	defer resp.body.Close()

	err = xmlUnmarshal(resp.body, &out)
	return out, err
}

//
// CopyObjectToBucket bucket间拷贝object。
//
// srcObjectKey   Copy的源对象。
// destBucket     Copy的目标Bucket。
// destObjectKey  Copy的目标Object。
// options        Copy选项，详见CopyObject的options。
//
// error  操作无错误为nil，非nil为错误信息。
//
func (bucket Bucket) CopyObjectToBucket(srcObjectKey, destBucketName, destObjectKey string, options ...Option) (CopyObjectResult, error) {
	var out CopyObjectResult
	options = append(options, CopySource(bucket.BucketName, srcObjectKey))
	headers := make(map[string]string)
	err := handleOptions(headers, options)
	if err != nil {
		return out, err
	}
	resp, err := bucket.Client.Conn.Do("PUT", destBucketName, destObjectKey, "", "", headers, nil)
	if err != nil {
		return out, err
	}
	defer resp.body.Close()

	err = xmlUnmarshal(resp.body, &out)
	return out, err
}

//
// AppendObject 追加方式上传。
//
// AppendObject参数必须包含position，其值指定从何处进行追加。首次追加操作的position必须为0，
// 后续追加操作的position是Object的当前长度。例如，第一次Append Object请求指定position值为0，
// content-length是65536；那么，第二次Append Object需要指定position为65536。
// 每次操作成功后，响应头部x-oss-next-append-position也会标明下一次追加的position。
//
// objectKey  需要追加的Object。
// reader     io.Reader，读取追的内容。
// appendPosition  object追加的起始位置。
// destObjectProperties  第一次追加时指定新对象的属性，如CacheControl、ContentDisposition、ContentEncoding、
// Expires、ServerSideEncryption、ObjectACL。
// int64 下次追加的开始位置，error为nil空时有效。
//
// error 操作无错误为nil，非nil为错误信息。
//
func (bucket Bucket) AppendObject(objectKey string, reader io.Reader, appendPosition int64, options ...Option) (int64, error) {
	var nextAppendPosition int64
	params := "append&position=" + strconv.Itoa(int(appendPosition))
	opts := addContentType(options, objectKey)
	resp, err := bucket.do("POST", objectKey, params, params, opts, reader)
	if err != nil {
		return nextAppendPosition, err
	}
	defer resp.body.Close()

	napint, err := strconv.Atoi(resp.headers.Get(HTTPHeaderOssNextAppendPosition))
	if err != nil {
		return nextAppendPosition, err
	}
	nextAppendPosition = int64(napint)
	return nextAppendPosition, nil
}

//
// DeleteObject 删除Object。
//
// objectKey 待删除Object。
//
// error 操作无错误为nil，非nil为错误信息。
//
func (bucket Bucket) DeleteObject(objectKey string) error {
	resp, err := bucket.do("DELETE", objectKey, "", "", nil, nil)
	if err != nil {
		return err
	}
	defer resp.body.Close()
	return checkRespCode(resp.statusCode, []int{http.StatusNoContent})
}

//
// DeleteObjects 批量删除object。
//
// objectKeys 待删除object类表。
// options 删除选项，DeleteObjectsQuiet，是否是安静模式，默认不使用。
//
// DeleteObjectsResult 非安静模式的的返回值。
// error 操作无错误为nil，非nil为错误信息。
//
func (bucket Bucket) DeleteObjects(objectKeys []string, options ...Option) (DeleteObjectsResult, error) {
	out := DeleteObjectsResult{}
	dxml := deleteXML{}
	for _, key := range objectKeys {
		dxml.Objects = append(dxml.Objects, DeleteObject{Key: key})
	}
	isQuietStr, _ := findOption(options, deleteObjectsQuiet, "FALSE")
	isQuiet, _ := strconv.ParseBool(isQuietStr)
	dxml.Quiet = isQuiet
	encode := "&encoding-type=url"

	bs, err := xml.Marshal(dxml)
	if err != nil {
		return out, err
	}
	buffer := new(bytes.Buffer)
	buffer.Write(bs)

	contentType := http.DetectContentType(buffer.Bytes())
	options = append(options, ContentType(contentType))
	resp, err := bucket.do("POST", "", "delete"+encode, "delete", options, buffer)
	if err != nil {
		return out, err
	}
	defer resp.body.Close()

	if !dxml.Quiet {
		if err = xmlUnmarshal(resp.body, &out); err == nil {
			err = decodeDeleteObjectsResult(&out)
		}
	}
	return out, err
}

//
// IsObjectExist object是否存在。
//
// bool  object是否存在，true存在，false不存在。error为nil时有效。
//
// error 操作无错误为nil，非nil为错误信息。
//
func (bucket Bucket) IsObjectExist(objectKey string) (bool, error) {
	listRes, err := bucket.ListObjects(Prefix(objectKey), MaxKeys(1))
	if err != nil {
		return false, err
	}

	if len(listRes.Objects) == 1 && listRes.Objects[0].Key == objectKey {
		return true, nil
	}
	return false, nil
}

//
// ListObjects 获得Bucket下筛选后所有的object的列表。
//
// options  ListObject的筛选行为。Prefix指定的前缀、MaxKeys最大数目、Marker第一个开始、Delimiter对Object名字进行分组的字符。
//
// 您有如下8个object，my-object-1, my-object-11, my-object-2, my-object-21,
// my-object-22, my-object-3, my-object-31, my-object-32。如果您指定了Prefix为my-object-2,
// 则返回my-object-2, my-object-21, my-object-22三个object。如果您指定了Marker为my-object-22，
// 则返回my-object-3, my-object-31, my-object-32三个object。如果您指定MaxKeys则每次最多返回MaxKeys个，
// 最后一次可能不足。这三个参数可以组合使用，实现分页等功能。如果把prefix设为某个文件夹名，就可以罗列以此prefix开头的文件，
// 即该文件夹下递归的所有的文件和子文件夹。如果再把delimiter设置为"/"时，返回值就只罗列该文件夹下的文件，该文件夹下的子文件名
// 返回在CommonPrefixes部分，子文件夹下递归的文件和文件夹不被显示。例如一个bucket存在三个object，fun/test.jpg、
// fun/movie/001.avi、fun/movie/007.avi。若设定prefix为"fun/"，则返回三个object；如果增加设定
// delimiter为"/"，则返回文件"fun/test.jpg"和前缀"fun/movie/"，即实现了文件夹的逻辑。
//
// 常用场景，请参数示例sample/list_object.go。
//
// ListObjectsResponse  操作成功后的返回值，成员Objects为bucket中对象列表。error为nil时该返回值有效。
//
func (bucket Bucket) ListObjects(options ...Option) (ListObjectsResult, error) {
	var out ListObjectsResult

	options = append(options, EncodingType("url"))
	params, err := handleParams(options)
	if err != nil {
		return out, err
	}

	resp, err := bucket.do("GET", "", params, "", nil, nil)
	if err != nil {
		return out, err
	}
	defer resp.body.Close()

	err = xmlUnmarshal(resp.body, &out)
	if err != nil {
		return out, err
	}

	err = decodeListObjectsResult(&out)
	return out, err
}

//
// SetObjectMeta 设置Object的Meta。
//
// objectKey object
// options 指定对象的属性，有以下可选项CacheControl、ContentDisposition、ContentEncoding、Expires、
// ServerSideEncryption、Meta。
//
// error 操作无错误时error为nil，非nil为错误信息。
//
func (bucket Bucket) SetObjectMeta(objectKey string, options ...Option) error {
	options = append(options, MetadataDirective(MetaReplace))
	_, err := bucket.CopyObject(objectKey, objectKey, options...)
	return err
}

//
// GetObjectDetailedMeta 查询Object的头信息。
//
// objectKey object名称。
// objectPropertyConstraints 对象的属性限制项，满足时正常返回，不满足时返回错误。现在项有IfModifiedSince、IfUnmodifiedSince、
// IfMatch、IfNoneMatch。具体含义请参看 https://help.aliyun.com/document_detail/oss/api-reference/object/HeadObject.html
//
// http.Header  对象的meta，error为nil时有效。
// error  操作无错误为nil，非nil为错误信息。
//
func (bucket Bucket) GetObjectDetailedMeta(objectKey string, options ...Option) (http.Header, error) {
	resp, err := bucket.do("HEAD", objectKey, "", "", options, nil)
	if err != nil {
		return nil, err
	}
	defer resp.body.Close()

	return resp.headers, nil
}

//
// GetObjectMeta 查询Object的头信息。
//
// GetObjectMeta相比GetObjectDetailedMeta更轻量，仅返回指定Object的少量基本meta信息，
// 包括该Object的ETag、Size（对象大小）、LastModified，其中Size由响应头Content-Length的数值表示。
//
// objectKey object名称。
//
// http.Header 对象的meta，error为nil时有效。
// error 操作无错误为nil，非nil为错误信息。
//
func (bucket Bucket) GetObjectMeta(objectKey string) (http.Header, error) {
	resp, err := bucket.do("GET", objectKey, "?objectMeta", "", nil, nil)
	if err != nil {
		return nil, err
	}
	defer resp.body.Close()

	return resp.headers, nil
}

//
// SetObjectACL 修改Object的ACL权限。
//
// 只有Bucket Owner才有权限调用PutObjectACL来修改Object的ACL。Object ACL优先级高于Bucket ACL。
// 例如Bucket ACL是private的，而Object ACL是public-read-write的，则访问这个Object时，
// 先判断Object的ACL，所以所有用户都拥有这个Object的访问权限，即使这个Bucket是private bucket。
// 如果某个Object从来没设置过ACL，则访问权限遵循Bucket ACL。
//
// Object的读操作包括GetObject，HeadObject，CopyObject和UploadPartCopy中的对source object的读；
// Object的写操作包括：PutObject，PostObject，AppendObject，DeleteObject，
// DeleteMultipleObjects，CompleteMultipartUpload以及CopyObject对新的Object的写。
//
// objectKey 设置权限的object。
// objectAcl 对象权限。可选值PrivateACL(私有读写)、PublicReadACL(公共读私有写)、PublicReadWriteACL(公共读写)。
//
// error 操作无错误为nil，非nil为错误信息。
//
func (bucket Bucket) SetObjectACL(objectKey string, objectACL ACLType) error {
	options := []Option{ObjectACL(objectACL)}
	resp, err := bucket.do("PUT", objectKey, "", "", options, nil)
	if err != nil {
		return err
	}
	defer resp.body.Close()
	return checkRespCode(resp.statusCode, []int{http.StatusOK})
}

//
// GetObjectACL 获取对象的ACL权限。
//
// objectKey 获取权限的object。
//
// GetObjectAclResponse 获取权限操作返回值，error为nil时有效。GetObjectAclResponse.Acl为对象的权限。
// error 操作无错误为nil，非nil为错误信息。
//
func (bucket Bucket) GetObjectACL(objectKey string) (GetObjectACLResult, error) {
	var out GetObjectACLResult
	resp, err := bucket.do("GET", objectKey, "acl", "acl", nil, nil)
	if err != nil {
		return out, err
	}
	defer resp.body.Close()

	err = xmlUnmarshal(resp.body, &out)
	return out, err
}

// Private
func (bucket Bucket) do(method, objectName, urlParams, subResource string,
	options []Option, data io.Reader) (*Response, error) {
	headers := make(map[string]string)
	err := handleOptions(headers, options)
	if err != nil {
		return nil, err
	}
	return bucket.Client.Conn.Do(method, bucket.BucketName, objectName,
		urlParams, subResource, headers, data)
}

func (bucket Bucket) getConfig() *Config {
	return bucket.Client.Config
}

// Private
func addContentType(options []Option, keys ...string) []Option {
	typ := TypeByExtension("")
	for _, key := range keys {
		typ = TypeByExtension(key)
		if typ != "" {
			break
		}
	}

	if typ == "" {
		typ = "application/octet-stream"
	}

	opts := []Option{ContentType(typ)}
	opts = append(opts, options...)

	return opts
}
