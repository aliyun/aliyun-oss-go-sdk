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

	// creates the bucket with default parameters
	err = client.CreateBucket(bucketName)
	if err != nil {
		HandleError(err)
	}
	// creates Target bucket for storing the logging files.
	var targetBucketName = "target-bucket"
	err = client.CreateBucket(targetBucketName)
	if err != nil {
		HandleError(err)
	}

	// case 1：sets Logging for object prefixed with "prefix-1" and save their access logs to the target bucket
	err = client.SetBucketLogging(bucketName, targetBucketName, "prefix-1", true)
	if err != nil {
		HandleError(err)
	}

	// case 2：sets the logging for the object prefixed with "prefix-2" and save their logs to the same bucket
	// Note: the rule will overwrite other rules if they have same bucket and prefix
	err = client.SetBucketLogging(bucketName, bucketName, "prefix-2", true)
	if err != nil {
		HandleError(err)
	}

	// deletes the bucket's logging configuration
	err = client.DeleteBucketLogging(bucketName)
	if err != nil {
		HandleError(err)
	}

	// case 3：sets the logging without enabling it
	err = client.SetBucketLogging(bucketName, targetBucketName, "prefix-3", false)
	if err != nil {
		HandleError(err)
	}

	// gets bucket's logging config
	gbl, err := client.GetBucketLogging(bucketName)
	if err != nil {
		HandleError(err)
	}
	fmt.Println("Bucket Logging:", gbl.LoggingEnabled)

	err = client.SetBucketLogging(bucketName, bucketName, "prefix2", true)
	if err != nil {
		HandleError(err)
	}

	// gets the Bucket logging config
	gbl, err = client.GetBucketLogging(bucketName)
	if err != nil {
		HandleError(err)
	}
	fmt.Println("Bucket Logging:", gbl.LoggingEnabled)

	// deletes Bucket's logging config
	err = client.DeleteBucketLogging(bucketName)
	if err != nil {
		HandleError(err)
	}

	// deletes bucket
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
