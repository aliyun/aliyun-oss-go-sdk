// Package sample examples
package sample

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

// AppendObjectSample shows the append file's usage
func AppendObjectSample() {
	// Create bucket
	bucket, err := GetTestBucket(bucketName)
	if err != nil {
		HandleError(err)
	}

	err = bucket.DeleteObject(objectKey)

	var str = "弃我去者，昨日之日不可留。 乱我心者，今日之日多烦忧！"
	var nextPos int64

	// Case 1: Append a string to the object
	// The first append position is 0 and the return value is for the next append's position.
	nextPos, err = bucket.AppendObject(objectKey, strings.NewReader(str), nextPos)
	if err != nil {
		HandleError(err)
	}

	// Second append
	nextPos, err = bucket.AppendObject(objectKey, strings.NewReader(str), nextPos)
	if err != nil {
		HandleError(err)
	}

	// Download
	body, err := bucket.GetObject(objectKey)
	if err != nil {
		HandleError(err)
	}
	data, err := ioutil.ReadAll(body)
	body.Close()
	if err != nil {
		HandleError(err)
	}
	fmt.Println(objectKey, ":", string(data))

	err = bucket.DeleteObject(objectKey)
	if err != nil {
		HandleError(err)
	}

	// Case 2: Append byte array to the object
	nextPos = 0
	// The first append position is 0, and the return value is for the next append's position.
	nextPos, err = bucket.AppendObject(objectKey, bytes.NewReader([]byte(str)), nextPos)
	if err != nil {
		HandleError(err)
	}

	// Second append
	nextPos, err = bucket.AppendObject(objectKey, bytes.NewReader([]byte(str)), nextPos)
	if err != nil {
		HandleError(err)
	}

	// Download
	body, err = bucket.GetObject(objectKey)
	if err != nil {
		HandleError(err)
	}
	data, err = ioutil.ReadAll(body)
	body.Close()
	if err != nil {
		HandleError(err)
	}
	fmt.Println(objectKey, ":", string(data))

	err = bucket.DeleteObject(objectKey)
	if err != nil {
		HandleError(err)
	}

	// Case 3: Append a local file to the object
	fd, err := os.Open(localFile)
	if err != nil {
		HandleError(err)
	}
	defer fd.Close()

	nextPos = 0
	nextPos, err = bucket.AppendObject(objectKey, fd, nextPos)
	if err != nil {
		HandleError(err)
	}

	// Case 4: Get the next append position by GetObjectDetailedMeta
	props, err := bucket.GetObjectDetailedMeta(objectKey)
	nextPos, err = strconv.ParseInt(props.Get(oss.HTTPHeaderOssNextAppendPosition), 10, 64)
	if err != nil {
		HandleError(err)
	}

	nextPos, err = bucket.AppendObject(objectKey, strings.NewReader(str), nextPos)
	if err != nil {
		HandleError(err)
	}

	err = bucket.DeleteObject(objectKey)
	if err != nil {
		HandleError(err)
	}

	// Case 5: Specify the object properties for the first append, including the "x-oss-meta"'s custom metadata.
	options := []oss.Option{
		oss.Expires(futureDate),
		oss.ObjectACL(oss.ACLPublicRead),
		oss.Meta("myprop", "mypropval")}
	nextPos = 0
	fd.Seek(0, os.SEEK_SET)
	nextPos, err = bucket.AppendObject(objectKey, strings.NewReader(str), nextPos, options...)
	if err != nil {
		HandleError(err)
	}
	// Second append
	fd.Seek(0, os.SEEK_SET)
	nextPos, err = bucket.AppendObject(objectKey, strings.NewReader(str), nextPos)
	if err != nil {
		HandleError(err)
	}

	props, err = bucket.GetObjectDetailedMeta(objectKey)
	if err != nil {
		HandleError(err)
	}
	fmt.Println("myprop:", props.Get("x-oss-meta-myprop"))

	goar, err := bucket.GetObjectACL(objectKey)
	if err != nil {
		HandleError(err)
	}
	fmt.Println("Object ACL:", goar.ACL)

	// Case 6: Set the storage classes.OSS provides three storage classes: Standard, Infrequent Access, and Archive.
	// Upload a strings, and you can append some strings in the behind of object. but the object is 'Archive' storange class.
	// An object created with the AppendObject operation is an appendable object. set the object storange-class to 'Archive'.
	nextPos, err = bucket.AppendObject(appendObjectKey, strings.NewReader("昨夜雨疏风骤，浓睡不消残酒。试问卷帘人，"), nextPos, oss.ObjectStorageClass("Archive"))
	if err != nil {
		HandleError(err)
	}

	// Delete the object and bucket
	err = DeleteTestBucketAndObject(bucketName)
	if err != nil {
		HandleError(err)
	}

	fmt.Println("AppendObjectSample completed")
}
