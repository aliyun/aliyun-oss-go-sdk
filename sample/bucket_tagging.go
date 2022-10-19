package sample

import (
	"fmt"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

// BucketTaggingSample shows how to set,get and  the bucket stat.
func BucketTaggingSample() {
	// New client
	client, err := oss.New(endpoint, accessID, accessKey)
	if err != nil {
		HandleError(err)
	}
	// Set bucket tagging
	tag1 := oss.Tag{
		Key:   "key1",
		Value: "value1",
	}
	tag2 := oss.Tag{
		Key:   "key2",
		Value: "value2",
	}
	tag3 := oss.Tag{
		Key:   "key3",
		Value: "value2",
	}
	tagging := oss.Tagging{
		Tags: []oss.Tag{tag1, tag2, tag3},
	}
	err = client.SetBucketTagging(bucketName, tagging)
	if err != nil {
		HandleError(err)
	}

	//Get bucket tagging
	ret, err := client.GetBucketTagging(bucketName)
	if err != nil {
		HandleError(err)
	}
	fmt.Println("Tag length: ", len(ret.Tags))
	for _, tag := range ret.Tags {
		fmt.Printf("Tag Key: %s\n", tag.Key)
		fmt.Printf("Tag Value: %s\n", tag.Value)
	}
	//Delete one tagging
	err = client.DeleteBucketTagging(bucketName, oss.AddParam("tagging", "key1"))
	if err != nil {
		HandleError(err)
	}
	// Delete many tagging
	err = client.DeleteBucketTagging(bucketName, oss.AddParam("tagging", "key1,key2"))
	if err != nil {
		HandleError(err)
	}
	// Delete all tagging
	err = client.DeleteBucketTagging(bucketName)
	if err != nil {
		HandleError(err)
	}
	fmt.Println("BucketTaggingSample completed")
}
