package sample

import (
	"fmt"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

// BucketRefererSample shows how to set, get and delete the bucket referer.
func BucketRefererSample() {
	// New client
	client, err := oss.New(endpoint, accessID, accessKey)
	if err != nil {
		HandleError(err)
	}

	// Create the bucket with default parameters
	err = client.CreateBucket(bucketName)
	if err != nil {
		HandleError(err)
	}

	var referers = []string{
		"http://www.aliyun.com",
		"http://www.???.aliyuncs.com",
		"http://www.*.com",
	}

	// Case 1: Set referers. The referers are with wildcards ? and * which could represent one and zero to multiple characters
	err = client.SetBucketReferer(bucketName, referers, false)
	if err != nil {
		HandleError(err)
	}

	// Case 2: Clear referers
	referers = []string{}
	err = client.SetBucketReferer(bucketName, referers, true)
	if err != nil {
		HandleError(err)
	}

	// Case 3: Create Refer With SetBucketRefererV2
	var setBucketReferer oss.RefererXML
	setBucketReferer.RefererList = []string{
		"http://www.aliyun.com",
		"https://www.aliyun.com",
		"http://www.???.aliyuncs.com",
		"http://www.*.com",
	}
	referer1 := "http://www.refuse.com"
	referer2 := "https://*.hack.com"
	referer3 := "http://ban.*.com"
	referer4 := "https://www.?.deny.com"
	setBucketReferer.RefererBlacklist = &oss.RefererBlacklist{
		[]string{
			referer1, referer2, referer3, referer4,
		},
	}
	setBucketReferer.AllowEmptyReferer = false
	boolTrue := true
	setBucketReferer.AllowTruncateQueryString = &boolTrue
	err = client.SetBucketRefererV2(bucketName, setBucketReferer)
	if err != nil {
		HandleError(err)
	}

	fmt.Println("Set Bucket Referer Success")

	// Get bucket referer configuration
	refRes, err := client.GetBucketReferer(bucketName)
	if err != nil {
		HandleError(err)
	}
	fmt.Println("Allow Empty Referer: ", refRes.AllowEmptyReferer)
	if refRes.AllowTruncateQueryString != nil {
		fmt.Println("Allow Truncate QueryString: ", *refRes.AllowTruncateQueryString)
	}
	if len(refRes.RefererList) > 0 {
		for _, referer := range refRes.RefererList {
			fmt.Println("Referer List: ", referer)
		}
	}
	if refRes.RefererBlacklist != nil {
		for _, refererBlack := range refRes.RefererBlacklist.Referer {
			fmt.Println("Referer Black List: ", refererBlack)
		}
	}
	fmt.Println("Get Bucket Referer Success")

	// Delete bucket referer
	// Case 1:Delete Refer With SetBucketReferer
	err = client.SetBucketReferer(bucketName, []string{}, true)
	if err != nil {
		HandleError(err)
	}

	// Case 2:Delete Refer With SetBucketRefererV2
	var delBucketReferer oss.RefererXML
	delBucketReferer.RefererList = []string{}
	delBucketReferer.AllowEmptyReferer = true
	err = client.SetBucketRefererV2(bucketName, delBucketReferer)
	if err != nil {
		HandleError(err)
	}
	fmt.Println("Delete Bucket Referer Success")

	// Delete bucket
	err = client.DeleteBucket(bucketName)
	if err != nil {
		HandleError(err)
	}
	fmt.Println("BucketRefererSample completed")
}
