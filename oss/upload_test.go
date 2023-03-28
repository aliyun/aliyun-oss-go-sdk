package oss

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	. "gopkg.in/check.v1"
)

type OssUploadSuite struct {
	cloudBoxControlClient *Client
	client                *Client
	bucket                *Bucket
}

var _ = Suite(&OssUploadSuite{})

// SetUpSuite runs once when the suite starts running
func (s *OssUploadSuite) SetUpSuite(c *C) {
	bucketName := bucketNamePrefix + RandLowStr(6)
	if cloudboxControlEndpoint == "" {
		client, err := New(endpoint, accessID, accessKey)
		c.Assert(err, IsNil)
		s.client = client

		s.client.CreateBucket(bucketName)

		bucket, err := s.client.Bucket(bucketName)
		c.Assert(err, IsNil)
		s.bucket = bucket
	} else {
		client, err := New(cloudboxEndpoint, accessID, accessKey)
		c.Assert(err, IsNil)
		s.client = client

		controlClient, err := New(cloudboxControlEndpoint, accessID, accessKey)
		c.Assert(err, IsNil)
		s.cloudBoxControlClient = controlClient
		controlClient.CreateBucket(bucketName)

		bucket, err := s.client.Bucket(bucketName)
		c.Assert(err, IsNil)
		s.bucket = bucket
	}

	testLogger.Println("test upload started")
}

// TearDownSuite runs before each test or benchmark starts running
func (s *OssUploadSuite) TearDownSuite(c *C) {
	// Delete part
	keyMarker := KeyMarker("")
	uploadIDMarker := UploadIDMarker("")
	for {
		lmur, err := s.bucket.ListMultipartUploads(keyMarker, uploadIDMarker)
		c.Assert(err, IsNil)
		for _, upload := range lmur.Uploads {
			var imur = InitiateMultipartUploadResult{Bucket: s.bucket.BucketName,
				Key: upload.Key, UploadID: upload.UploadID}
			err = s.bucket.AbortMultipartUpload(imur)
			c.Assert(err, IsNil)
		}
		keyMarker = KeyMarker(lmur.NextKeyMarker)
		uploadIDMarker = UploadIDMarker(lmur.NextUploadIDMarker)
		if !lmur.IsTruncated {
			break
		}
	}

	// Delete objects
	marker := Marker("")
	for {
		lor, err := s.bucket.ListObjects(marker)
		c.Assert(err, IsNil)
		for _, object := range lor.Objects {
			err = s.bucket.DeleteObject(object.Key)
			c.Assert(err, IsNil)
		}
		marker = Marker(lor.NextMarker)
		if !lor.IsTruncated {
			break
		}
	}

	// Delete bucket
	if s.cloudBoxControlClient != nil {
		err := s.cloudBoxControlClient.DeleteBucket(s.bucket.BucketName)
		c.Assert(err, IsNil)
	} else {
		err := s.client.DeleteBucket(s.bucket.BucketName)
		c.Assert(err, IsNil)
	}

	testLogger.Println("test upload completed")
}

// SetUpTest runs after each test or benchmark runs
func (s *OssUploadSuite) SetUpTest(c *C) {
	err := removeTempFiles("../oss", ".jpg")
	c.Assert(err, IsNil)
}

// TearDownTest runs once after all tests or benchmarks have finished running
func (s *OssUploadSuite) TearDownTest(c *C) {
	err := removeTempFiles("../oss", ".jpg")
	c.Assert(err, IsNil)
}

// TestUploadRoutineWithoutRecovery tests multiroutineed upload without checkpoint
func (s *OssUploadSuite) TestUploadRoutineWithoutRecovery(c *C) {
	objectName := objectNamePrefix + RandStr(8)
	fileName := "../sample/BingWallpaper-2015-11-07.jpg"
	newFile := RandStr(8) + ".jpg"

	// Routines is not specified, by default single routine
	err := s.bucket.UploadFile(objectName, fileName, 100*1024)
	c.Assert(err, IsNil)

	os.Remove(newFile)
	err = s.bucket.GetObjectToFile(objectName, newFile)
	c.Assert(err, IsNil)

	eq, err := compareFiles(fileName, newFile)
	c.Assert(err, IsNil)
	c.Assert(eq, Equals, true)

	err = s.bucket.DeleteObject(objectName)
	c.Assert(err, IsNil)

	// Specify routine count as 1
	err = s.bucket.UploadFile(objectName, fileName, 100*1024, Routines(1))
	c.Assert(err, IsNil)

	os.Remove(newFile)
	err = s.bucket.GetObjectToFile(objectName, newFile)
	c.Assert(err, IsNil)

	eq, err = compareFiles(fileName, newFile)
	c.Assert(err, IsNil)
	c.Assert(eq, Equals, true)

	err = s.bucket.DeleteObject(objectName)
	c.Assert(err, IsNil)

	// Specify routine count as 3, which is smaller than parts count 5
	err = s.bucket.UploadFile(objectName, fileName, 100*1024, Routines(3))
	c.Assert(err, IsNil)

	os.Remove(newFile)
	err = s.bucket.GetObjectToFile(objectName, newFile)
	c.Assert(err, IsNil)

	eq, err = compareFiles(fileName, newFile)
	c.Assert(err, IsNil)
	c.Assert(eq, Equals, true)

	err = s.bucket.DeleteObject(objectName)
	c.Assert(err, IsNil)

	// Specify routine count as 5, which is same as the part count 5
	err = s.bucket.UploadFile(objectName, fileName, 100*1024, Routines(5))
	c.Assert(err, IsNil)

	os.Remove(newFile)
	err = s.bucket.GetObjectToFile(objectName, newFile)
	c.Assert(err, IsNil)

	eq, err = compareFiles(fileName, newFile)
	c.Assert(err, IsNil)
	c.Assert(eq, Equals, true)

	err = s.bucket.DeleteObject(objectName)
	c.Assert(err, IsNil)

	// Specify routine count as 10, which is bigger than the part count 5.
	err = s.bucket.UploadFile(objectName, fileName, 100*1024, Routines(10))
	c.Assert(err, IsNil)

	os.Remove(newFile)
	err = s.bucket.GetObjectToFile(objectName, newFile)
	c.Assert(err, IsNil)

	eq, err = compareFiles(fileName, newFile)
	c.Assert(err, IsNil)
	c.Assert(eq, Equals, true)

	err = s.bucket.DeleteObject(objectName)
	c.Assert(err, IsNil)

	// Invalid routine count, it will use 1 automatically.
	err = s.bucket.UploadFile(objectName, fileName, 100*1024, Routines(0))
	os.Remove(newFile)
	err = s.bucket.GetObjectToFile(objectName, newFile)
	c.Assert(err, IsNil)

	eq, err = compareFiles(fileName, newFile)
	c.Assert(err, IsNil)
	c.Assert(eq, Equals, true)

	err = s.bucket.DeleteObject(objectName)
	c.Assert(err, IsNil)

	// Invalid routine count, it will use 1 automatically
	err = s.bucket.UploadFile(objectName, fileName, 100*1024, Routines(-1))
	os.Remove(newFile)
	err = s.bucket.GetObjectToFile(objectName, newFile)
	c.Assert(err, IsNil)

	eq, err = compareFiles(fileName, newFile)
	c.Assert(err, IsNil)
	c.Assert(eq, Equals, true)

	err = s.bucket.DeleteObject(objectName)
	c.Assert(err, IsNil)

	// Option
	err = s.bucket.UploadFile(objectName, fileName, 100*1024, Routines(3), Meta("myprop", "mypropval"))

	meta, err := s.bucket.GetObjectDetailedMeta(objectName)
	c.Assert(err, IsNil)
	c.Assert(meta.Get("X-Oss-Meta-Myprop"), Equals, "mypropval")

	os.Remove(newFile)
	err = s.bucket.GetObjectToFile(objectName, newFile)
	c.Assert(err, IsNil)

	eq, err = compareFiles(fileName, newFile)
	c.Assert(err, IsNil)
	c.Assert(eq, Equals, true)

	err = s.bucket.DeleteObject(objectName)
	c.Assert(err, IsNil)
}

// ErrorHooker is a UploadPart hook---it will fail the 5th part's upload.
func ErrorHooker(id int, chunk FileChunk) error {
	if chunk.Number == 5 {
		time.Sleep(time.Second)
		return fmt.Errorf("ErrorHooker")
	}
	return nil
}

// TestUploadRoutineWithoutRecoveryNegative is multiroutineed upload without checkpoint
func (s *OssUploadSuite) TestUploadRoutineWithoutRecoveryNegative(c *C) {
	objectName := objectNamePrefix + RandStr(8)
	fileName := "../sample/BingWallpaper-2015-11-07.jpg"

	uploadPartHooker = ErrorHooker
	// Worker routine error
	err := s.bucket.UploadFile(objectName, fileName, 100*1024, Routines(2))
	c.Assert(err, NotNil)
	c.Assert(err.Error(), Equals, "ErrorHooker")
	uploadPartHooker = defaultUploadPart

	// Local file does not exist
	err = s.bucket.UploadFile(objectName, "NotExist", 100*1024, Routines(2))
	c.Assert(err, NotNil)

	// The part size is invalid
	err = s.bucket.UploadFile(objectName, fileName, 1024, Routines(2))
	c.Assert(err, NotNil)

	err = s.bucket.UploadFile(objectName, fileName, 1024*1024*1024*100, Routines(2))
	c.Assert(err, NotNil)
}

// TestUploadRoutineWithRecovery is multi-routine upload with resumable recovery
func (s *OssUploadSuite) TestUploadRoutineWithRecovery(c *C) {
	objectName := objectNamePrefix + RandStr(8)
	fileName := "../sample/BingWallpaper-2015-11-07.jpg"
	newFile := "upload-new-file-2.jpg"

	// Use default routines and default CP file path (fileName+.cp)
	// First upload for 4 parts
	uploadPartHooker = ErrorHooker
	err := s.bucket.UploadFile(objectName, fileName, 100*1024, Checkpoint(true, fileName+".cp"))
	c.Assert(err, NotNil)
	c.Assert(err.Error(), Equals, "ErrorHooker")
	uploadPartHooker = defaultUploadPart

	// Check CP
	ucp := uploadCheckpoint{}
	err = ucp.load(fileName + ".cp")
	c.Assert(err, IsNil)
	c.Assert(ucp.Magic, Equals, uploadCpMagic)
	c.Assert(len(ucp.MD5), Equals, len("LC34jZU5xK4hlxi3Qn3XGQ=="))
	c.Assert(ucp.FilePath, Equals, fileName)
	c.Assert(ucp.FileStat.Size, Equals, int64(482048))
	c.Assert(len(ucp.FileStat.LastModified.String()) > 0, Equals, true)
	c.Assert(ucp.FileStat.MD5, Equals, "")
	c.Assert(ucp.ObjectKey, Equals, objectName)
	c.Assert(len(ucp.UploadID), Equals, len("3F79722737D1469980DACEDCA325BB52"))
	c.Assert(len(ucp.Parts), Equals, 5)
	c.Assert(len(ucp.todoParts()), Equals, 1)
	c.Assert(len(ucp.allParts()), Equals, 5)

	// Second upload, finish the remaining part
	err = s.bucket.UploadFile(objectName, fileName, 100*1024, Checkpoint(true, fileName+".cp"))
	c.Assert(err, IsNil)

	os.Remove(newFile)
	err = s.bucket.GetObjectToFile(objectName, newFile)
	c.Assert(err, IsNil)

	eq, err := compareFiles(fileName, newFile)
	c.Assert(err, IsNil)
	c.Assert(eq, Equals, true)

	err = s.bucket.DeleteObject(objectName)
	c.Assert(err, IsNil)

	err = ucp.load(fileName + ".cp")
	c.Assert(err, NotNil)

	// Resumable upload with empty checkpoint path
	uploadPartHooker = ErrorHooker
	err = s.bucket.UploadFile(objectName, fileName, 100*1024, Routines(3), CheckpointDir(true, ""))
	c.Assert(err, NotNil)
	c.Assert(err.Error(), Equals, "ErrorHooker")
	uploadPartHooker = defaultUploadPart
	ucp = uploadCheckpoint{}
	err = ucp.load(fileName + ".cp")
	c.Assert(err, NotNil)

	// Resumable upload with checkpoint dir
	uploadPartHooker = ErrorHooker
	err = s.bucket.UploadFile(objectName, fileName, 100*1024, Routines(3), CheckpointDir(true, "./"))
	c.Assert(err, NotNil)
	c.Assert(err.Error(), Equals, "ErrorHooker")
	uploadPartHooker = defaultUploadPart

	// Check CP
	ucp = uploadCheckpoint{}
	cpConf := cpConfig{IsEnable: true, DirPath: "./"}
	cpFilePath := getUploadCpFilePath(&cpConf, fileName, s.bucket.BucketName, objectName)
	err = ucp.load(cpFilePath)
	c.Assert(err, IsNil)
	c.Assert(ucp.Magic, Equals, uploadCpMagic)
	c.Assert(len(ucp.MD5), Equals, len("LC34jZU5xK4hlxi3Qn3XGQ=="))
	c.Assert(ucp.FilePath, Equals, fileName)
	c.Assert(ucp.FileStat.Size, Equals, int64(482048))
	c.Assert(len(ucp.FileStat.LastModified.String()) > 0, Equals, true)
	c.Assert(ucp.FileStat.MD5, Equals, "")
	c.Assert(ucp.ObjectKey, Equals, objectName)
	c.Assert(len(ucp.UploadID), Equals, len("3F79722737D1469980DACEDCA325BB52"))
	c.Assert(len(ucp.Parts), Equals, 5)
	c.Assert(len(ucp.todoParts()), Equals, 1)
	c.Assert(len(ucp.allParts()), Equals, 5)

	err = s.bucket.UploadFile(objectName, fileName, 100*1024, Routines(3), CheckpointDir(true, "./"))
	c.Assert(err, IsNil)

	os.Remove(newFile)
	err = s.bucket.GetObjectToFile(objectName, newFile)
	c.Assert(err, IsNil)

	eq, err = compareFiles(fileName, newFile)
	c.Assert(err, IsNil)
	c.Assert(eq, Equals, true)

	err = s.bucket.DeleteObject(objectName)
	c.Assert(err, IsNil)

	err = ucp.load(cpFilePath)
	c.Assert(err, NotNil)

	// Upload all 5 parts without error
	err = s.bucket.UploadFile(objectName, fileName, 100*1024, Routines(3), Checkpoint(true, objectName+".cp"))
	c.Assert(err, IsNil)

	os.Remove(newFile)
	err = s.bucket.GetObjectToFile(objectName, newFile)
	c.Assert(err, IsNil)

	eq, err = compareFiles(fileName, newFile)
	c.Assert(err, IsNil)
	c.Assert(eq, Equals, true)

	err = s.bucket.DeleteObject(objectName)
	c.Assert(err, IsNil)

	// Upload all 5 parts with 10 routines without error
	err = s.bucket.UploadFile(objectName, fileName, 100*1024, Routines(10), Checkpoint(true, objectName+".cp"))
	c.Assert(err, IsNil)

	os.Remove(newFile)
	err = s.bucket.GetObjectToFile(objectName, newFile)
	c.Assert(err, IsNil)

	eq, err = compareFiles(fileName, newFile)
	c.Assert(err, IsNil)
	c.Assert(eq, Equals, true)

	err = s.bucket.DeleteObject(objectName)
	c.Assert(err, IsNil)

	// Option
	err = s.bucket.UploadFile(objectName, fileName, 100*1024, Routines(3), Checkpoint(true, objectName+".cp"), Meta("myprop", "mypropval"))

	meta, err := s.bucket.GetObjectDetailedMeta(objectName)
	c.Assert(err, IsNil)
	c.Assert(meta.Get("X-Oss-Meta-Myprop"), Equals, "mypropval")

	os.Remove(newFile)
	err = s.bucket.GetObjectToFile(objectName, newFile)
	c.Assert(err, IsNil)

	eq, err = compareFiles(fileName, newFile)
	c.Assert(err, IsNil)
	c.Assert(eq, Equals, true)

	err = s.bucket.DeleteObject(objectName)
	c.Assert(err, IsNil)
}

// TestUploadRoutineWithRecoveryNegative is multiroutineed upload without checkpoint
func (s *OssUploadSuite) TestUploadRoutineWithRecoveryNegative(c *C) {
	objectName := objectNamePrefix + RandStr(8)
	fileName := "../sample/BingWallpaper-2015-11-07.jpg"

	// The local file does not exist
	err := s.bucket.UploadFile(objectName, "NotExist", 100*1024, Checkpoint(true, "NotExist.cp"))
	c.Assert(err, NotNil)

	err = s.bucket.UploadFile(objectName, "NotExist", 100*1024, Routines(2), Checkpoint(true, "NotExist.cp"))
	c.Assert(err, NotNil)

	// Specified part size is invalid
	err = s.bucket.UploadFile(objectName, fileName, 1024, Checkpoint(true, fileName+".cp"))
	c.Assert(err, NotNil)

	err = s.bucket.UploadFile(objectName, fileName, 1024, Routines(2), Checkpoint(true, fileName+".cp"))
	c.Assert(err, NotNil)

	err = s.bucket.UploadFile(objectName, fileName, 1024*1024*1024*100, Checkpoint(true, fileName+".cp"))
	c.Assert(err, NotNil)

	err = s.bucket.UploadFile(objectName, fileName, 1024*1024*1024*100, Routines(2), Checkpoint(true, fileName+".cp"))
	c.Assert(err, NotNil)
}

// TestUploadLocalFileChange tests the file is updated while being uploaded
func (s *OssUploadSuite) TestUploadLocalFileChange(c *C) {
	objectName := objectNamePrefix + RandStr(8)
	fileName := "../sample/BingWallpaper-2015-11-07.jpg"
	localFile := RandStr(8) + ".jpg"
	newFile := RandStr(8) + ".jpg"

	os.Remove(localFile)
	err := copyFile(fileName, localFile)
	c.Assert(err, IsNil)

	// First upload for 4 parts
	uploadPartHooker = ErrorHooker
	err = s.bucket.UploadFile(objectName, localFile, 100*1024, Checkpoint(true, localFile+".cp"))
	c.Assert(err, NotNil)
	c.Assert(err.Error(), Equals, "ErrorHooker")
	uploadPartHooker = defaultUploadPart

	os.Remove(localFile)
	err = copyFile(fileName, localFile)
	c.Assert(err, IsNil)

	// Updating the file. The second upload will re-upload all 5 parts.
	err = s.bucket.UploadFile(objectName, localFile, 100*1024, Checkpoint(true, localFile+".cp"))
	c.Assert(err, IsNil)

	os.Remove(newFile)
	err = s.bucket.GetObjectToFile(objectName, newFile)
	c.Assert(err, IsNil)

	eq, err := compareFiles(fileName, newFile)
	c.Assert(err, IsNil)
	c.Assert(eq, Equals, true)

	err = s.bucket.DeleteObject(objectName)
	c.Assert(err, IsNil)
}

// TestUploadPartArchiveObject
func (s *OssUploadSuite) TestUploadPartArchiveObject(c *C) {
	// create archive bucket
	client, err := New(endpoint, accessID, accessKey)
	c.Assert(err, IsNil)

	bucketName := bucketNamePrefix + RandLowStr(6)
	err = client.CreateBucket(bucketName, StorageClass(StorageArchive))
	c.Assert(err, IsNil)
	bucket, err := client.Bucket(bucketName)
	objectName := objectNamePrefix + RandStr(8)

	fileName := "../sample/BingWallpaper-2015-11-07.jpg"
	fileInfo, err := os.Stat(fileName)
	c.Assert(err, IsNil)

	// Updating the file,archive object
	err = bucket.UploadFile(objectName, fileName, fileInfo.Size()/2, ObjectStorageClass(StorageArchive))
	c.Assert(err, IsNil)

	// Updating the file,archive object,checkpoint
	err = bucket.UploadFile(objectName, fileName, fileInfo.Size()/2, ObjectStorageClass(StorageArchive), Checkpoint(true, fileName+".cp"))
	c.Assert(err, IsNil)
	ForceDeleteBucket(client, bucketName, c)
}

func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

func (s *OssUploadSuite) TestVersioningUploadRoutineWithRecovery(c *C) {
	// create a bucket with default proprety
	client, err := New(endpoint, accessID, accessKey)
	c.Assert(err, IsNil)

	bucketName := bucketNamePrefix + RandLowStr(6)
	err = client.CreateBucket(bucketName)
	c.Assert(err, IsNil)

	bucket, err := client.Bucket(bucketName)

	// put bucket version:enabled
	var versioningConfig VersioningConfig
	versioningConfig.Status = string(VersionEnabled)
	err = client.SetBucketVersioning(bucketName, versioningConfig)
	c.Assert(err, IsNil)

	// begin test
	objectName := objectNamePrefix + RandStr(8)
	fileName := "test-file-" + RandStr(8)
	fileData := RandStr(500 * 1024)
	CreateFile(fileName, fileData, c)
	newFile := "test-file-" + RandStr(8)

	// Use default routines and default CP file path (fileName+.cp)Header
	// First upload for 4 parts
	var respHeader http.Header
	uploadPartHooker = ErrorHooker
	options := []Option{Checkpoint(true, fileName+".cp"), GetResponseHeader(&respHeader)}
	err = bucket.UploadFile(objectName, fileName, 100*1024, options...)
	c.Assert(err, NotNil)
	c.Assert(err.Error(), Equals, "ErrorHooker")
	c.Assert(GetVersionId(respHeader), Equals, "")

	uploadPartHooker = defaultUploadPart

	// Second upload, finish the remaining part
	options = []Option{Checkpoint(true, fileName+".cp"), GetResponseHeader(&respHeader)}
	err = bucket.UploadFile(objectName, fileName, 100*1024, options...)
	c.Assert(err, IsNil)
	versionIdUp := GetVersionId(respHeader)
	c.Assert(len(versionIdUp) > 0, Equals, true)

	os.Remove(newFile)
	var respHeaderDown http.Header
	err = bucket.GetObjectToFile(objectName, newFile, GetResponseHeader(&respHeaderDown))
	versionIdDown := GetVersionId(respHeaderDown)
	c.Assert(err, IsNil)
	c.Assert(versionIdUp, Equals, versionIdDown)

	eq, err := compareFiles(fileName, newFile)
	c.Assert(err, IsNil)
	c.Assert(eq, Equals, true)

	os.Remove(fileName)
	os.Remove(newFile)
	bucket.DeleteObject(objectName)
	ForceDeleteBucket(client, bucketName, c)
}

// TestUploadFileChoiceOptions
func (s *OssUploadSuite) TestUploadFileChoiceOptions(c *C) {
	// create a bucket with default proprety
	client, err := New(endpoint, accessID, accessKey)
	c.Assert(err, IsNil)

	bucketName := bucketNamePrefix + RandLowStr(6)
	err = client.CreateBucket(bucketName)
	c.Assert(err, IsNil)
	bucket, err := client.Bucket(bucketName)

	fileName := "../sample/BingWallpaper-2015-11-07.jpg"
	fileInfo, err := os.Stat(fileName)
	c.Assert(err, IsNil)

	objectName := objectNamePrefix + RandStr(8)

	// UploadFile with properties
	options := []Option{
		ObjectACL(ACLPublicRead),
		RequestPayer(Requester),
		TrafficLimitHeader(1024 * 1024 * 8),
		ServerSideEncryption("AES256"),
		ObjectStorageClass(StorageArchive),
	}

	// Updating the file
	err = bucket.UploadFile(objectName, fileName, fileInfo.Size()/2, options...)
	c.Assert(err, IsNil)

	// GetMetaDetail
	headerResp, err := bucket.GetObjectDetailedMeta(objectName)
	c.Assert(err, IsNil)

	c.Assert(headerResp.Get("X-Oss-Server-Side-Encryption"), Equals, "AES256")
	aclResult, err := bucket.GetObjectACL(objectName)
	c.Assert(aclResult.ACL, Equals, "public-read")
	c.Assert(err, IsNil)
	ForceDeleteBucket(client, bucketName, c)
}

// TestUploadFileWithCpChoiceOptions
func (s *OssUploadSuite) TestUploadFileWithCpChoiceOptions(c *C) {
	// create a bucket with default proprety
	client, err := New(endpoint, accessID, accessKey)
	c.Assert(err, IsNil)

	bucketName := bucketNamePrefix + RandLowStr(6)
	err = client.CreateBucket(bucketName)
	c.Assert(err, IsNil)
	bucket, err := client.Bucket(bucketName)

	fileName := "../sample/BingWallpaper-2015-11-07.jpg"
	fileInfo, err := os.Stat(fileName)
	c.Assert(err, IsNil)

	objectName := objectNamePrefix + RandStr(8)

	// UploadFile with properties
	options := []Option{
		ObjectACL(ACLPublicRead),
		RequestPayer(Requester),
		TrafficLimitHeader(1024 * 1024 * 8),
		ServerSideEncryption("AES256"),
		ObjectStorageClass(StorageArchive),
		Checkpoint(true, fileName+".cp"), // with checkpoint
	}

	// Updating the file
	err = bucket.UploadFile(objectName, fileName, fileInfo.Size()/2, options...)
	c.Assert(err, IsNil)

	// GetMetaDetail
	headerResp, err := bucket.GetObjectDetailedMeta(objectName)
	c.Assert(err, IsNil)

	c.Assert(headerResp.Get("X-Oss-Server-Side-Encryption"), Equals, "AES256")
	c.Assert(headerResp.Get("X-Oss-Storage-Class"), Equals, "Archive")

	aclResult, err := bucket.GetObjectACL(objectName)
	c.Assert(aclResult.ACL, Equals, "public-read")
	c.Assert(err, IsNil)

	ForceDeleteBucket(client, bucketName, c)
}

// TestUploadFileWithForbidOverWrite
func (s *OssUploadSuite) TestUploadFileWithForbidOverWrite(c *C) {
	// create a bucket with default proprety
	client, err := New(endpoint, accessID, accessKey)
	c.Assert(err, IsNil)

	bucketName := bucketNamePrefix + RandLowStr(6)
	err = client.CreateBucket(bucketName)
	c.Assert(err, IsNil)
	bucket, err := client.Bucket(bucketName)

	fileName := "../sample/BingWallpaper-2015-11-07.jpg"
	fileInfo, err := os.Stat(fileName)
	c.Assert(err, IsNil)

	objectName := objectNamePrefix + RandStr(8)

	// UploadFile with properties
	options := []Option{
		ObjectACL(ACLPublicRead),
		RequestPayer(Requester),
		TrafficLimitHeader(1024 * 1024 * 8),
		ServerSideEncryption("AES256"),
		ObjectStorageClass(StorageArchive),
		ForbidOverWrite(true),
		Checkpoint(true, fileName+".cp"),
	}

	// Updating the file
	err = bucket.UploadFile(objectName, fileName, fileInfo.Size()/2, options...)
	c.Assert(err, IsNil)

	// Updating the file with ForbidOverWrite(true)
	err = bucket.UploadFile(objectName, fileName, fileInfo.Size()/2, options...)
	c.Assert(err, NotNil)

	// without Checkpoint
	options = []Option{
		ObjectACL(ACLPublicRead),
		RequestPayer(Requester),
		TrafficLimitHeader(1024 * 1024 * 8),
		ServerSideEncryption("AES256"),
		ObjectStorageClass(StorageArchive),
		ForbidOverWrite(true),
	}

	// Updating the file with ForbidOverWrite(true)
	err = bucket.UploadFile(objectName, fileName, fileInfo.Size()/2, options...)
	c.Assert(err, NotNil)

	ForceDeleteBucket(client, bucketName, c)
}

// TestUploadFileWithSequential
func (s *OssUploadSuite) TestUploadFileWithSequential(c *C) {
	// create a bucket with default proprety
	client, err := New(endpoint, accessID, accessKey)
	c.Assert(err, IsNil)

	bucketName := bucketNamePrefix + RandLowStr(6)
	err = client.CreateBucket(bucketName)
	c.Assert(err, IsNil)
	bucket, err := client.Bucket(bucketName)

	fileName := "../sample/BingWallpaper-2015-11-07.jpg"
	fileInfo, err := os.Stat(fileName)
	c.Assert(err, IsNil)

	objectName := objectNamePrefix + RandStr(8)

	var respHeader http.Header

	// UploadFile with properties
	options := []Option{
		Sequential(),
		GetResponseHeader(&respHeader),
		Checkpoint(true, fileName+".cp"),
	}

	// Updating the file
	err = bucket.UploadFile(objectName, fileName, fileInfo.Size()/2, options...)
	c.Assert(err, IsNil)

	respHeader, err = bucket.GetObjectDetailedMeta(objectName)
	c.Assert(err, IsNil)

	strMD5 := respHeader.Get("Content-MD5")
	c.Assert(len(strMD5) > 0, Equals, true)

	ForceDeleteBucket(client, bucketName, c)
}

// TestUploadFileWithSequential
func (s *OssUploadSuite) TestUploadFileWithCallbackResult(c *C) {
	// create a bucket with default proprety
	client, err := New(endpoint, accessID, accessKey)
	c.Assert(err, IsNil)

	bucketName := bucketNamePrefix + RandLowStr(6)
	err = client.CreateBucket(bucketName)
	c.Assert(err, IsNil)
	bucket, err := client.Bucket(bucketName)

	fileName := "../sample/BingWallpaper-2015-11-07.jpg"
	fileInfo, err := os.Stat(fileName)
	c.Assert(err, IsNil)

	objectName := objectNamePrefix + RandStr(8)

	var respHeader http.Header
	callbackMap := map[string]string{}
	callbackMap["callbackUrl"] = "www.aliyuncs.com"
	callbackMap["callbackBody"] = "filename=${object}&size=${size}&mimeType=${mimeType}"
	callbackMap["callbackBodyType"] = "application/x-www-form-urlencoded"
	callbackBuffer := bytes.NewBuffer([]byte{})
	callbackEncoder := json.NewEncoder(callbackBuffer)
	callbackEncoder.SetEscapeHTML(false)
	err = callbackEncoder.Encode(callbackMap)
	c.Assert(err, IsNil)

	callbackVal := base64.StdEncoding.EncodeToString(callbackBuffer.Bytes())
	var pBody []byte

	// UploadFile with properties
	options := []Option{
		Sequential(),
		GetResponseHeader(&respHeader),
		Checkpoint(true, fileName+".cp"),
		Callback(callbackVal),
		CallbackResult(&pBody),
	}

	// Updating the file
	err = bucket.UploadFile(objectName, fileName, fileInfo.Size()/2, options...)
	c.Assert(err, NotNil)

	e, ok := err.(ServiceError)
	c.Assert(ok, Equals, true)
	c.Assert(e.StatusCode, Equals, 203)
	c.Assert(pBody, IsNil)

	respHeader, err = bucket.GetObjectDetailedMeta(objectName)
	c.Assert(err, IsNil)

	strMD5 := respHeader.Get("Content-MD5")
	c.Assert(len(strMD5) > 0, Equals, true)

	ForceDeleteBucket(client, bucketName, c)
}
