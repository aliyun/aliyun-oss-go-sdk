package sample

import (
	"fmt"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

// BucketLifecycleSample shows how to set, get and delete bucket's lifecycle.
func BucketLifecycleSample() {
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

	// Case 1: Set the lifecycle. The rule ID is id1 and the applied objects' prefix is one and expired time is 11/11/2015
	//var rule1 = oss.BuildLifecycleRuleByDate("id1", "one", true, 2015, 11, 11)
	rule1, err := oss.NewLifecycleRuleByCreateBeforeDate("id1", "one", true, 2015, 11, 11, oss.LRTExpriration)
	if err != nil {
		HandleError(err)
	}
	var rules = []oss.LifecycleRule{*rule1}
	err = client.SetBucketLifecycle(bucketName, rules)
	if err != nil {
		HandleError(err)
	}

	// Case 2: Set the lifecycle, The rule ID is id2 and the applied objects' prefix is two and the expired time is three days after the object created.
	//var rule2 = oss.BuildLifecycleRuleByDays("id2", "two", true, 3)
	rule2, err := oss.NewLifecleRuleByDays("id2", "two", true, 3, oss.LRTTransition, oss.StorageIA)
	if err != nil {
		HandleError(err)
	}
	rules = []oss.LifecycleRule{*rule2}
	err = client.SetBucketLifecycle(bucketName, rules)
	if err != nil {
		HandleError(err)
	}

	// Get the bucket's lifecycle
	lc, err := client.GetBucketLifecycle(bucketName)
	if err != nil {
		HandleError(err)
	}
	fmt.Println("Bucket Lifecycle:", lc.Rules)

	rule3, err := oss.NewLifecleRuleByDays("id3", "three", true, 3, oss.LRTAbortMultiPartUpload)
	if err != nil {
		HandleError(err)
	}
	rules = append(lc.Rules, *rule3)
	err = client.SetBucketLifecycle(bucketName, rules)
	if err != nil {
		HandleError(err)
	}

	// Get the bucket's lifecycle
	lc, err = client.GetBucketLifecycle(bucketName)
	if err != nil {
		HandleError(err)
	}
	fmt.Println("Bucket Lifecycle:", lc.Rules)

	// Delete bucket's Lifecycle
	err = client.DeleteBucketLifecycle(bucketName)
	if err != nil {
		HandleError(err)
	}

	// Delete bucket
	err = client.DeleteBucket(bucketName)
	if err != nil {
		HandleError(err)
	}

	fmt.Println("BucketLifecycleSample completed")
}
