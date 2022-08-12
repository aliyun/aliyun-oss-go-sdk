package sample

import (
	"fmt"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"os"
)

func BucketMetaQuerySample() {
	// New client
	client, err := oss.New(endpoint, accessID, accessKey)
	if err != nil {
		HandleError(err)
	}
	// Open data indexing
	err = client.OpenMetaQuery(bucketName)
	if err != nil {
		HandleError(err)
	}

	// Get meta data indexing query
	result, err := client.GetMetaQueryStatus(bucketName)
	if err != nil {
		HandleError(err)
	}
	fmt.Printf("State:%s\n", result.State)
	fmt.Printf("Phase%s\n", result.Phase)
	fmt.Printf("CreateTime:%s\n", result.CreateTime)
	fmt.Printf("UpdateTime:%s\n", result.UpdateTime)

	// Do data indexing query
	// 1. Simple query
	query := oss.MetaQuery{
		NextToken:  "",
		MaxResults: 10,
		Query:      `{"Field": "Size","Value": "30","Operation": "gt"}`,
		Sort:       "Size",
		Order:      "asc",
	}
	queryResult, err := client.DoMetaQuery(bucketName, query)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(-1)
	}
	fmt.Printf("NextToken:%s\n", queryResult.NextToken)
	for _, file := range queryResult.Files {
		fmt.Printf("File name: %s\n", file.Filename)
		fmt.Printf("size: %d\n", file.Size)
		fmt.Printf("File Modified Time:%s\n", file.FileModifiedTime)
		fmt.Printf("Oss Object Type:%s\n", file.OssObjectType)
		fmt.Printf("Oss Storage Class:%s\n", file.OssStorageClass)
		fmt.Printf("Object ACL:%s\n", file.ObjectACL)
		fmt.Printf("ETag:%s\n", file.ETag)
		fmt.Printf("Oss CRC64:%s\n", file.OssCRC64)
		fmt.Printf("Oss Tagging Count:%d\n", file.OssTaggingCount)
		for _, tagging := range file.OssTagging {
			fmt.Printf("Oss Tagging Key:%s\n", tagging.Key)
			fmt.Printf("Oss Tagging Value:%s\n", tagging.Value)
		}
		for _, userMeta := range file.OssUserMeta {
			fmt.Printf("Oss User Meta Key:%s\n", userMeta.Key)
			fmt.Printf("Oss User Meta Key Value:%s\n", userMeta.Value)
		}
	}

	//2. Aggregate query

	query = oss.MetaQuery{
		NextToken:  "",
		MaxResults: 100,
		Query:      `{"Field": "Size","Value": "30","Operation": "gt"}`,
		Sort:       "Size",
		Order:      "asc",
		Aggregations: []oss.MetaQueryAggregationRequest{
			{
				Field:     "Size",
				Operation: "max",
			},
			{
				Field:     "Size",
				Operation: "sum",
			},
			{
				Field:     "Size",
				Operation: "group",
			},
		},
	}
	queryResult, err = client.DoMetaQuery(bucketName, query)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(-1)
	}
	fmt.Printf("Next Token:%s\n", queryResult.NextToken)
	for _, aggregation := range queryResult.Aggregations {
		fmt.Printf("Aggregation Field:%s\n", aggregation.Field)
		fmt.Printf("Aggregation Operation:%s\n", aggregation.Operation)
		fmt.Printf("Aggregation Value:%d\n", aggregation.Value)
		for _, group := range aggregation.Groups {
			fmt.Printf("Group Value:%s\n", group.Value)
			fmt.Printf("Group Count:%d\n", group.Count)
		}
	}

	// 3.Query all
	query = oss.MetaQuery{
		NextToken:  "",
		MaxResults: 10,
		Query:      `{"Field": "Size","Value": "30","Operation": "gt"}`,
		Sort:       "Size",
		Order:      "asc",
	}
	for {
		queryResult, err = client.DoMetaQuery(bucketName, query)
		if err != nil {
			fmt.Println("Error:", err)
			os.Exit(-1)
		}
		fmt.Printf("NextToken:%s\n", queryResult.NextToken)
		for _, file := range queryResult.Files {
			fmt.Printf("File name: %s\n", file.Filename)
			fmt.Printf("size: %d\n", file.Size)
			fmt.Printf("File Modified Time:%s\n", file.FileModifiedTime)
			fmt.Printf("Oss Object Type:%s\n", file.OssObjectType)
			fmt.Printf("Oss Storage Class:%s\n", file.OssStorageClass)
			fmt.Printf("Object ACL:%s\n", file.ObjectACL)
			fmt.Printf("ETag:%s\n", file.ETag)
			fmt.Printf("Oss CRC64:%s\n", file.OssCRC64)
			fmt.Printf("Oss Tagging Count:%d\n", file.OssTaggingCount)
			for _, tagging := range file.OssTagging {
				fmt.Printf("Oss Tagging Key:%s\n", tagging.Key)
				fmt.Printf("Oss Tagging Value:%s\n", tagging.Value)
			}
			for _, userMeta := range file.OssUserMeta {
				fmt.Printf("Oss User Meta Key:%s\n", userMeta.Key)
				fmt.Printf("Oss User Meta Key Value:%s\n", userMeta.Value)
			}
		}
		if queryResult.NextToken != "" {
			query.NextToken = queryResult.NextToken
		} else {
			break
		}
	}

	//4.Do Meta query use xml
	xml := `<?xml version="1.0" encoding="UTF-8"?>
<MetaQuery>
  <NextToken></NextToken>
  <MaxResults>5</MaxResults>
  <Query>{"Field": "Size","Value": "1048576","Operation": "gt"}</Query>
  <Sort>Size</Sort>
  <Order>asc</Order>
</MetaQuery>`
	queryResult, err = client.DoMetaQueryXml(bucketName, xml)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(-1)
	}
	fmt.Printf("NextToken:%s\n", queryResult.NextToken)
	for _, file := range queryResult.Files {
		fmt.Printf("File name: %s\n", file.Filename)
		fmt.Printf("size: %d\n", file.Size)
		fmt.Printf("File Modified Time:%s\n", file.FileModifiedTime)
		fmt.Printf("Oss Object Type:%s\n", file.OssObjectType)
		fmt.Printf("Oss Storage Class:%s\n", file.OssStorageClass)
		fmt.Printf("Object ACL:%s\n", file.ObjectACL)
		fmt.Printf("ETag:%s\n", file.ETag)
		fmt.Printf("Oss CRC64:%s\n", file.OssCRC64)
		fmt.Printf("Oss Tagging Count:%d\n", file.OssTaggingCount)
		for _, tagging := range file.OssTagging {
			fmt.Printf("Oss Tagging Key:%s\n", tagging.Key)
			fmt.Printf("Oss Tagging Value:%s\n", tagging.Value)
		}
		for _, userMeta := range file.OssUserMeta {
			fmt.Printf("Oss User Meta Key:%s\n", userMeta.Key)
			fmt.Printf("Oss User Meta Key Value:%s\n", userMeta.Value)
		}
	}

	// Close meta data indexing
	err = client.CloseMetaQuery(bucketName)
	if err != nil {
		HandleError(err)
	}

	fmt.Println("BucketDataIndexingSample completed")
}
