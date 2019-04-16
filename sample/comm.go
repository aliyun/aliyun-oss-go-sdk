package sample

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

var (
	pastDate   = time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)
	futureDate = time.Date(2049, time.January, 10, 23, 0, 0, 0, time.UTC)
)

// HandleError is the error handling method in the sample code
func HandleError(err error) {
	fmt.Println("occurred error:", err)
	os.Exit(-1)
}

// GetTestBucket creates the test bucket
func GetTestBucket(bucketName string) (*oss.Bucket, error) {
	// New client
	client, err := oss.New(endpoint, accessID, accessKey)
	if err != nil {
		return nil, err
	}

	// Create bucket
	err = client.CreateBucket(bucketName)
	if err != nil {
		return nil, err
	}

	// Get bucket
	bucket, err := client.Bucket(bucketName)
	if err != nil {
		return nil, err
	}

	return bucket, nil
}

// DeleteTestBucketAndLiveChannel 删除sample的channelname和bucket，该函数为了简化sample，让sample代码更明了
func DeleteTestBucketAndLiveChannel(bucketName string) error {
	// New Client
	client, err := oss.New(endpoint, accessID, accessKey)
	if err != nil {
		return err
	}

	// Get Bucket
	bucket, err := client.Bucket(bucketName)
	if err != nil {
		return err
	}

	marker := ""
	for {
		result, err := bucket.ListLiveChannel(oss.Marker(marker))
		if err != nil {
			HandleError(err)
		}

		for _, channel := range result.LiveChannel {
			err := bucket.DeleteLiveChannel(channel.Name)
			if err != nil {
				HandleError(err)
			}
		}

		if result.IsTruncated {
			marker = result.NextMarker
		} else {
			break
		}
	}

	// Delete Bucket
	err = client.DeleteBucket(bucketName)
	if err != nil {
		return err
	}

	return nil
}

// DeleteTestBucketAndObject deletes the test bucket and its objects
func DeleteTestBucketAndObject(bucketName string) error {
	// New client
	client, err := oss.New(endpoint, accessID, accessKey)
	if err != nil {
		return err
	}

	// Get bucket
	bucket, err := client.Bucket(bucketName)
	if err != nil {
		return err
	}

	// Delete part
	keyMarker := oss.KeyMarker("")
	uploadIDMarker := oss.UploadIDMarker("")
	for {
		lmur, err := bucket.ListMultipartUploads(keyMarker, uploadIDMarker)
		if err != nil {
			return err
		}
		for _, upload := range lmur.Uploads {
			var imur = oss.InitiateMultipartUploadResult{Bucket: bucket.BucketName,
				Key: upload.Key, UploadID: upload.UploadID}
			err = bucket.AbortMultipartUpload(imur)
			if err != nil {
				return err
			}
		}
		keyMarker = oss.KeyMarker(lmur.NextKeyMarker)
		uploadIDMarker = oss.UploadIDMarker(lmur.NextUploadIDMarker)
		if !lmur.IsTruncated {
			break
		}
	}

	// Delete objects
	marker := oss.Marker("")
	for {
		lor, err := bucket.ListObjects(marker)
		if err != nil {
			return err
		}
		for _, object := range lor.Objects {
			err = bucket.DeleteObject(object.Key)
			if err != nil {
				return err
			}
		}
		marker = oss.Marker(lor.NextMarker)
		if !lor.IsTruncated {
			break
		}
	}

	// Delete bucket
	err = client.DeleteBucket(bucketName)
	if err != nil {
		return err
	}

	return nil
}

// Object defines pair of key and value
type Object struct {
	Key   string
	Value string
}

// CreateObjects creates some objects
func CreateObjects(bucket *oss.Bucket, objects []Object) error {
	for _, object := range objects {
		err := bucket.PutObject(object.Key, strings.NewReader(object.Value))
		if err != nil {
			return err
		}
	}
	return nil
}

// DeleteObjects deletes some objects.
func DeleteObjects(bucket *oss.Bucket, objects []Object) error {
	for _, object := range objects {
		err := bucket.DeleteObject(object.Key)
		if err != nil {
			return err
		}
	}
	return nil
}
