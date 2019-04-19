package sample

import (
	"fmt"
	"strings"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

// ObjectTaggingSample shows how to set and get object Tagging
func ObjectTaggingSample() {
	// Create bucket
	bucket, err := GetTestBucket(bucketName)
	if err != nil {
		HandleError(err)
	}

	// Create object
	err = bucket.PutObject(objectKey, strings.NewReader("ObjectTaggingSample"))
	if err != nil {
		HandleError(err)
	}

	// Case 1: Set Tagging of object
	tag1 := oss.Tag{
		Key:   "key1",
		Value: "value1",
	}
	tag2 := oss.Tag{
		Key:   "key2",
		Value: "value2",
	}
	tagging := oss.ObjectTagging{
		Tags: []oss.Tag{tag1, tag2},
	}
	err = bucket.PutObjectTagging(objectKey, tagging)
	if err != nil {
		HandleError(err)
	}

	// Case 2: Get Tagging of object
	tagging, err = bucket.GetObjectTagging(objectKey)
	if err != nil {
		HandleError(err)
	}
	fmt.Printf("Object Tagging: %v\n", tagging)

	tag3 := oss.Tag{
		Key:   "key3",
		Value: "value3",
	}

	// Case 3: Put object with tagging
	tagging = oss.ObjectTagging{
		Tags: []oss.Tag{tag1, tag2, tag3},
	}
	err = bucket.PutObject(objectKey, strings.NewReader("ObjectTaggingSample"), oss.Tagging(tagging))
	if err != nil {
		HandleError(err)
	}

	// Case 4: Delete Tagging of object
	err = bucket.DeleteObjectTagging(objectKey)
	if err != nil {
		HandleError(err)
	}

	// Delete object and bucket
	err = DeleteTestBucketAndObject(bucketName)
	if err != nil {
		HandleError(err)
	}

	fmt.Println("ObjectACLSample completed")
}
