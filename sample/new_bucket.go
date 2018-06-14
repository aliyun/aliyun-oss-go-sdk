package sample

import (
	"fmt"
	"strings"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

// NewBucketSample shows how to initialize client and bucket
func NewBucketSample() {
	// New client
	client, err := oss.New(endpoint, accessID, accessKey)
	if err != nil {
		HandleError(err)
	}

	// Create bucket
	err = client.CreateBucket(bucketName)
	if err != nil {
		HandleError(err)
	}

	// New bucket
	bucket, err := client.Bucket(bucketName)
	if err != nil {
		HandleError(err)
	}

	// Put object, uploads an object
	var objectName = "myobject"
	err = bucket.PutObject(objectName, strings.NewReader("MyObjectValue"))
	if err != nil {
		HandleError(err)
	}

	// Delete object, deletes an object
	err = bucket.DeleteObject(objectName)
	if err != nil {
		HandleError(err)
	}

	// Delete bucket
	err = client.DeleteBucket(bucketName)
	if err != nil {
		HandleError(err)
	}

	fmt.Println("NewBucketSample completed")
}
