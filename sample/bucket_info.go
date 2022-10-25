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
	fmt.Println("Bucket Info AccessMonitor: ", res.BucketInfo.AccessMonitor)
	fmt.Println("Bucket Info Location: ", res.BucketInfo.Location)
	fmt.Println("Bucket Info CreationDate: ", res.BucketInfo.CreationDate)
	fmt.Println("Bucket Info ACL: ", res.BucketInfo.ACL)
	fmt.Println("Bucket Info Owner Id: ", res.BucketInfo.Owner.ID)
	fmt.Println("Bucket Info Owner DisplayName: ", res.BucketInfo.Owner.DisplayName)
	fmt.Println("Bucket Info StorageClass: ", res.BucketInfo.StorageClass)
	fmt.Println("Bucket Info RedundancyType: ", res.BucketInfo.RedundancyType)
	fmt.Println("Bucket Info ExtranetEndpoint: ", res.BucketInfo.ExtranetEndpoint)
	fmt.Println("Bucket Info IntranetEndpoint: ", res.BucketInfo.IntranetEndpoint)
	fmt.Println("Bucket Info CrossRegionReplication: ", res.BucketInfo.CrossRegionReplication)
	if res.BucketInfo.Versioning != "" {
		fmt.Println("Bucket Info Versioning: ", res.BucketInfo.Versioning)
	}
	if res.BucketInfo.SseRule.KMSDataEncryption != "" {
		fmt.Println("Bucket Info SseRule KMSDataEncryption: ", res.BucketInfo.SseRule.KMSDataEncryption)
	}
	if res.BucketInfo.SseRule.KMSMasterKeyID != "" {
		fmt.Println("Bucket Info SseRule KMSMasterKeyID: ", res.BucketInfo.SseRule.KMSMasterKeyID)
	}
	if res.BucketInfo.SseRule.SSEAlgorithm != "" {
		fmt.Println("Bucket Info SseRule SSEAlgorithm: ", res.BucketInfo.SseRule.SSEAlgorithm)
	}

}
