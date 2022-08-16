package oss

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	. "gopkg.in/check.v1"
)

type OssCopySuite struct {
	cloudBoxControlClient *Client
	client                *Client
	bucket                *Bucket
}

var _ = Suite(&OssCopySuite{})

// SetUpSuite runs once when the suite starts running
func (s *OssCopySuite) SetUpSuite(c *C) {
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

	testLogger.Println("test copy started")
}

// TearDownSuite runs before each test or benchmark starts running
func (s *OssCopySuite) TearDownSuite(c *C) {
	// Delete Part
	keyMarker := KeyMarker("")
	uploadIDMarker := UploadIDMarker("")
	for {
		lmur, err := s.bucket.ListMultipartUploads(keyMarker, uploadIDMarker)
		c.Assert(err, IsNil)
		for _, upload := range lmur.Uploads {
			var imur = InitiateMultipartUploadResult{Bucket: bucketName,
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

	testLogger.Println("test copy completed")
}

// SetUpTest runs after each test or benchmark runs
func (s *OssCopySuite) SetUpTest(c *C) {
	err := removeTempFiles("../oss", ".jpg")
	c.Assert(err, IsNil)
}

// TearDownTest runs once after all tests or benchmarks have finished running
func (s *OssCopySuite) TearDownTest(c *C) {
	err := removeTempFiles("../oss", ".jpg")
	c.Assert(err, IsNil)
}

// TestCopyRoutineWithoutRecovery is multi-routine copy without resumable recovery
func (s *OssCopySuite) TestCopyRoutineWithoutRecovery(c *C) {
	srcObjectName := objectNamePrefix + RandStr(8)
	destObjectName := srcObjectName + "-dest"
	fileName := "../sample/BingWallpaper-2015-11-07.jpg"
	newFile := "copy-new-file.jpg"

	// Upload source file
	err := s.bucket.UploadFile(srcObjectName, fileName, 100*1024, Routines(3))
	c.Assert(err, IsNil)
	os.Remove(newFile)

	// Does not specify parameter 'routines', by default it's single routine
	err = s.bucket.CopyFile(s.bucket.BucketName, srcObjectName, destObjectName, 100*1024)
	c.Assert(err, IsNil)

	err = s.bucket.GetObjectToFile(destObjectName, newFile)
	c.Assert(err, IsNil)

	eq, err := compareFiles(fileName, newFile)
	c.Assert(err, IsNil)
	c.Assert(eq, Equals, true)

	err = s.bucket.DeleteObject(destObjectName)
	c.Assert(err, IsNil)
	os.Remove(newFile)

	// Specify one routine.
	err = s.bucket.CopyFile(s.bucket.BucketName, srcObjectName, destObjectName, 100*1024, Routines(1))
	c.Assert(err, IsNil)

	err = s.bucket.GetObjectToFile(destObjectName, newFile)
	c.Assert(err, IsNil)

	eq, err = compareFiles(fileName, newFile)
	c.Assert(err, IsNil)
	c.Assert(eq, Equals, true)

	err = s.bucket.DeleteObject(destObjectName)
	c.Assert(err, IsNil)
	os.Remove(newFile)

	// Specify three routines, which is less than parts count 5
	err = s.bucket.CopyFile(s.bucket.BucketName, srcObjectName, destObjectName, 100*1024, Routines(3))
	c.Assert(err, IsNil)

	err = s.bucket.GetObjectToFile(destObjectName, newFile)
	c.Assert(err, IsNil)

	eq, err = compareFiles(fileName, newFile)
	c.Assert(err, IsNil)
	c.Assert(eq, Equals, true)

	err = s.bucket.DeleteObject(destObjectName)
	c.Assert(err, IsNil)
	os.Remove(newFile)

	// Specify 5 routines which is the same as parts count
	err = s.bucket.CopyFile(s.bucket.BucketName, srcObjectName, destObjectName, 100*1024, Routines(5))
	c.Assert(err, IsNil)

	err = s.bucket.GetObjectToFile(destObjectName, newFile)
	c.Assert(err, IsNil)

	eq, err = compareFiles(fileName, newFile)
	c.Assert(err, IsNil)
	c.Assert(eq, Equals, true)

	err = s.bucket.DeleteObject(destObjectName)
	c.Assert(err, IsNil)
	os.Remove(newFile)

	// Specify routine count 10, which is more than parts count
	err = s.bucket.CopyFile(s.bucket.BucketName, srcObjectName, destObjectName, 100*1024, Routines(10))
	c.Assert(err, IsNil)

	err = s.bucket.GetObjectToFile(destObjectName, newFile)
	c.Assert(err, IsNil)

	eq, err = compareFiles(fileName, newFile)
	c.Assert(err, IsNil)
	c.Assert(eq, Equals, true)

	err = s.bucket.DeleteObject(destObjectName)
	c.Assert(err, IsNil)
	os.Remove(newFile)

	// Invalid routine count, will use single routine
	err = s.bucket.CopyFile(s.bucket.BucketName, srcObjectName, destObjectName, 100*1024, Routines(-1))
	c.Assert(err, IsNil)

	err = s.bucket.GetObjectToFile(destObjectName, newFile)
	c.Assert(err, IsNil)

	eq, err = compareFiles(fileName, newFile)
	c.Assert(err, IsNil)
	c.Assert(eq, Equals, true)

	err = s.bucket.DeleteObject(destObjectName)
	c.Assert(err, IsNil)
	os.Remove(newFile)

	// Option
	err = s.bucket.CopyFile(s.bucket.BucketName, srcObjectName, destObjectName, 100*1024, Routines(3), Meta("myprop", "mypropval"))

	meta, err := s.bucket.GetObjectDetailedMeta(destObjectName)
	c.Assert(err, IsNil)
	c.Assert(meta.Get("X-Oss-Meta-Myprop"), Equals, "mypropval")

	err = s.bucket.GetObjectToFile(destObjectName, newFile)
	c.Assert(err, IsNil)

	eq, err = compareFiles(fileName, newFile)
	c.Assert(err, IsNil)
	c.Assert(eq, Equals, true)

	err = s.bucket.DeleteObject(destObjectName)
	c.Assert(err, IsNil)
	os.Remove(newFile)

	err = s.bucket.DeleteObject(srcObjectName)
	c.Assert(err, IsNil)
}

// CopyErrorHooker is a copypart request hook
func CopyErrorHooker(part copyPart) error {
	if part.Number == 5 {
		time.Sleep(time.Second)
		return fmt.Errorf("ErrorHooker")
	}
	return nil
}

// TestCopyRoutineWithoutRecoveryNegative is a multiple routines copy without checkpoint
func (s *OssCopySuite) TestCopyRoutineWithoutRecoveryNegative(c *C) {
	srcObjectName := objectNamePrefix + RandStr(8)
	destObjectName := srcObjectName + "-dest"
	fileName := "../sample/BingWallpaper-2015-11-07.jpg"

	// Upload source file
	err := s.bucket.UploadFile(srcObjectName, fileName, 100*1024, Routines(3))
	c.Assert(err, IsNil)

	copyPartHooker = CopyErrorHooker
	// Worker routine errors
	err = s.bucket.CopyFile(s.bucket.BucketName, srcObjectName, destObjectName, 100*1024, Routines(2))

	c.Assert(err, NotNil)
	c.Assert(err.Error(), Equals, "ErrorHooker")
	copyPartHooker = defaultCopyPartHook

	// Source bucket does not exist
	err = s.bucket.CopyFile("notexist", srcObjectName, destObjectName, 100*1024, Routines(2))
	c.Assert(err, NotNil)

	// Target object does not exist
	err = s.bucket.CopyFile(s.bucket.BucketName, "notexist", destObjectName, 100*1024, Routines(2))

	// The part size is invalid
	err = s.bucket.CopyFile(s.bucket.BucketName, srcObjectName, destObjectName, 1024, Routines(2))
	c.Assert(err, NotNil)

	err = s.bucket.CopyFile(s.bucket.BucketName, srcObjectName, destObjectName, 1024*1024*1024*100, Routines(2))
	c.Assert(err, NotNil)

	// Delete the source file
	err = s.bucket.DeleteObject(srcObjectName)
	c.Assert(err, IsNil)
}

// TestCopyRoutineWithRecovery is a multiple routines copy with resumable recovery
func (s *OssCopySuite) TestCopyRoutineWithRecovery(c *C) {
	srcObjectName := objectNamePrefix + RandStr(8)
	destObjectName := srcObjectName + "-dest"
	fileName := "../sample/BingWallpaper-2015-11-07.jpg"
	newFile := RandStr(8) + ".jpg"

	// Upload source file
	err := s.bucket.UploadFile(srcObjectName, fileName, 100*1024, Routines(3))
	c.Assert(err, IsNil)
	os.Remove(newFile)

	// Routines default value, CP's default path is destObjectName+.cp
	// Copy object with checkpoint enabled, single runtine.
	// Copy 4 parts---the CopyErrorHooker makes sure the copy of part 5 will fail.
	copyPartHooker = CopyErrorHooker
	err = s.bucket.CopyFile(s.bucket.BucketName, srcObjectName, destObjectName, 1024*100, Checkpoint(true, destObjectName+".cp"))
	c.Assert(err, NotNil)
	c.Assert(err.Error(), Equals, "ErrorHooker")
	copyPartHooker = defaultCopyPartHook

	// Check CP
	ccp := copyCheckpoint{}
	err = ccp.load(destObjectName + ".cp")
	c.Assert(err, IsNil)
	c.Assert(ccp.Magic, Equals, copyCpMagic)
	c.Assert(len(ccp.MD5), Equals, len("LC34jZU5xK4hlxi3Qn3XGQ=="))
	c.Assert(ccp.SrcBucketName, Equals, s.bucket.BucketName)
	c.Assert(ccp.SrcObjectKey, Equals, srcObjectName)
	c.Assert(ccp.DestBucketName, Equals, s.bucket.BucketName)
	c.Assert(ccp.DestObjectKey, Equals, destObjectName)
	c.Assert(len(ccp.CopyID), Equals, len("3F79722737D1469980DACEDCA325BB52"))
	c.Assert(ccp.ObjStat.Size, Equals, int64(482048))
	c.Assert(len(ccp.ObjStat.LastModified), Equals, len("2015-12-17 18:43:03 +0800 CST"))
	c.Assert(ccp.ObjStat.Etag, Equals, "\"2351E662233817A7AE974D8C5B0876DD-5\"")
	c.Assert(len(ccp.Parts), Equals, 5)
	c.Assert(len(ccp.todoParts()), Equals, 1)
	c.Assert(ccp.PartStat[4], Equals, false)

	// Second copy, finish the last part
	err = s.bucket.CopyFile(s.bucket.BucketName, srcObjectName, destObjectName, 1024*100, Checkpoint(true, destObjectName+".cp"))
	c.Assert(err, IsNil)

	err = s.bucket.GetObjectToFile(destObjectName, newFile)
	c.Assert(err, IsNil)

	eq, err := compareFiles(fileName, newFile)
	c.Assert(err, IsNil)
	c.Assert(eq, Equals, true)

	err = s.bucket.DeleteObject(destObjectName)
	c.Assert(err, IsNil)
	os.Remove(newFile)

	err = ccp.load(fileName + ".cp")
	c.Assert(err, NotNil)

	//multicopy with empty checkpoint path
	copyPartHooker = CopyErrorHooker
	err = s.bucket.CopyFile(s.bucket.BucketName, srcObjectName, destObjectName, 1024*100, Checkpoint(true, ""))
	c.Assert(err, NotNil)
	c.Assert(err.Error(), Equals, "ErrorHooker")
	copyPartHooker = defaultCopyPartHook
	ccp = copyCheckpoint{}
	err = ccp.load(destObjectName + ".cp")
	c.Assert(err, NotNil)

	//multi copy with checkpoint dir
	copyPartHooker = CopyErrorHooker
	err = s.bucket.CopyFile(s.bucket.BucketName, srcObjectName, destObjectName, 1024*100, Routines(2), CheckpointDir(true, "./"))
	c.Assert(err, NotNil)
	c.Assert(err.Error(), Equals, "ErrorHooker")
	copyPartHooker = defaultCopyPartHook

	// Check CP
	ccp = copyCheckpoint{}
	cpConf := cpConfig{IsEnable: true, DirPath: "./"}
	cpFilePath := getCopyCpFilePath(&cpConf, s.bucket.BucketName, srcObjectName, s.bucket.BucketName, destObjectName, "")
	err = ccp.load(cpFilePath)
	c.Assert(err, IsNil)
	c.Assert(ccp.Magic, Equals, copyCpMagic)
	c.Assert(len(ccp.MD5), Equals, len("LC34jZU5xK4hlxi3Qn3XGQ=="))
	c.Assert(ccp.SrcBucketName, Equals, s.bucket.BucketName)
	c.Assert(ccp.SrcObjectKey, Equals, srcObjectName)
	c.Assert(ccp.DestBucketName, Equals, s.bucket.BucketName)
	c.Assert(ccp.DestObjectKey, Equals, destObjectName)
	c.Assert(len(ccp.CopyID), Equals, len("3F79722737D1469980DACEDCA325BB52"))
	c.Assert(ccp.ObjStat.Size, Equals, int64(482048))
	c.Assert(len(ccp.ObjStat.LastModified), Equals, len("2015-12-17 18:43:03 +0800 CST"))
	c.Assert(ccp.ObjStat.Etag, Equals, "\"2351E662233817A7AE974D8C5B0876DD-5\"")
	c.Assert(len(ccp.Parts), Equals, 5)
	c.Assert(len(ccp.todoParts()), Equals, 1)
	c.Assert(ccp.PartStat[4], Equals, false)

	// Second copy, finish the last part.
	err = s.bucket.CopyFile(s.bucket.BucketName, srcObjectName, destObjectName, 1024*100, Routines(2), CheckpointDir(true, "./"))
	c.Assert(err, IsNil)

	err = s.bucket.GetObjectToFile(destObjectName, newFile)
	c.Assert(err, IsNil)

	eq, err = compareFiles(fileName, newFile)
	c.Assert(err, IsNil)
	c.Assert(eq, Equals, true)

	err = s.bucket.DeleteObject(destObjectName)
	c.Assert(err, IsNil)
	os.Remove(newFile)

	err = ccp.load(srcObjectName + ".cp")
	c.Assert(err, NotNil)

	// First copy without error.
	err = s.bucket.CopyFile(s.bucket.BucketName, srcObjectName, destObjectName, 1024*100, Routines(3), Checkpoint(true, destObjectName+".cp"))
	c.Assert(err, IsNil)

	err = s.bucket.GetObjectToFile(destObjectName, newFile)
	c.Assert(err, IsNil)

	eq, err = compareFiles(fileName, newFile)
	c.Assert(err, IsNil)
	c.Assert(eq, Equals, true)

	err = s.bucket.DeleteObject(destObjectName)
	c.Assert(err, IsNil)
	os.Remove(newFile)

	// Copy with multiple coroutines, no errors.
	err = s.bucket.CopyFile(s.bucket.BucketName, srcObjectName, destObjectName, 1024*100, Routines(10), Checkpoint(true, destObjectName+".cp"))
	c.Assert(err, IsNil)

	err = s.bucket.GetObjectToFile(destObjectName, newFile)
	c.Assert(err, IsNil)

	eq, err = compareFiles(fileName, newFile)
	c.Assert(err, IsNil)
	c.Assert(eq, Equals, true)

	err = s.bucket.DeleteObject(destObjectName)
	c.Assert(err, IsNil)
	os.Remove(newFile)

	// Option
	err = s.bucket.CopyFile(s.bucket.BucketName, srcObjectName, destObjectName, 1024*100, Routines(5), Checkpoint(true, destObjectName+".cp"), Meta("myprop", "mypropval"))
	c.Assert(err, IsNil)

	meta, err := s.bucket.GetObjectDetailedMeta(destObjectName)
	c.Assert(err, IsNil)
	c.Assert(meta.Get("X-Oss-Meta-Myprop"), Equals, "mypropval")

	err = s.bucket.GetObjectToFile(destObjectName, newFile)
	c.Assert(err, IsNil)

	eq, err = compareFiles(fileName, newFile)
	c.Assert(err, IsNil)
	c.Assert(eq, Equals, true)

	err = s.bucket.DeleteObject(destObjectName)
	c.Assert(err, IsNil)
	os.Remove(newFile)

	// Delete the source file
	err = s.bucket.DeleteObject(srcObjectName)
	c.Assert(err, IsNil)
}

// TestCopyRoutineWithRecoveryNegative is a multiple routineed copy without checkpoint
func (s *OssCopySuite) TestCopyRoutineWithRecoveryNegative(c *C) {
	srcObjectName := objectNamePrefix + RandStr(8)
	destObjectName := srcObjectName + "-dest"

	// Source bucket does not exist
	err := s.bucket.CopyFile("notexist", srcObjectName, destObjectName, 100*1024, Checkpoint(true, destObjectName+".cp"))
	c.Assert(err, NotNil)
	c.Assert(err, NotNil)

	// Source object does not exist
	err = s.bucket.CopyFile(s.bucket.BucketName, "notexist", destObjectName, 100*1024, Routines(2), Checkpoint(true, destObjectName+".cp"))
	c.Assert(err, NotNil)

	// Specify part size is invalid.
	err = s.bucket.CopyFile(s.bucket.BucketName, srcObjectName, destObjectName, 1024, Checkpoint(true, destObjectName+".cp"))
	c.Assert(err, NotNil)

	err = s.bucket.CopyFile(s.bucket.BucketName, srcObjectName, destObjectName, 1024*1024*1024*100, Routines(2), Checkpoint(true, destObjectName+".cp"))
	c.Assert(err, NotNil)
}

// TestCopyFileCrossBucket is a cross bucket's direct copy.
func (s *OssCopySuite) TestCopyFileCrossBucket(c *C) {
	destBucketName := s.bucket.BucketName + "-cross-b"
	srcObjectName := objectNamePrefix + RandStr(8)
	destObjectName := srcObjectName + "-dest"
	fileName := "../sample/BingWallpaper-2015-11-07.jpg"
	newFile := RandStr(8) + ".jpg"

	destBucket, err := s.client.Bucket(destBucketName)
	c.Assert(err, IsNil)

	// Create a target bucket
	err = s.client.CreateBucket(destBucketName)

	// Upload source file
	err = s.bucket.UploadFile(srcObjectName, fileName, 100*1024, Routines(3))
	c.Assert(err, IsNil)
	os.Remove(newFile)

	// Copy files
	err = destBucket.CopyFile(s.bucket.BucketName, srcObjectName, destObjectName, 1024*100, Routines(5), Checkpoint(true, destObjectName+".cp"))
	c.Assert(err, IsNil)

	err = destBucket.GetObjectToFile(destObjectName, newFile)
	c.Assert(err, IsNil)

	eq, err := compareFiles(fileName, newFile)
	c.Assert(err, IsNil)
	c.Assert(eq, Equals, true)

	err = destBucket.DeleteObject(destObjectName)
	c.Assert(err, IsNil)
	os.Remove(newFile)

	// Copy file with options
	err = destBucket.CopyFile(s.bucket.BucketName, srcObjectName, destObjectName, 1024*100, Routines(10), Checkpoint(true, "copy.cp"), Meta("myprop", "mypropval"))
	c.Assert(err, IsNil)

	err = destBucket.GetObjectToFile(destObjectName, newFile)
	c.Assert(err, IsNil)

	eq, err = compareFiles(fileName, newFile)
	c.Assert(err, IsNil)
	c.Assert(eq, Equals, true)

	err = destBucket.DeleteObject(destObjectName)
	c.Assert(err, IsNil)
	os.Remove(newFile)

	// Delete target bucket
	ForceDeleteBucket(s.client, destBucketName, c)
}

func (s *OssCopySuite) TestVersioningCopyFileCrossBucket(c *C) {
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
	destBucketName := bucketName + "-desc"
	srcObjectName := objectNamePrefix + RandStr(8)
	destObjectName := srcObjectName + "-dest"

	// Create dest bucket
	err = client.CreateBucket(destBucketName)
	c.Assert(err, IsNil)
	destBucket, err := client.Bucket(destBucketName)
	c.Assert(err, IsNil)

	err = client.SetBucketVersioning(destBucketName, versioningConfig)
	c.Assert(err, IsNil)

	// Upload source file
	var respHeader http.Header
	options := []Option{Routines(3), GetResponseHeader(&respHeader)}
	err = bucket.UploadFile(srcObjectName, fileName, 100*1024, options...)
	versionId := GetVersionId(respHeader)
	c.Assert(len(versionId) > 0, Equals, true)

	c.Assert(err, IsNil)
	os.Remove(newFile)

	// overwrite emtpy object
	err = bucket.PutObject(srcObjectName, strings.NewReader(""))
	c.Assert(err, IsNil)

	// Copy files
	var respCopyHeader http.Header
	options = []Option{Routines(5), Checkpoint(true, destObjectName+".cp"), GetResponseHeader(&respCopyHeader), VersionId(versionId)}
	err = destBucket.CopyFile(bucketName, srcObjectName, destObjectName, 1024*100, options...)
	c.Assert(err, IsNil)
	versionIdCopy := GetVersionId(respCopyHeader)
	c.Assert(len(versionIdCopy) > 0, Equals, true)

	err = destBucket.GetObjectToFile(destObjectName, newFile, VersionId(versionIdCopy))
	c.Assert(err, IsNil)

	eq, err := compareFiles(fileName, newFile)
	c.Assert(err, IsNil)
	c.Assert(eq, Equals, true)

	err = destBucket.DeleteObject(destObjectName)
	c.Assert(err, IsNil)
	os.Remove(newFile)

	// Copy file with options meta
	options = []Option{Routines(10), Checkpoint(true, "copy.cp"), Meta("myprop", "mypropval"), GetResponseHeader(&respCopyHeader), VersionId(versionId)}
	err = destBucket.CopyFile(bucketName, srcObjectName, destObjectName, 1024*100, options...)
	c.Assert(err, IsNil)
	versionIdCopy = GetVersionId(respCopyHeader)

	err = destBucket.GetObjectToFile(destObjectName, newFile, VersionId(versionIdCopy))
	c.Assert(err, IsNil)

	eq, err = compareFiles(fileName, newFile)
	c.Assert(err, IsNil)
	c.Assert(eq, Equals, true)

	os.Remove(fileName)
	os.Remove(newFile)
	destBucket.DeleteObject(destObjectName)
	bucket.DeleteObject(objectName)
	ForceDeleteBucket(client, bucketName, c)
	ForceDeleteBucket(client, destBucketName, c)
}

// TestCopyFileChoiceOptions
func (s *OssCopySuite) TestCopyFileChoiceOptions(c *C) {
	destBucketName := s.bucket.BucketName + "-desc"
	srcObjectName := objectNamePrefix + RandStr(8)
	destObjectName := srcObjectName + "-dest"
	fileName := "../sample/BingWallpaper-2015-11-07.jpg"
	newFile := RandStr(8) + ".jpg"

	destBucket, err := s.client.Bucket(destBucketName)
	c.Assert(err, IsNil)

	// Create a target bucket
	err = s.client.CreateBucket(destBucketName)

	// Upload source file
	err = s.bucket.UploadFile(srcObjectName, fileName, 100*1024, Routines(3))
	c.Assert(err, IsNil)
	os.Remove(newFile)

	// copyfile with properties
	options := []Option{
		ObjectACL(ACLPublicRead),
		RequestPayer(Requester),
		TrafficLimitHeader(1024 * 1024 * 8),
		ObjectStorageClass(StorageArchive),
		ServerSideEncryption("AES256"),
		Routines(5), // without checkpoint
	}

	// Copy files
	err = destBucket.CopyFile(s.bucket.BucketName, srcObjectName, destObjectName, 1024*100, options...)
	c.Assert(err, IsNil)

	// check object
	meta, err := destBucket.GetObjectDetailedMeta(destObjectName)
	c.Assert(err, IsNil)
	c.Assert(meta.Get("X-Oss-Storage-Class"), Equals, "Archive")
	c.Assert(meta.Get("X-Oss-Server-Side-Encryption"), Equals, "AES256")

	aclResult, err := destBucket.GetObjectACL(destObjectName)
	c.Assert(aclResult.ACL, Equals, "public-read")
	c.Assert(err, IsNil)

	err = destBucket.DeleteObject(destObjectName)
	c.Assert(err, IsNil)
	os.Remove(newFile)

	// Copy file with options
	options = []Option{
		ObjectACL(ACLPublicRead),
		RequestPayer(Requester),
		TrafficLimitHeader(1024 * 1024 * 8),
		ObjectStorageClass(StorageArchive),
		ServerSideEncryption("AES256"),
		Routines(10),
		Checkpoint(true, "copy.cp"), // with checkpoint
	}

	err = destBucket.CopyFile(s.bucket.BucketName, srcObjectName, destObjectName, 1024*100, options...)
	c.Assert(err, IsNil)

	// check object
	meta, err = destBucket.GetObjectDetailedMeta(destObjectName)
	c.Assert(err, IsNil)
	c.Assert(meta.Get("X-Oss-Storage-Class"), Equals, "Archive")
	c.Assert(meta.Get("X-Oss-Server-Side-Encryption"), Equals, "AES256")

	aclResult, err = destBucket.GetObjectACL(destObjectName)
	c.Assert(aclResult.ACL, Equals, "public-read")
	c.Assert(err, IsNil)

	err = destBucket.DeleteObject(destObjectName)
	c.Assert(err, IsNil)
	os.Remove(newFile)

	// Delete target bucket
	err = s.client.DeleteBucket(destBucketName)
	c.Assert(err, IsNil)
}
