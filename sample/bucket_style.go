package sample

import (
	"fmt"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

// BucketStyleSample shows how to set,get list and delete the bucket's style.
func BucketStyleSample() {
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

	// Get bucket's style.
	styleName := "image-style"
	style, err := client.GetBucketStyle(bucketName, styleName)
	if err != nil {
		HandleError(err)
	}
	fmt.Printf("Style Name:%s\n", style.Name)
	fmt.Printf("Style Name:%s\n", style.Content)
	fmt.Printf("Style Create Time:%s\n", style.CreateTime)
	fmt.Printf("Style Last Modify Time:%s\n", style.LastModifyTime)

	// Set bucket's style.
	styleContent := "image/resize,p_50"
	err = client.PutBucketStyle(bucketName, styleName, styleContent)
	if err != nil {
		HandleError(err)
	}
	fmt.Println("Put Bucket Style Success!")

	// List bucket's style
	list, err := client.ListBucketStyle(bucketName)
	if err != nil {
		HandleError(err)
	}

	for _, styleInfo := range list.Style {
		fmt.Printf("Style Name:%s\n", styleInfo.Name)
		fmt.Printf("Style Name:%s\n", styleInfo.Content)
		fmt.Printf("Style Create Time:%s\n", styleInfo.CreateTime)
		fmt.Printf("Style Last Modify Time:%s\n", styleInfo.LastModifyTime)
	}

	// Delete bucket's style
	err = client.DeleteBucketStyle(bucketName, styleName)
	if err != nil {
		HandleError(err)
	}
	fmt.Println("Bucket Style Delete Success!")

	fmt.Println("BucketStyleSample completed")
}
