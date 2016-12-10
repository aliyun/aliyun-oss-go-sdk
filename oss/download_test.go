package oss

import (
	"fmt"
	"os"
	"time"

	. "gopkg.in/check.v1"
)

type OssDownloadSuite struct {
	client *Client
	bucket *Bucket
}

var _ = Suite(&OssDownloadSuite{})

// Run once when the suite starts running
func (s *OssDownloadSuite) SetUpSuite(c *C) {
	client, err := New(endpoint, accessID, accessKey)
	c.Assert(err, IsNil)
	s.client = client

	s.client.CreateBucket(bucketName)
	time.Sleep(5 * time.Second)

	bucket, err := s.client.Bucket(bucketName)
	c.Assert(err, IsNil)
	s.bucket = bucket

	testLogger.Println("test download started")
}

// Run before each test or benchmark starts running
func (s *OssDownloadSuite) TearDownSuite(c *C) {
	// Delete Part
	lmur, err := s.bucket.ListMultipartUploads()
	c.Assert(err, IsNil)

	for _, upload := range lmur.Uploads {
		var imur = InitiateMultipartUploadResult{Bucket: s.bucket.BucketName,
			Key: upload.Key, UploadID: upload.UploadID}
		err = s.bucket.AbortMultipartUpload(imur)
		c.Assert(err, IsNil)
	}

	// Delete Objects
	lor, err := s.bucket.ListObjects()
	c.Assert(err, IsNil)

	for _, object := range lor.Objects {
		err = s.bucket.DeleteObject(object.Key)
		c.Assert(err, IsNil)
	}

	testLogger.Println("test download completed")
}

// Run after each test or benchmark runs
func (s *OssDownloadSuite) SetUpTest(c *C) {
	err := removeTempFiles("../oss", ".jpg")
	c.Assert(err, IsNil)
}

// Run once after all tests or benchmarks have finished running
func (s *OssDownloadSuite) TearDownTest(c *C) {
	err := removeTempFiles("../oss", ".jpg")
	c.Assert(err, IsNil)

	err = removeTempFiles("../oss", ".temp")
	c.Assert(err, IsNil)
}

// TestUploadRoutineWithoutRecovery 多线程无断点恢复的下载
func (s *OssDownloadSuite) TestDownloadRoutineWithoutRecovery(c *C) {
	objectName := objectNamePrefix + "tdrwr"
	fileName := "../sample/BingWallpaper-2015-11-07.jpg"
	newFile := "down-new-file.jpg"

	// 上传文件
	err := s.bucket.UploadFile(objectName, fileName, 100*1024, Routines(3))
	c.Assert(err, IsNil)

	// 使用默认值下载
	err = s.bucket.DownloadFile(objectName, newFile, 100*1024)
	c.Assert(err, IsNil)

	// check
	eq, err := compareFiles(fileName, newFile)
	c.Assert(err, IsNil)
	c.Assert(eq, Equals, true)

	// 使用2个协程下载，小于总分片数5
	os.Remove(newFile)
	err = s.bucket.DownloadFile(objectName, newFile, 100*1024, Routines(2))
	c.Assert(err, IsNil)

	// check
	eq, err = compareFiles(fileName, newFile)
	c.Assert(err, IsNil)
	c.Assert(eq, Equals, true)

	// 使用5个协程下载，等于总分片数5
	os.Remove(newFile)
	err = s.bucket.DownloadFile(objectName, newFile, 100*1024, Routines(5))
	c.Assert(err, IsNil)

	// check
	eq, err = compareFiles(fileName, newFile)
	c.Assert(err, IsNil)
	c.Assert(eq, Equals, true)

	// 使用10个协程下载，大于总分片数5
	os.Remove(newFile)
	err = s.bucket.DownloadFile(objectName, newFile, 100*1024, Routines(10))
	c.Assert(err, IsNil)

	// check
	eq, err = compareFiles(fileName, newFile)
	c.Assert(err, IsNil)
	c.Assert(eq, Equals, true)

	err = s.bucket.DeleteObject(objectName)
	c.Assert(err, IsNil)
}

// ErrorHooker DownloadPart请求Hook
func DownErrorHooker(part downloadPart) error {
	if part.Index == 4 {
		time.Sleep(time.Second)
		return fmt.Errorf("ErrorHooker")
	}
	return nil
}

// TestDownloadRoutineWithRecovery 多线程有断点恢复的下载
func (s *OssDownloadSuite) TestDownloadRoutineWithRecovery(c *C) {
	objectName := objectNamePrefix + "tdrtr"
	fileName := "../sample/BingWallpaper-2015-11-07.jpg"
	newFile := "down-new-file-2.jpg"

	// 上传文件
	err := s.bucket.UploadFile(objectName, fileName, 100*1024, Routines(3))
	c.Assert(err, IsNil)

	// 下载，CP使用默认值
	downloadPartHooker = DownErrorHooker
	err = s.bucket.DownloadFile(objectName, newFile, 100*1024, Checkpoint(true, ""))
	c.Assert(err, NotNil)
	c.Assert(err.Error(), Equals, "ErrorHooker")
	downloadPartHooker = defaultDownloadPartHook

	// check
	dcp := downloadCheckpoint{}
	err = dcp.load(newFile + ".cp")
	c.Assert(err, IsNil)
	c.Assert(dcp.Magic, Equals, downloadCpMagic)
	c.Assert(len(dcp.MD5), Equals, len("LC34jZU5xK4hlxi3Qn3XGQ=="))
	c.Assert(dcp.FilePath, Equals, newFile)
	c.Assert(dcp.ObjStat.Size, Equals, int64(482048))
	c.Assert(len(dcp.ObjStat.LastModified), Equals, len("2015-12-17 18:43:03 +0800 CST"))
	c.Assert(dcp.ObjStat.Etag, Equals, "\"2351E662233817A7AE974D8C5B0876DD-5\"")
	c.Assert(dcp.Object, Equals, objectName)
	c.Assert(len(dcp.Parts), Equals, 5)
	c.Assert(len(dcp.todoParts()), Equals, 1)

	err = s.bucket.DownloadFile(objectName, newFile, 100*1024, Checkpoint(true, ""))
	c.Assert(err, IsNil)

	err = dcp.load(newFile + ".cp")
	c.Assert(err, NotNil)

	eq, err := compareFiles(fileName, newFile)
	c.Assert(err, IsNil)
	c.Assert(eq, Equals, true)

	// 下载，指定CP
	os.Remove(newFile)
	downloadPartHooker = DownErrorHooker
	err = s.bucket.DownloadFile(objectName, newFile, 100*1024, Checkpoint(true, objectName+".cp"))
	c.Assert(err, NotNil)
	c.Assert(err.Error(), Equals, "ErrorHooker")
	downloadPartHooker = defaultDownloadPartHook

	// check
	dcp = downloadCheckpoint{}
	err = dcp.load(objectName + ".cp")
	c.Assert(err, IsNil)
	c.Assert(dcp.Magic, Equals, downloadCpMagic)
	c.Assert(len(dcp.MD5), Equals, len("LC34jZU5xK4hlxi3Qn3XGQ=="))
	c.Assert(dcp.FilePath, Equals, newFile)
	c.Assert(dcp.ObjStat.Size, Equals, int64(482048))
	c.Assert(len(dcp.ObjStat.LastModified), Equals, len("2015-12-17 18:43:03 +0800 CST"))
	c.Assert(dcp.ObjStat.Etag, Equals, "\"2351E662233817A7AE974D8C5B0876DD-5\"")
	c.Assert(dcp.Object, Equals, objectName)
	c.Assert(len(dcp.Parts), Equals, 5)
	c.Assert(len(dcp.todoParts()), Equals, 1)

	err = s.bucket.DownloadFile(objectName, newFile, 100*1024, Checkpoint(true, objectName+".cp"))
	c.Assert(err, IsNil)

	err = dcp.load(objectName + ".cp")
	c.Assert(err, NotNil)

	eq, err = compareFiles(fileName, newFile)
	c.Assert(err, IsNil)
	c.Assert(eq, Equals, true)

	// 一次完成下载，中间没有错误
	os.Remove(newFile)
	err = s.bucket.DownloadFile(objectName, newFile, 100*1024, Checkpoint(true, ""))
	c.Assert(err, IsNil)

	err = dcp.load(newFile + ".cp")
	c.Assert(err, NotNil)

	eq, err = compareFiles(fileName, newFile)
	c.Assert(err, IsNil)
	c.Assert(eq, Equals, true)

	// 一次完成下载，中间没有错误
	os.Remove(newFile)
	err = s.bucket.DownloadFile(objectName, newFile, 100*1024, Routines(10), Checkpoint(true, ""))
	c.Assert(err, IsNil)

	err = dcp.load(newFile + ".cp")
	c.Assert(err, NotNil)

	eq, err = compareFiles(fileName, newFile)
	c.Assert(err, IsNil)
	c.Assert(eq, Equals, true)

	err = s.bucket.DeleteObject(objectName)
	c.Assert(err, IsNil)
}

// TestDownloadOption 选项
func (s *OssDownloadSuite) TestDownloadOption(c *C) {
	objectName := objectNamePrefix + "tdmo"
	fileName := "../sample/BingWallpaper-2015-11-07.jpg"
	newFile := "down-new-file-3.jpg"

	// 上传文件
	err := s.bucket.UploadFile(objectName, fileName, 100*1024, Routines(3))
	c.Assert(err, IsNil)

	meta, err := s.bucket.GetObjectDetailedMeta(objectName)
	c.Assert(err, IsNil)

	// IfMatch
	os.Remove(newFile)
	err = s.bucket.DownloadFile(objectName, newFile, 100*1024, Routines(3), IfMatch(meta.Get("Etag")))
	c.Assert(err, IsNil)

	eq, err := compareFiles(fileName, newFile)
	c.Assert(err, IsNil)
	c.Assert(eq, Equals, true)

	// IfNoneMatch
	os.Remove(newFile)
	err = s.bucket.DownloadFile(objectName, newFile, 100*1024, Routines(3), IfNoneMatch(meta.Get("Etag")))
	c.Assert(err, NotNil)

	// IfMatch
	err = s.bucket.DownloadFile(objectName, newFile, 100*1024, Routines(3), Checkpoint(true, ""), IfMatch(meta.Get("Etag")))
	c.Assert(err, IsNil)

	eq, err = compareFiles(fileName, newFile)
	c.Assert(err, IsNil)
	c.Assert(eq, Equals, true)

	// IfNoneMatch
	err = s.bucket.DownloadFile(objectName, newFile, 100*1024, Routines(3), Checkpoint(true, ""), IfNoneMatch(meta.Get("Etag")))
	c.Assert(err, NotNil)
}

// TestDownloadObjectChange 上传过程中文件修改了
func (s *OssDownloadSuite) TestDownloadObjectChange(c *C) {
	objectName := objectNamePrefix + "tdloc"
	fileName := "../sample/BingWallpaper-2015-11-07.jpg"
	newFile := "down-new-file-4.jpg"

	// 上传文件
	err := s.bucket.UploadFile(objectName, fileName, 100*1024, Routines(3))
	c.Assert(err, IsNil)

	// 下载，CP使用默认值
	downloadPartHooker = DownErrorHooker
	err = s.bucket.DownloadFile(objectName, newFile, 100*1024, Checkpoint(true, ""))
	c.Assert(err, NotNil)
	c.Assert(err.Error(), Equals, "ErrorHooker")
	downloadPartHooker = defaultDownloadPartHook

	err = s.bucket.UploadFile(objectName, fileName, 100*1024, Routines(3))
	c.Assert(err, IsNil)

	err = s.bucket.DownloadFile(objectName, newFile, 100*1024, Checkpoint(true, ""))
	c.Assert(err, IsNil)

	eq, err := compareFiles(fileName, newFile)
	c.Assert(err, IsNil)
	c.Assert(eq, Equals, true)
}

// TestDownloadNegative Download Negative
func (s *OssDownloadSuite) TestDownloadNegative(c *C) {
	objectName := objectNamePrefix + "tdn"
	fileName := "../sample/BingWallpaper-2015-11-07.jpg"
	newFile := "down-new-file-3.jpg"

	// 上传文件
	err := s.bucket.UploadFile(objectName, fileName, 100*1024, Routines(3))
	c.Assert(err, IsNil)

	// worker线程错误
	downloadPartHooker = DownErrorHooker
	err = s.bucket.DownloadFile(objectName, newFile, 100*1024, Routines(2))
	c.Assert(err, NotNil)
	c.Assert(err.Error(), Equals, "ErrorHooker")
	downloadPartHooker = defaultDownloadPartHook

	// 本地文件不存在
	err = s.bucket.DownloadFile(objectName, "/tmp/", 100*1024, Routines(2))
	c.Assert(err, NotNil)

	// 指定的分片大小无效
	err = s.bucket.DownloadFile(objectName, newFile, 0, Routines(2))
	c.Assert(err, NotNil)

	err = s.bucket.DownloadFile(objectName, newFile, 1024*1024*1024*100, Routines(2))
	c.Assert(err, NotNil)

	err = s.bucket.DeleteObject(objectName)
	c.Assert(err, IsNil)

	// 本地文件不存在
	err = s.bucket.DownloadFile(objectName, "/tmp/", 100*1024, Checkpoint(true, ""))
	c.Assert(err, NotNil)

	err = s.bucket.DownloadFile(objectName, "/tmp/", 100*1024, Routines(2), Checkpoint(true, ""))
	c.Assert(err, NotNil)

	// 指定的分片大小无效
	err = s.bucket.DownloadFile(objectName, newFile, -1, Checkpoint(true, ""))
	c.Assert(err, NotNil)

	err = s.bucket.DownloadFile(objectName, newFile, 0, Routines(2), Checkpoint(true, ""))
	c.Assert(err, NotNil)

	err = s.bucket.DownloadFile(objectName, newFile, 1024*1024*1024*100, Checkpoint(true, ""))
	c.Assert(err, NotNil)

	err = s.bucket.DownloadFile(objectName, newFile, 1024*1024*1024*100, Routines(2), Checkpoint(true, ""))
	c.Assert(err, NotNil)
}
