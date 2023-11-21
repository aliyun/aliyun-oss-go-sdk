package sample

import (
	"fmt"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"net/http"
)

// ReservedCapacitySample shows how to create,update,get,list the reserved capacity and list buckets under the reserved capacity.
func ReservedCapacitySample() {
	// New client
	client, err := oss.New(endpoint, accessID, accessKey)
	if err != nil {
		HandleError(err)
	}
	// Create the reserved capacity
	// case 1:
	var ctcConfig oss.CreateReservedCapacity
	ctcConfig.Name = "example-name"
	ctcConfig.ReservedCapacity = 10240
	ctcConfig.DataRedundancyType = string(oss.RedundancyLRS)
	err = client.CreateReservedCapacity(ctcConfig)
	if err != nil {
		HandleError(err)
	}

	// case 2:Get header information
	var respHeader http.Header
	err = client.CreateReservedCapacity(ctcConfig, oss.GetResponseHeader(&respHeader))
	if err != nil {
		HandleError(err)
	}
	fmt.Printf("Reserved Capacity Id:%s\n", respHeader.Get("X-Oss-Reserved-Capacity-Id"))

	fmt.Printf("Create Reserved Capacity Success")

	//Update Reserved Capacity
	id := respHeader.Get("X-Oss-Reserved-Capacity-Id")
	var urcConfig oss.UpdateReservedCapacity
	urcConfig.Status = "Enabled"
	urcConfig.ReservedCapacity = 10240
	urcConfig.AutoExpansionSize = 100
	urcConfig.AutoExpansionMaxSize = 20480
	err = client.UpdateReservedCapacity(id, urcConfig)
	if err != nil {
		HandleError(err)
	}
	fmt.Printf("Update Reserved Capacity Success")
	//Get Reserved Capacity By Id
	record, err := client.GetReservedCapacity(id)
	if err != nil {
		HandleError(err)
	}
	fmt.Printf("Reserved Capacity Record Instance Id:%s\n", record.InstanceId)
	fmt.Printf("Reserved Capacity Record Name:%s\n", record.Name)
	fmt.Printf("Reserved Capacity Record Owner Id:%s\n", record.Owner.ID)
	fmt.Printf("Reserved Capacity Record Owner Display Name:%s\n", record.Owner.DisplayName)
	fmt.Printf("Reserved Capacity Record Region:%s\n", record.Region)
	fmt.Printf("Reserved Capacity Record Status:%s\n", record.Status)
	fmt.Printf("Reserved Capacity Record Data Redundancy Type:%s\n", record.DataRedundancyType)
	fmt.Printf("Reserved Capacity Record Reserved Capacity:%d\n", record.ReservedCapacity)
	if record.AutoExpansionSize != 0 {
		fmt.Printf("Reserved Capacity Record Auto Expansion Size:%d\n", record.AutoExpansionSize)
	}
	if record.AutoExpansionMaxSize != 0 {
		fmt.Printf("Reserved Capacity Record Auto Expansion Max Size:%d\n", record.AutoExpansionMaxSize)
	}
	fmt.Printf("Reserved Capacity Record Reserved Create Time:%d\n", record.CreateTime)
	fmt.Printf("Reserved Capacity Record Reserved Last Modify Time:%d\n", record.LastModifyTime)
	if record.Status == "Enabled" {
		fmt.Printf("Reserved Capacity Record Reserved Enable Time:%d\n", record.EnableTime)
	}
	fmt.Printf("Get Reserved Capacity Success")
	//List Reserved Capacity
	rs, err := client.ListReservedCapacity()
	if err != nil {
		HandleError(err)
	}
	for _, record := range rs.ReservedCapacityRecord {
		fmt.Printf("Reserved Capacity Record Instance Id:%s\n", record.InstanceId)
		fmt.Printf("Reserved Capacity Record Name:%s\n", record.Name)
		fmt.Printf("Reserved Capacity Record Owner Id:%s\n", record.Owner.ID)
		fmt.Printf("Reserved Capacity Record Owner Display Name:%s\n", record.Owner.DisplayName)
		fmt.Printf("Reserved Capacity Record Region:%s\n", record.Region)
		fmt.Printf("Reserved Capacity Record Status:%s\n", record.Status)
		fmt.Printf("Reserved Capacity Record Data Redundancy Type:%s\n", record.DataRedundancyType)
		fmt.Printf("Reserved Capacity Record Reserved Capacity:%d\n", record.ReservedCapacity)
		if record.AutoExpansionSize != 0 {
			fmt.Printf("Reserved Capacity Record Auto Expansion Size:%d\n", record.AutoExpansionSize)
		}
		if record.AutoExpansionMaxSize != 0 {
			fmt.Printf("Reserved Capacity Record Auto Expansion Max Size:%d\n", record.AutoExpansionMaxSize)
		}
		fmt.Printf("Reserved Capacity Record Reserved Create Time:%d\n", record.CreateTime)
		fmt.Printf("Reserved Capacity Record Reserved Last Modify Time:%d\n", record.LastModifyTime)
		if record.Status == "Enabled" {
			fmt.Printf("Reserved Capacity Record Reserved Enable Time:%d\n", record.EnableTime)
		}
	}
	fmt.Printf("List Reserved Capacity Success")
	// Query the list of buckets created on the reserved space
	list, err := client.ListBucketWithReservedCapacity(id)
	if err != nil {
		HandleError(err)
	}
	fmt.Printf("Reserved Capacity Record Instance Id:%s\n", list.InstanceId)
	for _, bucket := range list.BucketList {
		fmt.Printf("Reserved Capacity Record Bucket Name:%s\n", bucket)
	}
	fmt.Printf("List Bucket Of Reserved Capacity Success")

	fmt.Println("ReservedCapacitySample completed")
}
