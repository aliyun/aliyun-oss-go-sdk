package sample

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

// PutObjectSample illustrates two methods for uploading a file: simple upload and multipart upload.
func PutObjectSample() {
	// Create bucket
	bucket, err := GetTestBucket(bucketName)
	if err != nil {
		HandleError(err)
	}

	var val = "花间一壶酒，独酌无相亲。 举杯邀明月，对影成三人。"

	// Case 1: Upload an object from a string
	err = bucket.PutObject(objectKey, strings.NewReader(val))
	if err != nil {
		HandleError(err)
	}

	// Case 2: Upload an object whose value is a byte[]
	err = bucket.PutObject(objectKey, bytes.NewReader([]byte(val)))
	if err != nil {
		HandleError(err)
	}

	// Case 3: Upload the local file with file handle, user should open the file at first.
	fd, err := os.Open(localFile)
	if err != nil {
		HandleError(err)
	}
	defer fd.Close()

	err = bucket.PutObject(objectKey, fd)
	if err != nil {
		HandleError(err)
	}

	// Case 4: Upload an object with local file name, user need not open the file.
	err = bucket.PutObjectFromFile(objectKey, localFile)
	if err != nil {
		HandleError(err)
	}

	// Case 5: Upload an object with specified properties, PutObject/PutObjectFromFile/UploadFile also support this feature.
	options := []oss.Option{
		oss.Expires(futureDate),
		oss.ObjectACL(oss.ACLPublicRead),
		oss.Meta("myprop", "mypropval"),
	}
	err = bucket.PutObject(objectKey, strings.NewReader(val), options...)
	if err != nil {
		HandleError(err)
	}

	props, err := bucket.GetObjectDetailedMeta(objectKey)
	if err != nil {
		HandleError(err)
	}
	fmt.Println("Object Meta:", props)

	// Case 6: Upload an object with sever side encrpytion kms and kms id specified
	err = bucket.PutObject(objectKey, strings.NewReader(val), oss.ServerSideEncryption("KMS"), oss.ServerSideEncryptionKeyID(kmsID))
	if err != nil {
		HandleError(err)
	}

	// Case 7: Upload an object with callback
	callbackMap := map[string]string{}
	callbackMap["callbackUrl"] = "http://oss-demo.aliyuncs.com:23450"
	callbackMap["callbackHost"] = "oss-cn-hangzhou.aliyuncs.com"
	callbackMap["callbackBody"] = "filename=${object}&size=${size}&mimeType=${mimeType}"
	callbackMap["callbackBodyType"] = "application/x-www-form-urlencoded"

	callbackBuffer := bytes.NewBuffer([]byte{})
	callbackEncoder := json.NewEncoder(callbackBuffer)
	//do not encode '&' to "\u0026"
	callbackEncoder.SetEscapeHTML(false)
	err = callbackEncoder.Encode(callbackMap)
	if err != nil {
		HandleError(err)
	}

	callbackVal := base64.StdEncoding.EncodeToString(callbackBuffer.Bytes())
	err = bucket.PutObject(objectKey, strings.NewReader(val), oss.Callback(callbackVal))
	if err != nil {
		HandleError(err)
	}

	// Case 7-2: Upload an object with callback and get callback body
	callbackMap = map[string]string{}
	callbackMap["callbackUrl"] = "http://oss-demo.aliyuncs.com:23450"
	callbackMap["callbackHost"] = "oss-cn-hangzhou.aliyuncs.com"
	callbackMap["callbackBody"] = "filename=${object}&size=${size}&mimeType=${mimeType}"
	callbackMap["callbackBodyType"] = "application/x-www-form-urlencoded"

	callbackBuffer = bytes.NewBuffer([]byte{})
	callbackEncoder = json.NewEncoder(callbackBuffer)
	//do not encode '&' to "\u0026"
	callbackEncoder.SetEscapeHTML(false)
	err = callbackEncoder.Encode(callbackMap)
	if err != nil {
		HandleError(err)
	}

	callbackVal = base64.StdEncoding.EncodeToString(callbackBuffer.Bytes())
	var body []byte
	err = bucket.PutObject(objectKey, strings.NewReader(val), oss.Callback(callbackVal), oss.CallbackResult(&body))

	if err != nil {
		e, ok := err.(oss.UnexpectedStatusCodeError)
		if !(ok && e.Got() == 203) {
			HandleError(err)
		}
	}

	fmt.Printf("callback body:%s\n", body)

	// Case 8: Big file's multipart upload. It supports concurrent upload with resumable upload.
	// multipart upload with 100K as part size. By default 1 coroutine is used and no checkpoint is used.
	err = bucket.UploadFile(objectKey, localFile, 100*1024)
	if err != nil {
		HandleError(err)
	}

	// Part size is 100K and 3 coroutines are used
	err = bucket.UploadFile(objectKey, localFile, 100*1024, oss.Routines(3))
	if err != nil {
		HandleError(err)
	}

	// Part size is 100K and 3 coroutines with checkpoint
	err = bucket.UploadFile(objectKey, localFile, 100*1024, oss.Routines(3), oss.Checkpoint(true, ""))
	if err != nil {
		HandleError(err)
	}

	// Specify the local file path for checkpoint files.
	// the 2nd parameter of Checkpoint can specify the file path, when the file path is empty, it will upload the directory.
	err = bucket.UploadFile(objectKey, localFile, 100*1024, oss.Checkpoint(true, localFile+".cp"))
	if err != nil {
		HandleError(err)
	}

	// Case 8-1:Big file's multipart upload. Set callback and get callback body

	//The local file is partitioned, and the number of partitions is specified as 3.
	chunks, err := oss.SplitFileByPartNum(localFile, 3)
	fd, err = os.Open(localFile)
	defer fd.Close()
	//Specify the expiration time.
	expires := time.Date(2049, time.January, 10, 23, 0, 0, 0, time.UTC)
	// If you need to set the request header when initializing fragmentation, please refer to the following example code.
	options = []oss.Option{
		oss.MetadataDirective(oss.MetaReplace),
		oss.Expires(expires),
		// Specifies the web page caching behavior when the object is downloaded.
		// oss.CacheControl("no-cache"),
		// Specifies the name of the object when it is downloaded.
		// oss.ContentDisposition("attachment;filename=FileName.txt"),
		// Specifies the content encoding format of the object.
		// oss.ContentEncoding("gzip"),
		// Specifies to encode the returned key. Currently, URL encoding is supported.
		// oss.EncodingType("url"),
		// Specifies the storage type of the object.
		// oss.ObjectStorageClass(oss.StorageStandard),
	}
	// Step 1: initialize a fragment upload event and specify the storage type as standard storage
	imur, err := bucket.InitiateMultipartUpload(objectKey, options...)
	// Step 2: upload fragments.
	var parts []oss.UploadPart
	for _, chunk := range chunks {
		fd.Seek(chunk.Offset, os.SEEK_SET)
		// Call the uploadpart method to upload each fragment.
		part, err := bucket.UploadPart(imur, fd, chunk.Size, chunk.Number)
		if err != nil {
			HandleError(err)
		}
		parts = append(parts, part)
	}

	// Construct callback map
	callbackMap = map[string]string{}
	callbackMap["callbackUrl"] = "www.aliyuncs.com"
	callbackMap["callbackBody"] = "filename=demo.go&name=golang"
	callbackMap["callbackBodyType"] = "application/x-www-form-urlencoded"

	callbackBuffer = bytes.NewBuffer([]byte{})
	callbackEncoder = json.NewEncoder(callbackBuffer)
	//do not encode '&' to "\u0026"
	callbackEncoder.SetEscapeHTML(false)
	err = callbackEncoder.Encode(callbackMap)
	if err != nil {
		HandleError(err)
	}

	callbackVal = base64.StdEncoding.EncodeToString(callbackBuffer.Bytes())

	var pbody []byte
	// Step 3: complete fragment uploading
	_, err = bucket.CompleteMultipartUpload(imur, parts, oss.Callback(callbackVal), oss.CallbackResult(&pbody))
	if err != nil {
		HandleError(err)
	}
	fmt.Printf("callback body:%s/n", pbody)

	// Case 9: Set the storage classes.OSS provides three storage classes: Standard, Infrequent Access, and Archive.
	// Supported APIs: PutObject, CopyObject, UploadFile, AppendObject...
	err = bucket.PutObject(objectKey, strings.NewReader(val), oss.ObjectStorageClass("IA"))
	if err != nil {
		HandleError(err)
	}

	// Upload a local file, and set the object's storage-class to 'Archive'.
	err = bucket.UploadFile(objectKey, localFile, 100*1024, oss.ObjectStorageClass("Archive"))
	if err != nil {
		HandleError(err)
	}

	// Delete object and bucket
	err = DeleteTestBucketAndObject(bucketName)
	if err != nil {
		HandleError(err)
	}

	fmt.Println("PutObjectSample completed")
}
