package oss

import (
	"fmt"
	"os"
	"time"

	. "gopkg.in/check.v1"
)

type OssCopySuite struct {
	client *Client
	bucket *Bucket
}

var _ = Suite(&OssCopySuite{})

// SetUpSuite runs once when the suite starts running
func (s *OssCopySuite) SetUpSuite(c *C) {
	client, err := New(endpoint, accessID, accessKey)
	c.Assert(err, IsNil)
	s.client = client

	s.client.CreateBucket(bucketName)
	time.Sleep(5 * time.Second)

	bucket, err := s.client.Bucket(bucketName)
	c.Assert(err, IsNil)
	s.bucket = bucket

	testLogger.Println("test copy started")
}

// TearDownSuite runs before each test or benchmark starts running
func (s *OssCopySuite) TearDownSuite(c *C) {
	// Delete Part
	lmur, err := s.bucket.ListMultipartUploads()
	c.Assert(err, IsNil)

	for _, upload := range lmur.Uploads {
		var imur = InitiateMultipartUploadResult{Bucket: bucketName,
			Key: upload.Key, UploadID: upload.UploadID}
		err = s.bucket.AbortMultipartUpload(imur)
		c.Assert(err, IsNil)
	}

	// Delete objects
	lor, err := s.bucket.ListObjects()
	c.Assert(err, IsNil)

	for _, object := range lor.Objects {
		err = s.bucket.DeleteObject(object.Key)
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
	srcObjectName := objectNamePrefix + "tcrwr"
	destObjectName := srcObjectName + "-copy"
	fileName := "../sample/BingWallpaper-2015-11-07.jpg"
	newFile := "copy-new-file.jpg"

	// Upload source file
	err := s.bucket.UploadFile(srcObjectName, fileName, 100*1024, Routines(3))
	c.Assert(err, IsNil)
	os.Remove(newFile)

	// Does not specify parameter 'routines', by default it's single routine
	err = s.bucket.CopyFile(bucketName, srcObjectName, destObjectName, 100*1024)
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
	err = s.bucket.CopyFile(bucketName, srcObjectName, destObjectName, 100*1024, Routines(1))
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
	err = s.bucket.CopyFile(bucketName, srcObjectName, destObjectName, 100*1024, Routines(3))
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
	err = s.bucket.CopyFile(bucketName, srcObjectName, destObjectName, 100*1024, Routines(5))
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
	err = s.bucket.CopyFile(bucketName, srcObjectName, destObjectName, 100*1024, Routines(10))
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
	err = s.bucket.CopyFile(bucketName, srcObjectName, destObjectName, 100*1024, Routines(-1))
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
	err = s.bucket.CopyFile(bucketName, srcObjectName, destObjectName, 100*1024, Routines(3), Meta("myprop", "mypropval"))

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
	srcObjectName := objectNamePrefix + "tcrwrn"
	destObjectName := srcObjectName + "-copy"
	fileName := "../sample/BingWallpaper-2015-11-07.jpg"

	// Upload source file
	err := s.bucket.UploadFile(srcObjectName, fileName, 100*1024, Routines(3))
	c.Assert(err, IsNil)

	copyPartHooker = CopyErrorHooker
	// Worker routine errors
	err = s.bucket.CopyFile(bucketName, srcObjectName, destObjectName, 100*1024, Routines(2))

	c.Assert(err, NotNil)
	c.Assert(err.Error(), Equals, "ErrorHooker")
	copyPartHooker = defaultCopyPartHook

	// Source bucket does not exist
	err = s.bucket.CopyFile("NotExist", srcObjectName, destObjectName, 100*1024, Routines(2))
	c.Assert(err, NotNil)

	// Target object does not exist
	err = s.bucket.CopyFile(bucketName, "NotExist", destObjectName, 100*1024, Routines(2))

	// The part size is invalid
	err = s.bucket.CopyFile(bucketName, srcObjectName, destObjectName, 1024, Routines(2))
	c.Assert(err, NotNil)

	err = s.bucket.CopyFile(bucketName, srcObjectName, destObjectName, 1024*1024*1024*100, Routines(2))
	c.Assert(err, NotNil)

	// Delete the source file
	err = s.bucket.DeleteObject(srcObjectName)
	c.Assert(err, IsNil)
}

// TestCopyRoutineWithRecovery is a multiple routines copy with resumable recovery
func (s *OssCopySuite) TestCopyRoutineWithRecovery(c *C) {
	srcObjectName := objectNamePrefix + "tcrtr"
	destObjectName := srcObjectName + "-copy"
	fileName := "../sample/BingWallpaper-2015-11-07.jpg"
	newFile := "copy-new-file.jpg"

	// Upload source file
	err := s.bucket.UploadFile(srcObjectName, fileName, 100*1024, Routines(3))
	c.Assert(err, IsNil)
	os.Remove(newFile)

	// Routines default value, CP's default path is destObjectName+.cp
	// Copy object with checkpoint enabled, single runtine.
	// Copy 4 parts---the CopyErrorHooker makes sure the copy of part 5 will fail.
	copyPartHooker = CopyErrorHooker
	err = s.bucket.CopyFile(bucketName, srcObjectName, destObjectName, 1024*100, Checkpoint(true, ""))
	c.Assert(err, NotNil)
	c.Assert(err.Error(), Equals, "ErrorHooker")
	copyPartHooker = defaultCopyPartHook

	// Check CP
	ccp := copyCheckpoint{}
	err = ccp.load(destObjectName + ".cp")
	c.Assert(err, IsNil)
	c.Assert(ccp.Magic, Equals, copyCpMagic)
	c.Assert(len(ccp.MD5), Equals, len("LC34jZU5xK4hlxi3Qn3XGQ=="))
	c.Assert(ccp.SrcBucketName, Equals, bucketName)
	c.Assert(ccp.SrcObjectKey, Equals, srcObjectName)
	c.Assert(ccp.DestBucketName, Equals, bucketName)
	c.Assert(ccp.DestObjectKey, Equals, destObjectName)
	c.Assert(len(ccp.CopyID), Equals, len("3F79722737D1469980DACEDCA325BB52"))
	c.Assert(ccp.ObjStat.Size, Equals, int64(482048))
	c.Assert(len(ccp.ObjStat.LastModified), Equals, len("2015-12-17 18:43:03 +0800 CST"))
	c.Assert(ccp.ObjStat.Etag, Equals, "\"2351E662233817A7AE974D8C5B0876DD-5\"")
	c.Assert(len(ccp.Parts), Equals, 5)
	c.Assert(len(ccp.todoParts()), Equals, 1)
	c.Assert(ccp.PartStat[4], Equals, false)

	// Second copy, finish the last part
	err = s.bucket.CopyFile(bucketName, srcObjectName, destObjectName, 1024*100, Checkpoint(true, ""))
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

	// Specify Routine and CP's path
	copyPartHooker = CopyErrorHooker
	err = s.bucket.CopyFile(bucketName, srcObjectName, destObjectName, 1024*100, Routines(2), Checkpoint(true, srcObjectName+".cp"))
	c.Assert(err, NotNil)
	c.Assert(err.Error(), Equals, "ErrorHooker")
	copyPartHooker = defaultCopyPartHook

	// Check CP
	ccp = copyCheckpoint{}
	err = ccp.load(srcObjectName + ".cp")
	c.Assert(err, IsNil)
	c.Assert(ccp.Magic, Equals, copyCpMagic)
	c.Assert(len(ccp.MD5), Equals, len("LC34jZU5xK4hlxi3Qn3XGQ=="))
	c.Assert(ccp.SrcBucketName, Equals, bucketName)
	c.Assert(ccp.SrcObjectKey, Equals, srcObjectName)
	c.Assert(ccp.DestBucketName, Equals, bucketName)
	c.Assert(ccp.DestObjectKey, Equals, destObjectName)
	c.Assert(len(ccp.CopyID), Equals, len("3F79722737D1469980DACEDCA325BB52"))
	c.Assert(ccp.ObjStat.Size, Equals, int64(482048))
	c.Assert(len(ccp.ObjStat.LastModified), Equals, len("2015-12-17 18:43:03 +0800 CST"))
	c.Assert(ccp.ObjStat.Etag, Equals, "\"2351E662233817A7AE974D8C5B0876DD-5\"")
	c.Assert(len(ccp.Parts), Equals, 5)
	c.Assert(len(ccp.todoParts()), Equals, 1)
	c.Assert(ccp.PartStat[4], Equals, false)

	// Second copy, finish the last part.
	err = s.bucket.CopyFile(bucketName, srcObjectName, destObjectName, 1024*100, Routines(2), Checkpoint(true, srcObjectName+".cp"))
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
	err = s.bucket.CopyFile(bucketName, srcObjectName, destObjectName, 1024*100, Routines(3), Checkpoint(true, ""))
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
	err = s.bucket.CopyFile(bucketName, srcObjectName, destObjectName, 1024*100, Routines(10), Checkpoint(true, ""))
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
	err = s.bucket.CopyFile(bucketName, srcObjectName, destObjectName, 1024*100, Routines(5), Checkpoint(true, ""), Meta("myprop", "mypropval"))
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
	srcObjectName := objectNamePrefix + "tcrwrn"
	destObjectName := srcObjectName + "-copy"

	// Source bucket does not exist
	err := s.bucket.CopyFile("NotExist", srcObjectName, destObjectName, 100*1024, Checkpoint(true, ""))
	c.Assert(err, NotNil)
	c.Assert(err, NotNil)

	// Source object does not exist
	err = s.bucket.CopyFile(bucketName, "NotExist", destObjectName, 100*1024, Routines(2), Checkpoint(true, ""))
	c.Assert(err, NotNil)

	// Specify part size is invalid.
	err = s.bucket.CopyFile(bucketName, srcObjectName, destObjectName, 1024, Checkpoint(true, ""))
	c.Assert(err, NotNil)

	err = s.bucket.CopyFile(bucketName, srcObjectName, destObjectName, 1024*1024*1024*100, Routines(2), Checkpoint(true, ""))
	c.Assert(err, NotNil)
}

// TestCopyFileCrossBucket is a cross bucket's direct copy.
func (s *OssCopySuite) TestCopyFileCrossBucket(c *C) {
	destBucketName := bucketName + "-cfcb-desc"
	srcObjectName := objectNamePrefix + "tcrtr"
	destObjectName := srcObjectName + "-copy"
	fileName := "../sample/BingWallpaper-2015-11-07.jpg"
	newFile := "copy-new-file.jpg"

	destBucket, err := s.client.Bucket(destBucketName)
	c.Assert(err, IsNil)

	// Create a target bucket
	err = s.client.CreateBucket(destBucketName)

	// Upload source file
	err = s.bucket.UploadFile(srcObjectName, fileName, 100*1024, Routines(3))
	c.Assert(err, IsNil)
	os.Remove(newFile)

	// Copy files
	err = destBucket.CopyFile(bucketName, srcObjectName, destObjectName, 1024*100, Routines(5), Checkpoint(true, ""))
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
	err = destBucket.CopyFile(bucketName, srcObjectName, destObjectName, 1024*100, Routines(10), Checkpoint(true, "copy.cp"), Meta("myprop", "mypropval"))
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
	err = s.client.DeleteBucket(destBucketName)
	c.Assert(err, IsNil)
}
