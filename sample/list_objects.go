package sample

import (
	"fmt"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

// ListObjectsSample demos the file list
func ListObjectsSample() {
	var myObjects = []Object{
		{"my-object-1", ""},
		{"my-object-11", ""},
		{"my-object-2", ""},
		{"my-object-21", ""},
		{"my-object-22", ""},
		{"my-object-3", ""},
		{"my-object-31", ""},
		{"my-object-32", ""}}

	// creates Bucket
	bucket, err := GetTestBucket(bucketName)
	if err != nil {
		HandleError(err)
	}

	// creates object
	err = CreateObjects(bucket, myObjects)
	if err != nil {
		HandleError(err)
	}

	// case 1：uses default parameters
	lor, err := bucket.ListObjects()
	if err != nil {
		HandleError(err)
	}
	fmt.Println("my objects:", getObjectsFormResponse(lor))

	// case 2：specifies max keys
	lor, err = bucket.ListObjects(oss.MaxKeys(3))
	if err != nil {
		HandleError(err)
	}
	fmt.Println("my objects max num:", getObjectsFormResponse(lor))

	// case 3：specifies prefix of objects
	lor, err = bucket.ListObjects(oss.Prefix("my-object-2"))
	if err != nil {
		HandleError(err)
	}
	fmt.Println("my objects prefix :", getObjectsFormResponse(lor))

	// case4：specifies the mark
	lor, err = bucket.ListObjects(oss.Marker("my-object-22"))
	if err != nil {
		HandleError(err)
	}
	fmt.Println("my objects marker :", getObjectsFormResponse(lor))

	// case5：list object with paging. each page has 3 objects
	marker := oss.Marker("")
	for {
		lor, err = bucket.ListObjects(oss.MaxKeys(3), marker)
		if err != nil {
			HandleError(err)
		}
		marker = oss.Marker(lor.NextMarker)
		fmt.Println("my objects page :", getObjectsFormResponse(lor))
		if !lor.IsTruncated {
			break
		}
	}

	// case 6：list object with paging , marker and max keys
	marker = oss.Marker("my-object-22")
	for {
		lor, err = bucket.ListObjects(oss.MaxKeys(3), marker)
		if err != nil {
			HandleError(err)
		}
		marker = oss.Marker(lor.NextMarker)
		fmt.Println("my objects marker&page :", getObjectsFormResponse(lor))
		if !lor.IsTruncated {
			break
		}
	}

	// case 7：list object with paging , with prefix and max keys
	pre := oss.Prefix("my-object-2")
	marker = oss.Marker("")
	for {
		lor, err = bucket.ListObjects(oss.MaxKeys(2), marker, pre)
		if err != nil {
			HandleError(err)
		}
		pre = oss.Prefix(lor.Prefix)
		marker = oss.Marker(lor.NextMarker)
		fmt.Println("my objects prefix&page :", getObjectsFormResponse(lor))
		if !lor.IsTruncated {
			break
		}
	}

	err = DeleteObjects(bucket, myObjects)
	if err != nil {
		HandleError(err)
	}

	// case 8：combile the prefix and delimiter for grouping. ListObjectsResponse.Objects is the objects returned.
	// ListObjectsResponse.CommonPrefixes is the common prefixes returned.
	myObjects = []Object{
		{"fun/test.txt", ""},
		{"fun/test.jpg", ""},
		{"fun/movie/001.avi", ""},
		{"fun/movie/007.avi", ""},
		{"fun/music/001.mp3", ""},
		{"fun/music/001.mp3", ""}}

	// creates object
	err = CreateObjects(bucket, myObjects)
	if err != nil {
		HandleError(err)
	}

	lor, err = bucket.ListObjects(oss.Prefix("fun/"), oss.Delimiter("/"))
	if err != nil {
		HandleError(err)
	}
	fmt.Println("my objects prefix :", getObjectsFormResponse(lor),
		"common prefixes:", lor.CommonPrefixes)

	// deletes object and bucket
	err = DeleteTestBucketAndObject(bucketName)
	if err != nil {
		HandleError(err)
	}

	fmt.Println("ListObjectsSample completed")
}

func getObjectsFormResponse(lor oss.ListObjectsResult) string {
	var output string
	for _, object := range lor.Objects {
		output += object.Key + "  "
	}
	return output
}
