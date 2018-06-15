package sample

import (
	"fmt"
	"strings"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

// ObjectACLSample shows how to set and get object ACL
func ObjectACLSample() {
	// Create bucket
	bucket, err := GetTestBucket(bucketName)
	if err != nil {
		HandleError(err)
	}

	// Create object
	err = bucket.PutObject(objectKey, strings.NewReader("YoursObjectValue"))
	if err != nil {
		HandleError(err)
	}

	// Case 1: Set bucket ACL, valid ACLs are ACLPrivate、ACLPublicRead、ACLPublicReadWrite
	err = bucket.SetObjectACL(objectKey, oss.ACLPrivate)
	if err != nil {
		HandleError(err)
	}

	// Get object ACL, returns one of the three values: private、public-read、public-read-write
	goar, err := bucket.GetObjectACL(objectKey)
	if err != nil {
		HandleError(err)
	}
	fmt.Println("Object ACL:", goar.ACL)

	// Delete object and bucket
	err = DeleteTestBucketAndObject(bucketName)
	if err != nil {
		HandleError(err)
	}

	fmt.Println("ObjectACLSample completed")
}
