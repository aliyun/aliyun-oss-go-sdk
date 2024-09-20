package sample

import (
	"fmt"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

// BucketAccessPointSample  how to set, get or delete the bucket access point.
func BucketAccessPointSample() {
	// New client
	client, err := oss.New(endpoint, accessID, accessKey)
	if err != nil {
		HandleError(err)
	}

	// Create a bucket with default parameters
	err = client.CreateBucket(bucketName)
	if err != nil {
		HandleError(err)
	}

	// Case 1:Create Bucket Access Point
	apName := "access-point-name-1"
	var create oss.CreateBucketAccessPoint
	create.AccessPointName = apName
	create.NetworkOrigin = "internet"
	resp, err := client.CreateBucketAccessPoint(bucketName, create)
	if err != nil {
		HandleError(err)
	}
	fmt.Printf("Create Access Point Result Access Point Arn:%#v\n", resp.AccessPointArn)
	fmt.Printf("Create Access Point Result Alias:%#v\n", resp.Alias)

	// Case 2:Get Bucket Access Point
	result, err := client.GetBucketAccessPoint(bucketName, apName)
	if err != nil {
		HandleError(err)
	}
	fmt.Printf("Get Access Point Result Bucket:%s\n", result.Bucket)
	fmt.Printf("Get Access Point Result Access Point Name:%s\n", result.AccessPointName)
	fmt.Printf("Get Access Point Result AccountId:%s\n", result.AccountId)
	fmt.Printf("Get Access Point Result Network Origin:%s\n", result.NetworkOrigin)
	fmt.Printf("Get Access Point Result Vpc Id:%s\n", result.VpcId)
	fmt.Printf("Get Access Point Result Access Point Arn:%s\n", result.AccessPointArn)
	fmt.Printf("Get Access Point Result Creation Date:%s\n", result.CreationDate)
	fmt.Printf("Get Access Point Result Alias:%s\n", result.Alias)
	fmt.Printf("Get Access Point Result Status:%s\n", result.Status)
	fmt.Printf("Get Access Point Result Internal Endpoint :%s\n", result.Endpoints.InternalEndpoint)
	fmt.Printf("Get Access Point Result Public Endpoint :%s\n", result.Endpoints.PublicEndpoint)

	// Case 3:List Bucket Access Point
	token := ""
	for {
		ls, err := client.ListBucketAccessPoint(bucketName, oss.ContinuationToken(token), oss.MaxKeys(100))
		if err != nil {
			HandleError(err)
		}
		fmt.Printf("List Bucket Access Point Is Truncated:%t\n", ls.IsTruncated)
		fmt.Printf("List Bucket Access Point Next Continuation Token:%s\n", ls.NextContinuationToken)
		fmt.Printf("List Bucket Access Point AccountId:%s\n", ls.AccountId)
		for _, access := range ls.AccessPoints {
			fmt.Printf("Access Point Bucket:%s\n", access.Bucket)
			fmt.Printf("Access Point Name:%s\n", access.AccessPointName)
			fmt.Printf("Access Point Alias:%s\n", access.Alias)
			fmt.Printf("Access Point Network Origin:%s\n", access.NetworkOrigin)
			fmt.Printf("Access Point Vpc Id:%s\n", access.VpcId)
			fmt.Printf("Access Point Status:%s\n", access.Status)
		}
		if ls.IsTruncated {
			token = ls.NextContinuationToken
		} else {
			break
		}
	}

	// Case 4:Create Bucket Access Point Policy
	policy := `{
   "Version":"1",
   "Statement":[
   {
     "Action":[
       "oss:PutObject",
       "oss:GetObject"
    ],
    "Effect":"Deny",
    "Principal":["1234567890"],
    "Resource":[
       "acs:oss:*:1234567890:accesspoint/` + apName + `",
       "acs:oss:*:1234567890:accesspoint/` + apName + `/object/*"
     ]
   }
  ]
 }`
	err = client.PutAccessPointPolicy(bucketName, apName, policy)
	if err != nil {
		HandleError(err)
	}
	fmt.Println("Create Access Point Policy Success!")

	// Case 5:Get Bucket Access Point Policy
	policyInfo, err := client.GetAccessPointPolicy(bucketName, apName)
	if err != nil {
		HandleError(err)
	}
	fmt.Printf("Access Point Ploicy:%s\n", policyInfo)

	// Case 6:Delete Bucket Access Point Policy
	err = client.DeleteAccessPointPolicy(bucketName, apName)
	if err != nil {
		HandleError(err)
	}
	fmt.Println("Delete Access Point Policy Success!")

	// Case 7:Delete Bucket Access Point
	err = client.DeleteBucketAccessPoint(bucketName, apName)
	if err != nil {
		HandleError(err)
	}
	fmt.Println("Delete Bucket Replication Success!")
	fmt.Println("BucketReplicationSample completed")
}
