package sample

import (
	"fmt"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

// BucketQoSInfoSample shows how to set, get and delete the bucket QoS configuration
func BucketQoSInfoSample() {
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
	// Initial QoS Configuration
	five := 5
	four := 4
	totalQps := 200
	qosConf := oss.BucketQoSConfiguration{
		TotalUploadBandwidth:      &five,
		IntranetUploadBandwidth:   &four,
		ExtranetUploadBandwidth:   &four,
		TotalDownloadBandwidth:    &four,
		IntranetDownloadBandwidth: &four,
		ExtranetDownloadBandwidth: &four,
		TotalQPS:                  &totalQps,
		IntranetQPS:               &totalQps,
		ExtranetQPS:               &totalQps,
	}

	// Set Qos Info
	err = client.SetBucketQoSInfo(bucketName, qosConf)
	if err != nil {
		HandleError(err)
	}

	// Get Qos Info
	ret, err := client.GetBucketQosInfo(bucketName)
	if err != nil {
		HandleError(err)
	}
	fmt.Printf("Bucket QoSInfo\n  TotalUploadBandwidth: %d\n  IntranetUploadBandwidth: %d\n  ExtranetUploadBandwidth: %d\n",
		*ret.TotalUploadBandwidth, *ret.IntranetUploadBandwidth, *ret.ExtranetUploadBandwidth)

	// Delete QosInfo
	err = client.DeleteBucketQosInfo(bucketName)
	if err != nil {
		HandleError(err)
	}

	// Delete bucket
	err = client.DeleteBucket(bucketName)
	if err != nil {
		HandleError(err)
	}

	fmt.Println("BucketPolicySample completed")
}
