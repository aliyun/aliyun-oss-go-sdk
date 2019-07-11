package sample

import (
	"fmt"
	"strings"
	"io/ioutil"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

// BucketrRequestPaymentSample shows how to set, get the bucket request payment.
func BucketrRequestPaymentSample() {
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

	reqPayConf := oss.RequestPaymentConfiguration{
		Payer: string(oss.Requester),
	}

	// Case 1: Set bucket request payment.
	err = client.SetBucketRequestPayment(bucketName, reqPayConf)
	if err != nil {
		HandleError(err)
	}

	// Get bucket request payment configuration
	ret, err := client.GetBucketRequestPayment(bucketName)
	if err != nil {
		HandleError(err)
	}
	fmt.Println("Bucket request payer:", ret.Payer)

	if credentialUID == "" {
		fmt.Println("Please enter a credential User ID, if you want to test credential user.")
		clearData(client, bucketName)
		return
	}
	// Credential other User
	policyInfo := `
	{
		"Version":"1",
		"Statement":[
			{
				"Action":[
					"oss:*"
				],
				"Effect":"Allow",
				"Principal":["` + credentialUID + `"],
				"Resource":["acs:oss:*:*:` + bucketName + `", "acs:oss:*:*:` + bucketName + `/*"]
			}
		]
	}`

	err = client.SetBucketPolicy(bucketName, policyInfo)
	if err != nil {
		HandleError(err)
	}

	// New a Credential client
	creClient, err := oss.New(endpoint, credentialAccessID, credentialAccessKey)
	if err != nil {
		HandleError(err)
	}

	// Get credential bucket
	creBucket, err := creClient.Bucket(bucketName)
	if err != nil {
		HandleError(err)
	}

	// Put object by credential User
	key := "testCredentialObject"
	objectValue := "this is a test string."
	// Put object
	err = creBucket.PutObject(key, strings.NewReader(objectValue), oss.RequestPayer(oss.Requester))
	if err != nil {
		HandleError(err)
	}
	// Get object
	body, err := creBucket.GetObject(key, oss.RequestPayer(oss.Requester))
	if err != nil {
		HandleError(err)
	}
	defer body.Close()

	data, err := ioutil.ReadAll(body)
	if err != nil {
		HandleError(err)
	}
	fmt.Println(string(data))
	
	// Delete object
	err = creBucket.DeleteObject(key, oss.RequestPayer(oss.Requester))
	if err != nil {
		HandleError(err)
	}

	clearData(client, bucketName)
}

func clearData(client *oss.Client, bucketName string) {
	// Delete bucket
	err := client.DeleteBucket(bucketName)
	if err != nil {
		HandleError(err)
	}

	fmt.Println("BucketrRequestPaymentSample completed")
}
