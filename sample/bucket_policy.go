package sample

import (
	"fmt"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

// BucketPolicySample shows how to set, get and delete the bucket policy configuration
func BucketPolicySample() {
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

	// the policy string
	var policyInfo string
	policyInfo = `
	{
		"Version":"1",
		"Statement":[
			{
				"Action":[
					"oss:GetObject",
					"oss:PutObject"
				],
				"Effect":"Deny",
				"Principal":"[123456790]",
				"Resource":["acs:oss:*:1234567890:*/*"]
			}
		]
	}`

	// Set policy
	err = client.SetBucketPolicy(bucketName, policyInfo)
	if err != nil {
		HandleError(err)
	}

	// Get Bucket policy
	ret, err := client.GetBucketPolicy(bucketName)
	if err != nil {
		HandleError(err)
	}
	fmt.Println("Bucket policy:", ret)

	// Delete Bucket policy
	err = client.DeleteBucketPolicy(bucketName)
	if err != nil {
		HandleError(err)
	}

	// Delete bucket
	err = client.DeleteBucket(bucketName)
	if err != nil {
		HandleError(err)
	}

	fmt.Println("BucketPolicySample completed")
}
