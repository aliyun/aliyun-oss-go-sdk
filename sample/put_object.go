package sample

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

// PutObjectSample illustrates two methods for uploading a file: simple upload and multipart upload.
func PutObjectSample() {
	// creates Bucket
	bucket, err := GetTestBucket(bucketName)
	if err != nil {
		HandleError(err)
	}

	var val = "花间一壶酒，独酌无相亲。 举杯邀明月，对影成三人。"

	// case 1：uploads a object from a string
	err = bucket.PutObject(objectKey, strings.NewReader(val))
	if err != nil {
		HandleError(err)
	}

	// case 2：uploads a object from a byte[]
	err = bucket.PutObject(objectKey, bytes.NewReader([]byte(val)))
	if err != nil {
		HandleError(err)
	}

	// case3：uploads the local file with file handle, user should open the file at first.
	fd, err := os.Open(localFile)
	if err != nil {
		HandleError(err)
	}
	defer fd.Close()

	err = bucket.PutObject(objectKey, fd)
	if err != nil {
		HandleError(err)
	}

	// case 4：uploads an object with local file name, user need not open the file.
	err = bucket.PutObjectFromFile(objectKey, localFile)
	if err != nil {
		HandleError(err)
	}

	// case5：uploads a object with specified properties, PutObject/PutObjectFromFile/UploadFile also support this feature.
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

	// case 6：big file's multipart upload. It supports concurrent upload with checkpoint
	// multipart upload with 100K as part size. By default 1 coroutine is used and no checkpoint is used.
	err = bucket.UploadFile(objectKey, localFile, 100*1024)
	if err != nil {
		HandleError(err)
	}

	// part size is 100K and 3 coroutines are used
	err = bucket.UploadFile(objectKey, localFile, 100*1024, oss.Routines(3))
	if err != nil {
		HandleError(err)
	}

	// part size is 100K and 3 coroutines are used with checkpoint
	err = bucket.UploadFile(objectKey, localFile, 100*1024, oss.Routines(3), oss.Checkpoint(true, ""))
	if err != nil {
		HandleError(err)
	}

	// specifies the local file path for checkpoint files. 
	// the 2nd paramter of Checkpoint can specify the file path, when the file path is empty, it will upload the directory.
	err = bucket.UploadFile(objectKey, localFile, 100*1024, oss.Checkpoint(true, localFile+".cp"))
	if err != nil {
		HandleError(err)
	}

	// deletes object and bucket
	err = DeleteTestBucketAndObject(bucketName)
	if err != nil {
		HandleError(err)
	}

	fmt.Println("PutObjectSample completed")
}
