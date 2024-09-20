package sample

import (
	"encoding/base64"
	"fmt"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"os"
	"strings"
)

// BucketCallbackPolicySample shows how to set, get and delete the bucket callback policy configuration
func BucketCallbackPolicySample() {
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
	var callbackPolicy oss.PutBucketCallbackPolicy
	callbackVal := base64.StdEncoding.EncodeToString([]byte(`{"callbackUrl":"http://www.aliyuncs.com", "callbackBody":"bucket=${bucket}&object=${object}"}`))

	callbackVal2 := base64.StdEncoding.EncodeToString([]byte(`{"callbackUrl":"http://www.aliyun.com", "callbackBody":"bucket=${bucket}&object=${object}"}`))

	callbackVar2 := base64.StdEncoding.EncodeToString([]byte(`{"x:a":"a", "x:b":"b"}`))
	callbackPolicy = oss.PutBucketCallbackPolicy{
		PolicyItem: []oss.PolicyItem{
			{
				PolicyName:  "first",
				Callback:    callbackVal,
				CallbackVar: "",
			},
			{
				PolicyName:  "second",
				Callback:    callbackVal2,
				CallbackVar: callbackVar2,
			},
		},
	}

	// Set callback policy
	err = client.PutBucketCallbackPolicy(bucketName, callbackPolicy)
	if err != nil {
		HandleError(err)
	}

	fmt.Println("Put Bucket Callback Policy Success!")

	// Get Bucket policy
	result, err := client.GetBucketCallbackPolicy(bucketName)
	if err != nil {
		HandleError(err)
	}
	for _, policy := range result.PolicyItem {
		fmt.Printf("Callback Policy Name:%s\n", policy.PolicyName)

		callback, _ := base64.StdEncoding.DecodeString(policy.Callback)
		fmt.Printf("Callback Policy Callback:%s\n", callback)
		if policy.CallbackVar != "" {
			callbackVar, _ := base64.StdEncoding.DecodeString(policy.CallbackVar)
			fmt.Printf("Callback Policy Callback Var:%s\n", callbackVar)
		}
	}
	fmt.Println("Get Bucket Callback Policy Success!")

	// Delete Bucket policy
	err = client.DeleteBucketCallbackPolicy(bucketName)
	if err != nil {
		HandleError(err)
	}
	fmt.Println("Delete Bucket Callback Policy Success!")

	// put object by use callback policy
	bucket, err := client.Bucket(bucketName)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(-1)
	}
	name := base64.StdEncoding.EncodeToString([]byte(`{"callbackPolicy":"first"}`))
	err = bucket.PutObject("example-object.txt", strings.NewReader("Hello OSS"), oss.Callback(name))
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(-1)
	}
	fmt.Println("Use Callback Policy Success!")

	// Delete bucket
	err = client.DeleteBucket(bucketName)
	if err != nil {
		HandleError(err)
	}

	fmt.Println("BucketCallbackPolicySample completed")
}
