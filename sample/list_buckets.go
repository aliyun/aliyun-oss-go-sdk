package sample

import (
	"fmt"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

// ListBucketsSample shows the list bucket, including default and specified parameters.
func ListBucketsSample() {
	var myBuckets = []string{
		"my-bucket-1",
		"my-bucket-11",
		"my-bucket-2",
		"my-bucket-21",
		"my-bucket-22",
		"my-bucket-3",
		"my-bucket-31",
		"my-bucket-32"}

	// New Client
	client, err := oss.New(endpoint, accessID, accessKey)
	if err != nil {
		HandleError(err)
	}

	// remove other bucket
	lbr, err := client.ListBuckets()
	if err != nil {
		HandleError(err)
	}

	for _, bucket := range lbr.Buckets {
		err = client.DeleteBucket(bucket.Name)
		if err != nil {
			//HandleError(err)
		}
	}

	// creates bucket
	for _, bucketName := range myBuckets {
		err = client.CreateBucket(bucketName)
		if err != nil {
			HandleError(err)
		}
	}

	// case 1：use default parameter
	lbr, err = client.ListBuckets()
	if err != nil {
		HandleError(err)
	}
	fmt.Println("my buckets:", lbr.Buckets)

	// case 2：specifies the max keys : 3
	lbr, err = client.ListBuckets(oss.MaxKeys(3))
	if err != nil {
		HandleError(err)
	}
	fmt.Println("my buckets max num:", lbr.Buckets)

	// case3：specifies the prefix of buckets.
	lbr, err = client.ListBuckets(oss.Prefix("my-bucket-2"))
	if err != nil {
		HandleError(err)
	}
	fmt.Println("my buckets prefix :", lbr.Buckets)

	// case 4：specifies the marker
	lbr, err = client.ListBuckets(oss.Marker("my-bucket-22"))
	if err != nil {
		HandleError(err)
	}
	fmt.Println("my buckets marker :", lbr.Buckets)

	// case 5：specifies max key and list all buckets with paging
	marker := oss.Marker("")
	for {
		lbr, err = client.ListBuckets(oss.MaxKeys(3), marker)
		if err != nil {
			HandleError(err)
		}
		marker = oss.Marker(lbr.NextMarker)
		fmt.Println("my buckets page :", lbr.Buckets)
		if !lbr.IsTruncated {
			break
		}
	}

	// case 6：list bucket with marker and max key
	marker = oss.Marker("my-bucket-22")
	for {
		lbr, err = client.ListBuckets(oss.MaxKeys(3), marker)
		if err != nil {
			HandleError(err)
		}
		marker = oss.Marker(lbr.NextMarker)
		fmt.Println("my buckets marker&page :", lbr.Buckets)
		if !lbr.IsTruncated {
			break
		}
	}

	// case 7：list bucket with prefix and max key
	pre := oss.Prefix("my-bucket-2")
	marker = oss.Marker("")
	for {
		lbr, err = client.ListBuckets(oss.MaxKeys(3), pre, marker)
		if err != nil {
			HandleError(err)
		}
		pre = oss.Prefix(lbr.Prefix)
		marker = oss.Marker(lbr.NextMarker)
		fmt.Println("my buckets prefix&page :", lbr.Buckets)
		if !lbr.IsTruncated {
			break
		}
	}

	// deletes bucket
	for _, bucketName := range myBuckets {
		err = client.DeleteBucket(bucketName)
		if err != nil {
			HandleError(err)
		}
	}

	fmt.Println("ListsBucketSample completed")
}
