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

	// Case 1: Set the lifecycle. The rule ID is rule1 and the applied objects' prefix is one and expired time is 11/11/2015
	expriation := oss.LifecycleExpiration{
		CreatedBeforeDate: "2015-11-11T00:00:00.000Z",
	}
	rule1, err := oss.NewLifecycleRule("rule1", "one", true, &expriation, nil)
	if err != nil {
		HandleError(err)
	}
	var rules = []oss.LifecycleRule{*rule1}
	err = client.SetBucketLifecycle(bucketName, rules)
	if err != nil {
		HandleError(err)
	}

	// Get the bucket's lifecycle
	lc, err := client.GetBucketLifecycle(bucketName)
	if err != nil {
		HandleError(err)
	}
	fmt.Printf("Bucket Lifecycle:%v, %v\n", lc.Rules, *lc.Rules[0].Expiration)

	// Case 2: Set the lifecycle, The rule ID is id2 and the applied objects' prefix is two and the expired time is three days after the object created.
	transitionIA := oss.LifecycleTransition{
		Days:         3,
		StorageClass: oss.StorageIA,
	}
	transitionArch := oss.LifecycleTransition{
		Days:         30,
		StorageClass: oss.StorageArchive,
	}
	rule2, err := oss.NewLifecycleRule("rule2", "two", true, nil, nil, &transitionIA, &transitionArch)
	if err != nil {
		HandleError(err)
	}
	rules = []oss.LifecycleRule{*rule2}
	err = client.SetBucketLifecycle(bucketName, rules)
	if err != nil {
		HandleError(err)
	}

	// Get the bucket's lifecycle
	lc, err = client.GetBucketLifecycle(bucketName)
	if err != nil {
		HandleError(err)
	}
	fmt.Printf("Bucket Lifecycle:%v, %v, %v\n", lc.Rules, *lc.Rules[0].Transition[0], *lc.Rules[0].Transition[1])

	abortMPU := oss.LifecycleAbortMultipartUpload{
		Days: 3,
	}
	rule3, err := oss.NewLifecycleRule("rule3", "three", true, nil, &abortMPU)
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
	fmt.Printf("Bucket Lifecycle:%v, %v, %v, %v\n", lc.Rules, *lc.Rules[0].Transition[0], *lc.Rules[0].Transition[1], *lc.Rules[1].AbortMultipartUpload)

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
