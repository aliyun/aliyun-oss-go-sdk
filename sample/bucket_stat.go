package sample

import (
	"fmt"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

// BucketStatSample shows how to get the bucket stat.
func BucketStatSample() {
	// New client
	client, err := oss.New(endpoint, accessID, accessKey)
	if err != nil {
		HandleError(err)
	}
	// Get bucket stat
	stat, err := client.GetBucketStat(bucketName)
	if err != nil {
		HandleError(err)
	}
	fmt.Println("Bucket Stat Storage:", stat.Storage)
	fmt.Println("Bucket Stat Object Count:", stat.ObjectCount)
	fmt.Println("Bucket Stat Multipart Upload Count:", stat.MultipartUploadCount)
	fmt.Println("Bucket Stat Live Channel Count:", stat.LiveChannelCount)
	fmt.Println("Bucket Stat Last Modified Time:", stat.LastModifiedTime)
	fmt.Println("Bucket Stat Standard Storage:", stat.StandardStorage)
	fmt.Println("Bucket Stat Standard Object Count:", stat.StandardObjectCount)
	fmt.Println("Bucket Stat Infrequent Access Storage:", stat.InfrequentAccessStorage)
	fmt.Println("Bucket Stat Infrequent Access Real Storage:", stat.InfrequentAccessRealStorage)
	fmt.Println("Bucket Stat Infrequent Access Object Count:", stat.InfrequentAccessObjectCount)
	fmt.Println("Bucket Stat Archive Storage:", stat.ArchiveStorage)
	fmt.Println("Bucket Stat Archive Real Storage:", stat.ArchiveRealStorage)
	fmt.Println("Bucket Stat Archive Object Count:", stat.ArchiveObjectCount)
	fmt.Println("Bucket Stat Cold Archive Storage:", stat.ColdArchiveStorage)
	fmt.Println("Bucket Stat Cold Archive Real Storage:", stat.ColdArchiveRealStorage)
	fmt.Println("Bucket Stat Cold Archive Object Count:", stat.ColdArchiveObjectCount)
	fmt.Println("BucketStatSample completed")
}
