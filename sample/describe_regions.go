package sample

import (
	"fmt"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

//DescribeRegionsSample shows how to get or list describe regions
func DescribeRegionsSample() {
	// Create archive bucket
	client, err := oss.New(endpoint, accessID, accessKey)
	if err != nil {
		HandleError(err)
	}

	// Get describe regions
	regionEndpoint := "oss-cn-hangzhou"
	list, err := client.DescribeRegions(oss.AddParam("regions", regionEndpoint))
	if err != nil {
		HandleError(err)
	}
	for _, region := range list.Regions {
		fmt.Printf("Region:%s\n", region.Region)
		fmt.Printf("Region Internet Endpoint:%s\n", region.InternetEndpoint)
		fmt.Printf("Region Internal Endpoint:%s\n", region.InternalEndpoint)
		fmt.Printf("Region Accelerate Endpoint:%s\n", region.AccelerateEndpoint)
	}
	fmt.Println("Get Describe Regions Success")

	// List describe regions

	list, err = client.DescribeRegions()
	if err != nil {
		HandleError(err)
	}
	for _, region := range list.Regions {
		fmt.Printf("Region:%s\n", region.Region)
		fmt.Printf("Region Internet Endpoint:%s\n", region.InternetEndpoint)
		fmt.Printf("Region Internal Endpoint:%s\n", region.InternalEndpoint)
		fmt.Printf("Region Accelerate Endpoint:%s\n", region.AccelerateEndpoint)
	}
	fmt.Println("List Describe Regions Success")

	fmt.Println("DescribeRegionsSample completed")
}
