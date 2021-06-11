package sample

import (
	"fmt"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

// BucketInventorySample shows how to set, get, list and delete the bucket inventory configuration
func BucketInventorySample() {
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

	// the inventory configuration,not use any encryption
	bl := true
	invConfig := oss.InventoryConfiguration{
		Id:        "report1",
		IsEnabled: &bl,
		Prefix:    "filterPrefix/",
		OSSBucketDestination: oss.OSSBucketDestination{
			Format:    "CSV",
			AccountId: accountID,
			RoleArn:   stsARN,
			Bucket:    "acs:oss:::" + bucketName,
			Prefix:    "prefix1",
		},
		Frequency:              "Daily",
		IncludedObjectVersions: "All",
		OptionalFields: oss.OptionalFields{
			Field: []string{
				"Size", "LastModifiedDate", "ETag", "StorageClass", "IsMultipartUploaded", "EncryptionStatus",
			},
		},
	}

	// case 1: Set inventory
	err = client.SetBucketInventory(bucketName, invConfig)
	if err != nil {
		HandleError(err)
	}

	// case 2: Get Bucket inventory
	ret, err := client.GetBucketInventory(bucketName, invConfig.Id)
	if err != nil {
		HandleError(err)
	}
	fmt.Println("Bucket inventory:", ret)

	// case 3: List Bucket inventory
	invConfig2 := oss.InventoryConfiguration{
		Id:        "report2",
		IsEnabled: &bl,
		Prefix:    "filterPrefix/",
		OSSBucketDestination: oss.OSSBucketDestination{
			Format:    "CSV",
			AccountId: accountID,
			RoleArn:   stsARN,
			Bucket:    "acs:oss:::" + bucketName,
			Prefix:    "prefix1",
		},
		Frequency:              "Daily",
		IncludedObjectVersions: "All",
		OptionalFields: oss.OptionalFields{
			Field: []string{
				"Size", "LastModifiedDate", "ETag", "StorageClass", "IsMultipartUploaded", "EncryptionStatus",
			},
		},
	}
	err = client.SetBucketInventory(bucketName, invConfig2)
	if err != nil {
		HandleError(err)
	}
	NextContinuationToken := ""
	listInvConf, err := client.ListBucketInventory(bucketName, NextContinuationToken)
	if err != nil {
		HandleError(err)
	}
	fmt.Println("Bucket inventory list:", listInvConf)

	// case 4: Delete Bucket inventory
	err = client.DeleteBucketInventory(bucketName, invConfig.Id)
	if err != nil {
		HandleError(err)
	}

	// Delete bucket
	err = client.DeleteBucket(bucketName)
	if err != nil {
		HandleError(err)
	}

	fmt.Println("BucketInventorySample completed")
}
