package oss

import (
	"crypto/md5"
	"encoding/base64"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"time"

	. "gopkg.in/check.v1"
)

type OssCrcSuite struct {
	client *Client
	bucket *Bucket
}

var _ = Suite(&OssCrcSuite{})

// Run once when the suite starts running
func (s *OssCrcSuite) SetUpSuite(c *C) {
	client, err := New(endpoint, accessID, accessKey)
	c.Assert(err, IsNil)
	s.client = client

	s.client.CreateBucket(bucketName)
	time.Sleep(5 * time.Second)

	bucket, err := s.client.Bucket(bucketName)
	c.Assert(err, IsNil)
	s.bucket = bucket

	testLogger.Println("test crc started")
}

// Run before each test or benchmark starts running
func (s *OssCrcSuite) TearDownSuite(c *C) {
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

	testLogger.Println("test crc completed")
}

// Run after each test or benchmark runs
func (s *OssCrcSuite) SetUpTest(c *C) {
	err := removeTempFiles("../oss", ".jpg")
	c.Assert(err, IsNil)
}

// Run once after all tests or benchmarks have finished running
func (s *OssCrcSuite) TearDownTest(c *C) {
	err := removeTempFiles("../oss", ".jpg")
	c.Assert(err, IsNil)
}

// TestEnableCRCAndMD5 开启MD5和CRC校验
func (s *OssCrcSuite) TestEnableCRCAndMD5(c *C) {
	objectName := objectNamePrefix + "tecam"
	fileName := "../sample/BingWallpaper-2015-11-07.jpg"
	newFileName := "BingWallpaper-2015-11-07-2.jpg"
	objectValue := "大江东去，浪淘尽，千古风流人物。 故垒西边，人道是、三国周郎赤壁。 乱石穿空，惊涛拍岸，卷起千堆雪。 江山如画，一时多少豪杰。" +
		"遥想公谨当年，小乔初嫁了，雄姿英发。 羽扇纶巾，谈笑间、樯橹灰飞烟灭。故国神游，多情应笑我，早生华发，人生如梦，一尊还酹江月。"

	client, err := New(endpoint, accessID, accessKey, EnableCRC(true), EnableMD5(true), MD5ThresholdCalcInMemory(200*1024))
	c.Assert(err, IsNil)
	bucket, err := client.Bucket(bucketName)
	c.Assert(err, IsNil)

	// PutObject
	err = bucket.PutObject(objectName, strings.NewReader(objectValue))
	c.Assert(err, IsNil)

	// GetObject
	body, err := bucket.GetObject(objectName)
	c.Assert(err, IsNil)
	_, err = ioutil.ReadAll(body)
	c.Assert(err, IsNil)
	body.Close()

	// GetObjectWithCRC
	body, calcCRC, srvCRC, err := bucket.GetObjectWithCRC(objectName)
	c.Assert(err, IsNil)
	str, err := readBody(body)
	c.Assert(err, IsNil)
	c.Assert(str, Equals, objectValue)
	c.Assert(calcCRC.Sum64(), Equals, srvCRC)

	// PutObjectFromFile
	err = bucket.PutObjectFromFile(objectName, fileName)
	c.Assert(err, IsNil)

	// GetObjectToFile
	err = bucket.GetObjectToFile(objectName, newFileName)
	c.Assert(err, IsNil)
	eq, err := compareFiles(fileName, newFileName)
	c.Assert(err, IsNil)
	c.Assert(eq, Equals, true)

	// DeleteObject
	err = bucket.DeleteObject(objectName)
	c.Assert(err, IsNil)

	// AppendObject
	var nextPos int64
	nextPos, err = s.bucket.AppendObject(objectName, strings.NewReader(objectValue), nextPos)
	c.Assert(err, IsNil)

	err = s.bucket.DeleteObject(objectName)
	c.Assert(err, IsNil)

	//	MultipartUpload
	chunks, err := SplitFileByPartSize(fileName, 100*1024)
	imurUpload, err := bucket.InitiateMultipartUpload(objectName)
	c.Assert(err, IsNil)
	var partsUpload []UploadPart

	for _, chunk := range chunks {
		part, err := bucket.UploadPartFromFile(imurUpload, fileName, chunk.Offset, chunk.Size, (int)(chunk.Number))
		c.Assert(err, IsNil)
		partsUpload = append(partsUpload, part)
	}

	_, err = bucket.CompleteMultipartUpload(imurUpload, partsUpload)
	c.Assert(err, IsNil)

	// Check MultipartUpload
	err = bucket.GetObjectToFile(objectName, newFileName)
	c.Assert(err, IsNil)
	eq, err = compareFiles(fileName, newFileName)
	c.Assert(err, IsNil)
	c.Assert(eq, Equals, true)

	// DeleteObjects
	_, err = bucket.DeleteObjects([]string{objectName})
	c.Assert(err, IsNil)
}

// TestDisableCRCAndMD5 关闭MD5和CRC校验
func (s *OssCrcSuite) TestDisableCRCAndMD5(c *C) {
	objectName := objectNamePrefix + "tdcam"
	fileName := "../sample/BingWallpaper-2015-11-07.jpg"
	newFileName := "BingWallpaper-2015-11-07-3.jpg"
	objectValue := "大江东去，浪淘尽，千古风流人物。 故垒西边，人道是、三国周郎赤壁。 乱石穿空，惊涛拍岸，卷起千堆雪。 江山如画，一时多少豪杰。" +
		"遥想公谨当年，小乔初嫁了，雄姿英发。 羽扇纶巾，谈笑间、樯橹灰飞烟灭。故国神游，多情应笑我，早生华发，人生如梦，一尊还酹江月。"

	client, err := New(endpoint, accessID, accessKey, EnableCRC(false), EnableMD5(false))
	c.Assert(err, IsNil)
	bucket, err := client.Bucket(bucketName)
	c.Assert(err, IsNil)

	// PutObject
	err = bucket.PutObject(objectName, strings.NewReader(objectValue))
	c.Assert(err, IsNil)

	// GetObject
	body, err := bucket.GetObject(objectName)
	c.Assert(err, IsNil)
	_, err = ioutil.ReadAll(body)
	c.Assert(err, IsNil)
	body.Close()

	// GetObjectWithCRC
	body, _, _, err = bucket.GetObjectWithCRC(objectName)
	c.Assert(err, IsNil)
	str, err := readBody(body)
	c.Assert(err, IsNil)
	c.Assert(str, Equals, objectValue)

	// PutObjectFromFile
	err = bucket.PutObjectFromFile(objectName, fileName)
	c.Assert(err, IsNil)

	// GetObjectToFile
	err = bucket.GetObjectToFile(objectName, newFileName)
	c.Assert(err, IsNil)
	eq, err := compareFiles(fileName, newFileName)
	c.Assert(err, IsNil)
	c.Assert(eq, Equals, true)

	// DeleteObject
	err = bucket.DeleteObject(objectName)
	c.Assert(err, IsNil)

	// AppendObject
	var nextPos int64
	nextPos, err = s.bucket.AppendObject(objectName, strings.NewReader(objectValue), nextPos)
	c.Assert(err, IsNil)

	err = s.bucket.DeleteObject(objectName)
	c.Assert(err, IsNil)

	//	MultipartUpload
	chunks, err := SplitFileByPartSize(fileName, 100*1024)
	imurUpload, err := bucket.InitiateMultipartUpload(objectName)
	c.Assert(err, IsNil)
	var partsUpload []UploadPart

	for _, chunk := range chunks {
		part, err := bucket.UploadPartFromFile(imurUpload, fileName, chunk.Offset, chunk.Size, (int)(chunk.Number))
		c.Assert(err, IsNil)
		partsUpload = append(partsUpload, part)
	}

	_, err = bucket.CompleteMultipartUpload(imurUpload, partsUpload)
	c.Assert(err, IsNil)

	// Check MultipartUpload
	err = bucket.GetObjectToFile(objectName, newFileName)
	c.Assert(err, IsNil)
	eq, err = compareFiles(fileName, newFileName)
	c.Assert(err, IsNil)
	c.Assert(eq, Equals, true)

	// DeleteObjects
	_, err = bucket.DeleteObjects([]string{objectName})
	c.Assert(err, IsNil)
}

// TestDisableCRCAndMD5 关闭MD5和CRC校验
func (s *OssCrcSuite) TestSpecifyContentMD5(c *C) {
	objectName := objectNamePrefix + "tdcam"
	fileName := "../sample/BingWallpaper-2015-11-07.jpg"
	objectValue := "大江东去，浪淘尽，千古风流人物。 故垒西边，人道是、三国周郎赤壁。 乱石穿空，惊涛拍岸，卷起千堆雪。 江山如画，一时多少豪杰。" +
		"遥想公谨当年，小乔初嫁了，雄姿英发。 羽扇纶巾，谈笑间、樯橹灰飞烟灭。故国神游，多情应笑我，早生华发，人生如梦，一尊还酹江月。"

	mh := md5.Sum([]byte(objectValue))
	md5B64 := base64.StdEncoding.EncodeToString(mh[:])

	// PutObject
	err := s.bucket.PutObject(objectName, strings.NewReader(objectValue), ContentMD5(md5B64))
	c.Assert(err, IsNil)

	// PutObjectFromFile
	file, err := os.Open(fileName)
	md5 := md5.New()
	io.Copy(md5, file)
	mdHex := base64.StdEncoding.EncodeToString(md5.Sum(nil)[:])
	err = s.bucket.PutObjectFromFile(objectName, fileName, ContentMD5(mdHex))
	c.Assert(err, IsNil)

	err = s.bucket.DeleteObject(objectName)
	c.Assert(err, IsNil)

	// AppendObject
	var nextPos int64
	nextPos, err = s.bucket.AppendObject(objectName, strings.NewReader(objectValue), nextPos, ContentMD5(md5B64))
	c.Assert(err, IsNil)

	err = s.bucket.DeleteObject(objectName)
	c.Assert(err, IsNil)

	//	MultipartUpload
	imurUpload, err := s.bucket.InitiateMultipartUpload(objectName)
	c.Assert(err, IsNil)

	var partsUpload []UploadPart
	part, err := s.bucket.UploadPart(imurUpload, strings.NewReader(objectValue), (int64)(len([]byte(objectValue))), 1)
	c.Assert(err, IsNil)
	partsUpload = append(partsUpload, part)

	_, err = s.bucket.CompleteMultipartUpload(imurUpload, partsUpload)
	c.Assert(err, IsNil)

	// DeleteObject
	err = s.bucket.DeleteObject(objectName)
	c.Assert(err, IsNil)
}
