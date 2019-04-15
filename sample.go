// main of samples

package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/aliyun/aliyun-oss-go-sdk/sample"
)

// sampleMap contains all samples
var sampleMap = map[string]interface{}{
	"CreateBucketSample":    sample.CreateBucketSample,
	"NewBucketSample":       sample.NewBucketSample,
	"ListBucketsSample":     sample.ListBucketsSample,
	"BucketACLSample":       sample.BucketACLSample,
	"BucketLifecycleSample": sample.BucketLifecycleSample,
	"BucketRefererSample":   sample.BucketRefererSample,
	"BucketLoggingSample":   sample.BucketLoggingSample,
	"BucketCORSSample":      sample.BucketCORSSample,
	"ObjectACLSample":       sample.ObjectACLSample,
	"ObjectMetaSample":      sample.ObjectMetaSample,
	"ListObjectsSample":     sample.ListObjectsSample,
	"DeleteObjectSample":    sample.DeleteObjectSample,
	"AppendObjectSample":    sample.AppendObjectSample,
	"CopyObjectSample":      sample.CopyObjectSample,
	"PutObjectSample":       sample.PutObjectSample,
	"GetObjectSample":       sample.GetObjectSample,
	"CnameSample":           sample.CnameSample,
	"SignURLSample":         sample.SignURLSample,
	"ArchiveSample":         sample.ArchiveSample,
}

func main() {
	var name string
	flag.StringVar(&name, "name", "", "Waiting for a sample of execution")
	flag.Parse()

	if len(name) <= 0 {
		fmt.Println("please enter your sample's name. like '-name CreateBucketSample'")
		os.Exit(-1)
	} else {
		if sampleMap[name] == nil {
			fmt.Println("the " + name + "is not exist.")
			os.Exit(-1)
		}
		sampleMap[name].(func())()
	}

	// sample.CreateBucketSample()
	// sample.NewBucketSample()
	// sample.ListBucketsSample()
	// sample.BucketACLSample()
	// sample.BucketLifecycleSample()
	// sample.BucketRefererSample()
	// sample.BucketLoggingSample()
	// sample.BucketCORSSample()

	// sample.ObjectACLSample()
	// sample.ObjectMetaSample()
	// sample.ListObjectsSample()
	// sample.DeleteObjectSample()
	// sample.AppendObjectSample()
	// sample.CopyObjectSample()
	// sample.PutObjectSample()
	// sample.GetObjectSample()

	// sample.CnameSample()
	// sample.SignURLSample()

	// sample.ArchiveSample()

	// fmt.Println("All samples completed")
}
