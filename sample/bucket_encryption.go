package sample

import (
	"fmt"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

// BucketEncryptionSample shows how to get and set the bucket encryption Algorithm
func BucketEncryptionSample() {
	// New client
	client, err := oss.New(endpoint, accessID, accessKey)
	if err != nil {
		HandleError(err)
	}

	// Create a bucket with default parameters
	err = client.CreateBucket(bucketName)
	if err != nil {
		HandleError(err)
	}

	// SetBucketEncryption:AES256 ,"123"
	encryptionRule := oss.ServerEncryptionRule{}
	encryptionRule.SSEDefault.SSEAlgorithm = string(oss.AESAlgorithm)
	err = client.SetBucketEncryption(bucketName, encryptionRule)
	if err != nil {
		HandleError(err)
	}

	// Get bucket encryption
	encryptionResult, err := client.GetBucketEncryption(bucketName)
	if err != nil {
		HandleError(err)
	}
	fmt.Println("Bucket Encryption:", encryptionResult)

	// Delete the bucket
	err = client.DeleteBucketEncryption(bucketName)
	if err != nil {
		HandleError(err)
	}

	// Delete the object and bucket
	err = DeleteTestBucketAndObject(bucketName)
	if err != nil {
		HandleError(err)
	}

	fmt.Println("BucketEncryptionSample completed")
}
