package sample

import (
	"fmt"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

// CreateBucketSample shows how to create bucket
func CreateBucketSample() {
	// New Client
	client, err := oss.New(endpoint, accessID, accessKey)
	if err != nil {
		HandleError(err)
	}

	DeleteTestBucketAndObject(bucketName)

	// Case 1：creates a bucket with default parameters
	err = client.CreateBucket(bucketName)
	if err != nil {
		HandleError(err)
	}

	// Deletes bucket
	err = client.DeleteBucket(bucketName)
	if err != nil {
		HandleError(err)
	}

	// Case 2：creates the bucket with ACL
	err = client.CreateBucket(bucketName, oss.ACL(oss.ACLPublicRead))
	if err != nil {
		HandleError(err)
	}

	// Case 3：repeat the same bucket. OSS will not return error, but just no op. The ACL is not updated.
	err = client.CreateBucket(bucketName, oss.ACL(oss.ACLPublicReadWrite))
	if err != nil {
		HandleError(err)
	}

	// Deletes bucket
	err = client.DeleteBucket(bucketName)
	if err != nil {
		HandleError(err)
	}

	fmt.Println("CreateBucketSample completed")
}
