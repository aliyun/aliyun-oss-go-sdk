package sample

import (
	"fmt"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

// BucketLifecycleSample shows how to set, get and delete bucket's lifecycle.
func BucketLifecycleSample() {
	// New Client
	client, err := oss.New(endpoint, accessID, accessKey)
	if err != nil {
		HandleError(err)
	}

	// Creates the bucket with default parameters
	err = client.CreateBucket(bucketName)
	if err != nil {
		HandleError(err)
	}

	// Case 1：sets the lifecycle. The Rule Id is id1 and the applied objects' prefix is one and expired time is 11/11/2015
	var rule1 = oss.BuildLifecycleRuleByDate("id1", "one", true, 2015, 11, 11)
	var rules = []oss.LifecycleRule{rule1}
	err = client.SetBucketLifecycle(bucketName, rules)
	if err != nil {
		HandleError(err)
	}

	// Case 2：Sets the Lifecycle，The rule Id is id2 and the applied objects' prefix is two and the expired time is three days after the object created.
	var rule2 = oss.BuildLifecycleRuleByDays("id2", "two", true, 3)
	rules = []oss.LifecycleRule{rule2}
	err = client.SetBucketLifecycle(bucketName, rules)
	if err != nil {
		HandleError(err)
	}

	// Case 3：creates two rules in the bucket for different objects. The rule with the same Id will be overwritten.
	var rule3 = oss.BuildLifecycleRuleByDays("id1", "two", true, 365)
	var rule4 = oss.BuildLifecycleRuleByDate("id2", "one", true, 2016, 11, 11)
	rules = []oss.LifecycleRule{rule3, rule4}
	err = client.SetBucketLifecycle(bucketName, rules)
	if err != nil {
		HandleError(err)
	}

	// Get the bucket's lifecycle
	gbl, err := client.GetBucketLifecycle(bucketName)
	if err != nil {
		HandleError(err)
	}
	fmt.Println("Bucket Lifecycle:", gbl.Rules)

	// Delete Bucket's Lifecycle
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
