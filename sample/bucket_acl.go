package sample

import (
	"fmt"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

// BucketACLSample Set/Get Bucket ACL Sample
func BucketACLSample() {
	// New Client
	client, err := oss.New(endpoint, accessID, accessKey)
	if err != nil {
		HandleError(err)
	}

	// 使用默认参数创建bucket
	err = client.CreateBucket(bucketName)
	if err != nil {
		HandleError(err)
	}

	// 场景：设置Bucket ACL，可选权限有ACLPrivate、ACLPublicRead、ACLPublicReadWrite
	err = client.SetBucketACL(bucketName, oss.ACLPublicRead)
	if err != nil {
		HandleError(err)
	}

	// 查看Bucket ACL
	gbar, err := client.GetBucketACL(bucketName)
	if err != nil {
		HandleError(err)
	}
	fmt.Println("Bucket ACL:", gbar.ACL)

	// 删除bucket
	err = client.DeleteBucket(bucketName)
	if err != nil {
		HandleError(err)
	}

	fmt.Println("BucketACLSample completed")
}
