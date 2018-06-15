package sample

import (
	"fmt"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

// BucketLoggingSample shows how to set, get and delete the bucket logging configuration
func BucketLoggingSample() {
	// New client
	client, err := oss.New(endpoint, accessID, accessKey)
	if err != nil {
		HandleError(err)
	}

	// Create the bucket with default parameters
	err = client.CreateBucket(bucketName)
	if err != nil {
		HandleError(err)
	}
	// Create target bucket to store the logging files.
	var targetBucketName = "target-bucket"
	err = client.CreateBucket(targetBucketName)
	if err != nil {
		HandleError(err)
	}

	// Case 1: Set the logging for the object prefixed with "prefix-1" and save their access logs to the target bucket
	err = client.SetBucketLogging(bucketName, targetBucketName, "prefix-1", true)
	if err != nil {
		HandleError(err)
	}

	// Case 2: Set the logging for the object prefixed with "prefix-2" and save their logs to the same bucket
	// Note: the rule will overwrite other rules if they have same bucket and prefix
	err = client.SetBucketLogging(bucketName, bucketName, "prefix-2", true)
	if err != nil {
		HandleError(err)
	}

	// Delete the bucket's logging configuration
	err = client.DeleteBucketLogging(bucketName)
	if err != nil {
		HandleError(err)
	}

	// Case 3: Set the logging without enabling it
	err = client.SetBucketLogging(bucketName, targetBucketName, "prefix-3", false)
	if err != nil {
		HandleError(err)
	}

	// Get the bucket's logging configuration
	gbl, err := client.GetBucketLogging(bucketName)
	if err != nil {
		HandleError(err)
	}
	fmt.Println("Bucket Logging:", gbl.LoggingEnabled)

	err = client.SetBucketLogging(bucketName, bucketName, "prefix2", true)
	if err != nil {
		HandleError(err)
	}

	// Get the bucket's logging configuration
	gbl, err = client.GetBucketLogging(bucketName)
	if err != nil {
		HandleError(err)
	}
	fmt.Println("Bucket Logging:", gbl.LoggingEnabled)

	// Delete the bucket's logging configuration
	err = client.DeleteBucketLogging(bucketName)
	if err != nil {
		HandleError(err)
	}

	// Delete bucket
	err = client.DeleteBucket(bucketName)
	if err != nil {
		HandleError(err)
	}
	err = client.DeleteBucket(targetBucketName)
	if err != nil {
		HandleError(err)
	}

	fmt.Println("BucketLoggingSample completed")
}
