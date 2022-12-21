package sample

import (
	"fmt"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

// BucketResourceGroupSample shows how to set and get the bucket's resource group.
func BucketResourceGroupSample() {
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

	// Get bucket's resource group.
	result, err := client.GetBucketResourceGroup(bucketName)
	if err != nil {
		HandleError(err)
	}

	fmt.Printf("Resource Group Id:%s\n", result.ResourceGroupId)

	// Set bucket's resource group.
	resourceGroup := oss.PutBucketResourceGroup{
		ResourceGroupId: "rg-aek27tc********",
	}
	err = client.PutBucketResourceGroup(bucketName, resourceGroup)
	if err != nil {
		HandleError(err)
	}

	fmt.Println("Bucket Resource Group Set Success!")

	fmt.Println("BucketResourceGroupSample completed")
}
