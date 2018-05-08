package sample

import (
	"fmt"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

// BucketACLSample shows how to get and set the bucket ACL
func BucketACLSample() {
	// New Client
	client, err := oss.New(endpoint, accessID, accessKey)
	if err != nil {
		HandleError(err)
	}

	// creates a bucket with default parameters
	err = client.CreateBucket(bucketName)
	if err != nil {
		HandleError(err)
	}

	// sets Bucket ACL. The valid ACLs are ACLPrivate、ACLPublicRead、ACLPublicReadWrite
	err = client.SetBucketACL(bucketName, oss.ACLPublicRead)
	if err != nil {
		HandleError(err)
	}

	// gets Bucket ACL
	gbar, err := client.GetBucketACL(bucketName)
	if err != nil {
		HandleError(err)
	}
	fmt.Println("Bucket ACL:", gbar.ACL)

	// deletes the bucket
	err = client.DeleteBucket(bucketName)
	if err != nil {
		HandleError(err)
	}

	fmt.Println("BucketACLSample completed")
}
