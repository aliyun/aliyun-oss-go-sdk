package sample

import (
	"fmt"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

// BucketHttpsConfigSample  how to set, get the bucket https config.
func BucketHttpsConfigSample() {
	// New client
	client, err := oss.New(endpoint, accessID, accessKey)
	if err != nil {
		HandleError(err)
	}
	// put bucket https config
	config := oss.PutBucketHttpsConfig{
		TLS: oss.HttpsConfigTLS{
			Enable:     true,
			TLSVersion: []string{"TLSv1.2", "TLSv1.3"},
		},
	}
	err = client.PutBucketHttpsConfig(bucketName, config)
	if err != nil {
		HandleError(err)
	}
	fmt.Println("Put Bucket Https Config Success!")

	// get bucket https config
	result, err := client.GetBucketHttpsConfig(bucketName)
	if err != nil {
		HandleError(err)
	}
	fmt.Printf("TLS Enable:%t\n", result.TLS.Enable)
	for _, tls := range result.TLS.TLSVersion {
		fmt.Printf("TLS Version:%s\n", tls)
	}

	fmt.Println("BucketHttpsConfigSample completed")

}
