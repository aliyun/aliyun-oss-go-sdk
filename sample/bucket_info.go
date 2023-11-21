package sample

import (
	"fmt"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

// BucketInfoSample shows how to get the bucket info.
func BucketInfoSample() {
	// New client
	client, err := oss.New(endpoint, accessID, accessKey)
	if err != nil {
		HandleError(err)
	}
	// Get bucket info
	res, err := client.GetBucketInfo(bucketName)
	if err != nil {
		HandleError(err)
	}
	fmt.Println("Bucket Info Name: ", res.BucketInfo.Name)
	fmt.Println("Bucket Info Access Monitor: ", res.BucketInfo.AccessMonitor)
	fmt.Println("Bucket Info Location: ", res.BucketInfo.Location)
	fmt.Println("Bucket Info Creation Date: ", res.BucketInfo.CreationDate)
	fmt.Println("Bucket Info ACL: ", res.BucketInfo.ACL)
	fmt.Println("Bucket Info Owner Id: ", res.BucketInfo.Owner.ID)
	fmt.Println("Bucket Info Owner Display Name: ", res.BucketInfo.Owner.DisplayName)
	fmt.Println("Bucket Info Storage Class: ", res.BucketInfo.StorageClass)
	fmt.Println("Bucket Info Redundancy Type: ", res.BucketInfo.RedundancyType)
	if res.BucketInfo.ReservedCapacityInstanceId != "" {
		fmt.Println("Bucket Info Reserved Capacity Instance Id: ", res.BucketInfo.ReservedCapacityInstanceId)
	}
	fmt.Println("Bucket Info Extranet Endpoint: ", res.BucketInfo.ExtranetEndpoint)
	fmt.Println("Bucket Info Intranet Endpoint: ", res.BucketInfo.IntranetEndpoint)
	fmt.Println("Bucket Info Cross Region Replication: ", res.BucketInfo.CrossRegionReplication)
	if res.BucketInfo.Versioning != "" {
		fmt.Println("Bucket Info Versioning: ", res.BucketInfo.Versioning)
	}
	if res.BucketInfo.SseRule.KMSDataEncryption != "" {
		fmt.Println("Bucket Info SseRule KMS Data Encryption: ", res.BucketInfo.SseRule.KMSDataEncryption)
	}
	if res.BucketInfo.SseRule.KMSMasterKeyID != "" {
		fmt.Println("Bucket Info SseRule KMS Master Key ID: ", res.BucketInfo.SseRule.KMSMasterKeyID)
	}
	if res.BucketInfo.SseRule.SSEAlgorithm != "" {
		fmt.Println("Bucket Info SseRule SSE Algorithm: ", res.BucketInfo.SseRule.SSEAlgorithm)
	}

}
