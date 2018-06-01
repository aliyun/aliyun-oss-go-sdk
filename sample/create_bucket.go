package sample

import (
	"fmt"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

// CreateBucketSample shows how to create bucket
func CreateBucketSample() {
	// New client
	client, err := oss.New(endpoint, accessID, accessKey)
	if err != nil {
		HandleError(err)
	}

	DeleteTestBucketAndObject(bucketName)

	// Case 1: Create a bucket with default parameters
	err = client.CreateBucket(bucketName)
	if err != nil {
		HandleError(err)
	}

	// Delete bucket
	err = client.DeleteBucket(bucketName)
	if err != nil {
		HandleError(err)
	}

	// Case 2: Create the bucket with ACL
	err = client.CreateBucket(bucketName, oss.ACL(oss.ACLPublicRead))
	if err != nil {
		HandleError(err)
	}

	// Case 3: Repeat the same bucket. OSS will not return error, but just no op. The ACL is not updated.
	err = client.CreateBucket(bucketName, oss.ACL(oss.ACLPublicReadWrite))
	if err != nil {
		HandleError(err)
	}

	// Delete bucket
	err = client.DeleteBucket(bucketName)
	if err != nil {
		HandleError(err)
	}

	fmt.Println("CreateBucketSample completed")
}
