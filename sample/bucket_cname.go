package sample

import (
	"fmt"
	"os"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

// BucketCnameSample shows how to get,put,list or delete the bucket cname.
func BucketCnameSample() {
	// New client
	client, err := oss.New(endpoint, accessID, accessKey)
	if err != nil {
		HandleError(err)
	}

	// Create a bucket with default parameters
	err = client.CreateBucket(bucketName)
	if err != nil {
		HandleError(err)
	}

	// case1:Create the bucket cname token
	cname := "www.example.com"
	cbResult, err := client.CreateBucketCnameToken(bucketName, cname)
	if err != nil {
		HandleError(err)
	}
	fmt.Printf("Cname: %s\n", cbResult.Cname)
	fmt.Printf("Token: %s\n", cbResult.Token)
	fmt.Printf("ExpireTime: %s\n", cbResult.ExpireTime)

	// case2: Get the bucket cname token
	ctResult, err := client.GetBucketCnameToken(bucketName, cname)
	if err != nil {
		HandleError(err)
	}
	fmt.Printf("Cname: %s\n", ctResult.Cname)
	fmt.Printf("Token: %s\n", ctResult.Token)
	fmt.Printf("ExpireTime: %s\n", ctResult.ExpireTime)

	// case3: Add the bucket cname

	// case 3-1: Add bucket cname
	err = client.PutBucketCname(bucketName, cname)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(-1)
	}
	fmt.Println("Put Bucket Cname Success!")

	// case 3-2: Bind certificate
	var bindCnameConfig oss.PutBucketCname
	var bindCertificateConfig oss.CertificateConfiguration
	bindCnameConfig.Cname = "www.example.com"
	bindCertificate := "-----BEGIN CERTIFICATE-----MIIGeDCCBOCgAwIBAgIRAPj4FWpW5XN6kwgU7*******-----END CERTIFICATE-----"
	privateKey := "-----BEGIN CERTIFICATE-----MIIFBzCCA++gT2H2hT6Wb3nwxjpLIfXmSVcV*****-----END CERTIFICATE-----"
	bindCertificateConfig.CertId = "92******-cn-hangzhou"
	bindCertificateConfig.Certificate = bindCertificate
	bindCertificateConfig.PrivateKey = privateKey
	bindCertificateConfig.Force = true
	bindCnameConfig.CertificateConfiguration = &bindCertificateConfig
	err = client.PutBucketCnameWithCertificate(bucketName, bindCnameConfig)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(-1)
	}
	fmt.Println("Bind Certificate Success!")

	// case 3-3: Unbind certificate
	var putCnameConfig oss.PutBucketCname
	var CertificateConfig oss.CertificateConfiguration
	putCnameConfig.Cname = "www.example.com"
	CertificateConfig.DeleteCertificate = true
	putCnameConfig.CertificateConfiguration = &CertificateConfig

	err = client.PutBucketCnameWithCertificate(bucketName, putCnameConfig)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(-1)
	}
	fmt.Println("Unbind Certificate Success!")

	// case4: List the bucket cname
	cnResult, err := client.ListBucketCname(bucketName)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(-1)
	}
	var certificate oss.Certificate
	fmt.Printf("Bucket:%s\n", cnResult.Bucket)
	fmt.Printf("Owner:%s\n", cnResult.Owner)
	if len(cnResult.Cname) > 0 {
		for _, cnameInfo := range cnResult.Cname {
			fmt.Printf("Domain:%s\n", cnameInfo.Domain)
			fmt.Printf("LastModified:%s\n", cnameInfo.LastModified)
			fmt.Printf("Status:%s\n", cnameInfo.Status)
			if cnameInfo.Certificate != certificate {
				fmt.Printf("Type:%s\n", cnameInfo.Certificate.Type)
				fmt.Printf("CertId:%s\n", cnameInfo.Certificate.CertId)
				fmt.Printf("Status:%s\n", cnameInfo.Certificate.Status)
				fmt.Printf("CreationDate:%s\n", cnameInfo.Certificate.CreationDate)
				fmt.Printf("Fingerprint:%s\n", cnameInfo.Certificate.Fingerprint)
				fmt.Printf("ValidStartDate:%s\n", cnameInfo.Certificate.ValidStartDate)
				fmt.Printf("ValidEndDate:%s\n", cnameInfo.Certificate.ValidEndDate)
			}
		}
	}

	// case5: Delete the bucket cname
	err = client.DeleteBucketCname(bucketName, cname)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(-1)
	}
	fmt.Println("Delete Bucket Cname Success!")

	fmt.Println("BucketCnameSample completed")
}
