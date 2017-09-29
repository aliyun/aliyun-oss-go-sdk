package sample

import (
	"fmt"
	"strings"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

// ObjectMetaSample demos how to get and set the object metadata
func ObjectMetaSample() {
	// creates Bucket
	bucket, err := GetTestBucket(bucketName)
	if err != nil {
		HandleError(err)
	}

	// deletes object
	err = bucket.PutObject(objectKey, strings.NewReader("YoursObjectValue"))
	if err != nil {
		HandleError(err)
	}

	// case 0：sets Bucket Meta. one or more properties could be set
	// note：Meta is case insensitive
	options := []oss.Option{
		oss.Expires(futureDate),
		oss.Meta("myprop", "mypropval")}
	err = bucket.SetObjectMeta(objectKey, options...)
	if err != nil {
		HandleError(err)
	}

	// case 1：gets the object metadata. Only return basic meta information includes ETag, size and last modified.
	props, err := bucket.GetObjectMeta(objectKey)
	if err != nil {
		HandleError(err)
	}
	fmt.Println("Object Meta:", props)

	// case 2：gets the Object meta as well as custom meta
	props, err = bucket.GetObjectDetailedMeta(objectKey)
	if err != nil {
		HandleError(err)
	}
	fmt.Println("Expires:", props.Get("Expires"))

	// case 3：gets the object's all metadata with contraints. When constraints are met, return the metadata.
	props, err = bucket.GetObjectDetailedMeta(objectKey, oss.IfUnmodifiedSince(futureDate))
	if err != nil {
		HandleError(err)
	}
	fmt.Println("MyProp:", props.Get("X-Oss-Meta-Myprop"))

	_, err = bucket.GetObjectDetailedMeta(objectKey, oss.IfModifiedSince(futureDate))
	if err == nil {
		HandleError(err)
	}

	goar, err := bucket.GetObjectACL(objectKey)
	if err != nil {
		HandleError(err)
	}
	fmt.Println("Object ACL:", goar.ACL)

	// deletes object and bucket
	err = DeleteTestBucketAndObject(bucketName)
	if err != nil {
		HandleError(err)
	}

	fmt.Println("ObjectMetaSample completed")
}
