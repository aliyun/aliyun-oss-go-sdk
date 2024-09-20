package sample

import (
	"fmt"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

// BucketArchiveDirectReadSample shows how to set, get the bucket archive direct read.
func BucketArchiveDirectReadSample() {
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

	// Set bucket's archive direct read
	var req oss.PutBucketArchiveDirectRead
	req.Enabled = true
	err = client.PutBucketArchiveDirectRead(bucketName, req)
	if err != nil {
		HandleError(err)
	}
	fmt.Println("Set Bucket Archive Direct Read Success!")

	// Get bucket's archive direct read
	res, err := client.GetBucketArchiveDirectRead(bucketName)
	if err != nil {
		HandleError(err)
	}
	fmt.Println("Bucket Archive Direct Read Enabled:", res.Enabled)

	// Delete bucket
	err = client.DeleteBucket(bucketName)
	if err != nil {
		HandleError(err)
	}

	fmt.Println("BucketArchiveDirectReadSample completed")
}
