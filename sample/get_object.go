package sample

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

// GetObjectSample shows the streaming download, range download and resumable download.
func GetObjectSample() {
	// Create bucket
	bucket, err := GetTestBucket(bucketName)
	if err != nil {
		HandleError(err)
	}

	// Upload the object
	err = bucket.PutObjectFromFile(objectKey, localFile)
	if err != nil {
		HandleError(err)
	}

	// Case 1: Download the object into ReadCloser(). The body needs to be closed
	body, err := bucket.GetObject(objectKey)
	if err != nil {
		HandleError(err)
	}

	data, err := ioutil.ReadAll(body)
	body.Close()
	if err != nil {
		HandleError(err)
	}
	fmt.Println("size of data is: ", len(data))

	// Case 2: Download in the range of object.
	body, err = bucket.GetObject(objectKey, oss.Range(15, 19))
	if err != nil {
		HandleError(err)
	}
	data, err = ioutil.ReadAll(body)
	body.Close()
	fmt.Println("the range of data is: ", string(data))

	// Case 3: Download the object to byte array. This is for small object.
	buf := new(bytes.Buffer)
	body, err = bucket.GetObject(objectKey)
	if err != nil {
		HandleError(err)
	}
	io.Copy(buf, body)
	body.Close()

	// Case 4: Download the object to local file. The file handle needs to be specified
	fd, err := os.OpenFile("mynewfile-1.jpg", os.O_WRONLY|os.O_CREATE, 0660)
	if err != nil {
		HandleError(err)
	}
	defer fd.Close()

	body, err = bucket.GetObject(objectKey)
	if err != nil {
		HandleError(err)
	}
	io.Copy(fd, body)
	body.Close()

	// Case 5: Download the object to local file with file name specified
	err = bucket.GetObjectToFile(objectKey, "mynewfile-2.jpg")
	if err != nil {
		HandleError(err)
	}

	// Case 6: Get the object with contraints. When contraints are met, download the file. Otherwise return precondition error
	// last modified time constraint is met, download the file
	body, err = bucket.GetObject(objectKey, oss.IfModifiedSince(pastDate))
	if err != nil {
		HandleError(err)
	}
	body.Close()

	// Last modified time contraint is not met, do not download the file
	_, err = bucket.GetObject(objectKey, oss.IfUnmodifiedSince(pastDate))
	if err == nil {
		HandleError(fmt.Errorf("This result is not the expected result"))
	}
	// body.Close()

	meta, err := bucket.GetObjectDetailedMeta(objectKey)
	if err != nil {
		HandleError(err)
	}
	etag := meta.Get(oss.HTTPHeaderEtag)
	// Check the content, etag contraint is met, download the file
	body, err = bucket.GetObject(objectKey, oss.IfMatch(etag))
	if err != nil {
		HandleError(err)
	}
	body.Close()

	// Check the content, etag contraint is not met, do not download the file
	_, err = bucket.GetObject(objectKey, oss.IfNoneMatch(etag))
	if err == nil {
		HandleError(fmt.Errorf("This result is not the expected result"))
	}
	// body.Close()

	// Case 7: Big file's multipart download, concurrent and resumable download is supported.
	// multipart download with part size 100KB. By default single coroutine is used and no checkpoint
	err = bucket.DownloadFile(objectKey, "mynewfile-3.jpg", 100*1024)
	if err != nil {
		HandleError(err)
	}

	// Part size is 100K and 3 coroutines are used
	err = bucket.DownloadFile(objectKey, "mynewfile-3.jpg", 100*1024, oss.Routines(3))
	if err != nil {
		HandleError(err)
	}

	// Part size is 100K and 3 coroutines with checkpoint
	err = bucket.DownloadFile(objectKey, "mynewfile-3.jpg", 100*1024, oss.Routines(3), oss.Checkpoint(true, ""))
	if err != nil {
		HandleError(err)
	}

	// Specify the checkpoint file path to record which parts have been downloaded.
	// This file path can be specified by the 2nd parameter of Checkpoint, it will be the download directory if the file path is empty.
	err = bucket.DownloadFile(objectKey, "mynewfile-3.jpg", 100*1024, oss.Checkpoint(true, "mynewfile.cp"))
	if err != nil {
		HandleError(err)
	}

	// Case 8: Use GZIP encoding for downloading the file, GetObject/GetObjectToFile are the same.
	err = bucket.PutObjectFromFile(objectKey, htmlLocalFile)
	if err != nil {
		HandleError(err)
	}

	err = bucket.GetObjectToFile(objectKey, "myhtml.gzip", oss.AcceptEncoding("gzip"))
	if err != nil {
		HandleError(err)
	}

	// Delete the object and bucket
	err = DeleteTestBucketAndObject(bucketName)
	if err != nil {
		HandleError(err)
	}

	fmt.Println("GetObjectSample completed")
}
