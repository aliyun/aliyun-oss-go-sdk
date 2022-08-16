package oss

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	. "gopkg.in/check.v1"
)

type OssDownloadSuite struct {
	cloudBoxControlClient *Client
	client                *Client
	bucket                *Bucket
}

var _ = Suite(&OssDownloadSuite{})

// SetUpSuite runs once when the suite starts running
func (s *OssDownloadSuite) SetUpSuite(c *C) {
	bucketName := bucketNamePrefix + RandLowStr(6)
	if cloudboxControlEndpoint == "" {
		client, err := New(endpoint, accessID, accessKey)
		c.Assert(err, IsNil)
		s.client = client

		s.client.CreateBucket(bucketName)

		bucket, err := s.client.Bucket(bucketName)
		c.Assert(err, IsNil)
		s.bucket = bucket

		testLogger.Println("test crc started")
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

	testLogger.Println("test download started")
}

// TearDownSuite runs before each test or benchmark starts running
func (s *OssDownloadSuite) TearDownSuite(c *C) {
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

	testLogger.Println("test download completed")
}

// SetUpTest runs after each test or benchmark runs
func (s *OssDownloadSuite) SetUpTest(c *C) {
	err := removeTempFiles("../oss", ".jpg")
	c.Assert(err, IsNil)
}

// TearDownTest runs once after all tests or benchmarks have finished running
func (s *OssDownloadSuite) TearDownTest(c *C) {
	err := removeTempFiles("../oss", ".jpg")
	c.Assert(err, IsNil)

	err = removeTempFiles("../oss", ".temp")
	c.Assert(err, IsNil)
}

// TestDownloadRoutineWithoutRecovery multipart downloads without checkpoint
func (s *OssDownloadSuite) TestDownloadRoutineWithoutRecovery(c *C) {
	objectName := objectNamePrefix + RandStr(8)
	fileName := "../sample/BingWallpaper-2015-11-07.jpg"
	newFile := RandStr(8) + ".jpg"

	// Upload a file
	err := s.bucket.UploadFile(objectName, fileName, 100*1024, Routines(3))
	c.Assert(err, IsNil)

	// Download the file by default
	err = s.bucket.DownloadFile(objectName, newFile, 100*1024)
	c.Assert(err, IsNil)

	// Check
	eq, err := compareFiles(fileName, newFile)
	c.Assert(err, IsNil)
	c.Assert(eq, Equals, true)

	// Use 2 coroutines to download the file and total parts count is 5
	os.Remove(newFile)
	err = s.bucket.DownloadFile(objectName, newFile, 100*1024, Routines(2))
	c.Assert(err, IsNil)

	// Check
	eq, err = compareFiles(fileName, newFile)
	c.Assert(err, IsNil)
	c.Assert(eq, Equals, true)

	// Use 5 coroutines to download the file and the total parts count is 5.
	os.Remove(newFile)
	err = s.bucket.DownloadFile(objectName, newFile, 100*1024, Routines(5))
	c.Assert(err, IsNil)

	// Check
	eq, err = compareFiles(fileName, newFile)
	c.Assert(err, IsNil)
	c.Assert(eq, Equals, true)

	// Use 10 coroutines to download the file and the total parts count is 5.
	os.Remove(newFile)
	err = s.bucket.DownloadFile(objectName, newFile, 100*1024, Routines(10))
	c.Assert(err, IsNil)

	// Check
	eq, err = compareFiles(fileName, newFile)
	c.Assert(err, IsNil)
	c.Assert(eq, Equals, true)

	err = s.bucket.DeleteObject(objectName)
	c.Assert(err, IsNil)
}

// DownErrorHooker requests hook by downloadPart
func DownErrorHooker(part downloadPart) error {
	if part.Index == 4 {
		time.Sleep(time.Second)
		return fmt.Errorf("ErrorHooker")
	}
	return nil
}

// TestDownloadRoutineWithRecovery multi-routine resumable download
func (s *OssDownloadSuite) TestDownloadRoutineWithRecovery(c *C) {
	objectName := objectNamePrefix + RandStr(8)
	fileName := "../sample/BingWallpaper-2015-11-07.jpg"
	newFile := RandStr(8) + ".jpg"

	// Upload a file
	err := s.bucket.UploadFile(objectName, fileName, 100*1024, Routines(3))
	c.Assert(err, IsNil)

	// Download a file with default checkpoint
	downloadPartHooker = DownErrorHooker
	err = s.bucket.DownloadFile(objectName, newFile, 100*1024, Checkpoint(true, newFile+".cp"))
	c.Assert(err, NotNil)
	c.Assert(err.Error(), Equals, "ErrorHooker")
	downloadPartHooker = defaultDownloadPartHook

	// Check
	dcp := downloadCheckpoint{}
	err = dcp.load(newFile + ".cp")
	c.Assert(err, IsNil)
	c.Assert(dcp.Magic, Equals, downloadCpMagic)
	c.Assert(len(dcp.MD5), Equals, len("LC34jZU5xK4hlxi3Qn3XGQ=="))
	c.Assert(dcp.FilePath, Equals, newFile)
	c.Assert(dcp.ObjStat.Size, Equals, int64(482048))
	c.Assert(len(dcp.ObjStat.LastModified) > 0, Equals, true)
	c.Assert(dcp.ObjStat.Etag, Equals, "\"2351E662233817A7AE974D8C5B0876DD-5\"")
	c.Assert(dcp.Object, Equals, objectName)
	c.Assert(len(dcp.Parts), Equals, 5)
	c.Assert(len(dcp.todoParts()), Equals, 1)

	err = s.bucket.DownloadFile(objectName, newFile, 100*1024, Checkpoint(true, newFile+".cp"))
	c.Assert(err, IsNil)
	//download success, checkpoint file has been deleted
	err = dcp.load(newFile + ".cp")
	c.Assert(err, NotNil)

	eq, err := compareFiles(fileName, newFile)
	c.Assert(err, IsNil)
	c.Assert(eq, Equals, true)

	// Resumable download with empty checkpoint file path
	downloadPartHooker = DownErrorHooker
	err = s.bucket.DownloadFile(objectName, newFile, 100*1024, Checkpoint(true, ""))
	c.Assert(err, NotNil)
	c.Assert(err.Error(), Equals, "ErrorHooker")
	downloadPartHooker = defaultDownloadPartHook

	dcp = downloadCheckpoint{}
	err = dcp.load(newFile + ".cp")
	c.Assert(err, NotNil)

	// Resumable download with checkpoint dir
	os.Remove(newFile)
	downloadPartHooker = DownErrorHooker
	err = s.bucket.DownloadFile(objectName, newFile, 100*1024, CheckpointDir(true, "./"))
	c.Assert(err, NotNil)
	c.Assert(err.Error(), Equals, "ErrorHooker")
	downloadPartHooker = defaultDownloadPartHook

	// Check
	dcp = downloadCheckpoint{}
	cpConf := cpConfig{IsEnable: true, DirPath: "./"}
	cpFilePath := getDownloadCpFilePath(&cpConf, s.bucket.BucketName, objectName, "", newFile)
	err = dcp.load(cpFilePath)
	c.Assert(err, IsNil)
	c.Assert(dcp.Magic, Equals, downloadCpMagic)
	c.Assert(len(dcp.MD5), Equals, len("LC34jZU5xK4hlxi3Qn3XGQ=="))
	c.Assert(dcp.FilePath, Equals, newFile)
	c.Assert(dcp.ObjStat.Size, Equals, int64(482048))
	c.Assert(len(dcp.ObjStat.LastModified) > 0, Equals, true)
	c.Assert(dcp.ObjStat.Etag, Equals, "\"2351E662233817A7AE974D8C5B0876DD-5\"")
	c.Assert(dcp.Object, Equals, objectName)
	c.Assert(len(dcp.Parts), Equals, 5)
	c.Assert(len(dcp.todoParts()), Equals, 1)

	err = s.bucket.DownloadFile(objectName, newFile, 100*1024, CheckpointDir(true, "./"))
	c.Assert(err, IsNil)
	//download success, checkpoint file has been deleted
	err = dcp.load(cpFilePath)
	c.Assert(err, NotNil)

	eq, err = compareFiles(fileName, newFile)
	c.Assert(err, IsNil)
	c.Assert(eq, Equals, true)

	// Resumable download with checkpoint at a time. No error is expected in the download procedure.
	os.Remove(newFile)
	err = s.bucket.DownloadFile(objectName, newFile, 100*1024, Checkpoint(true, newFile+".cp"))
	c.Assert(err, IsNil)

	err = dcp.load(newFile + ".cp")
	c.Assert(err, NotNil)

	eq, err = compareFiles(fileName, newFile)
	c.Assert(err, IsNil)
	c.Assert(eq, Equals, true)

	// Resumable download with checkpoint at a time. No error is expected in the download procedure.
	os.Remove(newFile)
	err = s.bucket.DownloadFile(objectName, newFile, 100*1024, Routines(10), Checkpoint(true, newFile+".cp"))
	c.Assert(err, IsNil)

	err = dcp.load(newFile + ".cp")
	c.Assert(err, NotNil)

	eq, err = compareFiles(fileName, newFile)
	c.Assert(err, IsNil)
	c.Assert(eq, Equals, true)

	err = s.bucket.DeleteObject(objectName)
	c.Assert(err, IsNil)
}

// TestDownloadOption options
func (s *OssDownloadSuite) TestDownloadOption(c *C) {
	objectName := objectNamePrefix + RandStr(8)
	fileName := "../sample/BingWallpaper-2015-11-07.jpg"
	newFile := RandStr(8) + ".jpg"

	// Upload the file
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
	err = s.bucket.DownloadFile(objectName, newFile, 100*1024, Routines(3), IfMatch(meta.Get("Etag")))
	c.Assert(err, IsNil)

	eq, err = compareFiles(fileName, newFile)
	c.Assert(err, IsNil)
	c.Assert(eq, Equals, true)

	// IfNoneMatch
	err = s.bucket.DownloadFile(objectName, newFile, 100*1024, Routines(3), IfNoneMatch(meta.Get("Etag")))
	c.Assert(err, NotNil)
}

// TestDownloadObjectChange tests the file is updated during the upload
func (s *OssDownloadSuite) TestDownloadObjectChange(c *C) {
	objectName := objectNamePrefix + RandStr(8)
	fileName := "../sample/BingWallpaper-2015-11-07.jpg"
	newFile := RandStr(8) + ".jpg"

	// Upload a file
	err := s.bucket.UploadFile(objectName, fileName, 100*1024, Routines(3))
	c.Assert(err, IsNil)

	// Download with default checkpoint
	downloadPartHooker = DownErrorHooker
	err = s.bucket.DownloadFile(objectName, newFile, 100*1024, Checkpoint(true, newFile+".cp"))
	c.Assert(err, NotNil)
	c.Assert(err.Error(), Equals, "ErrorHooker")
	downloadPartHooker = defaultDownloadPartHook

	err = s.bucket.UploadFile(objectName, fileName, 100*1024, Routines(3))
	c.Assert(err, IsNil)

	err = s.bucket.DownloadFile(objectName, newFile, 100*1024, Checkpoint(true, newFile+".cp"))
	c.Assert(err, IsNil)

	eq, err := compareFiles(fileName, newFile)
	c.Assert(err, IsNil)
	c.Assert(eq, Equals, true)
}

// TestDownloadNegative tests downloading negative
func (s *OssDownloadSuite) TestDownloadNegative(c *C) {
	objectName := objectNamePrefix + RandStr(8)
	fileName := "../sample/BingWallpaper-2015-11-07.jpg"
	newFile := RandStr(8) + ".jpg"

	// Upload a file
	err := s.bucket.UploadFile(objectName, fileName, 100*1024, Routines(3))
	c.Assert(err, IsNil)

	// Worker routine error
	downloadPartHooker = DownErrorHooker
	err = s.bucket.DownloadFile(objectName, newFile, 100*1024, Routines(2))
	c.Assert(err, NotNil)
	c.Assert(err.Error(), Equals, "ErrorHooker")
	downloadPartHooker = defaultDownloadPartHook

	// Local file does not exist
	err = s.bucket.DownloadFile(objectName, "/tmp/", 100*1024, Routines(2))
	c.Assert(err, NotNil)

	// Invalid part size
	err = s.bucket.DownloadFile(objectName, newFile, 0, Routines(2))
	c.Assert(err, NotNil)

	err = s.bucket.DownloadFile(objectName, newFile, 1024*1024*1024*100, Routines(2))
	c.Assert(err, IsNil)

	err = s.bucket.DeleteObject(objectName)
	c.Assert(err, IsNil)

	// Local file does not exist
	err = s.bucket.DownloadFile(objectName, "/tmp/", 100*1024)
	c.Assert(err, NotNil)

	err = s.bucket.DownloadFile(objectName, "/tmp/", 100*1024, Routines(2))
	c.Assert(err, NotNil)

	// Invalid part size
	err = s.bucket.DownloadFile(objectName, newFile, -1)
	c.Assert(err, NotNil)

	err = s.bucket.DownloadFile(objectName, newFile, 0, Routines(2))
	c.Assert(err, NotNil)

	err = s.bucket.DownloadFile(objectName, newFile, 1024*1024*1024*100)
	c.Assert(err, NotNil)

	err = s.bucket.DownloadFile(objectName, newFile, 1024*1024*1024*100, Routines(2))
	c.Assert(err, NotNil)
}

// TestDownloadWithRange tests concurrent downloading with range specified and checkpoint enabled
func (s *OssDownloadSuite) TestDownloadWithRange(c *C) {
	objectName := objectNamePrefix + RandStr(8)
	fileName := "../sample/BingWallpaper-2015-11-07.jpg"
	newFile := RandStr(8) + ".jpg"
	newFileGet := RandStr(8) + "-.jpg"

	// Upload a file
	err := s.bucket.UploadFile(objectName, fileName, 100*1024, Routines(3))
	c.Assert(err, IsNil)

	fileSize, err := getFileSize(fileName)
	c.Assert(err, IsNil)

	// Download with range, from 1024 to 4096
	os.Remove(newFile)
	err = s.bucket.DownloadFile(objectName, newFile, 100*1024, Routines(3), Range(1024, 4095))
	c.Assert(err, IsNil)

	// Check
	eq, err := compareFilesWithRange(fileName, 1024, newFile, 0, 3072)
	c.Assert(err, IsNil)
	c.Assert(eq, Equals, true)

	os.Remove(newFileGet)
	err = s.bucket.GetObjectToFile(objectName, newFileGet, Range(1024, 4095))
	c.Assert(err, IsNil)

	// Compare get and download
	eq, err = compareFiles(newFile, newFileGet)
	c.Assert(err, IsNil)
	c.Assert(eq, Equals, true)

	// Download with range, from 1024 to 4096
	os.Remove(newFile)
	err = s.bucket.DownloadFile(objectName, newFile, 1024, Routines(3), NormalizedRange("1024-4095"))
	c.Assert(err, IsNil)

	// Check
	eq, err = compareFilesWithRange(fileName, 1024, newFile, 0, 3072)
	c.Assert(err, IsNil)
	c.Assert(eq, Equals, true)

	os.Remove(newFileGet)
	err = s.bucket.GetObjectToFile(objectName, newFileGet, NormalizedRange("1024-4095"))
	c.Assert(err, IsNil)

	// Compare get and download
	eq, err = compareFiles(newFile, newFileGet)
	c.Assert(err, IsNil)
	c.Assert(eq, Equals, true)

	// Download with range, from 2048 to the end
	os.Remove(newFile)
	err = s.bucket.DownloadFile(objectName, newFile, 1024*1024, Routines(3), NormalizedRange("2048-"))
	c.Assert(err, IsNil)

	// Check
	eq, err = compareFilesWithRange(fileName, 2048, newFile, 0, fileSize-2048)
	c.Assert(err, IsNil)
	c.Assert(eq, Equals, true)

	os.Remove(newFileGet)
	err = s.bucket.GetObjectToFile(objectName, newFileGet, NormalizedRange("2048-"))
	c.Assert(err, IsNil)

	// Compare get and download
	eq, err = compareFiles(newFile, newFileGet)
	c.Assert(err, IsNil)
	c.Assert(eq, Equals, true)

	// Download with range, the last 4096
	os.Remove(newFile)
	err = s.bucket.DownloadFile(objectName, newFile, 1024, Routines(3), NormalizedRange("-4096"))
	c.Assert(err, IsNil)

	// Check
	eq, err = compareFilesWithRange(fileName, fileSize-4096, newFile, 0, 4096)
	c.Assert(err, IsNil)
	c.Assert(eq, Equals, true)

	os.Remove(newFileGet)
	err = s.bucket.GetObjectToFile(objectName, newFileGet, NormalizedRange("-4096"))
	c.Assert(err, IsNil)

	// Compare get and download
	eq, err = compareFiles(newFile, newFileGet)
	c.Assert(err, IsNil)
	c.Assert(eq, Equals, true)

	err = s.bucket.DeleteObject(objectName)
	c.Assert(err, IsNil)
}

// TestDownloadWithCheckoutAndRange tests concurrent downloading with range specified and checkpoint enabled
func (s *OssDownloadSuite) TestDownloadWithCheckoutAndRange(c *C) {
	objectName := objectNamePrefix + RandStr(8)
	fileName := "../sample/BingWallpaper-2015-11-07.jpg"
	newFile := RandStr(8) + ".jpg"
	newFileGet := RandStr(8) + "-get.jpg"

	// Upload a file
	err := s.bucket.UploadFile(objectName, fileName, 100*1024, Routines(3), Checkpoint(true, fileName+".cp"))
	c.Assert(err, IsNil)

	fileSize, err := getFileSize(fileName)
	c.Assert(err, IsNil)

	// Download with range, from 1024 to 4096
	os.Remove(newFile)
	err = s.bucket.DownloadFile(objectName, newFile, 100*1024, Routines(3), Checkpoint(true, newFile+".cp"), Range(1024, 4095))
	c.Assert(err, IsNil)

	// Check
	eq, err := compareFilesWithRange(fileName, 1024, newFile, 0, 3072)
	c.Assert(err, IsNil)
	c.Assert(eq, Equals, true)

	os.Remove(newFileGet)
	err = s.bucket.GetObjectToFile(objectName, newFileGet, Range(1024, 4095))
	c.Assert(err, IsNil)

	// Compare get and download
	eq, err = compareFiles(newFile, newFileGet)
	c.Assert(err, IsNil)
	c.Assert(eq, Equals, true)

	// Download with range, from 1024 to 4096
	os.Remove(newFile)
	err = s.bucket.DownloadFile(objectName, newFile, 1024, Routines(3), Checkpoint(true, newFile+".cp"), NormalizedRange("1024-4095"))
	c.Assert(err, IsNil)

	// Check
	eq, err = compareFilesWithRange(fileName, 1024, newFile, 0, 3072)
	c.Assert(err, IsNil)
	c.Assert(eq, Equals, true)

	os.Remove(newFileGet)
	err = s.bucket.GetObjectToFile(objectName, newFileGet, NormalizedRange("1024-4095"))
	c.Assert(err, IsNil)

	// Compare get and download
	eq, err = compareFiles(newFile, newFileGet)
	c.Assert(err, IsNil)
	c.Assert(eq, Equals, true)

	// Download with range, from 2048 to the end
	os.Remove(newFile)
	err = s.bucket.DownloadFile(objectName, newFile, 1024*1024, Routines(3), Checkpoint(true, newFile+".cp"), NormalizedRange("2048-"))
	c.Assert(err, IsNil)

	// Check
	eq, err = compareFilesWithRange(fileName, 2048, newFile, 0, fileSize-2048)
	c.Assert(err, IsNil)
	c.Assert(eq, Equals, true)

	os.Remove(newFileGet)
	err = s.bucket.GetObjectToFile(objectName, newFileGet, NormalizedRange("2048-"))
	c.Assert(err, IsNil)

	// Compare get and download
	eq, err = compareFiles(newFile, newFileGet)
	c.Assert(err, IsNil)
	c.Assert(eq, Equals, true)

	// Download with range, the last 4096 bytes
	os.Remove(newFile)
	err = s.bucket.DownloadFile(objectName, newFile, 1024, Routines(3), Checkpoint(true, newFile+".cp"), NormalizedRange("-4096"))
	c.Assert(err, IsNil)

	// Check
	eq, err = compareFilesWithRange(fileName, fileSize-4096, newFile, 0, 4096)
	c.Assert(err, IsNil)
	c.Assert(eq, Equals, true)

	os.Remove(newFileGet)
	err = s.bucket.GetObjectToFile(objectName, newFileGet, NormalizedRange("-4096"))
	c.Assert(err, IsNil)

	// Compare get and download
	eq, err = compareFiles(newFile, newFileGet)
	c.Assert(err, IsNil)
	c.Assert(eq, Equals, true)

	err = s.bucket.DeleteObject(objectName)
	c.Assert(err, IsNil)
}

// TestCombineCRCInDownloadParts tests combineCRCInParts
func (s *OssDownloadSuite) TestCombineCRCInDownloadParts(c *C) {
	crc := combineCRCInParts(nil)
	c.Assert(crc == 0, Equals, true)

	crc = combineCRCInParts(make([]downloadPart, 0))
	c.Assert(crc == 0, Equals, true)

	parts := make([]downloadPart, 1)
	parts[0].CRC64 = 10278880121275185425
	crc = combineCRCInParts(parts)
	c.Assert(crc == 10278880121275185425, Equals, true)

	parts = make([]downloadPart, 2)
	parts[0].CRC64 = 6748440630437108969
	parts[0].Start = 0
	parts[0].End = 4
	parts[1].CRC64 = 10278880121275185425
	parts[1].Start = 5
	parts[1].End = 8
	crc = combineCRCInParts(parts)
	c.Assert(crc == 11051210869376104954, Equals, true)
}

func getFileSize(fileName string) (int64, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return 0, err
	}

	return stat.Size(), nil
}

// compareFilesWithRange compares the content between fileL and fileR with specified range
func compareFilesWithRange(fileL string, offsetL int64, fileR string, offsetR int64, size int64) (bool, error) {
	finL, err := os.Open(fileL)
	if err != nil {
		return false, err
	}
	defer finL.Close()
	finL.Seek(offsetL, os.SEEK_SET)

	finR, err := os.Open(fileR)
	if err != nil {
		return false, err
	}
	defer finR.Close()
	finR.Seek(offsetR, os.SEEK_SET)

	statL, err := finL.Stat()
	if err != nil {
		return false, err
	}

	statR, err := finR.Stat()
	if err != nil {
		return false, err
	}

	if (offsetL+size > statL.Size()) || (offsetR+size > statR.Size()) {
		return false, nil
	}

	part := statL.Size() - offsetL
	if part > 16*1024 {
		part = 16 * 1024
	}

	bufL := make([]byte, part)
	bufR := make([]byte, part)
	for readN := int64(0); readN < size; {
		n, _ := finL.Read(bufL)
		if 0 == n {
			break
		}

		n, _ = finR.Read(bufR)
		if 0 == n {
			break
		}

		tailer := part
		if tailer > size-readN {
			tailer = size - readN
		}
		readN += tailer

		if !bytes.Equal(bufL[0:tailer], bufR[0:tailer]) {
			return false, nil
		}
	}

	return true, nil
}

func (s *OssDownloadSuite) TestVersioningDownloadWithoutCheckPoint(c *C) {
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
	time.Sleep(timeoutInOperation)

	// begin test
	objectName := objectNamePrefix + RandStr(8)
	fileName := "test-file-" + RandStr(8)
	fileData := RandStr(500 * 1024)
	CreateFile(fileName, fileData, c)

	newFile := RandStr(8) + ".jpg"
	newFileGet := RandStr(8) + "-.jpg"

	// Upload a file
	var respHeader http.Header
	options := []Option{Routines(3), GetResponseHeader(&respHeader)}
	err = bucket.UploadFile(objectName, fileName, 100*1024, options...)
	c.Assert(err, IsNil)
	versionId := GetVersionId(respHeader)
	c.Assert(len(versionId) > 0, Equals, true)

	fileSize, err := getFileSize(fileName)
	c.Assert(err, IsNil)

	// overwrite emtpy object
	err = bucket.PutObject(objectName, strings.NewReader(""))
	c.Assert(err, IsNil)

	// Download with range, from 1024 to 4096
	os.Remove(newFile)
	options = []Option{Routines(3), Range(1024, 4095), VersionId(versionId)}
	err = bucket.DownloadFile(objectName, newFile, 100*1024, options...)
	c.Assert(err, IsNil)

	// Check
	eq, err := compareFilesWithRange(fileName, 1024, newFile, 0, 3072)
	c.Assert(err, IsNil)
	c.Assert(eq, Equals, true)

	os.Remove(newFileGet)
	options = []Option{Range(1024, 4095), VersionId(versionId)}
	err = bucket.GetObjectToFile(objectName, newFileGet, options...)
	c.Assert(err, IsNil)

	// Compare get and download
	eq, err = compareFiles(newFile, newFileGet)
	c.Assert(err, IsNil)
	c.Assert(eq, Equals, true)

	// Download with range, from 1024 to 4096
	os.Remove(newFile)
	options = []Option{Routines(3), NormalizedRange("1024-4095"), VersionId(versionId)}
	err = bucket.DownloadFile(objectName, newFile, 1024, options...)
	c.Assert(err, IsNil)

	// Check
	eq, err = compareFilesWithRange(fileName, 1024, newFile, 0, 3072)
	c.Assert(err, IsNil)
	c.Assert(eq, Equals, true)

	os.Remove(newFileGet)
	options = []Option{NormalizedRange("1024-4095"), VersionId(versionId)}
	err = bucket.GetObjectToFile(objectName, newFileGet, options...)
	c.Assert(err, IsNil)

	// Compare get and download
	eq, err = compareFiles(newFile, newFileGet)
	c.Assert(err, IsNil)
	c.Assert(eq, Equals, true)

	// Download with range, from 2048 to the end
	os.Remove(newFile)
	options = []Option{NormalizedRange("2048-"), VersionId(versionId)}
	err = bucket.DownloadFile(objectName, newFile, 1024*1024, options...)
	c.Assert(err, IsNil)

	// Check
	eq, err = compareFilesWithRange(fileName, 2048, newFile, 0, fileSize-2048)
	c.Assert(err, IsNil)
	c.Assert(eq, Equals, true)

	os.Remove(newFileGet)
	options = []Option{NormalizedRange("2048-"), VersionId(versionId)}
	err = bucket.GetObjectToFile(objectName, newFileGet, options...)
	c.Assert(err, IsNil)

	// Compare get and download
	eq, err = compareFiles(newFile, newFileGet)
	c.Assert(err, IsNil)
	c.Assert(eq, Equals, true)

	// Download with range, the last 4096
	os.Remove(newFile)
	options = []Option{Routines(3), NormalizedRange("-4096"), VersionId(versionId)}
	err = bucket.DownloadFile(objectName, newFile, 1024, options...)
	c.Assert(err, IsNil)

	// Check
	eq, err = compareFilesWithRange(fileName, fileSize-4096, newFile, 0, 4096)
	c.Assert(err, IsNil)
	c.Assert(eq, Equals, true)

	os.Remove(newFileGet)
	options = []Option{NormalizedRange("-4096"), VersionId(versionId)}
	err = bucket.GetObjectToFile(objectName, newFileGet, options...)
	c.Assert(err, IsNil)

	// Compare get and download
	eq, err = compareFiles(newFile, newFileGet)
	c.Assert(err, IsNil)
	c.Assert(eq, Equals, true)

	// download whole file
	os.Remove(newFileGet)
	options = []Option{Routines(3), VersionId(versionId)}
	err = bucket.GetObjectToFile(objectName, newFileGet, options...)
	c.Assert(err, IsNil)

	// Compare get and download
	eq, err = compareFiles(fileName, newFileGet)
	c.Assert(err, IsNil)
	c.Assert(eq, Equals, true)

	os.Remove(fileName)
	os.Remove(newFileGet)
	err = bucket.DeleteObject(objectName)
	c.Assert(err, IsNil)
	ForceDeleteBucket(client, bucketName, c)
}

func (s *OssDownloadSuite) TestVersioningDownloadWithCheckPoint(c *C) {
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
	time.Sleep(timeoutInOperation)

	// begin test
	objectName := objectNamePrefix + RandStr(8)
	fileName := "test-file-" + RandStr(8)
	fileData := RandStr(500 * 1024)
	CreateFile(fileName, fileData, c)
	newFile := RandStr(8) + ".jpg"

	// Upload a file
	var respHeader http.Header
	options := []Option{Routines(3), GetResponseHeader(&respHeader)}
	err = bucket.UploadFile(objectName, fileName, 100*1024, options...)
	c.Assert(err, IsNil)
	versionId := GetVersionId(respHeader)
	c.Assert(len(versionId) > 0, Equals, true)

	// Resumable download with checkpoint dir
	os.Remove(newFile)
	downloadPartHooker = DownErrorHooker
	options = []Option{CheckpointDir(true, "./"), VersionId(versionId)}

	strPayer := getPayer(options)
	c.Assert(strPayer, Equals, "")

	err = bucket.DownloadFile(objectName, newFile, 100*1024, options...)
	c.Assert(err, NotNil)
	c.Assert(err.Error(), Equals, "ErrorHooker")

	// download again
	downloadPartHooker = defaultDownloadPartHook
	options = []Option{CheckpointDir(true, "./"), VersionId(versionId), GetResponseHeader(&respHeader)}
	err = bucket.DownloadFile(objectName, newFile, 100*1024, options...)
	c.Assert(err, IsNil)
	c.Assert(GetVersionId(respHeader), Equals, versionId)

	eq, err := compareFiles(fileName, newFile)
	c.Assert(err, IsNil)
	c.Assert(eq, Equals, true)

	os.Remove(fileName)
	os.Remove(newFile)
	err = bucket.DeleteObject(objectName)
	c.Assert(err, IsNil)
	ForceDeleteBucket(client, bucketName, c)
}

func (s *OssDownloadSuite) TestdownloadFileChoiceOptions(c *C) {
	// create a bucket with default proprety
	client, err := New(endpoint, accessID, accessKey)
	c.Assert(err, IsNil)

	bucketName := bucketNamePrefix + RandLowStr(6)
	err = client.CreateBucket(bucketName)
	c.Assert(err, IsNil)

	bucket, err := client.Bucket(bucketName)

	// begin test
	objectName := objectNamePrefix + RandStr(8)
	fileName := "test-file-" + RandStr(8)
	fileData := RandStr(500 * 1024)
	CreateFile(fileName, fileData, c)
	newFile := RandStr(8) + ".jpg"

	// Upload a file
	var respHeader http.Header
	options := []Option{Routines(3), GetResponseHeader(&respHeader)}
	err = bucket.UploadFile(objectName, fileName, 100*1024, options...)
	c.Assert(err, IsNil)

	// Resumable download with checkpoint dir
	os.Remove(newFile)

	// downloadFile with properties
	options = []Option{
		ObjectACL(ACLPublicRead),
		RequestPayer(Requester),
		TrafficLimitHeader(1024 * 1024 * 8),
	}

	err = bucket.DownloadFile(objectName, newFile, 100*1024, options...)
	c.Assert(err, IsNil)

	eq, err := compareFiles(fileName, newFile)
	c.Assert(err, IsNil)
	c.Assert(eq, Equals, true)

	os.Remove(fileName)
	os.Remove(newFile)
	err = bucket.DeleteObject(objectName)
	c.Assert(err, IsNil)
	ForceDeleteBucket(client, bucketName, c)
}

func (s *OssDownloadSuite) TestdownloadFileWithCpChoiceOptions(c *C) {
	// create a bucket with default proprety
	client, err := New(endpoint, accessID, accessKey)
	c.Assert(err, IsNil)

	bucketName := bucketNamePrefix + RandLowStr(6)
	err = client.CreateBucket(bucketName)
	c.Assert(err, IsNil)

	bucket, err := client.Bucket(bucketName)

	// begin test
	objectName := objectNamePrefix + RandStr(8)
	fileName := "test-file-" + RandStr(8)
	fileData := RandStr(500 * 1024)
	CreateFile(fileName, fileData, c)
	newFile := RandStr(8) + ".jpg"

	// Upload a file
	var respHeader http.Header
	options := []Option{Routines(3), GetResponseHeader(&respHeader)}
	err = bucket.UploadFile(objectName, fileName, 100*1024, options...)
	c.Assert(err, IsNil)

	// Resumable download with checkpoint dir
	os.Remove(newFile)

	// DownloadFile with properties
	options = []Option{
		ObjectACL(ACLPublicRead),
		RequestPayer(Requester),
		TrafficLimitHeader(1024 * 1024 * 8),
		CheckpointDir(true, "./"),
	}

	err = bucket.DownloadFile(objectName, newFile, 100*1024, options...)
	c.Assert(err, IsNil)

	eq, err := compareFiles(fileName, newFile)
	c.Assert(err, IsNil)
	c.Assert(eq, Equals, true)

	os.Remove(fileName)
	os.Remove(newFile)
	err = bucket.DeleteObject(objectName)
	c.Assert(err, IsNil)
	ForceDeleteBucket(client, bucketName, c)
}
