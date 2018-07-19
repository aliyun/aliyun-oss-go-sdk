package sample

import (
	"fmt"
	"strings"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

// DeleteObjectSample shows how to delete single file or multiple files
func DeleteObjectSample() {
	// Create a bucket
	bucket, err := GetTestBucket(bucketName)
	if err != nil {
		HandleError(err)
	}

	var val = "抽刀断水水更流，举杯销愁愁更愁。 人生在世不称意，明朝散发弄扁舟。"

	// Case 1: Delete an object
	err = bucket.PutObject(objectKey, strings.NewReader(val))
	if err != nil {
		HandleError(err)
	}

	err = bucket.DeleteObject(objectKey)
	if err != nil {
		HandleError(err)
	}

	// Case 2: Delete multiple Objects
	err = bucket.PutObject(objectKey+"1", strings.NewReader(val))
	if err != nil {
		HandleError(err)
	}

	err = bucket.PutObject(objectKey+"2", strings.NewReader(val))
	if err != nil {
		HandleError(err)
	}

	delRes, err := bucket.DeleteObjects([]string{objectKey + "1", objectKey + "2"})
	if err != nil {
		HandleError(err)
	}
	fmt.Println("Del Res:", delRes)

	lsRes, err := bucket.ListObjects()
	if err != nil {
		HandleError(err)
	}
	fmt.Println("Objects:", getObjectsFormResponse(lsRes))

	// Case 3: Delete multiple objects and it will return deleted objects in detail mode which is by default.
	err = bucket.PutObject(objectKey+"1", strings.NewReader(val))
	if err != nil {
		HandleError(err)
	}

	err = bucket.PutObject(objectKey+"2", strings.NewReader(val))
	if err != nil {
		HandleError(err)
	}

	delRes, err = bucket.DeleteObjects([]string{objectKey + "1", objectKey + "2"},
		oss.DeleteObjectsQuiet(false))
	if err != nil {
		HandleError(err)
	}
	fmt.Println("Detail Del Res:", delRes)

	lsRes, err = bucket.ListObjects()
	if err != nil {
		HandleError(err)
	}
	fmt.Println("Objects:", getObjectsFormResponse(lsRes))

	// Case 4: Delete multiple objects and returns undeleted objects in quiet mode
	err = bucket.PutObject(objectKey+"1", strings.NewReader(val))
	if err != nil {
		HandleError(err)
	}

	err = bucket.PutObject(objectKey+"2", strings.NewReader(val))
	if err != nil {
		HandleError(err)
	}

	delRes, err = bucket.DeleteObjects([]string{objectKey + "1", objectKey + "2"}, oss.DeleteObjectsQuiet(true))
	if err != nil {
		HandleError(err)
	}
	fmt.Println("Sample Del Res:", delRes)

	lsRes, err = bucket.ListObjects()
	if err != nil {
		HandleError(err)
	}
	fmt.Println("Objects:", getObjectsFormResponse(lsRes))

	// Delete object and bucket
	err = DeleteTestBucketAndObject(bucketName)
	if err != nil {
		HandleError(err)
	}

	fmt.Println("DeleteObjectSample completed")
}
