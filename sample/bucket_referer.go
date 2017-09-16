package sample

import (
	"fmt"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

// BucketRefererSample demos how to set, get and delete the bucket referer.
func BucketRefererSample() {
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

	var referers = []string{
		"http://www.aliyun.com",
		"http://www.???.aliyuncs.com",
		"http://www.*.com",
	}

	// case 1：sets referers. The referers are with wildcards ? and * which could represent one and zero to multiple characters
	err = client.SetBucketReferer(bucketName, referers, false)
	if err != nil {
		HandleError(err)
	}

	// case 2：clear referers
	referers = []string{}
	err = client.SetBucketReferer(bucketName, referers, true)
	if err != nil {
		HandleError(err)
	}

	// gets Bucket referer config
	gbr, err := client.GetBucketReferer(bucketName)
	if err != nil {
		HandleError(err)
	}
	fmt.Println("Bucket Referers:", gbr.RefererList,
		"AllowEmptyReferer:", gbr.AllowEmptyReferer)

	// deletes bucket
	err = client.DeleteBucket(bucketName)
	if err != nil {
		HandleError(err)
	}

	fmt.Println("BucketRefererSample completed")
}
