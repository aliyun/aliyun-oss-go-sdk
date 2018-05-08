package sample

import (
	"fmt"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

// CopyObjectSample shows the copy files usage
func CopyObjectSample() {
	// creates a Bucket
	bucket, err := GetTestBucket(bucketName)
	if err != nil {
		HandleError(err)
	}

	// creates an Object
	err = bucket.PutObjectFromFile(objectKey, localFile)
	if err != nil {
		HandleError(err)
	}

	// case 1：copy an existing object
	var descObjectKey = "descobject"
	_, err = bucket.CopyObject(objectKey, descObjectKey)
	if err != nil {
		HandleError(err)
	}

	// case 2：copy an existing object to another existing object
	_, err = bucket.CopyObject(objectKey, descObjectKey)
	if err != nil {
		HandleError(err)
	}

	err = bucket.DeleteObject(descObjectKey)
	if err != nil {
		HandleError(err)
	}

	// case 3：copy file with constraints. When the constraints are met, the copy executes. otherwise the copy does not execute.
	// constraints are not met，copy does not execute
	_, err = bucket.CopyObject(objectKey, descObjectKey, oss.CopySourceIfModifiedSince(futureDate))
	if err == nil {
		HandleError(err)
	}
	fmt.Println("CopyObjectError:", err)
	// constraints are met, the copy executes
	_, err = bucket.CopyObject(objectKey, descObjectKey, oss.CopySourceIfUnmodifiedSince(futureDate))
	if err != nil {
		HandleError(err)
	}

	// case 4：specify the properties when copying. The MetadataDirective needs to be MetaReplace
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

	// case 5：when the source file is same as the target file, the copy could be used to update metadata
	options = []oss.Option{
		oss.Expires(futureDate),
		oss.Meta("myprop", "mypropval"),
		oss.MetadataDirective(oss.MetaReplace)}

	_, err = bucket.CopyObject(objectKey, objectKey, options...)
	if err != nil {
		HandleError(err)
	}
	fmt.Println("meta:", meta)

	// case 6：big file's multipart copy. It supports concurrent copy with resumable upload
	// copy file with multipart. The part size is 100K. By default one routine is used without resumable upload
	err = bucket.CopyFile(bucketName, objectKey, descObjectKey, 100*1024)
	if err != nil {
		HandleError(err)
	}

	// part size is 100K and three coroutines for the concurrent copy
	err = bucket.CopyFile(bucketName, objectKey, descObjectKey, 100*1024, oss.Routines(3))
	if err != nil {
		HandleError(err)
	}

	// part size is 100K and three coroutines for the concurrent copy with resumable upload
	err = bucket.CopyFile(bucketName, objectKey, descObjectKey, 100*1024, oss.Routines(3), oss.Checkpoint(true, ""))
	if err != nil {
		HandleError(err)
	}

	// specify the checkpoint file path. If the checkpoint file path is not specified, the current folder is used.
	err = bucket.CopyFile(bucketName, objectKey, descObjectKey, 100*1024, oss.Checkpoint(true, localFile+".cp"))
	if err != nil {
		HandleError(err)
	}

	// deletes object and bucket
	err = DeleteTestBucketAndObject(bucketName)
	if err != nil {
		HandleError(err)
	}

	fmt.Println("CopyObjectSample completed")
}
