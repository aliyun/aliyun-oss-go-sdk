package sample

import (
	"fmt"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

// BucketAccessMonitorSample  how to set, get the bucket access monitor.
func BucketAccessMonitorSample() {
	// New client
	client, err := oss.New(endpoint, accessID, accessKey)
	if err != nil {
		HandleError(err)
	}
	access := oss.PutBucketAccessMonitor{
		Status: "Enabled",
	}
	// put bucket access monitor
	err = client.PutBucketAccessMonitor(bucketName, access)
	if err != nil {
		HandleError(err)
	}
	fmt.Println("put bucket access monitor success!")

	// put bucket access monitor in xml format

	xml := `<?xml version="1.0" encoding="UTF-8"?>
<AccessMonitorConfiguration>
  <Status>Enabled</Status>
</AccessMonitorConfiguration>
`
	err = client.PutBucketAccessMonitorXml(bucketName, xml)
	if err != nil {
		HandleError(err)
	}
	fmt.Println("put bucket access monitor in xml format success!")

	// get bucket access monitor
	result, err := client.GetBucketAccessMonitor(bucketName)
	if err != nil {
		HandleError(err)
	}
	fmt.Printf("bucket access monitor config is:%s\n", result.Status)

	// get bucket access monitor in xml format
	xmlData, err := client.GetBucketAccessMonitorXml(bucketName)
	if err != nil {
		HandleError(err)
	}
	fmt.Printf("bucket access monitor config is:%s\n", xmlData)

	fmt.Println("BucketAccessMonitorSample completed")

}
