package sample

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

// SignURLSample signs URL sample
func SignURLSample() {
	// Create bucket
	bucket, err := GetTestBucket(bucketName)
	if err != nil {
		HandleError(err)
	}

	// Put object
	signedURL, err := bucket.SignURL(objectKey, oss.HTTPPut, 60)
	if err != nil {
		HandleError(err)
	}

	var val = "花间一壶酒，独酌无相亲。 举杯邀明月，对影成三人。"
	err = bucket.PutObjectWithURL(signedURL, strings.NewReader(val))
	if err != nil {
		HandleError(err)
	}

	// Put object with option
	options := []oss.Option{
		oss.Meta("myprop", "mypropval"),
		oss.ContentType("image/tiff"),
	}

	signedURL, err = bucket.SignURL(objectKey, oss.HTTPPut, 60, options...)
	if err != nil {
		HandleError(err)
	}

	err = bucket.PutObjectFromFileWithURL(signedURL, localFile, options...)
	if err != nil {
		HandleError(err)
	}

	// Get object
	signedURL, err = bucket.SignURL(objectKey, oss.HTTPGet, 60)
	if err != nil {
		HandleError(err)
	}

	body, err := bucket.GetObjectWithURL(signedURL)
	if err != nil {
		HandleError(err)
	}
	defer body.Close()

	// Read content
	data, err := ioutil.ReadAll(body)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(-1)
	}
	fmt.Println("data:", string(data))

	err = bucket.GetObjectToFileWithURL(signedURL, "mynewfile-1.jpg")
	if err != nil {
		HandleError(err)
	}

	// Delete the object and bucket
	err = DeleteTestBucketAndObject(bucketName)
	if err != nil {
		HandleError(err)
	}

	fmt.Println("SignURLSample completed")
}

func signCallbackUrl() (string, error) {
	bucket, err := GetTestBucket(bucketName)
	if err != nil {
		HandleError(err)
	}

	body := `{"object":${object},"bucket":${bucket},"size":${size},"mimeType":${mimeType},"my_var":${x:my_var}}`
	callbackMap := map[string]string{}
	callbackMap["callbackUrl"] = "http://oss-demo.aliyuncs.com:23450"
	callbackMap["callbackBody"] = body
	callbackMap["callbackBodyType"] = "application/json"

	callbackBuffer := bytes.NewBuffer([]byte{})
	callbackEncoder := json.NewEncoder(callbackBuffer)
	//do not encode '&' to "\u0026"
	callbackEncoder.SetEscapeHTML(false)
	err = callbackEncoder.Encode(callbackMap)
	if err != nil {
		return "", err
	}
	callbackVal := base64.StdEncoding.EncodeToString(callbackBuffer.Bytes())
	signURL, err := bucket.SignURL("foo/bar.png", oss.HTTPPut,
		3600,
		oss.ContentType("image/png"),
		oss.AddCallback(callbackVal),
		oss.AddCallbackVar(base64.StdEncoding.EncodeToString([]byte(`{"x:my_var": "1245"}`))),
	)
	if err != nil {
		return "", err
	}
	return signURL, err
}
