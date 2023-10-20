package sample

import (
	"fmt"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

// BucketResponseHeaderSample shows how to set, get and delete the bucket's response header.
func BucketResponseHeaderSample() {
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

	// Set bucket's response header.
	reqHeader := oss.PutBucketResponseHeader{
		Rule: []oss.ResponseHeaderRule{
			{
				Name: "name1",
				Filters: oss.ResponseHeaderRuleFilters{
					[]string{
						"Put", "GetObject",
					},
				},
				HideHeaders: oss.ResponseHeaderRuleHeaders{
					[]string{
						"Last-Modified",
					},
				},
			},
			{
				Name: "name2",
				Filters: oss.ResponseHeaderRuleFilters{
					[]string{
						"*",
					},
				},
				HideHeaders: oss.ResponseHeaderRuleHeaders{
					[]string{
						"Last-Modified",
					},
				},
			},
		},
	}
	err = client.PutBucketResponseHeader(bucketName, reqHeader)
	if err != nil {
		HandleError(err)
	}

	fmt.Println("Bucket Response Header Set Success!")

	// Get bucket's response header.
	header, err := client.GetBucketResponseHeader(bucketName)
	if err != nil {
		HandleError(err)
	}
	for _, rule := range header.Rule {
		fmt.Printf("Rule Name:%#v\n", rule.Name)
		if len(rule.Filters.Operation) > 0 {
			for _, Operation := range rule.Filters.Operation {
				fmt.Printf("Rule Filter Operation:%s\n", Operation)
			}
		}
		if len(rule.HideHeaders.Header) > 0 {
			for _, head := range rule.HideHeaders.Header {
				fmt.Printf("Rule Hide Headers Header:%s\n", head)
			}
		}
	}

	// Delete bucket's response header.
	err = client.DeleteBucketResponseHeader(bucketName)
	if err != nil {
		HandleError(err)
	}
	fmt.Println("Bucket Response Header Delete Success!")

	fmt.Println("BucketResponseHeaderSample completed")
}
