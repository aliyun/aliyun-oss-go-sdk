package sample

import (
	"fmt"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

// CopyObjectSample shows the copy files usage
func CopyObjectSample() {
	// Create a bucket
	bucket, err := GetTestBucket(bucketName)
	if err != nil {
		HandleError(err)
	}

	// Create an object
	err = bucket.PutObjectFromFile(objectKey, localFile)
	if err != nil {
		HandleError(err)
	}

	// Case 1: Copy an existing object
	var descObjectKey = "descobject"
	_, err = bucket.CopyObject(objectKey, descObjectKey)
	if err != nil {
		HandleError(err)
	}

	// Case 2: Copy an existing object to another existing object
	_, err = bucket.CopyObject(objectKey, descObjectKey)
	if err != nil {
		HandleError(err)
	}

	err = bucket.DeleteObject(descObjectKey)
	if err != nil {
		HandleError(err)
	}

	// Case 3: Copy file with constraints. When the constraints are met, the copy executes. otherwise the copy does not execute.
	// constraints are not met, copy does not execute
	_, err = bucket.CopyObject(objectKey, descObjectKey, oss.CopySourceIfModifiedSince(futureDate))
	if err == nil {
		HandleError(err)
	}
	fmt.Println("CopyObjectError:", err)
	// Constraints are met, the copy executes
	_, err = bucket.CopyObject(objectKey, descObjectKey, oss.CopySourceIfUnmodifiedSince(futureDate))
	if err != nil {
		HandleError(err)
	}

	// Case 4: Specify the properties when copying. The MetadataDirective needs to be MetaReplace
	options := []oss.Option{
		oss.Expires(futureDate),
		oss.Meta("myprop", "mypropval"),
		oss.MetadataDirective(oss.MetaReplace)}
	_, err = bucket.CopyObject(objectKey, descObjectKey, options...)
	if err != nil {
		HandleError(err)
	}

	meta, err := bucket.GetObjectDetailedMeta(descObjectKey)
	if err != nil {
		HandleError(err)
	}
	fmt.Println("meta:", meta)

	// Case 5: When the source file is the same as the target file, the copy could be used to update metadata
	options = []oss.Option{
		oss.Expires(futureDate),
		oss.Meta("myprop", "mypropval"),
		oss.MetadataDirective(oss.MetaReplace)}

	_, err = bucket.CopyObject(objectKey, objectKey, options...)
	if err != nil {
		HandleError(err)
	}
	fmt.Println("meta:", meta)

	// Case 6: Big file's multipart copy. It supports concurrent copy with resumable upload
	// copy file with multipart. The part size is 100K. By default one routine is used without resumable upload
	err = bucket.CopyFile(bucketName, objectKey, descObjectKey, 100*1024)
	if err != nil {
		HandleError(err)
	}

	// Part size is 100K and three coroutines for the concurrent copy
	err = bucket.CopyFile(bucketName, objectKey, descObjectKey, 100*1024, oss.Routines(3))
	if err != nil {
		HandleError(err)
	}

	// Part size is 100K and three coroutines for the concurrent copy with resumable upload
	err = bucket.CopyFile(bucketName, objectKey, descObjectKey, 100*1024, oss.Routines(3), oss.Checkpoint(true, ""))
	if err != nil {
		HandleError(err)
	}

	// Specify the checkpoint file path. If the checkpoint file path is not specified, the current folder is used.
	err = bucket.CopyFile(bucketName, objectKey, descObjectKey, 100*1024, oss.Checkpoint(true, localFile+".cp"))
	if err != nil {
		HandleError(err)
	}

	// Delete object and bucket
	err = DeleteTestBucketAndObject(bucketName)
	if err != nil {
		HandleError(err)
	}

	fmt.Println("CopyObjectSample completed")
}
