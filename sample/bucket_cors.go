package sample

import (
	"fmt"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

// BucketCORSSample shows how to get or set the bucket CORS.
func BucketCORSSample() {
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

	rule1 := oss.CORSRule{
		AllowedOrigin: []string{"*"},
		AllowedMethod: []string{"PUT", "GET", "POST"},
		AllowedHeader: []string{},
		ExposeHeader:  []string{},
		MaxAgeSeconds: 100,
	}
	rule2 := oss.CORSRule{
		AllowedOrigin: []string{"http://www.a.com", "http://www.b.com"},
		AllowedMethod: []string{"GET"},
		AllowedHeader: []string{"Authorization"},
		ExposeHeader:  []string{"x-oss-test", "x-oss-test1"},
		MaxAgeSeconds: 100,
	}

	// Case 1: Set the bucket CORS rules
	err = client.SetBucketCORS(bucketName, []oss.CORSRule{rule1})
	if err != nil {
		HandleError(err)
	}

	// Case 2: Set the bucket CORS rules. if CORS rules exist, they will be overwritten.
	err = client.SetBucketCORS(bucketName, []oss.CORSRule{rule1, rule2})
	if err != nil {
		HandleError(err)
	}

	// Case 3: Set the bucket CORS rules. if CORS rules exist, they will be overwritten.
	isTrue := true
	put := oss.PutBucketCORS{}
	put.CORSRules = []oss.CORSRule{rule1, rule2}
	put.ResponseVary = &isTrue
	err = client.SetBucketCORSV2(bucketName, put)
	if err != nil {
		HandleError(err)
	}

	// Get the bucket's CORS
	corsRes, err := client.GetBucketCORS(bucketName)
	if err != nil {
		HandleError(err)
	}
	for _, rule := range corsRes.CORSRules {
		fmt.Printf("Cors Rules Allowed Origin:%s\n", rule.AllowedOrigin)
		fmt.Printf("Cors Rules Allowed Method:%s\n", rule.AllowedMethod)
		fmt.Printf("Cors Rules Allowed Header:%s\n", rule.AllowedHeader)
		fmt.Printf("Cors Rules Expose Header:%s\n", rule.ExposeHeader)
		fmt.Printf("Cors Rules Max Age Seconds:%d\n", rule.MaxAgeSeconds)
	}
	if corsRes.ResponseVary != nil {
		fmt.Printf("Cors Rules Response Vary:%t\n", *corsRes.ResponseVary)
	}

	// Delete bucket's CORS
	err = client.DeleteBucketCORS(bucketName)
	if err != nil {
		HandleError(err)
	}

	// Delete bucket
	err = client.DeleteBucket(bucketName)
	if err != nil {
		HandleError(err)
	}

	fmt.Println("Bucket CORS Sample completed")
}
