package sample

import (
	"fmt"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

// BucketLoggingSample shows how to set, get and delete the bucket logging configuration
func BucketLoggingSample() {
	// New Client
	client, err := oss.New(endpoint, accessID, accessKey)
	if err != nil {
		HandleError(err)
	}

	// Creates the bucket with default parameters
	err = client.CreateBucket(bucketName)
	if err != nil {
		HandleError(err)
	}
	// Creates Target bucket for storing the logging files.
	var targetBucketName = "target-bucket"
	err = client.CreateBucket(targetBucketName)
	if err != nil {
		HandleError(err)
	}

	// Case 1: sets Logging for object prefixed with "prefix-1" and save their access logs to the target bucket
	err = client.SetBucketLogging(bucketName, targetBucketName, "prefix-1", true)
	if err != nil {
		HandleError(err)
	}

	// Case 2: sets the logging for the object prefixed with "prefix-2" and save their logs to the same bucket
	// Note: the rule will overwrite other rules if they have same bucket and prefix
	err = client.SetBucketLogging(bucketName, bucketName, "prefix-2", true)
	if err != nil {
		HandleError(err)
	}

	// Deletes the bucket's logging configuration
	err = client.DeleteBucketLogging(bucketName)
	if err != nil {
		HandleError(err)
	}

	// Case 3: sets the logging without enabling it
	err = client.SetBucketLogging(bucketName, targetBucketName, "prefix-3", false)
	if err != nil {
		HandleError(err)
	}

	// Gets bucket's logging configuration
	gbl, err := client.GetBucketLogging(bucketName)
	if err != nil {
		HandleError(err)
	}
	fmt.Println("Bucket Logging:", gbl.LoggingEnabled)

	err = client.SetBucketLogging(bucketName, bucketName, "prefix2", true)
	if err != nil {
		HandleError(err)
	}

	// Gets the Bucket logging configuration
	gbl, err = client.GetBucketLogging(bucketName)
	if err != nil {
		HandleError(err)
	}
	fmt.Println("Bucket Logging:", gbl.LoggingEnabled)

	// Deletes Bucket's logging configuration
	err = client.DeleteBucketLogging(bucketName)
	if err != nil {
		HandleError(err)
	}

	// Deletes bucket
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
