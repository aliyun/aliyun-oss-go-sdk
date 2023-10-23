// client test
// use gocheck, install gocheck to execute "go get gopkg.in/check.v1",
// see https://labix.org/gocheck

package oss

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	. "gopkg.in/check.v1"
)

// Test hooks up gocheck into the "go test" runner.
func Test(t *testing.T) {
	TestingT(t)
}

type OssClientSuite struct{}

var _ = Suite(&OssClientSuite{})

var (
	// Endpoint/ID/Key
	endpoint  = os.Getenv("OSS_TEST_ENDPOINT")
	accessID  = os.Getenv("OSS_TEST_ACCESS_KEY_ID")
	accessKey = os.Getenv("OSS_TEST_ACCESS_KEY_SECRET")
	accountID = os.Getenv("OSS_TEST_ACCOUNT_ID")

	// Proxy
	proxyHost   = os.Getenv("OSS_TEST_PROXY_HOST")
	proxyUser   = os.Getenv("OSS_TEST_PROXY_USER")
	proxyPasswd = os.Getenv("OSS_TEST_PROXY_PASSWORD")

	// STS
	stsaccessID  = os.Getenv("OSS_TEST_STS_ID")
	stsaccessKey = os.Getenv("OSS_TEST_STS_KEY")
	stsARN       = os.Getenv("OSS_TEST_STS_ARN")

	// Credential
	credentialAccessID  = os.Getenv("OSS_CREDENTIAL_KEY_ID")
	credentialAccessKey = os.Getenv("OSS_CREDENTIAL_KEY_SECRET")
	credentialUID       = os.Getenv("OSS_CREDENTIAL_UID")

	// cloud box endpoint
	cloudboxControlEndpoint = os.Getenv("OSS_TEST_CLOUDBOX_CONTROL_ENDPOINT")
	cloudboxEndpoint        = os.Getenv("OSS_TEST_CLOUDBOX_ENDPOINT")

	// for v4 signature
	envRegion = os.Getenv("OSS_TEST_REGION")

	// for cloud box ID
	cloudBoxID = os.Getenv("OSS_TEST_CLOUDBOX_ID")

	kmsID = os.Getenv("OSS_TEST_KMS_ID")
)

var (
	// prefix of bucket name for bucket ops test
	bucketNamePrefix = "go-sdk-test-bucket-"
	// bucket name for object ops test
	bucketName        = bucketNamePrefix + RandLowStr(6)
	archiveBucketName = bucketNamePrefix + "arch-" + RandLowStr(6)
	// object name for object ops test
	objectNamePrefix = "go-sdk-test-object-"
	// sts region is one and only hangzhou
	stsRegion = "cn-hangzhou"
	// Credentials
	credentialBucketName = bucketNamePrefix + RandLowStr(6)
)

var (
	logPath            = "go_sdk_test_" + time.Now().Format("20060102_150405") + ".log"
	testLogFile, _     = os.OpenFile(logPath, os.O_RDWR|os.O_CREATE, 0664)
	testLogger         = log.New(testLogFile, "", log.Ldate|log.Ltime|log.Lshortfile)
	letters            = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	timeoutInOperation = 3 * time.Second
)

// structs for replication get test
type GetResult struct {
	Rules []Rule `xml:"Rule"`
}

type Rule struct {
	Action                      string          `xml:"Action,omitempty"`                      // The replication action (ALL or PUT)
	ID                          string          `xml:"ID,omitempty"`                          // The rule ID
	Destination                 DestinationType `xml:"Destination"`                           // Container for storing target bucket information
	HistoricalObjectReplication string          `xml:"HistoricalObjectReplication,omitempty"` // Whether to copy copy historical data (enabled or not)
	Status                      string          `xml:"Status,omitempty"`                      // The replication status (starting, doing or closing)
}

type DestinationType struct {
	Bucket   string `xml:"Bucket"`
	Location string `xml:"Location"`
}

func RandStr(n int) string {
	b := make([]rune, n)
	randMarker := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := range b {
		b[i] = letters[randMarker.Intn(len(letters))]
	}
	return string(b)
}

func CreateFile(fileName, content string, c *C) {
	fout, err := os.Create(fileName)
	defer fout.Close()
	c.Assert(err, IsNil)
	_, err = fout.WriteString(content)
	c.Assert(err, IsNil)
}

func RandLowStr(n int) string {
	return strings.ToLower(RandStr(n))
}

func ForceDeleteBucket(client *Client, bucketName string, c *C) {
	bucket, err := client.Bucket(bucketName)
	c.Assert(err, IsNil)

	// Delete Object
	marker := Marker("")
	for {
		lor, err := bucket.ListObjects(marker)
		c.Assert(err, IsNil)
		for _, object := range lor.Objects {
			err = bucket.DeleteObject(object.Key)
			c.Assert(err, IsNil)
		}
		marker = Marker(lor.NextMarker)
		if !lor.IsTruncated {
			break
		}
	}

	// Delete Object Versions and DeleteMarks
	keyMarker := KeyMarker("")
	versionIdMarker := VersionIdMarker("")
	options := []Option{keyMarker, versionIdMarker}
	for {
		lor, err := bucket.ListObjectVersions(options...)
		if err != nil {
			break
		}

		for _, object := range lor.ObjectDeleteMarkers {
			err = bucket.DeleteObject(object.Key, VersionId(object.VersionId))
			c.Assert(err, IsNil)
		}

		for _, object := range lor.ObjectVersions {
			err = bucket.DeleteObject(object.Key, VersionId(object.VersionId))
			c.Assert(err, IsNil)
		}

		keyMarker = KeyMarker(lor.NextKeyMarker)
		versionIdMarker := VersionIdMarker(lor.NextVersionIdMarker)
		options = []Option{keyMarker, versionIdMarker}

		if !lor.IsTruncated {
			break
		}
	}

	// Delete Part
	keyMarker = KeyMarker("")
	uploadIDMarker := UploadIDMarker("")
	for {
		lmur, err := bucket.ListMultipartUploads(keyMarker, uploadIDMarker)
		c.Assert(err, IsNil)
		for _, upload := range lmur.Uploads {
			var imur = InitiateMultipartUploadResult{Bucket: bucketName,
				Key: upload.Key, UploadID: upload.UploadID}
			err = bucket.AbortMultipartUpload(imur)
			c.Assert(err, IsNil)
		}
		keyMarker = KeyMarker(lmur.NextKeyMarker)
		uploadIDMarker = UploadIDMarker(lmur.NextUploadIDMarker)
		if !lmur.IsTruncated {
			break
		}
	}

	// delete live channel
	strMarker := ""
	for {
		result, err := bucket.ListLiveChannel(Marker(strMarker))
		c.Assert(err, IsNil)

		for _, channel := range result.LiveChannel {
			err := bucket.DeleteLiveChannel(channel.Name)
			c.Assert(err, IsNil)
		}

		if result.IsTruncated {
			strMarker = result.NextMarker
		} else {
			break
		}
	}

	// Delete Bucket
	err = client.DeleteBucket(bucketName)
	c.Assert(err, IsNil)
}

// SetUpSuite runs once when the suite starts running
func (s *OssClientSuite) SetUpSuite(c *C) {
	client, err := New(endpoint, accessID, accessKey)
	c.Assert(err, IsNil)

	lbr, err := client.ListBuckets(Prefix(bucketNamePrefix), MaxKeys(1000))
	c.Assert(err, IsNil)

	for _, bucket := range lbr.Buckets {
		ForceDeleteBucket(client, bucket.Name, c)
	}
	time.Sleep(timeoutInOperation)
	testLogger.Println("test client started")
}

// TearDownSuite runs before each test or benchmark starts running
func (s *OssClientSuite) TearDownSuite(c *C) {
	client, err := New(endpoint, accessID, accessKey)
	c.Assert(err, IsNil)

	lbr, err := client.ListBuckets(Prefix(bucketNamePrefix), MaxKeys(1000))
	c.Assert(err, IsNil)

	for _, bucket := range lbr.Buckets {
		s.deleteBucket(client, bucket.Name, c)
	}
	time.Sleep(timeoutInOperation)
	testLogger.Println("test client completed")
}

func (s *OssClientSuite) deleteBucket(client *Client, bucketName string, c *C) {
	ForceDeleteBucket(client, bucketName, c)
}

// SetUpTest runs after each test or benchmark runs
func (s *OssClientSuite) SetUpTest(c *C) {
}

// TearDownTest runs once after all tests or benchmarks have finished running
func (s *OssClientSuite) TearDownTest(c *C) {
}

// TestCreateBucket
func (s *OssClientSuite) TestCreateBucket(c *C) {
	var bucketNameTest = bucketNamePrefix + RandLowStr(6)

	client, err := New(endpoint, accessID, accessKey)
	c.Assert(err, IsNil)

	// Create
	client.DeleteBucket(bucketNameTest)
	err = client.CreateBucket(bucketNameTest)
	c.Assert(err, IsNil)
	//sleep 3 seconds after create bucket
	time.Sleep(timeoutInOperation)

	// verify bucket is exist
	found, err := client.IsBucketExist(bucketNameTest)
	c.Assert(err, IsNil)
	c.Assert(found, Equals, true)

	res, err := client.GetBucketACL(bucketNameTest)
	c.Assert(err, IsNil)
	c.Assert(res.ACL, Equals, string(ACLPrivate))

	err = client.DeleteBucket(bucketNameTest)
	c.Assert(err, IsNil)
	time.Sleep(timeoutInOperation)

	// CreateBucket creates with ACLPublicRead
	err = client.CreateBucket(bucketNameTest, ACL(ACLPublicRead))
	c.Assert(err, IsNil)
	time.Sleep(timeoutInOperation)

	res, err = client.GetBucketACL(bucketNameTest)
	c.Assert(err, IsNil)
	c.Assert(res.ACL, Equals, string(ACLPublicRead))

	err = client.DeleteBucket(bucketNameTest)
	c.Assert(err, IsNil)
	time.Sleep(timeoutInOperation)

	// ACLPublicReadWrite
	err = client.CreateBucket(bucketNameTest, ACL(ACLPublicReadWrite))
	c.Assert(err, IsNil)
	time.Sleep(timeoutInOperation)

	res, err = client.GetBucketACL(bucketNameTest)
	c.Assert(err, IsNil)
	c.Assert(res.ACL, Equals, string(ACLPublicReadWrite))

	err = client.DeleteBucket(bucketNameTest)
	c.Assert(err, IsNil)
	time.Sleep(timeoutInOperation)

	// ACLPrivate
	err = client.CreateBucket(bucketNameTest, ACL(ACLPrivate))
	c.Assert(err, IsNil)
	time.Sleep(timeoutInOperation)

	res, err = client.GetBucketACL(bucketNameTest)
	c.Assert(err, IsNil)
	c.Assert(res.ACL, Equals, string(ACLPrivate))

	// Delete
	err = client.DeleteBucket(bucketNameTest)
	c.Assert(err, IsNil)

	// Create bucket with configuration and test GetBucketInfo
	for _, storage := range []StorageClassType{StorageStandard, StorageIA, StorageArchive, StorageColdArchive} {
		bucketNameTest := bucketNamePrefix + RandLowStr(6)
		err = client.CreateBucket(bucketNameTest, StorageClass(storage), ACL(ACLPublicRead))
		c.Assert(err, IsNil)
		time.Sleep(timeoutInOperation)

		res, err := client.GetBucketInfo(bucketNameTest)
		c.Assert(err, IsNil)
		c.Assert(res.BucketInfo.Name, Equals, bucketNameTest)
		c.Assert(res.BucketInfo.StorageClass, Equals, string(storage))
		c.Assert(res.BucketInfo.ACL, Equals, string(ACLPublicRead))

		// Delete
		err = client.DeleteBucket(bucketNameTest)
		c.Assert(err, IsNil)
	}

	// Error put bucket with configuration
	err = client.CreateBucket("ERRORBUCKETNAME", StorageClass(StorageArchive))
	c.Assert(err, NotNil)

	// Create bucket with configuration and test ListBuckets
	for _, storage := range []StorageClassType{StorageStandard, StorageIA, StorageArchive, StorageColdArchive} {
		bucketNameTest := bucketNamePrefix + RandLowStr(6)
		err = client.CreateBucket(bucketNameTest, StorageClass(storage))
		c.Assert(err, IsNil)
		time.Sleep(timeoutInOperation)

		res, err := client.GetBucketInfo(bucketNameTest)
		c.Assert(err, IsNil)
		c.Assert(res.BucketInfo.Name, Equals, bucketNameTest)
		c.Assert(res.BucketInfo.StorageClass, Equals, string(storage))

		// Delete
		err = client.DeleteBucket(bucketNameTest)
		c.Assert(err, IsNil)
	}
}

// TestCreateBucketWithServerEncryption
func (s *OssClientSuite) TestCreateBucketWithServerEncryption(c *C) {
	var bucketNameTest = bucketNamePrefix + RandLowStr(6)

	client, err := New(endpoint, accessID, accessKey)
	c.Assert(err, IsNil)

	// Create
	client.DeleteBucket(bucketNameTest)
	err = client.CreateBucket(bucketNameTest, ServerSideEncryption("KMS"), ServerSideDataEncryption("SM4"))
	c.Assert(err, IsNil)
	//sleep 3 seconds after create bucket
	time.Sleep(timeoutInOperation)

	// verify bucket is exist
	rs, err := client.GetBucketEncryption(bucketNameTest)
	c.Assert(err, IsNil)
	c.Assert(rs.SSEDefault.SSEAlgorithm, Equals, "KMS")
	c.Assert(rs.SSEDefault.KMSDataEncryption, Equals, "SM4")

	client.DeleteBucket(bucketNameTest)
	err = client.CreateBucket(bucketNameTest, ServerSideEncryption("AES256"))
	c.Assert(err, IsNil)

	rs, err = client.GetBucketEncryption(bucketNameTest)
	c.Assert(err, IsNil)
	c.Assert(rs.SSEDefault.KMSDataEncryption, Equals, "")
	c.Assert(rs.SSEDefault.SSEAlgorithm, Equals, "AES256")

	client.DeleteBucket(bucketNameTest)
}

func (s *OssClientSuite) TestCreateBucketRedundancyType(c *C) {
	bucketNameTest := bucketNamePrefix + RandLowStr(6)
	client, err := New(endpoint, accessID, accessKey)
	c.Assert(err, IsNil)

	// CreateBucket creates without property
	err = client.CreateBucket(bucketNameTest)
	c.Assert(err, IsNil)
	client.DeleteBucket(bucketNameTest)
	time.Sleep(timeoutInOperation)

	// CreateBucket creates with RedundancyZRS
	err = client.CreateBucket(bucketNameTest, RedundancyType(RedundancyZRS))
	c.Assert(err, IsNil)

	res, err := client.GetBucketInfo(bucketNameTest)
	c.Assert(err, IsNil)
	c.Assert(res.BucketInfo.RedundancyType, Equals, string(RedundancyZRS))
	client.DeleteBucket(bucketNameTest)
	time.Sleep(timeoutInOperation)

	// CreateBucket creates with RedundancyLRS
	err = client.CreateBucket(bucketNameTest, RedundancyType(RedundancyLRS))
	c.Assert(err, IsNil)

	res, err = client.GetBucketInfo(bucketNameTest)
	c.Assert(err, IsNil)
	c.Assert(res.BucketInfo.RedundancyType, Equals, string(RedundancyLRS))
	c.Assert(res.BucketInfo.StorageClass, Equals, string(StorageStandard))
	client.DeleteBucket(bucketNameTest)
	time.Sleep(timeoutInOperation)

	// CreateBucket creates with ACLPublicRead RedundancyZRS
	err = client.CreateBucket(bucketNameTest, ACL(ACLPublicRead), RedundancyType(RedundancyZRS))
	c.Assert(err, IsNil)

	res, err = client.GetBucketInfo(bucketNameTest)
	c.Assert(err, IsNil)
	c.Assert(res.BucketInfo.RedundancyType, Equals, string(RedundancyZRS))
	c.Assert(res.BucketInfo.ACL, Equals, string(ACLPublicRead))
	client.DeleteBucket(bucketNameTest)
}

// TestCreateBucketNegative
func (s *OssClientSuite) TestCreateBucketNegative(c *C) {
	client, err := New(endpoint, accessID, accessKey)
	c.Assert(err, IsNil)

	// Bucket name invalid
	err = client.CreateBucket("xx")
	c.Assert(err, NotNil)

	err = client.CreateBucket("XXXX")
	c.Assert(err, NotNil)
	testLogger.Println(err)

	err = client.CreateBucket("_bucket")
	c.Assert(err, NotNil)
	testLogger.Println(err)

	// ACL invalid
	err = client.CreateBucket(bucketNamePrefix+RandLowStr(6), ACL("InvaldAcl"))
	c.Assert(err, NotNil)
	testLogger.Println(err)
}

// TestDeleteBucket
func (s *OssClientSuite) TestDeleteBucket(c *C) {
	var bucketNameTest = bucketNamePrefix + RandLowStr(6)

	client, err := New(endpoint, accessID, accessKey)
	c.Assert(err, IsNil)

	// Create
	err = client.CreateBucket(bucketNameTest)
	c.Assert(err, IsNil)
	time.Sleep(timeoutInOperation)

	// Check
	found, err := client.IsBucketExist(bucketNameTest)
	c.Assert(err, IsNil)
	c.Assert(found, Equals, true)

	// Delete
	err = client.DeleteBucket(bucketNameTest)
	c.Assert(err, IsNil)
	time.Sleep(timeoutInOperation)

	// Check
	found, err = client.IsBucketExist(bucketNameTest)
	c.Assert(err, IsNil)
	c.Assert(found, Equals, false)

	err = client.DeleteBucket(bucketNameTest)
	c.Assert(err, NotNil)
}

// TestDeleteBucketNegative
func (s *OssClientSuite) TestDeleteBucketNegative(c *C) {
	var bucketNameTest = bucketNamePrefix + RandLowStr(6)

	client, err := New(endpoint, accessID, accessKey)
	c.Assert(err, IsNil)

	// Bucket name invalid
	err = client.DeleteBucket("xx")
	c.Assert(err, NotNil)

	err = client.DeleteBucket("XXXX")
	c.Assert(err, NotNil)

	err = client.DeleteBucket("_bucket")
	c.Assert(err, NotNil)

	// Delete no exist bucket
	err = client.DeleteBucket("notexist")
	c.Assert(err, NotNil)

	// No permission to delete, this ak/sk for js sdk
	err = client.CreateBucket(bucketNameTest)
	c.Assert(err, IsNil)

	accessID := "<accessKeyId>"
	accessKey := "<accessKeySecret>"
	clientOtherUser, err := New(endpoint, accessID, accessKey)
	c.Assert(err, IsNil)

	err = clientOtherUser.DeleteBucket(bucketNameTest)
	c.Assert(err, NotNil)

	err = client.DeleteBucket(bucketNameTest)
	c.Assert(err, IsNil)
}

// TestListBucket
func (s *OssClientSuite) TestListBucket(c *C) {
	var prefix = bucketNamePrefix + RandLowStr(6)
	var bucketNameLbOne = prefix + "tlb1"
	var bucketNameLbTwo = prefix + "tlb2"
	var bucketNameLbThree = prefix + "tlb3"

	client, err := New(endpoint, accessID, accessKey)
	c.Assert(err, IsNil)

	// CreateBucket
	err = client.CreateBucket(bucketNameLbOne)
	c.Assert(err, IsNil)
	err = client.CreateBucket(bucketNameLbTwo)
	c.Assert(err, IsNil)
	err = client.CreateBucket(bucketNameLbThree)
	c.Assert(err, IsNil)

	// ListBuckets, specified prefix
	var respHeader http.Header
	lbr, err := client.ListBuckets(Prefix(prefix), MaxKeys(2), GetResponseHeader(&respHeader))
	c.Assert(GetRequestId(respHeader) != "", Equals, true)
	c.Assert(err, IsNil)
	c.Assert(len(lbr.Buckets), Equals, 2)

	// ListBuckets, specified max keys
	lbr, err = client.ListBuckets(MaxKeys(2))
	c.Assert(err, IsNil)
	c.Assert(len(lbr.Buckets), Equals, 2)

	// ListBuckets, specified max keys
	lbr, err = client.ListBuckets(Marker(bucketNameLbOne), MaxKeys(1))
	c.Assert(err, IsNil)
	c.Assert(len(lbr.Buckets), Equals, 1)

	// ListBuckets, specified max keys
	lbr, err = client.ListBuckets(Marker(bucketNameLbOne))
	c.Assert(err, IsNil)
	c.Assert(len(lbr.Buckets) >= 2, Equals, true)

	// DeleteBucket
	err = client.DeleteBucket(bucketNameLbOne)
	c.Assert(err, IsNil)
	err = client.DeleteBucket(bucketNameLbTwo)
	c.Assert(err, IsNil)
	err = client.DeleteBucket(bucketNameLbThree)
	c.Assert(err, IsNil)
}

// TestListBucket
func (s *OssClientSuite) TestIsBucketExist(c *C) {
	var prefix = bucketNamePrefix + RandLowStr(6)
	var bucketNameLbOne = prefix + "tibe1"
	var bucketNameLbTwo = prefix + "tibe11"
	var bucketNameLbThree = prefix + "tibe111"

	client, err := New(endpoint, accessID, accessKey)
	c.Assert(err, IsNil)

	// CreateBucket
	err = client.CreateBucket(bucketNameLbOne)
	c.Assert(err, IsNil)
	err = client.CreateBucket(bucketNameLbTwo)
	c.Assert(err, IsNil)
	err = client.CreateBucket(bucketNameLbThree)
	c.Assert(err, IsNil)

	// Exist
	exist, err := client.IsBucketExist(bucketNameLbTwo)
	c.Assert(err, IsNil)
	c.Assert(exist, Equals, true)

	exist, err = client.IsBucketExist(bucketNameLbThree)
	c.Assert(err, IsNil)
	c.Assert(exist, Equals, true)

	exist, err = client.IsBucketExist(bucketNameLbOne)
	c.Assert(err, IsNil)
	c.Assert(exist, Equals, true)

	// Not exist
	exist, err = client.IsBucketExist(prefix + "tibe")
	c.Assert(err, IsNil)
	c.Assert(exist, Equals, false)

	exist, err = client.IsBucketExist(prefix + "tibe1111")
	c.Assert(err, IsNil)
	c.Assert(exist, Equals, false)

	// Negative
	exist, err = client.IsBucketExist("BucketNameInvalid")
	c.Assert(err, NotNil)

	// DeleteBucket
	err = client.DeleteBucket(bucketNameLbOne)
	c.Assert(err, IsNil)
	err = client.DeleteBucket(bucketNameLbTwo)
	c.Assert(err, IsNil)
	err = client.DeleteBucket(bucketNameLbThree)
	c.Assert(err, IsNil)
}

// TestSetBucketAcl
func (s *OssClientSuite) TestSetBucketAcl(c *C) {
	var bucketNameTest = bucketNamePrefix + RandLowStr(6)

	client, err := New(endpoint, accessID, accessKey)
	c.Assert(err, IsNil)

	// Private
	err = client.CreateBucket(bucketNameTest)
	c.Assert(err, IsNil)

	res, err := client.GetBucketACL(bucketNameTest)
	c.Assert(err, IsNil)
	c.Assert(res.ACL, Equals, string(ACLPrivate))

	// Set ACL_PUBLIC_R
	err = client.SetBucketACL(bucketNameTest, ACLPublicRead)
	c.Assert(err, IsNil)
	time.Sleep(timeoutInOperation)

	res, err = client.GetBucketACL(bucketNameTest)
	c.Assert(err, IsNil)
	c.Assert(res.ACL, Equals, string(ACLPublicRead))

	// Set ACL_PUBLIC_RW
	err = client.SetBucketACL(bucketNameTest, ACLPublicReadWrite)
	c.Assert(err, IsNil)
	time.Sleep(timeoutInOperation)

	res, err = client.GetBucketACL(bucketNameTest)
	c.Assert(err, IsNil)
	c.Assert(res.ACL, Equals, string(ACLPublicReadWrite))

	// Set ACL_PUBLIC_RW
	err = client.SetBucketACL(bucketNameTest, ACLPrivate)
	c.Assert(err, IsNil)
	time.Sleep(timeoutInOperation)
	time.Sleep(timeoutInOperation)

	res, err = client.GetBucketACL(bucketNameTest)
	c.Assert(err, IsNil)
	c.Assert(res.ACL, Equals, string(ACLPrivate))

	err = client.DeleteBucket(bucketNameTest)
	c.Assert(err, IsNil)
}

// TestSetBucketAclNegative
func (s *OssClientSuite) TestBucketAclNegative(c *C) {
	var bucketNameTest = bucketNamePrefix + RandLowStr(6)

	client, err := New(endpoint, accessID, accessKey)
	c.Assert(err, IsNil)

	err = client.CreateBucket(bucketNameTest)
	c.Assert(err, IsNil)

	err = client.SetBucketACL(bucketNameTest, "InvalidACL")
	c.Assert(err, NotNil)
	testLogger.Println(err)

	err = client.DeleteBucket(bucketNameTest)
	c.Assert(err, IsNil)
}

// TestGetBucketAcl
func (s *OssClientSuite) TestGetBucketAcl(c *C) {
	var bucketNameTest = bucketNamePrefix + RandLowStr(6)

	client, err := New(endpoint, accessID, accessKey)
	c.Assert(err, IsNil)

	// Private
	err = client.CreateBucket(bucketNameTest)
	c.Assert(err, IsNil)
	time.Sleep(timeoutInOperation)

	res, err := client.GetBucketACL(bucketNameTest)
	c.Assert(err, IsNil)
	c.Assert(res.ACL, Equals, string(ACLPrivate))

	err = client.DeleteBucket(bucketNameTest)
	c.Assert(err, IsNil)
	time.Sleep(timeoutInOperation)

	// PublicRead
	err = client.CreateBucket(bucketNameTest, ACL(ACLPublicRead))
	c.Assert(err, IsNil)
	time.Sleep(timeoutInOperation)

	res, err = client.GetBucketACL(bucketNameTest)
	c.Assert(err, IsNil)
	c.Assert(res.ACL, Equals, string(ACLPublicRead))

	err = client.DeleteBucket(bucketNameTest)
	c.Assert(err, IsNil)
	time.Sleep(timeoutInOperation)

	// PublicReadWrite
	err = client.CreateBucket(bucketNameTest, ACL(ACLPublicReadWrite))
	c.Assert(err, IsNil)
	time.Sleep(timeoutInOperation)

	res, err = client.GetBucketACL(bucketNameTest)
	c.Assert(err, IsNil)
	c.Assert(res.ACL, Equals, string(ACLPublicReadWrite))

	err = client.DeleteBucket(bucketNameTest)
	c.Assert(err, IsNil)
}

// TestGetBucketAcl
func (s *OssClientSuite) TestGetBucketLocation(c *C) {
	var bucketNameTest = bucketNamePrefix + RandLowStr(6)

	client, err := New(endpoint, accessID, accessKey)
	c.Assert(err, IsNil)

	// Private
	err = client.CreateBucket(bucketNameTest)
	c.Assert(err, IsNil)

	loc, err := client.GetBucketLocation(bucketNameTest)
	c.Assert(strings.HasPrefix(loc, "oss-"), Equals, true)

	err = client.DeleteBucket(bucketNameTest)
	c.Assert(err, IsNil)
}

// TestGetBucketLocationNegative
func (s *OssClientSuite) TestGetBucketLocationNegative(c *C) {
	var bucketNameTest = bucketNamePrefix + RandLowStr(6)

	client, err := New(endpoint, accessID, accessKey)
	c.Assert(err, IsNil)

	// Not exist
	_, err = client.GetBucketLocation(bucketNameTest)
	c.Assert(err, NotNil)

	// Not exist
	_, err = client.GetBucketLocation("InvalidBucketName_")
	c.Assert(err, NotNil)
}

// TestSetBucketLifecycle
func (s *OssClientSuite) TestSetBucketLifecycle(c *C) {
	var bucketNameTest = bucketNamePrefix + RandLowStr(6)
	var rule1 = BuildLifecycleRuleByDate("rule1", "one", true, 2015, 11, 11)
	var rule2 = BuildLifecycleRuleByDays("rule2", "two", true, 3)

	client, err := New(endpoint, accessID, accessKey)
	c.Assert(err, IsNil)

	err = client.CreateBucket(bucketNameTest)
	c.Assert(err, IsNil)

	// Set single rule
	var rules = []LifecycleRule{rule1}
	err = client.SetBucketLifecycle(bucketNameTest, rules)
	c.Assert(err, IsNil)
	// Double set rule
	err = client.SetBucketLifecycle(bucketNameTest, rules)
	c.Assert(err, IsNil)

	res, err := client.GetBucketLifecycle(bucketNameTest)
	c.Assert(err, IsNil)
	c.Assert(len(res.Rules), Equals, 1)
	c.Assert(res.Rules[0].ID, Equals, "rule1")

	err = client.DeleteBucketLifecycle(bucketNameTest)
	c.Assert(err, IsNil)

	// Set two rules
	rules = []LifecycleRule{rule1, rule2}
	err = client.SetBucketLifecycle(bucketNameTest, rules)
	c.Assert(err, IsNil)

	// Eliminate effect of cache
	time.Sleep(timeoutInOperation)

	res, err = client.GetBucketLifecycle(bucketNameTest)
	c.Assert(err, IsNil)
	c.Assert(len(res.Rules), Equals, 2)
	c.Assert(res.Rules[0].ID, Equals, "rule1")
	c.Assert(res.Rules[1].ID, Equals, "rule2")

	err = client.DeleteBucket(bucketNameTest)
	c.Assert(err, IsNil)
}

// TestSetBucketLifecycleNew
func (s *OssClientSuite) TestSetBucketLifecycleNew(c *C) {
	var bucketNameTest = bucketNamePrefix + RandLowStr(6)

	client, err := New(endpoint, accessID, accessKey)
	c.Assert(err, IsNil)

	err = client.CreateBucket(bucketNameTest)
	c.Assert(err, IsNil)

	//invalid status of lifecyclerule
	expiration := LifecycleExpiration{
		Days: 30,
	}
	rule := LifecycleRule{
		ID:         "rule1",
		Prefix:     "one",
		Status:     "Invalid",
		Expiration: &expiration,
	}
	rules := []LifecycleRule{rule}
	err = client.SetBucketLifecycle(bucketNameTest, rules)
	c.Assert(err, NotNil)

	//invalid value of CreatedBeforeDate
	expiration = LifecycleExpiration{
		CreatedBeforeDate: RandStr(10),
	}
	rule = LifecycleRule{
		ID:         "rule1",
		Prefix:     "one",
		Status:     "Enabled",
		Expiration: &expiration,
	}
	rules = []LifecycleRule{rule}
	err = client.SetBucketLifecycle(bucketNameTest, rules)
	c.Assert(err, NotNil)

	//invalid value of Days
	abortMPU := LifecycleAbortMultipartUpload{
		Days: -30,
	}
	rule = LifecycleRule{
		ID:                   "rule1",
		Prefix:               "one",
		Status:               "Enabled",
		AbortMultipartUpload: &abortMPU,
	}
	rules = []LifecycleRule{rule}
	err = client.SetBucketLifecycle(bucketNameTest, rules)
	c.Assert(err, NotNil)

	expiration = LifecycleExpiration{
		CreatedBeforeDate: "2015-11-11T00:00:00.000Z",
	}
	rule1 := LifecycleRule{
		ID:         "rule1",
		Prefix:     "one",
		Status:     "Enabled",
		Expiration: &expiration,
	}

	abortMPU = LifecycleAbortMultipartUpload{
		Days: 30,
	}
	rule2 := LifecycleRule{
		ID:                   "rule2",
		Prefix:               "two",
		Status:               "Enabled",
		Expiration:           &expiration,
		AbortMultipartUpload: &abortMPU,
	}

	transition1 := LifecycleTransition{
		Days:         3,
		StorageClass: StorageIA,
	}
	transition2 := LifecycleTransition{
		Days:         30,
		StorageClass: StorageArchive,
	}
	transitions := []LifecycleTransition{transition1, transition2}
	rule3 := LifecycleRule{
		ID:                   "rule3",
		Prefix:               "three",
		Status:               "Enabled",
		AbortMultipartUpload: &abortMPU,
		Transitions:          transitions,
	}

	// Set single rule
	rules = []LifecycleRule{rule1}
	err = client.SetBucketLifecycle(bucketNameTest, rules)
	c.Assert(err, IsNil)
	time.Sleep(timeoutInOperation)

	res, err := client.GetBucketLifecycle(bucketNameTest)
	c.Assert(err, IsNil)
	c.Assert(len(res.Rules), Equals, 1)
	c.Assert(res.Rules[0].ID, Equals, "rule1")
	c.Assert(res.Rules[0].Expiration, NotNil)
	c.Assert(res.Rules[0].Expiration.CreatedBeforeDate, Equals, "2015-11-11T00:00:00.000Z")

	err = client.DeleteBucketLifecycle(bucketNameTest)
	c.Assert(err, IsNil)

	// Set two rule: rule1 and rule2
	rules = []LifecycleRule{rule1, rule2}
	err = client.SetBucketLifecycle(bucketNameTest, rules)
	c.Assert(err, IsNil)
	time.Sleep(timeoutInOperation)

	res, err = client.GetBucketLifecycle(bucketNameTest)
	c.Assert(err, IsNil)
	c.Assert(len(res.Rules), Equals, 2)
	c.Assert(res.Rules[0].ID, Equals, "rule1")
	c.Assert(res.Rules[0].Expiration, NotNil)
	c.Assert(res.Rules[0].Expiration.CreatedBeforeDate, Equals, "2015-11-11T00:00:00.000Z")
	c.Assert(res.Rules[1].ID, Equals, "rule2")
	c.Assert(res.Rules[1].Expiration, NotNil)
	c.Assert(res.Rules[1].Expiration.CreatedBeforeDate, Equals, "2015-11-11T00:00:00.000Z")
	c.Assert(res.Rules[1].AbortMultipartUpload, NotNil)
	c.Assert(res.Rules[1].AbortMultipartUpload.Days, Equals, 30)

	err = client.DeleteBucketLifecycle(bucketNameTest)
	c.Assert(err, IsNil)

	// Set two rule: rule2 and rule3
	rules = []LifecycleRule{rule2, rule3}
	err = client.SetBucketLifecycle(bucketNameTest, rules)
	c.Assert(err, IsNil)
	time.Sleep(timeoutInOperation)

	res, err = client.GetBucketLifecycle(bucketNameTest)
	c.Assert(err, IsNil)
	c.Assert(len(res.Rules), Equals, 2)
	c.Assert(res.Rules[0].ID, Equals, "rule2")
	c.Assert(res.Rules[0].Expiration, NotNil)
	c.Assert(res.Rules[0].Expiration.CreatedBeforeDate, Equals, "2015-11-11T00:00:00.000Z")
	c.Assert(res.Rules[0].AbortMultipartUpload, NotNil)
	c.Assert(res.Rules[0].AbortMultipartUpload.Days, Equals, 30)
	c.Assert(res.Rules[1].ID, Equals, "rule3")
	c.Assert(res.Rules[1].AbortMultipartUpload, NotNil)
	c.Assert(res.Rules[1].AbortMultipartUpload.Days, Equals, 30)
	c.Assert(len(res.Rules[1].Transitions), Equals, 2)
	c.Assert(res.Rules[1].Transitions[0].StorageClass, Equals, StorageIA)
	c.Assert(res.Rules[1].Transitions[0].Days, Equals, 3)
	c.Assert(res.Rules[1].Transitions[1].StorageClass, Equals, StorageArchive)
	c.Assert(res.Rules[1].Transitions[1].Days, Equals, 30)

	err = client.DeleteBucketLifecycle(bucketNameTest)
	c.Assert(err, IsNil)

	// Set two rule: rule1 and rule3
	rules = []LifecycleRule{rule1, rule3}
	err = client.SetBucketLifecycle(bucketNameTest, rules)
	c.Assert(err, IsNil)
	time.Sleep(timeoutInOperation)

	res, err = client.GetBucketLifecycle(bucketNameTest)
	c.Assert(err, IsNil)
	c.Assert(len(res.Rules), Equals, 2)
	c.Assert(res.Rules[0].ID, Equals, "rule1")
	c.Assert(res.Rules[0].Expiration, NotNil)
	c.Assert(res.Rules[0].Expiration.CreatedBeforeDate, Equals, "2015-11-11T00:00:00.000Z")
	c.Assert(res.Rules[1].ID, Equals, "rule3")
	c.Assert(res.Rules[1].AbortMultipartUpload, NotNil)
	c.Assert(res.Rules[1].AbortMultipartUpload.Days, Equals, 30)
	c.Assert(len(res.Rules[1].Transitions), Equals, 2)
	c.Assert(res.Rules[1].Transitions[0].StorageClass, Equals, StorageIA)
	c.Assert(res.Rules[1].Transitions[0].Days, Equals, 3)
	c.Assert(res.Rules[1].Transitions[1].StorageClass, Equals, StorageArchive)
	c.Assert(res.Rules[1].Transitions[1].Days, Equals, 30)

	err = client.DeleteBucketLifecycle(bucketNameTest)
	c.Assert(err, IsNil)

	// Set three rules
	rules = []LifecycleRule{rule1, rule2, rule3}
	err = client.SetBucketLifecycle(bucketNameTest, rules)
	c.Assert(err, IsNil)
	time.Sleep(timeoutInOperation)

	res, err = client.GetBucketLifecycle(bucketNameTest)
	c.Assert(err, IsNil)
	c.Assert(len(res.Rules), Equals, 3)
	c.Assert(res.Rules[0].ID, Equals, "rule1")
	c.Assert(res.Rules[0].Expiration, NotNil)
	c.Assert(res.Rules[0].Expiration.CreatedBeforeDate, Equals, "2015-11-11T00:00:00.000Z")
	c.Assert(res.Rules[1].ID, Equals, "rule2")
	c.Assert(res.Rules[1].Expiration, NotNil)
	c.Assert(res.Rules[1].Expiration.CreatedBeforeDate, Equals, "2015-11-11T00:00:00.000Z")
	c.Assert(res.Rules[1].AbortMultipartUpload, NotNil)
	c.Assert(res.Rules[1].AbortMultipartUpload.Days, Equals, 30)
	c.Assert(res.Rules[2].ID, Equals, "rule3")
	c.Assert(res.Rules[2].AbortMultipartUpload, NotNil)
	c.Assert(res.Rules[2].AbortMultipartUpload.Days, Equals, 30)
	c.Assert(len(res.Rules[2].Transitions), Equals, 2)
	c.Assert(res.Rules[2].Transitions[0].StorageClass, Equals, StorageIA)
	c.Assert(res.Rules[2].Transitions[0].Days, Equals, 3)
	c.Assert(res.Rules[2].Transitions[1].StorageClass, Equals, StorageArchive)
	c.Assert(res.Rules[2].Transitions[1].Days, Equals, 30)

	err = client.DeleteBucket(bucketNameTest)
	c.Assert(err, IsNil)
}

// TestSetBucketLifecycleWithFilter
func (s *OssClientSuite) TestSetBucketLifecycleFilter(c *C) {
	var bucketNameTest = bucketNamePrefix + RandLowStr(6)

	client, err := New(endpoint, accessID, accessKey)
	c.Assert(err, IsNil)

	err = client.CreateBucket(bucketNameTest)
	c.Assert(err, IsNil)

	expiration := LifecycleExpiration{
		Days: 30,
	}
	tag := Tag{
		Key:   "key1",
		Value: "value1",
	}
	filter := LifecycleFilter{
		Not: []LifecycleFilterNot{
			{
				Prefix: "logs1",
				Tag:    &tag,
			},
		},
	}
	rule := LifecycleRule{
		ID:         "filter one",
		Prefix:     "logs",
		Status:     "Enabled",
		Expiration: &expiration,
		Transitions: []LifecycleTransition{
			{
				Days:         10,
				StorageClass: StorageIA,
			},
		},
		Filter: &filter,
	}
	rules := []LifecycleRule{rule}
	err = client.SetBucketLifecycle(bucketNameTest, rules)
	c.Assert(err, IsNil)
	res, err := client.GetBucketLifecycle(bucketNameTest)
	c.Assert(err, IsNil)
	c.Assert(len(res.Rules), Equals, 1)
	c.Assert(res.Rules[0].ID, Equals, "filter one")
	c.Assert(res.Rules[0].Expiration, NotNil)
	c.Assert(res.Rules[0].Expiration.CreatedBeforeDate, Equals, "")
	c.Assert(res.Rules[0].Expiration.Days, Equals, 30)
	c.Assert(res.Rules[0].Transitions[0].Days, Equals, 10)
	c.Assert(res.Rules[0].Transitions[0].StorageClass, Equals, StorageIA)
	c.Assert(res.Rules[0].Filter.Not[0].Prefix, Equals, "logs1")
	c.Assert(res.Rules[0].Filter.Not[0].Tag.Key, Equals, "key1")
	c.Assert(res.Rules[0].Filter.Not[0].Tag.Value, Equals, "value1")

	err = client.DeleteBucketLifecycle(bucketNameTest)
	c.Assert(err, IsNil)

	err = client.DeleteBucket(bucketNameTest)
	c.Assert(err, IsNil)
}

// TestSetBucketLifecycleOverLap
func (s *OssClientSuite) TestSetBucketLifecycleOverLap(c *C) {
	var bucketNameTest = bucketNamePrefix + RandLowStr(6)

	// rule1's prefix and rule2's prefix are overlap
	var rule1 = BuildLifecycleRuleByDate("rule1", "one", true, 2015, 11, 11)
	rule1.Prefix = "prefix1"
	var rule2 = BuildLifecycleRuleByDays("rule2", "two", true, 3)
	rule2.Prefix = "prefix1/prefix2"

	client, err := New(endpoint, accessID, accessKey)
	c.Assert(err, IsNil)

	err = client.CreateBucket(bucketNameTest)
	c.Assert(err, IsNil)

	// overlap is error
	var rules = []LifecycleRule{rule1, rule2}
	err = client.SetBucketLifecycle(bucketNameTest, rules)
	c.Assert(err, NotNil)

	//enable overlap,error
	options := []Option{AllowSameActionOverLap(true)}
	err = client.SetBucketLifecycle(bucketNameTest, rules, options...)
	c.Assert(err, IsNil)
	err = client.DeleteBucket(bucketNameTest)

}

// TestSetBucketLifecycleXml
func (s *OssClientSuite) TestSetBucketLifecycleXml(c *C) {
	var bucketNameTest = bucketNamePrefix + RandLowStr(6)

	client, err := New(endpoint, accessID, accessKey)
	c.Assert(err, IsNil)

	err = client.CreateBucket(bucketNameTest)
	c.Assert(err, IsNil)

	xmlBody := `<?xml version="1.0" encoding="UTF-8"?>
    <LifecycleConfiguration>
      <Rule>
        <ID>RuleID1</ID>
        <Prefix>Prefix</Prefix>
        <Status>Enabled</Status>
        <Expiration>
          <Days>65</Days>
        </Expiration>
        <Transition>
          <Days>45</Days>
          <StorageClass>IA</StorageClass>
        </Transition>
        <AbortMultipartUpload>
          <Days>30</Days>
        </AbortMultipartUpload>
      </Rule>
      <Rule>
        <ID>RuleID2</ID>
        <Prefix>Prefix/SubPrefix</Prefix>
        <Status>Enabled</Status>
        <Expiration>
          <Days>60</Days>
        </Expiration>
        <Transition>
          <Days>40</Days>
          <StorageClass>Archive</StorageClass>
        </Transition>
        <AbortMultipartUpload>
          <Days>40</Days>
        </AbortMultipartUpload>
      </Rule>
    </LifecycleConfiguration>`

	// overlap is error
	err = client.SetBucketLifecycleXml(bucketNameTest, xmlBody)
	c.Assert(err, NotNil)

	// enable overlap,error
	options := []Option{AllowSameActionOverLap(true)}
	err = client.SetBucketLifecycleXml(bucketNameTest, xmlBody, options...)
	c.Assert(err, NotNil)

	_, err = client.GetBucketLifecycleXml(bucketNameTest, options...)
	c.Assert(err, NotNil)

	err = client.DeleteBucket(bucketNameTest)
	c.Assert(err, IsNil)
}

// TestSetBucketLifecycleAboutVersionObject
func (s *OssClientSuite) TestSetBucketLifecycleAboutVersionObject(c *C) {
	var bucketNameTest = bucketNamePrefix + RandLowStr(6)

	client, err := New(endpoint, accessID, accessKey)
	c.Assert(err, IsNil)

	err = client.CreateBucket(bucketNameTest)
	c.Assert(err, IsNil)

	deleteMark := true
	expiration := LifecycleExpiration{
		ExpiredObjectDeleteMarker: &deleteMark,
	}

	versionExpiration := LifecycleVersionExpiration{
		NoncurrentDays: 20,
	}

	versionTransition := LifecycleVersionTransition{
		NoncurrentDays: 10,
		StorageClass:   "IA",
	}

	rule := LifecycleRule{
		Status:               "Enabled",
		Expiration:           &expiration,
		NonVersionExpiration: &versionExpiration,
		NonVersionTransition: &versionTransition,
	}
	rules := []LifecycleRule{rule}

	err = client.SetBucketLifecycle(bucketNameTest, rules)
	c.Assert(err, IsNil)

	res, err := client.GetBucketLifecycle(bucketNameTest)
	c.Assert(err, IsNil)

	c.Assert(res.Rules[0].Expiration, NotNil)
	c.Assert(res.Rules[0].Expiration.Days, Equals, 0)
	c.Assert(res.Rules[0].Expiration.Date, Equals, "")
	c.Assert(*(res.Rules[0].Expiration.ExpiredObjectDeleteMarker), Equals, true)

	c.Assert(res.Rules[0].NonVersionExpiration.NoncurrentDays, Equals, 20)
	c.Assert(res.Rules[0].NonVersionTransition.NoncurrentDays, Equals, 10)
	c.Assert(res.Rules[0].NonVersionTransition.StorageClass, Equals, StorageClassType("IA"))

	err = client.DeleteBucket(bucketNameTest)
	c.Assert(err, IsNil)
}

// TestSetBucketLifecycleAboutVersionObject
func (s *OssClientSuite) TestSetBucketLifecycleAboutVersionObjectError(c *C) {
	var bucketNameTest = bucketNamePrefix + RandLowStr(6)

	client, err := New(endpoint, accessID, accessKey)
	c.Assert(err, IsNil)

	err = client.CreateBucket(bucketNameTest)
	c.Assert(err, IsNil)

	deleteMark := true
	expiration := LifecycleExpiration{
		ExpiredObjectDeleteMarker: &deleteMark,
	}

	versionExpiration := LifecycleVersionExpiration{
		NoncurrentDays: 20,
	}

	versionTransition := LifecycleVersionTransition{
		NoncurrentDays: 10,
		StorageClass:   "IA",
	}

	// NonVersionTransition and NonVersionTransitions can not both have value
	rule := LifecycleRule{
		Status:                "Enabled",
		Expiration:            &expiration,
		NonVersionExpiration:  &versionExpiration,
		NonVersionTransition:  &versionTransition,
		NonVersionTransitions: []LifecycleVersionTransition{versionTransition},
	}
	rules := []LifecycleRule{rule}

	err = client.SetBucketLifecycle(bucketNameTest, rules)
	c.Assert(err, NotNil)

	err = client.DeleteBucket(bucketNameTest)
	c.Assert(err, IsNil)
}

// TestSetBucketLifecycleAboutVersionObject
func (s *OssClientSuite) TestSetBucketLifecycleAboutVersionObjectNew(c *C) {
	var bucketNameTest = bucketNamePrefix + RandLowStr(6)

	client, err := New(endpoint, accessID, accessKey)
	c.Assert(err, IsNil)

	err = client.CreateBucket(bucketNameTest)
	c.Assert(err, IsNil)

	deleteMark := true
	expiration := LifecycleExpiration{
		ExpiredObjectDeleteMarker: &deleteMark,
	}

	versionExpiration := LifecycleVersionExpiration{
		NoncurrentDays: 40,
	}

	versionTransition1 := LifecycleVersionTransition{
		NoncurrentDays: 25,
		StorageClass:   "IA",
	}

	versionTransition2 := LifecycleVersionTransition{
		NoncurrentDays: 30,
		StorageClass:   "ColdArchive",
	}

	rule := LifecycleRule{
		Status:                "Enabled",
		Expiration:            &expiration,
		NonVersionExpiration:  &versionExpiration,
		NonVersionTransitions: []LifecycleVersionTransition{versionTransition1, versionTransition2},
	}
	rules := []LifecycleRule{rule}

	err = client.SetBucketLifecycle(bucketNameTest, rules)
	c.Assert(err, IsNil)

	res, err := client.GetBucketLifecycle(bucketNameTest)
	c.Assert(err, IsNil)

	c.Assert(res.Rules[0].Expiration, NotNil)
	c.Assert(res.Rules[0].Expiration.Days, Equals, 0)
	c.Assert(res.Rules[0].Expiration.Date, Equals, "")
	c.Assert(*(res.Rules[0].Expiration.ExpiredObjectDeleteMarker), Equals, true)

	c.Assert(res.Rules[0].NonVersionExpiration.NoncurrentDays, Equals, 40)
	c.Assert(res.Rules[0].NonVersionTransition.NoncurrentDays, Equals, 25)
	c.Assert(res.Rules[0].NonVersionTransition.StorageClass, Equals, StorageClassType("IA"))
	c.Assert(len(res.Rules[0].NonVersionTransitions), Equals, 2)

	err = client.DeleteBucket(bucketNameTest)
	c.Assert(err, IsNil)
}

// TestDeleteBucketLifecycle
func (s *OssClientSuite) TestDeleteBucketLifecycle(c *C) {
	var bucketNameTest = bucketNamePrefix + RandLowStr(6)

	var rule1 = BuildLifecycleRuleByDate("rule1", "one", true, 2015, 11, 11)
	var rule2 = BuildLifecycleRuleByDays("rule2", "two", true, 3)
	var rules = []LifecycleRule{rule1, rule2}

	client, err := New(endpoint, accessID, accessKey)
	c.Assert(err, IsNil)

	err = client.CreateBucket(bucketNameTest)
	c.Assert(err, IsNil)
	//time.Sleep(timeoutInOperation)

	err = client.DeleteBucketLifecycle(bucketNameTest)
	c.Assert(err, IsNil)

	err = client.SetBucketLifecycle(bucketNameTest, rules)
	c.Assert(err, IsNil)
	//time.Sleep(timeoutInOperation)

	res, err := client.GetBucketLifecycle(bucketNameTest)
	c.Assert(err, IsNil)
	c.Assert(len(res.Rules), Equals, 2)

	// Delete
	err = client.DeleteBucketLifecycle(bucketNameTest)
	c.Assert(err, IsNil)

	//time.Sleep(timeoutInOperation)
	res, err = client.GetBucketLifecycle(bucketNameTest)
	c.Assert(err, NotNil)

	// Eliminate effect of cache
	//time.Sleep(timeoutInOperation)

	// Delete when not set
	err = client.DeleteBucketLifecycle(bucketNameTest)
	c.Assert(err, IsNil)

	err = client.DeleteBucket(bucketNameTest)
	c.Assert(err, IsNil)
}

// TestSetBucketLifecycleNegative
func (s *OssClientSuite) TestBucketLifecycleNegative(c *C) {
	var bucketNameTest = bucketNamePrefix + RandLowStr(6)
	var rules = []LifecycleRule{}

	client, err := New(endpoint, accessID, accessKey)
	c.Assert(err, IsNil)

	err = client.CreateBucket(bucketNameTest)
	c.Assert(err, IsNil)

	// Set with no rule
	err = client.SetBucketLifecycle(bucketNameTest, rules)
	c.Assert(err, NotNil)

	err = client.DeleteBucket(bucketNameTest)
	c.Assert(err, IsNil)

	// Not exist
	err = client.SetBucketLifecycle(bucketNameTest, rules)
	c.Assert(err, NotNil)

	// Not exist
	_, err = client.GetBucketLifecycle(bucketNameTest)
	c.Assert(err, NotNil)

	// Not exist
	err = client.DeleteBucketLifecycle(bucketNameTest)
	c.Assert(err, NotNil)
}

// TestBucketLifecycleWithFilterSize
func (s *OssClientSuite) TestBucketLifecycleWithFilterSize(c *C) {
	var bucketNameTest = bucketNamePrefix + RandLowStr(6)

	greater := int64(500)
	less := int64(645000)
	filter := LifecycleFilter{
		ObjectSizeGreaterThan: &greater,
		ObjectSizeLessThan:    &less,
	}
	rule1 := LifecycleRule{
		ID:     "rs1",
		Prefix: "logs",
		Status: "Enabled",
		Transitions: []LifecycleTransition{
			{
				Days:         30,
				StorageClass: StorageIA,
			},
		},
		Filter: &filter,
	}
	config4 := LifecycleConfiguration{
		Rules: []LifecycleRule{rule1},
	}
	xmlData4, err := xml.Marshal(config4)
	testLogger.Println(string(xmlData4))

	rules := []LifecycleRule{rule1}
	client, err := New(endpoint, accessID, accessKey)
	c.Assert(err, IsNil)

	err = client.CreateBucket(bucketNameTest)
	c.Assert(err, IsNil)

	err = client.SetBucketLifecycle(bucketNameTest, rules)
	c.Assert(err, IsNil)

	_, err = client.GetBucketLifecycle(bucketNameTest)
	c.Assert(err, IsNil)

	err = client.DeleteBucketLifecycle(bucketNameTest)
	c.Assert(err, IsNil)
	tag := Tag{Key: "key1", Value: "val1"}
	filter2 := LifecycleFilter{
		ObjectSizeGreaterThan: &greater,
		ObjectSizeLessThan:    &less,
		Not: []LifecycleFilterNot{
			{
				Tag: &tag,
			},
		},
	}
	rule2 := LifecycleRule{
		ID:     "rs2",
		Prefix: "",
		Status: "Enabled",
		Transitions: []LifecycleTransition{
			{
				Days:         30,
				StorageClass: StorageIA,
			},
		},
		Filter: &filter2,
	}
	rules2 := []LifecycleRule{rule2}

	err = client.SetBucketLifecycle(bucketNameTest, rules2)
	c.Assert(err, IsNil)

	_, err = client.GetBucketLifecycle(bucketNameTest)
	c.Assert(err, IsNil)

	err = client.DeleteBucketLifecycle(bucketNameTest)
	err = client.DeleteBucket(bucketNameTest)
	c.Assert(err, IsNil)

}

// TestSetBucketReferer
func (s *OssClientSuite) TestSetBucketReferer(c *C) {
	var bucketNameTest = bucketNamePrefix + RandLowStr(6)
	var referrers = []string{"http://www.aliyun.com", "https://www.aliyun.com"}

	client, err := New(endpoint, accessID, accessKey)
	c.Assert(err, IsNil)

	err = client.CreateBucket(bucketNameTest)
	c.Assert(err, IsNil)
	time.Sleep(timeoutInOperation)

	res, err := client.GetBucketReferer(bucketNameTest)
	c.Assert(res.AllowEmptyReferer, Equals, true)
	c.Assert(len(res.RefererList), Equals, 0)

	// Set referers
	err = client.SetBucketReferer(bucketNameTest, referrers, false)
	c.Assert(err, IsNil)
	time.Sleep(timeoutInOperation)

	res, err = client.GetBucketReferer(bucketNameTest)
	c.Assert(res.AllowEmptyReferer, Equals, false)
	c.Assert(len(res.RefererList), Equals, 2)
	c.Assert(res.RefererList[0], Equals, "http://www.aliyun.com")
	c.Assert(res.RefererList[1], Equals, "https://www.aliyun.com")

	// Reset referer, referers empty
	referrers = []string{""}
	err = client.SetBucketReferer(bucketNameTest, referrers, true)
	c.Assert(err, IsNil)

	referrers = []string{}
	err = client.SetBucketReferer(bucketNameTest, referrers, true)
	c.Assert(err, IsNil)

	res, err = client.GetBucketReferer(bucketNameTest)
	c.Assert(res.AllowEmptyReferer, Equals, true)
	c.Assert(len(res.RefererList), Equals, 0)

	err = client.DeleteBucket(bucketNameTest)
	c.Assert(err, IsNil)
}

// TestSetBucketRefererNegative
func (s *OssClientSuite) TestBucketRefererNegative(c *C) {
	var bucketNameTest = bucketNamePrefix + RandLowStr(6)
	var referers = []string{""}

	client, err := New(endpoint, accessID, accessKey)
	c.Assert(err, IsNil)

	// Not exist
	_, err = client.GetBucketReferer(bucketNameTest)
	c.Assert(err, NotNil)
	testLogger.Println(err)

	// Not exist
	err = client.SetBucketReferer(bucketNameTest, referers, true)
	c.Assert(err, NotNil)
	testLogger.Println(err)
}

// TestBucketRefererV2
func (s *OssClientSuite) TestBucketRefererV2(c *C) {
	var bucketNameTest = bucketNamePrefix + RandLowStr(6)
	var referrers = []string{"http://www.aliyun.com", "https://www.aliyun.com"}

	client, err := New(endpoint, accessID, accessKey)
	c.Assert(err, IsNil)

	err = client.CreateBucket(bucketNameTest)
	c.Assert(err, IsNil)
	time.Sleep(timeoutInOperation)

	res, err := client.GetBucketReferer(bucketNameTest)
	c.Assert(res.AllowEmptyReferer, Equals, true)
	c.Assert(len(res.RefererList), Equals, 0)
	c.Assert(*res.AllowTruncateQueryString, Equals, true)
	c.Assert(res.RefererBlacklist, IsNil)

	// Set referers
	var set RefererXML
	set.AllowEmptyReferer = false
	set.RefererList = referrers
	err = client.SetBucketRefererV2(bucketNameTest, set)
	c.Assert(err, IsNil)
	time.Sleep(timeoutInOperation)

	res, err = client.GetBucketReferer(bucketNameTest)
	c.Assert(res.AllowEmptyReferer, Equals, false)
	c.Assert(len(res.RefererList), Equals, 2)
	c.Assert(res.RefererList[0], Equals, "http://www.aliyun.com")
	c.Assert(res.RefererList[1], Equals, "https://www.aliyun.com")

	// Reset referer, referers empty
	var del RefererXML
	del.AllowEmptyReferer = true
	del.RefererList = []string{}
	err = client.SetBucketRefererV2(bucketNameTest, del)
	c.Assert(err, IsNil)

	res, err = client.GetBucketReferer(bucketNameTest)
	c.Assert(res.AllowEmptyReferer, Equals, true)
	c.Assert(len(res.RefererList), Equals, 0)

	// Set referers
	var setBucketReferer RefererXML
	setBucketReferer.AllowEmptyReferer = false
	setBucketReferer.RefererList = referrers
	referer1 := "http://www.refuse.com"
	referer2 := "https://*.hack.com"
	referer3 := "http://ban.*.com"
	referer4 := "https://www.?.deny.com"
	setBucketReferer.RefererBlacklist = &RefererBlacklist{
		[]string{
			referer1, referer2, referer3, referer4,
		},
	}
	setBucketReferer.AllowEmptyReferer = false
	boolTrue := true
	setBucketReferer.AllowTruncateQueryString = &boolTrue
	err = client.SetBucketRefererV2(bucketNameTest, setBucketReferer)
	c.Assert(err, IsNil)
	time.Sleep(timeoutInOperation)

	res, err = client.GetBucketReferer(bucketNameTest)
	c.Assert(res.AllowEmptyReferer, Equals, false)
	c.Assert(len(res.RefererList), Equals, 2)
	c.Assert(res.RefererList[0], Equals, "http://www.aliyun.com")
	c.Assert(res.RefererList[1], Equals, "https://www.aliyun.com")
	c.Assert(*res.AllowTruncateQueryString, Equals, true)
	c.Assert(res.RefererBlacklist, NotNil)
	c.Assert(len(res.RefererBlacklist.Referer), Equals, 4)
	c.Assert(res.RefererBlacklist.Referer[0], Equals, referer1)
	c.Assert(res.RefererBlacklist.Referer[3], Equals, referer4)

	del.AllowEmptyReferer = true
	del.RefererList = []string{}
	err = client.SetBucketRefererV2(bucketNameTest, del)
	c.Assert(err, IsNil)

	err = client.DeleteBucket(bucketNameTest)
	c.Assert(err, IsNil)
}

// TestSetBucketLogging
func (s *OssClientSuite) TestSetBucketLogging(c *C) {
	var bucketNameTest = bucketNamePrefix + RandLowStr(6)
	var bucketNameTarget = bucketNameTest + "-target"

	client, err := New(endpoint, accessID, accessKey)
	c.Assert(err, IsNil)

	err = client.CreateBucket(bucketNameTest)
	c.Assert(err, IsNil)
	err = client.CreateBucket(bucketNameTarget)
	c.Assert(err, IsNil)
	time.Sleep(timeoutInOperation)

	// Set logging
	err = client.SetBucketLogging(bucketNameTest, bucketNameTarget, "prefix", true)
	c.Assert(err, IsNil)
	// Reset
	err = client.SetBucketLogging(bucketNameTest, bucketNameTarget, "prefix", false)
	c.Assert(err, IsNil)

	time.Sleep(timeoutInOperation)
	res, err := client.GetBucketLogging(bucketNameTest)
	c.Assert(err, IsNil)
	c.Assert(res.LoggingEnabled.TargetBucket, Equals, "")
	c.Assert(res.LoggingEnabled.TargetPrefix, Equals, "")

	err = client.DeleteBucketLogging(bucketNameTest)
	c.Assert(err, IsNil)

	// Set to self
	err = client.SetBucketLogging(bucketNameTest, bucketNameTest, "prefix", true)
	c.Assert(err, IsNil)

	err = client.DeleteBucket(bucketNameTest)
	c.Assert(err, IsNil)
	err = client.DeleteBucket(bucketNameTarget)
	c.Assert(err, IsNil)
}

// TestDeleteBucketLogging
func (s *OssClientSuite) TestDeleteBucketLogging(c *C) {
	var bucketNameTest = bucketNamePrefix + RandLowStr(6)
	var bucketNameTarget = bucketNameTest + "-target"

	client, err := New(endpoint, accessID, accessKey)
	c.Assert(err, IsNil)

	err = client.CreateBucket(bucketNameTest)
	c.Assert(err, IsNil)
	err = client.CreateBucket(bucketNameTarget)
	c.Assert(err, IsNil)
	time.Sleep(timeoutInOperation)

	// Get when not set
	res, err := client.GetBucketLogging(bucketNameTest)
	c.Assert(err, IsNil)
	c.Assert(res.LoggingEnabled.TargetBucket, Equals, "")
	c.Assert(res.LoggingEnabled.TargetPrefix, Equals, "")

	// Set
	err = client.SetBucketLogging(bucketNameTest, bucketNameTarget, "prefix", true)
	c.Assert(err, IsNil)

	// Get
	time.Sleep(timeoutInOperation)
	res, err = client.GetBucketLogging(bucketNameTest)
	c.Assert(err, IsNil)
	c.Assert(res.LoggingEnabled.TargetBucket, Equals, bucketNameTarget)
	c.Assert(res.LoggingEnabled.TargetPrefix, Equals, "prefix")

	// Set
	err = client.SetBucketLogging(bucketNameTest, bucketNameTarget, "prefix", false)
	c.Assert(err, IsNil)

	// Get
	time.Sleep(timeoutInOperation)
	res, err = client.GetBucketLogging(bucketNameTest)
	c.Assert(err, IsNil)
	c.Assert(res.LoggingEnabled.TargetBucket, Equals, "")
	c.Assert(res.LoggingEnabled.TargetPrefix, Equals, "")

	// Delete
	err = client.DeleteBucketLogging(bucketNameTest)
	c.Assert(err, IsNil)

	// Get after delete
	time.Sleep(timeoutInOperation)
	res, err = client.GetBucketLogging(bucketNameTest)
	c.Assert(err, IsNil)
	c.Assert(res.LoggingEnabled.TargetBucket, Equals, "")
	c.Assert(res.LoggingEnabled.TargetPrefix, Equals, "")

	err = client.DeleteBucket(bucketNameTest)
	c.Assert(err, IsNil)
	err = client.DeleteBucket(bucketNameTarget)
	c.Assert(err, IsNil)
}

// TestSetBucketLoggingNegative
func (s *OssClientSuite) TestSetBucketLoggingNegative(c *C) {
	var bucketNameTest = bucketNamePrefix + RandLowStr(6)
	var bucketNameTarget = bucketNameTest + "-target"

	client, err := New(endpoint, accessID, accessKey)
	c.Assert(err, IsNil)

	// Not exist
	_, err = client.GetBucketLogging(bucketNameTest)
	c.Assert(err, NotNil)

	// Not exist
	err = client.SetBucketLogging(bucketNameTest, "targetbucket", "prefix", true)
	c.Assert(err, NotNil)

	// Not exist
	err = client.DeleteBucketLogging(bucketNameTest)
	c.Assert(err, NotNil)

	err = client.CreateBucket(bucketNameTest)
	c.Assert(err, IsNil)
	time.Sleep(timeoutInOperation)

	// Target bucket not exist
	err = client.SetBucketLogging(bucketNameTest, bucketNameTarget, "prefix", true)
	c.Assert(err, NotNil)

	// Parameter invalid
	err = client.SetBucketLogging(bucketNameTest, "XXXX", "prefix", true)
	c.Assert(err, NotNil)

	err = client.SetBucketLogging(bucketNameTest, "xx", "prefix", true)
	c.Assert(err, NotNil)

	err = client.DeleteBucket(bucketNameTest)
	c.Assert(err, IsNil)
}

// TestSetBucketWebsite
func (s *OssClientSuite) TestSetBucketWebsite(c *C) {
	var bucketNameTest = bucketNamePrefix + RandLowStr(6)
	var indexWebsite = "myindex.html"
	var errorWebsite = "myerror.html"

	client, err := New(endpoint, accessID, accessKey)
	c.Assert(err, IsNil)
	time.Sleep(timeoutInOperation)

	err = client.CreateBucket(bucketNameTest)
	c.Assert(err, IsNil)

	// Set
	err = client.SetBucketWebsite(bucketNameTest, indexWebsite, errorWebsite)
	c.Assert(err, IsNil)

	// Double set
	err = client.SetBucketWebsite(bucketNameTest, indexWebsite, errorWebsite)
	c.Assert(err, IsNil)

	res, err := client.GetBucketWebsite(bucketNameTest)
	c.Assert(err, IsNil)
	c.Assert(res.IndexDocument.Suffix, Equals, indexWebsite)
	c.Assert(res.ErrorDocument.Key, Equals, errorWebsite)

	// Reset
	err = client.SetBucketWebsite(bucketNameTest, "your"+indexWebsite, "your"+errorWebsite)
	c.Assert(err, IsNil)

	time.Sleep(timeoutInOperation)
	res, err = client.GetBucketWebsite(bucketNameTest)
	c.Assert(err, IsNil)
	c.Assert(res.IndexDocument.Suffix, Equals, "your"+indexWebsite)
	c.Assert(res.ErrorDocument.Key, Equals, "your"+errorWebsite)

	err = client.DeleteBucketWebsite(bucketNameTest)
	c.Assert(err, IsNil)

	// Set after delete
	err = client.SetBucketWebsite(bucketNameTest, indexWebsite, errorWebsite)
	c.Assert(err, IsNil)

	// Eliminate effect of cache
	time.Sleep(timeoutInOperation)

	res, err = client.GetBucketWebsite(bucketNameTest)
	c.Assert(err, IsNil)
	c.Assert(res.IndexDocument.Suffix, Equals, indexWebsite)
	c.Assert(res.ErrorDocument.Key, Equals, errorWebsite)

	err = client.DeleteBucket(bucketNameTest)
	c.Assert(err, IsNil)
}

// TestDeleteBucketWebsite
func (s *OssClientSuite) TestDeleteBucketWebsite(c *C) {
	var bucketNameTest = bucketNamePrefix + RandLowStr(6)
	var indexWebsite = "myindex.html"
	var errorWebsite = "myerror.html"

	client, err := New(endpoint, accessID, accessKey)
	c.Assert(err, IsNil)

	err = client.CreateBucket(bucketNameTest)
	c.Assert(err, IsNil)
	time.Sleep(timeoutInOperation)

	// Get
	res, err := client.GetBucketWebsite(bucketNameTest)
	c.Assert(err, NotNil)

	// Detele without set
	err = client.DeleteBucketWebsite(bucketNameTest)
	c.Assert(err, IsNil)

	// Set
	err = client.SetBucketWebsite(bucketNameTest, indexWebsite, errorWebsite)
	c.Assert(err, IsNil)

	time.Sleep(timeoutInOperation)
	res, err = client.GetBucketWebsite(bucketNameTest)
	c.Assert(err, IsNil)
	c.Assert(res.IndexDocument.Suffix, Equals, indexWebsite)
	c.Assert(res.ErrorDocument.Key, Equals, errorWebsite)

	// Detele
	time.Sleep(timeoutInOperation)
	err = client.DeleteBucketWebsite(bucketNameTest)
	c.Assert(err, IsNil)

	time.Sleep(timeoutInOperation)
	res, err = client.GetBucketWebsite(bucketNameTest)
	c.Assert(err, NotNil)

	// Detele after delete
	err = client.DeleteBucketWebsite(bucketNameTest)
	c.Assert(err, IsNil)

	err = client.DeleteBucket(bucketNameTest)
	c.Assert(err, IsNil)
}

// TestSetBucketWebsiteNegative
func (s *OssClientSuite) TestSetBucketWebsiteNegative(c *C) {
	var bucketNameTest = bucketNamePrefix + RandLowStr(6)
	var indexWebsite = "myindex.html"
	var errorWebsite = "myerror.html"

	client, err := New(endpoint, accessID, accessKey)
	c.Assert(err, IsNil)

	err = client.DeleteBucket(bucketNameTest)

	// Not exist
	_, err = client.GetBucketWebsite(bucketNameTest)
	c.Assert(err, NotNil)

	err = client.DeleteBucketWebsite(bucketNameTest)
	c.Assert(err, NotNil)

	err = client.SetBucketWebsite(bucketNameTest, indexWebsite, errorWebsite)
	c.Assert(err, NotNil)

	err = client.CreateBucket(bucketNameTest)
	c.Assert(err, IsNil)

	// Set
	time.Sleep(timeoutInOperation)
	err = client.SetBucketWebsite(bucketNameTest, "myindex", "myerror")
	c.Assert(err, IsNil)

	res, err := client.GetBucketWebsite(bucketNameTest)
	c.Assert(err, IsNil)
	c.Assert(res.IndexDocument.Suffix, Equals, "myindex")
	c.Assert(res.ErrorDocument.Key, Equals, "myerror")

	// Detele
	err = client.DeleteBucketWebsite(bucketNameTest)
	c.Assert(err, IsNil)

	time.Sleep(timeoutInOperation)
	_, err = client.GetBucketWebsite(bucketNameTest)
	c.Assert(err, NotNil)

	// Detele after delete
	err = client.DeleteBucketWebsite(bucketNameTest)
	c.Assert(err, IsNil)

	err = client.DeleteBucket(bucketNameTest)
	c.Assert(err, IsNil)
}

// TestSetBucketWebsiteDetail
func (s *OssClientSuite) TestSetBucketWebsiteDetail(c *C) {
	var bucketNameTest = bucketNamePrefix + RandLowStr(6)
	var indexWebsite = "myindex.html"
	var errorWebsite = "myerror.html"

	client, err := New(endpoint, accessID, accessKey)
	c.Assert(err, IsNil)

	err = client.CreateBucket(bucketNameTest)
	c.Assert(err, IsNil)
	time.Sleep(timeoutInOperation)

	btrue := true
	bfalse := false
	// Define one routing rule
	ruleOk := RoutingRule{
		RuleNumber: 1,
		Condition: Condition{
			KeyPrefixEquals:             "",
			HTTPErrorCodeReturnedEquals: 404,
		},
		Redirect: Redirect{
			RedirectType: "Mirror",
			// PassQueryString: &btrue, 		// set default value
			MirrorURL: "http://www.test.com/",
			// MirrorPassQueryString:&btrue, 	// set default value
			// MirrorFollowRedirect:&bfalse, 	// set default value
			// MirrorCheckMd5:&bfalse, 			// set default value
			MirrorHeaders: MirrorHeaders{
				// PassAll:&bfalse, 			// set default value
				Pass:   []string{"myheader-key1", "myheader-key2"},
				Remove: []string{"myheader-key3", "myheader-key4"},
				Set: []MirrorHeaderSet{
					{
						Key:   "myheader-key5",
						Value: "myheader-value5",
					},
				},
			},
		},
	}

	// Define array routing rule
	ruleArrOk := []RoutingRule{
		{
			RuleNumber: 2,
			Condition: Condition{
				KeyPrefixEquals:             "abc/",
				HTTPErrorCodeReturnedEquals: 404,
				IncludeHeader: []IncludeHeader{
					{
						Key:    "host",
						Equals: "test.oss-cn-beijing-internal.aliyuncs.com",
					},
				},
			},
			Redirect: Redirect{
				RedirectType:     "AliCDN",
				Protocol:         "http",
				HostName:         "www.test.com",
				PassQueryString:  &bfalse,
				ReplaceKeyWith:   "prefix/${key}.suffix",
				HttpRedirectCode: 301,
			},
		},
		{
			RuleNumber: 3,
			Condition: Condition{
				KeyPrefixEquals:             "",
				HTTPErrorCodeReturnedEquals: 404,
			},
			Redirect: Redirect{
				RedirectType:          "Mirror",
				PassQueryString:       &btrue,
				MirrorURL:             "http://www.test.com/",
				MirrorPassQueryString: &btrue,
				MirrorFollowRedirect:  &bfalse,
				MirrorCheckMd5:        &bfalse,
				MirrorHeaders: MirrorHeaders{
					PassAll: &btrue,
					Pass:    []string{"myheader-key1", "myheader-key2"},
					Remove:  []string{"myheader-key3", "myheader-key4"},
					Set: []MirrorHeaderSet{
						{
							Key:   "myheader-key5",
							Value: "myheader-value5",
						},
					},
				},
			},
		},
	}

	// Set one routing rule
	wxmlOne := WebsiteXML{}
	wxmlOne.RoutingRules = append(wxmlOne.RoutingRules, ruleOk)
	var responseHeader http.Header
	err = client.SetBucketWebsiteDetail(bucketNameTest, wxmlOne, GetResponseHeader(&responseHeader))
	c.Assert(err, IsNil)
	requestId := GetRequestId(responseHeader)
	c.Assert(len(requestId) > 0, Equals, true)

	res, err := client.GetBucketWebsite(bucketNameTest)
	c.Assert(err, IsNil)
	c.Assert(res.RoutingRules[0].Redirect.RedirectType, Equals, "Mirror")
	c.Assert(*res.RoutingRules[0].Redirect.PassQueryString, Equals, false)
	c.Assert(*res.RoutingRules[0].Redirect.MirrorPassQueryString, Equals, false)
	c.Assert(*res.RoutingRules[0].Redirect.MirrorFollowRedirect, Equals, true)
	c.Assert(*res.RoutingRules[0].Redirect.MirrorCheckMd5, Equals, false)
	c.Assert(*res.RoutingRules[0].Redirect.MirrorHeaders.PassAll, Equals, false)

	// Set one routing rule and IndexDocument, IndexDocument
	wxml := WebsiteXML{
		IndexDocument: IndexDocument{Suffix: indexWebsite},
		ErrorDocument: ErrorDocument{Key: errorWebsite},
	}
	wxml.RoutingRules = append(wxml.RoutingRules, ruleOk)
	err = client.SetBucketWebsiteDetail(bucketNameTest, wxml, GetResponseHeader(&responseHeader))
	c.Assert(err, IsNil)
	requestId = GetRequestId(responseHeader)
	c.Assert(len(requestId) > 0, Equals, true)

	res, err = client.GetBucketWebsite(bucketNameTest)
	c.Assert(err, IsNil)
	c.Assert(res.IndexDocument.Suffix, Equals, indexWebsite)
	c.Assert(res.ErrorDocument.Key, Equals, errorWebsite)
	c.Assert(res.RoutingRules[0].Redirect.RedirectType, Equals, "Mirror")

	// Set array routing rule
	wxml.RoutingRules = append(wxml.RoutingRules, ruleArrOk...)
	err = client.SetBucketWebsiteDetail(bucketNameTest, wxml, GetResponseHeader(&responseHeader))
	c.Assert(err, IsNil)
	requestId = GetRequestId(responseHeader)
	c.Assert(len(requestId) > 0, Equals, true)

	res, err = client.GetBucketWebsite(bucketNameTest)
	c.Assert(err, IsNil)
	c.Assert(res.IndexDocument.Suffix, Equals, indexWebsite)
	c.Assert(res.ErrorDocument.Key, Equals, errorWebsite)
	c.Assert(len(res.RoutingRules), Equals, 3)
	c.Assert(res.RoutingRules[1].Redirect.RedirectType, Equals, "AliCDN")
	c.Assert(*res.RoutingRules[2].Redirect.MirrorPassQueryString, Equals, true)
	c.Assert(*res.RoutingRules[2].Redirect.MirrorFollowRedirect, Equals, false)

	// Define one error routing rule
	ruleErr := RoutingRule{
		RuleNumber: 1,
		Redirect: Redirect{
			RedirectType:    "Mirror",
			PassQueryString: &btrue,
		},
	}
	// Define array error routing rule
	rulesErrArr := []RoutingRule{
		{
			RuleNumber: 1,
			Redirect: Redirect{
				RedirectType:    "Mirror",
				PassQueryString: &btrue,
			},
		},
		{
			RuleNumber: 2,
			Redirect: Redirect{
				RedirectType:    "Mirror",
				PassQueryString: &btrue,
			},
		},
	}

	ruleIntErr := RoutingRule{
		// RuleNumber:0,						// set NULL value
		Condition: Condition{
			KeyPrefixEquals:             "",
			HTTPErrorCodeReturnedEquals: 404,
		},
		Redirect: Redirect{
			RedirectType: "Mirror",
			// PassQueryString: &btrue, 		// set default value
			MirrorURL: "http://www.test.com/",
			// MirrorPassQueryString:&btrue, 	// set default value
			// MirrorFollowRedirect:&bfalse, 	// set default value
			// MirrorCheckMd5:&bfalse, 			// set default value
			MirrorHeaders: MirrorHeaders{
				// PassAll:&bfalse, 			// set default value
				Pass:   []string{"myheader-key1", "myheader-key2"},
				Remove: []string{"myheader-key3", "myheader-key4"},
				Set: []MirrorHeaderSet{
					{
						Key:   "myheader-key5",
						Value: "myheader-value5",
					},
				},
			},
		},
	}

	// Set one int type error rule
	wxmlIntErr := WebsiteXML{}
	wxmlIntErr.RoutingRules = append(wxmlIntErr.RoutingRules, ruleIntErr)
	err = client.SetBucketWebsiteDetail(bucketNameTest, wxmlIntErr)
	c.Assert(err, NotNil)

	// Set one error rule
	wxmlErr := WebsiteXML{}
	wxmlErr.RoutingRules = append(wxmlErr.RoutingRules, ruleErr)
	err = client.SetBucketWebsiteDetail(bucketNameTest, wxmlErr)
	c.Assert(err, NotNil)

	// Set one error rule and one correct rule
	wxmlErr.RoutingRules = append(wxmlErr.RoutingRules, ruleOk)
	err = client.SetBucketWebsiteDetail(bucketNameTest, wxmlErr)
	c.Assert(err, NotNil)

	wxmlErrRuleArr := WebsiteXML{}
	wxmlErrRuleArr.RoutingRules = append(wxmlErrRuleArr.RoutingRules, rulesErrArr...)
	// Set array error routing rule
	err = client.SetBucketWebsiteDetail(bucketNameTest, wxmlErrRuleArr)
	c.Assert(err, NotNil)

	err = client.DeleteBucketWebsite(bucketNameTest)
	c.Assert(err, IsNil)

	err = client.DeleteBucket(bucketNameTest)
	c.Assert(err, IsNil)
}

// TestSetBucketWebsiteXml
func (s *OssClientSuite) TestSetBucketWebsiteXml(c *C) {
	var bucketNameTest = bucketNamePrefix + RandLowStr(6)
	client, err := New(endpoint, accessID, accessKey)
	c.Assert(err, IsNil)

	err = client.CreateBucket(bucketNameTest)
	c.Assert(err, IsNil)
	time.Sleep(timeoutInOperation)

	// Define one routing rule
	ruleOk := RoutingRule{
		RuleNumber: 1,
		Condition: Condition{
			KeyPrefixEquals:             "",
			HTTPErrorCodeReturnedEquals: 404,
		},
		Redirect: Redirect{
			RedirectType: "Mirror",
			// PassQueryString: &btrue, 		// set default value
			MirrorURL: "http://www.test.com/",
			// MirrorPassQueryString:&btrue, 	// set default value
			// MirrorFollowRedirect:&bfalse, 	// set default value
			// MirrorCheckMd5:&bfalse, 			// set default value
			MirrorHeaders: MirrorHeaders{
				// PassAll:&bfalse, 			// set default value
				Pass:   []string{"myheader-key1", "myheader-key2"},
				Remove: []string{"myheader-key3", "myheader-key4"},
				Set: []MirrorHeaderSet{
					{
						Key:   "myheader-key5",
						Value: "myheader-value5",
					},
				},
			},
		},
	}

	// Set one routing rule
	wxmlOne := WebsiteXML{}
	wxmlOne.RoutingRules = append(wxmlOne.RoutingRules, ruleOk)
	bs, err := xml.Marshal(wxmlOne)
	c.Assert(err, IsNil)

	var responseHeader http.Header
	err = client.SetBucketWebsiteXml(bucketNameTest, string(bs), GetResponseHeader(&responseHeader))
	c.Assert(err, IsNil)

	requestId := GetRequestId(responseHeader)
	c.Assert(len(requestId) > 0, Equals, true)

	res, err := client.GetBucketWebsite(bucketNameTest)
	c.Assert(err, IsNil)
	c.Assert(res.RoutingRules[0].Redirect.RedirectType, Equals, "Mirror")
	c.Assert(*res.RoutingRules[0].Redirect.PassQueryString, Equals, false)
	c.Assert(*res.RoutingRules[0].Redirect.MirrorPassQueryString, Equals, false)
	c.Assert(*res.RoutingRules[0].Redirect.MirrorFollowRedirect, Equals, true)
	c.Assert(*res.RoutingRules[0].Redirect.MirrorCheckMd5, Equals, false)
	c.Assert(*res.RoutingRules[0].Redirect.MirrorHeaders.PassAll, Equals, false)

	// test GetBucketWebsite xml
	xmlText, err := client.GetBucketWebsiteXml(bucketNameTest)
	c.Assert(err, IsNil)

	c.Assert(strings.Contains(xmlText, "<Pass>myheader-key1</Pass>"), Equals, true)
	c.Assert(strings.Contains(xmlText, "<Pass>myheader-key2</Pass>"), Equals, true)

	err = client.DeleteBucket(bucketNameTest)
	c.Assert(err, IsNil)
}

// TestSetBucketCORS
func (s *OssClientSuite) TestSetBucketCORS(c *C) {
	var bucketNameTest = bucketNamePrefix + RandLowStr(6)
	var rule1 = CORSRule{
		AllowedOrigin: []string{"*"},
		AllowedMethod: []string{"PUT", "GET", "POST"},
		AllowedHeader: []string{},
		ExposeHeader:  []string{},
		MaxAgeSeconds: 100,
	}

	var rule2 = CORSRule{
		AllowedOrigin: []string{"http://www.a.com", "http://www.b.com"},
		AllowedMethod: []string{"GET"},
		AllowedHeader: []string{"Authorization"},
		ExposeHeader:  []string{"x-oss-test", "x-oss-test1"},
		MaxAgeSeconds: 200,
	}

	client, err := New(endpoint, accessID, accessKey)
	c.Assert(err, IsNil)

	err = client.CreateBucket(bucketNameTest)
	c.Assert(err, IsNil)
	time.Sleep(timeoutInOperation)

	// Set
	err = client.SetBucketCORS(bucketNameTest, []CORSRule{rule1})
	c.Assert(err, IsNil)
	time.Sleep(timeoutInOperation)

	gbcr, err := client.GetBucketCORS(bucketNameTest)
	c.Assert(err, IsNil)
	c.Assert(len(gbcr.CORSRules), Equals, 1)
	c.Assert(len(gbcr.CORSRules[0].AllowedOrigin), Equals, 1)
	c.Assert(len(gbcr.CORSRules[0].AllowedMethod), Equals, 3)
	c.Assert(len(gbcr.CORSRules[0].AllowedHeader), Equals, 0)
	c.Assert(len(gbcr.CORSRules[0].ExposeHeader), Equals, 0)
	c.Assert(gbcr.CORSRules[0].MaxAgeSeconds, Equals, 100)

	// Double set
	err = client.SetBucketCORS(bucketNameTest, []CORSRule{rule1})
	c.Assert(err, IsNil)
	time.Sleep(timeoutInOperation)

	gbcr, err = client.GetBucketCORS(bucketNameTest)
	c.Assert(err, IsNil)
	c.Assert(len(gbcr.CORSRules), Equals, 1)
	c.Assert(len(gbcr.CORSRules[0].AllowedOrigin), Equals, 1)
	c.Assert(len(gbcr.CORSRules[0].AllowedMethod), Equals, 3)
	c.Assert(len(gbcr.CORSRules[0].AllowedHeader), Equals, 0)
	c.Assert(len(gbcr.CORSRules[0].ExposeHeader), Equals, 0)
	c.Assert(gbcr.CORSRules[0].MaxAgeSeconds, Equals, 100)

	// Set rule2
	err = client.SetBucketCORS(bucketNameTest, []CORSRule{rule2})
	c.Assert(err, IsNil)

	time.Sleep(timeoutInOperation)
	gbcr, err = client.GetBucketCORS(bucketNameTest)
	c.Assert(err, IsNil)
	c.Assert(len(gbcr.CORSRules), Equals, 1)
	c.Assert(len(gbcr.CORSRules[0].AllowedOrigin), Equals, 2)
	c.Assert(len(gbcr.CORSRules[0].AllowedMethod), Equals, 1)
	c.Assert(len(gbcr.CORSRules[0].AllowedHeader), Equals, 1)
	c.Assert(len(gbcr.CORSRules[0].ExposeHeader), Equals, 2)
	c.Assert(gbcr.CORSRules[0].MaxAgeSeconds, Equals, 200)

	// Reset
	err = client.SetBucketCORS(bucketNameTest, []CORSRule{rule1, rule2})
	c.Assert(err, IsNil)

	time.Sleep(timeoutInOperation)
	gbcr, err = client.GetBucketCORS(bucketNameTest)
	c.Assert(err, IsNil)
	c.Assert(len(gbcr.CORSRules), Equals, 2)

	// Set after delete
	err = client.DeleteBucketCORS(bucketNameTest)
	c.Assert(err, IsNil)

	err = client.SetBucketCORS(bucketNameTest, []CORSRule{rule1, rule2})
	c.Assert(err, IsNil)

	time.Sleep(timeoutInOperation)
	gbcr, err = client.GetBucketCORS(bucketNameTest)
	c.Assert(err, IsNil)
	c.Assert(len(gbcr.CORSRules), Equals, 2)

	// GetBucketCORSXml
	xmlBody, err := client.GetBucketCORSXml(bucketNameTest)
	c.Assert(err, IsNil)

	err = client.SetBucketCORSXml(bucketNameTest, xmlBody)
	c.Assert(err, IsNil)

	// get again
	time.Sleep(timeoutInOperation)
	gbcr, err = client.GetBucketCORS(bucketNameTest)
	c.Assert(err, IsNil)
	c.Assert(len(gbcr.CORSRules), Equals, 2)

	isTrue := true
	put := PutBucketCORS{}
	put.CORSRules = []CORSRule{rule1, rule2}
	put.ResponseVary = &isTrue
	err = client.SetBucketCORSV2(bucketNameTest, put)
	c.Assert(err, IsNil)

	time.Sleep(timeoutInOperation)
	gbcr, err = client.GetBucketCORS(bucketNameTest)
	c.Assert(err, IsNil)
	c.Assert(len(gbcr.CORSRules), Equals, 2)
	c.Assert(*gbcr.ResponseVary, Equals, isTrue)

	err = client.DeleteBucketCORS(bucketNameTest)
	c.Assert(err, IsNil)

	err = client.DeleteBucket(bucketNameTest)
	c.Assert(err, IsNil)
}

// TestDeleteBucketCORS
func (s *OssClientSuite) TestDeleteBucketCORS(c *C) {
	var bucketNameTest = bucketNamePrefix + RandLowStr(6)
	var rule = CORSRule{
		AllowedOrigin: []string{"*"},
		AllowedMethod: []string{"PUT", "GET", "POST"},
		AllowedHeader: []string{},
		ExposeHeader:  []string{},
		MaxAgeSeconds: 100,
	}

	client, err := New(endpoint, accessID, accessKey)
	c.Assert(err, IsNil)

	err = client.CreateBucket(bucketNameTest)
	c.Assert(err, IsNil)
	time.Sleep(timeoutInOperation)

	// Delete not set
	err = client.DeleteBucketCORS(bucketNameTest)
	c.Assert(err, IsNil)

	// Set
	err = client.SetBucketCORS(bucketNameTest, []CORSRule{rule})
	c.Assert(err, IsNil)

	time.Sleep(timeoutInOperation)
	_, err = client.GetBucketCORS(bucketNameTest)
	c.Assert(err, IsNil)

	// Detele
	err = client.DeleteBucketCORS(bucketNameTest)
	c.Assert(err, IsNil)

	time.Sleep(timeoutInOperation)
	_, err = client.GetBucketCORS(bucketNameTest)
	c.Assert(err, NotNil)

	// Detele after deleting
	err = client.DeleteBucketCORS(bucketNameTest)
	c.Assert(err, IsNil)

	err = client.DeleteBucket(bucketNameTest)
	c.Assert(err, IsNil)
}

// TestSetBucketCORSNegative
func (s *OssClientSuite) TestSetBucketCORSNegative(c *C) {
	var bucketNameTest = bucketNamePrefix + RandLowStr(6)
	var rule = CORSRule{
		AllowedOrigin: []string{"*"},
		AllowedMethod: []string{"PUT", "GET", "POST"},
		AllowedHeader: []string{},
		ExposeHeader:  []string{},
		MaxAgeSeconds: 100,
	}

	client, err := New(endpoint, accessID, accessKey)
	c.Assert(err, IsNil)

	err = client.DeleteBucket(bucketNameTest)

	// Not exist
	_, err = client.GetBucketCORS(bucketNameTest)
	c.Assert(err, NotNil)

	err = client.DeleteBucketCORS(bucketNameTest)
	c.Assert(err, NotNil)

	err = client.SetBucketCORS(bucketNameTest, []CORSRule{rule})
	c.Assert(err, NotNil)

	err = client.CreateBucket(bucketNameTest)
	c.Assert(err, IsNil)
	time.Sleep(timeoutInOperation)

	_, err = client.GetBucketCORS(bucketNameTest)
	c.Assert(err, NotNil)

	// Set
	err = client.SetBucketCORS(bucketNameTest, []CORSRule{rule})
	c.Assert(err, IsNil)
	time.Sleep(timeoutInOperation)

	_, err = client.GetBucketCORS(bucketNameTest)
	c.Assert(err, IsNil)

	// Delete
	err = client.DeleteBucketCORS(bucketNameTest)
	c.Assert(err, IsNil)

	time.Sleep(timeoutInOperation)
	_, err = client.GetBucketCORS(bucketNameTest)
	c.Assert(err, NotNil)

	// Delete after deleting
	err = client.DeleteBucketCORS(bucketNameTest)
	c.Assert(err, IsNil)

	err = client.DeleteBucket(bucketNameTest)
	c.Assert(err, IsNil)
}

// TestGetBucketInfo
func (s *OssClientSuite) TestGetBucketInfo(c *C) {
	var bucketNameTest = bucketNamePrefix + RandLowStr(6)

	client, err := New(endpoint, accessID, accessKey)
	c.Assert(err, IsNil)

	err = client.CreateBucket(bucketNameTest)
	c.Assert(err, IsNil)

	res, err := client.GetBucketInfo(bucketNameTest)
	c.Assert(err, IsNil)
	c.Assert(res.BucketInfo.Name, Equals, bucketNameTest)
	c.Assert(strings.HasPrefix(res.BucketInfo.Location, "oss-"), Equals, true)
	c.Assert(res.BucketInfo.ACL, Equals, "private")
	c.Assert(strings.HasSuffix(res.BucketInfo.ExtranetEndpoint, ".com"), Equals, true)
	c.Assert(strings.HasSuffix(res.BucketInfo.IntranetEndpoint, ".com"), Equals, true)
	c.Assert(res.BucketInfo.CreationDate, NotNil)
	c.Assert(res.BucketInfo.TransferAcceleration, Equals, "Disabled")
	c.Assert(res.BucketInfo.CrossRegionReplication, Equals, "Disabled")

	err = client.DeleteBucket(bucketNameTest)
	c.Assert(err, IsNil)
}

// TestGetBucketInfoNegative
func (s *OssClientSuite) TestGetBucketInfoNegative(c *C) {
	var bucketNameTest = bucketNamePrefix + RandLowStr(6)

	client, err := New(endpoint, accessID, accessKey)
	c.Assert(err, IsNil)

	// Not exist
	_, err = client.GetBucketInfo(bucketNameTest)
	c.Assert(err, NotNil)

	// Bucket name invalid
	_, err = client.GetBucketInfo("InvalidBucketName_")
	c.Assert(err, NotNil)
}

// TestEndpointFormat
func (s *OssClientSuite) TestEndpointFormat(c *C) {
	var bucketNameTest = bucketNamePrefix + RandLowStr(6)

	// http://host
	client, err := New(endpoint, accessID, accessKey)
	c.Assert(err, IsNil)

	err = client.CreateBucket(bucketNameTest)
	c.Assert(err, IsNil)
	time.Sleep(timeoutInOperation)

	res, err := client.GetBucketACL(bucketNameTest)
	c.Assert(err, IsNil)
	c.Assert(res.ACL, Equals, string(ACLPrivate))

	err = client.DeleteBucket(bucketNameTest)
	c.Assert(err, IsNil)
	time.Sleep(timeoutInOperation)

	// http://host:port
	client, err = New(endpoint+":80", accessID, accessKey)
	c.Assert(err, IsNil)

	err = client.CreateBucket(bucketNameTest)
	c.Assert(err, IsNil)

	time.Sleep(timeoutInOperation)
	res, err = client.GetBucketACL(bucketNameTest)
	c.Assert(err, IsNil)
	c.Assert(res.ACL, Equals, string(ACLPrivate))

	err = client.DeleteBucket(bucketNameTest)
	c.Assert(err, IsNil)
}

// TestCname
func (s *OssClientSuite) _TestCname(c *C) {
	var bucketNameTest = "<my-bucket-cname>"

	client, err := New("<endpoint>", "<accessKeyId>", "<accessKeySecret>", UseCname(true))
	c.Assert(err, IsNil)

	err = client.CreateBucket(bucketNameTest)
	c.Assert(err, IsNil)

	_, err = client.ListBuckets()
	c.Assert(err, NotNil)

	res, err := client.GetBucketACL(bucketNameTest)
	c.Assert(err, IsNil)
	c.Assert(res.ACL, Equals, string(ACLPrivate))
}

// TestCnameNegative
func (s *OssClientSuite) _TestCnameNegative(c *C) {
	var bucketNameTest = "<my-bucket-cname>"

	client, err := New("<endpoint>", "<accessKeyId>", "<accessKeySecret>", UseCname(true))
	c.Assert(err, IsNil)

	err = client.CreateBucket(bucketNameTest)
	c.Assert(err, NotNil)

	_, err = client.ListBuckets()
	c.Assert(err, NotNil)

	_, err = client.GetBucketACL(bucketNameTest)
	c.Assert(err, NotNil)
}

// _TestHTTPS
func (s *OssClientSuite) _TestHTTPS(c *C) {
	var bucketNameTest = "<my-bucket-https>"

	client, err := New("<endpoint>", "<accessKeyId>", "<accessKeySecret>")
	c.Assert(err, IsNil)

	err = client.CreateBucket(bucketNameTest)
	c.Assert(err, IsNil)

	res, err := client.GetBucketACL(bucketNameTest)
	c.Assert(err, IsNil)
	c.Assert(res.ACL, Equals, string(ACLPrivate))

	err = client.DeleteBucket(bucketNameTest)
	c.Assert(err, IsNil)
}

// TestClientOption
func (s *OssClientSuite) TestClientOption(c *C) {
	var bucketNameTest = bucketNamePrefix + RandLowStr(6)

	client, err := New(endpoint, accessID, accessKey, UseCname(true),
		Timeout(11, 12), SecurityToken("token"), Proxy("http://127.0.0.1:8120"))
	c.Assert(err, IsNil)

	// CreateBucket timeout
	err = client.CreateBucket(bucketNameTest)
	c.Assert(err, NotNil)

	c.Assert(client.Conn.config.HTTPTimeout.ConnectTimeout, Equals, time.Second*11)
	c.Assert(client.Conn.config.HTTPTimeout.ReadWriteTimeout, Equals, time.Second*12)
	c.Assert(client.Conn.config.HTTPTimeout.HeaderTimeout, Equals, time.Second*12)
	c.Assert(client.Conn.config.HTTPTimeout.IdleConnTimeout, Equals, time.Second*12)
	c.Assert(client.Conn.config.HTTPTimeout.LongTimeout, Equals, time.Second*12*10)

	c.Assert(client.Conn.config.SecurityToken, Equals, "token")
	c.Assert(client.Conn.config.IsCname, Equals, true)

	c.Assert(client.Conn.config.IsUseProxy, Equals, true)
	c.Assert(client.Config.ProxyHost, Equals, "http://127.0.0.1:8120")

	client, err = New(endpoint, accessID, accessKey, AuthProxy("http://127.0.0.1:8120", "user", "passwd"))

	c.Assert(client.Conn.config.IsUseProxy, Equals, true)
	c.Assert(client.Config.ProxyHost, Equals, "http://127.0.0.1:8120")
	c.Assert(client.Conn.config.IsAuthProxy, Equals, true)
	c.Assert(client.Conn.config.ProxyUser, Equals, "user")
	c.Assert(client.Conn.config.ProxyPassword, Equals, "passwd")

	client, err = New(endpoint, accessID, accessKey, UserAgent("go sdk user agent"))
	c.Assert(client.Conn.config.UserAgent, Equals, "go sdk user agent")

	// Check we can overide the http.Client
	httpClient := new(http.Client)
	client, err = New(endpoint, accessID, accessKey, HTTPClient(httpClient))
	c.Assert(client.HTTPClient, Equals, httpClient)
	c.Assert(client.Conn.client, Equals, httpClient)
	client, err = New(endpoint, accessID, accessKey)
	c.Assert(client.HTTPClient, IsNil)
}

// TestProxy
func (s *OssClientSuite) ProxyTestFunc(c *C, authVersion AuthVersionType, extraHeaders []string) {
	bucketNameTest := bucketNamePrefix + RandLowStr(6)
	objectName := "体育/奥运/首金"
	objectValue := "大江东去，浪淘尽，千古风流人物。 故垒西边，人道是、三国周郎赤壁。"

	client, err := New(endpoint, accessID, accessKey, AuthProxy(proxyHost, proxyUser, proxyPasswd))

	oldType := client.Config.AuthVersion
	oldHeaders := client.Config.AdditionalHeaders
	client.Config.AuthVersion = authVersion
	client.Config.AdditionalHeaders = extraHeaders

	// Create bucket
	err = client.CreateBucket(bucketNameTest)
	c.Assert(err, IsNil)

	// Get bucket info
	_, err = client.GetBucketInfo(bucketNameTest)
	c.Assert(err, IsNil)

	bucket, err := client.Bucket(bucketNameTest)

	// Sign URL
	str, err := bucket.SignURL(objectName, HTTPPut, 60)
	c.Assert(err, IsNil)
	if bucket.Client.Config.AuthVersion == AuthV1 {
		c.Assert(strings.Contains(str, HTTPParamExpires+"="), Equals, true)
		c.Assert(strings.Contains(str, HTTPParamAccessKeyID+"="), Equals, true)
		c.Assert(strings.Contains(str, HTTPParamSignature+"="), Equals, true)
	} else if bucket.Client.Config.AuthVersion == AuthV2 {
		c.Assert(strings.Contains(str, HTTPParamSignatureVersion+"=OSS2"), Equals, true)
		c.Assert(strings.Contains(str, HTTPParamExpiresV2+"="), Equals, true)
		c.Assert(strings.Contains(str, HTTPParamAccessKeyIDV2+"="), Equals, true)
		c.Assert(strings.Contains(str, HTTPParamSignatureV2+"="), Equals, true)
	}

	// Put object with URL
	err = bucket.PutObjectWithURL(str, strings.NewReader(objectValue))
	c.Assert(err, IsNil)

	// Sign URL for get object
	str, err = bucket.SignURL(objectName, HTTPGet, 60)
	c.Assert(err, IsNil)
	if bucket.Client.Config.AuthVersion == AuthV1 {
		c.Assert(strings.Contains(str, HTTPParamExpires+"="), Equals, true)
		c.Assert(strings.Contains(str, HTTPParamAccessKeyID+"="), Equals, true)
		c.Assert(strings.Contains(str, HTTPParamSignature+"="), Equals, true)
	} else {
		c.Assert(strings.Contains(str, HTTPParamSignatureVersion+"=OSS2"), Equals, true)
		c.Assert(strings.Contains(str, HTTPParamExpiresV2+"="), Equals, true)
		c.Assert(strings.Contains(str, HTTPParamAccessKeyIDV2+"="), Equals, true)
		c.Assert(strings.Contains(str, HTTPParamSignatureV2+"="), Equals, true)
	}

	// Get object with URL
	body, err := bucket.GetObjectWithURL(str)
	c.Assert(err, IsNil)
	str, err = readBody(body)
	c.Assert(err, IsNil)
	c.Assert(str, Equals, objectValue)

	// Put object
	err = bucket.PutObject(objectName, strings.NewReader(objectValue))
	c.Assert(err, IsNil)

	// Get object
	_, err = bucket.GetObject(objectName)
	c.Assert(err, IsNil)

	// List objects
	_, err = bucket.ListObjects()
	c.Assert(err, IsNil)

	// Delete object
	err = bucket.DeleteObject(objectName)
	c.Assert(err, IsNil)

	// Delete bucket
	err = client.DeleteBucket(bucketNameTest)
	c.Assert(err, IsNil)

	client.Config.AuthVersion = oldType
	client.Config.AdditionalHeaders = oldHeaders
}

func (s *OssClientSuite) TestProxy(c *C) {
	s.ProxyTestFunc(c, AuthV1, []string{})
	s.ProxyTestFunc(c, AuthV2, []string{})
	s.ProxyTestFunc(c, AuthV2, []string{"host", "range", "user-agent"})
}

// TestProxy for https endpoint
func (s *OssClientSuite) TestHttpsEndpointProxy(c *C) {
	bucketNameTest := bucketNamePrefix + RandLowStr(6)
	objectName := objectNamePrefix + RandLowStr(6)
	objectValue := RandLowStr(100)

	httpsEndPoint := ""
	if strings.HasPrefix(endpoint, "http://") {
		httpsEndPoint = strings.Replace(endpoint, "http://", "https://", 1)
	} else if !strings.HasPrefix(endpoint, "https://") {
		httpsEndPoint = "https://" + endpoint
	} else {
		httpsEndPoint = endpoint
	}

	client, err := New(httpsEndPoint, accessID, accessKey, AuthProxy(proxyHost, proxyUser, proxyPasswd))

	// Create bucket
	err = client.CreateBucket(bucketNameTest)
	c.Assert(err, IsNil)

	bucket, err := client.Bucket(bucketNameTest)

	// Put object
	err = bucket.PutObject(objectName, strings.NewReader(objectValue))
	c.Assert(err, IsNil)

	// Get object
	_, err = bucket.GetObject(objectName)
	c.Assert(err, IsNil)

	// List objects
	_, err = bucket.ListObjects()
	c.Assert(err, IsNil)

	// Delete object
	err = bucket.DeleteObject(objectName)
	c.Assert(err, IsNil)

	// Delete bucket
	err = client.DeleteBucket(bucketNameTest)
	c.Assert(err, IsNil)
}

func (s *OssBucketSuite) TestProxyNavigate(c *C) {
	client, err := New(endpoint, accessID, accessKey, AuthProxy("http://127.0.0.1:8120", "user", "passwd"))
	c.Assert(err, IsNil)
	_, err = client.GetBucketInfo(bucketName)
	c.Assert(strings.Contains(err.Error(), "proxyconnect tcp: dial tcp 127.0.0.1:8120"), Equals, true)
}

// Private
func (s *OssClientSuite) checkBucket(buckets []BucketProperties, bucket string) bool {
	for _, v := range buckets {
		if v.Name == bucket {
			return true
		}
	}
	return false
}

func (s *OssClientSuite) getBucket(buckets []BucketProperties, bucket string) (bool, BucketProperties) {
	for _, v := range buckets {
		if v.Name == bucket {
			return true, v
		}
	}
	return false, BucketProperties{}
}

func (s *OssClientSuite) TestHttpLogNotSignUrl(c *C) {
	logName := "." + string(os.PathSeparator) + "test-go-sdk-httpdebug.log" + RandStr(5)
	f, err := os.OpenFile(logName, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0660)
	c.Assert(err, IsNil)

	client, err := New(endpoint, accessID, accessKey)
	client.Config.LogLevel = Debug

	client.Config.Logger = log.New(f, "", log.LstdFlags)

	var testBucketName = bucketNamePrefix + RandLowStr(6)

	// CreateBucket
	err = client.CreateBucket(testBucketName)
	f.Close()

	// read log file,get http info
	contents, err := ioutil.ReadFile(logName)
	c.Assert(err, IsNil)

	httpContent := string(contents)
	//fmt.Println(httpContent)

	c.Assert(strings.Contains(httpContent, "signStr"), Equals, true)
	c.Assert(strings.Contains(httpContent, "Method:"), Equals, true)

	// delete test bucket and log
	os.Remove(logName)
	client.DeleteBucket(testBucketName)
}

func (s *OssClientSuite) HttpLogSignUrlTestFunc(c *C, authVersion AuthVersionType, extraHeaders []string) {
	logName := "." + string(os.PathSeparator) + "test-go-sdk-httpdebug-signurl.log" + RandStr(5)
	f, err := os.OpenFile(logName, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0660)
	c.Assert(err, IsNil)

	client, err := New(endpoint, accessID, accessKey)
	client.Config.LogLevel = Debug
	client.Config.Logger = log.New(f, "", log.LstdFlags)

	oldType := client.Config.AuthVersion
	oldHeaders := client.Config.AdditionalHeaders
	client.Config.AuthVersion = authVersion
	client.Config.AdditionalHeaders = extraHeaders

	var testBucketName = bucketNamePrefix + RandLowStr(6)

	// CreateBucket
	err = client.CreateBucket(testBucketName)
	f.Close()

	// clear log
	f, err = os.OpenFile(logName, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0660)
	client.Config.Logger = log.New(f, "", log.LstdFlags)

	bucket, _ := client.Bucket(testBucketName)
	objectName := objectNamePrefix + RandStr(8)
	objectValue := RandStr(20)

	// Sign URL for put
	str, err := bucket.SignURL(objectName, HTTPPut, 60)
	c.Assert(err, IsNil)
	if bucket.Client.Config.AuthVersion == AuthV1 {
		c.Assert(strings.Contains(str, HTTPParamExpires+"="), Equals, true)
		c.Assert(strings.Contains(str, HTTPParamAccessKeyID+"="), Equals, true)
		c.Assert(strings.Contains(str, HTTPParamSignature+"="), Equals, true)
	} else {
		c.Assert(strings.Contains(str, HTTPParamSignatureVersion+"=OSS2"), Equals, true)
		c.Assert(strings.Contains(str, HTTPParamExpiresV2+"="), Equals, true)
		c.Assert(strings.Contains(str, HTTPParamAccessKeyIDV2+"="), Equals, true)
		c.Assert(strings.Contains(str, HTTPParamSignatureV2+"="), Equals, true)
	}

	// Error put object with URL
	err = bucket.PutObjectWithURL(str, strings.NewReader(objectValue), ContentType("image/tiff"))
	f.Close()

	// read log file,get http info
	contents, err := ioutil.ReadFile(logName)
	c.Assert(err, IsNil)

	httpContent := string(contents)
	//fmt.Println(httpContent)

	c.Assert(strings.Contains(httpContent, "signStr"), Equals, true)
	c.Assert(strings.Contains(httpContent, "Method:"), Equals, true)

	// delete test bucket and log
	os.Remove(logName)
	client.DeleteBucket(testBucketName)

	client.Config.AuthVersion = oldType
	client.Config.AdditionalHeaders = oldHeaders
}

func (s *OssClientSuite) TestHttpLogSignUrl(c *C) {
	s.HttpLogSignUrlTestFunc(c, AuthV1, []string{})
	s.HttpLogSignUrlTestFunc(c, AuthV2, []string{})
	s.HttpLogSignUrlTestFunc(c, AuthV2, []string{"host", "range", "user-agent"})
}

func (s *OssClientSuite) TestSetLimitUploadSpeed(c *C) {
	client, err := New(endpoint, accessID, accessKey)
	c.Assert(err, IsNil)

	err = client.LimitUploadSpeed(100)

	goVersion := runtime.Version()
	pSlice := strings.Split(strings.ToLower(goVersion), ".")

	// compare with go1.7
	if len(pSlice) >= 2 {
		if pSlice[0] > "go1" {
			c.Assert(err, IsNil)
		} else if pSlice[0] == "go1" {
			subVersion, _ := strconv.Atoi(pSlice[1])
			if subVersion >= 7 {
				c.Assert(err, IsNil)
			} else {
				c.Assert(err, NotNil)
			}
		} else {
			c.Assert(err, NotNil)
		}
	}
}

func (s *OssClientSuite) TestBucketEncyptionError(c *C) {
	client, err := New(endpoint, accessID, accessKey)
	c.Assert(err, IsNil)

	bucketName := bucketNamePrefix + RandLowStr(5)
	err = client.CreateBucket(bucketName)
	c.Assert(err, IsNil)

	// SetBucketEncryption:AES256 ,"123"
	encryptionRule := ServerEncryptionRule{}
	encryptionRule.SSEDefault.SSEAlgorithm = string(AESAlgorithm)
	encryptionRule.SSEDefault.KMSMasterKeyID = "123"

	var responseHeader http.Header
	err = client.SetBucketEncryption(bucketName, encryptionRule, GetResponseHeader(&responseHeader))
	c.Assert(err, NotNil)
	requestId := GetRequestId(responseHeader)
	c.Assert(len(requestId) > 0, Equals, true)

	// GetBucketEncryption
	_, err = client.GetBucketEncryption(bucketName, GetResponseHeader(&responseHeader))
	c.Assert(err, NotNil)
	requestId = GetRequestId(responseHeader)
	c.Assert(len(requestId) > 0, Equals, true)

	// Get default bucket info
	bucketResult, err := client.GetBucketInfo(bucketName)
	c.Assert(err, IsNil)

	c.Assert(bucketResult.BucketInfo.SseRule.SSEAlgorithm, Equals, "")
	c.Assert(bucketResult.BucketInfo.SseRule.KMSMasterKeyID, Equals, "")
	c.Assert(bucketResult.BucketInfo.Versioning, Equals, "")

	err = client.DeleteBucket(bucketName)
	c.Assert(err, IsNil)
}

func (s *OssClientSuite) TestBucketEncryptionPutAndGetAndDelete(c *C) {
	client, err := New(endpoint, accessID, accessKey)
	c.Assert(err, IsNil)

	bucketName := bucketNamePrefix + RandLowStr(5)
	err = client.CreateBucket(bucketName)
	c.Assert(err, IsNil)

	// SetBucketEncryption:KMS ,""
	encryptionRule := ServerEncryptionRule{}
	encryptionRule.SSEDefault.SSEAlgorithm = string(KMSAlgorithm)

	var responseHeader http.Header
	err = client.SetBucketEncryption(bucketName, encryptionRule, GetResponseHeader(&responseHeader))
	c.Assert(err, IsNil)
	requestId := GetRequestId(responseHeader)
	c.Assert(len(requestId) > 0, Equals, true)

	// GetBucketEncryption
	getResult, err := client.GetBucketEncryption(bucketName, GetResponseHeader(&responseHeader))
	c.Assert(err, IsNil)
	requestId = GetRequestId(responseHeader)
	c.Assert(len(requestId) > 0, Equals, true)

	// check encryption value
	c.Assert(encryptionRule.SSEDefault.SSEAlgorithm, Equals, getResult.SSEDefault.SSEAlgorithm)
	c.Assert(encryptionRule.SSEDefault.KMSMasterKeyID, Equals, getResult.SSEDefault.KMSMasterKeyID)

	// delete bucket encyption
	err = client.DeleteBucketEncryption(bucketName, GetResponseHeader(&responseHeader))
	c.Assert(err, IsNil)
	requestId = GetRequestId(responseHeader)
	c.Assert(len(requestId) > 0, Equals, true)

	// GetBucketEncryption failure
	_, err = client.GetBucketEncryption(bucketName, GetResponseHeader(&responseHeader))
	c.Assert(err, NotNil)
	requestId = GetRequestId(responseHeader)
	c.Assert(len(requestId) > 0, Equals, true)

	// Get default bucket info
	bucketResult, err := client.GetBucketInfo(bucketName)
	c.Assert(err, IsNil)

	c.Assert(bucketResult.BucketInfo.SseRule.SSEAlgorithm, Equals, "")
	c.Assert(bucketResult.BucketInfo.SseRule.KMSMasterKeyID, Equals, "")
	c.Assert(bucketResult.BucketInfo.Versioning, Equals, "")

	err = client.DeleteBucket(bucketName)
	c.Assert(err, IsNil)
}

func (s *OssClientSuite) TestBucketEncryptionWithSm4(c *C) {
	client, err := New(endpoint, accessID, accessKey)
	c.Assert(err, IsNil)

	bucketName := bucketNamePrefix + RandLowStr(5)
	err = client.CreateBucket(bucketName)
	c.Assert(err, IsNil)

	// SetBucketEncryption:SM4 ,""
	encryptionRule := ServerEncryptionRule{}
	encryptionRule.SSEDefault.SSEAlgorithm = string(SM4Algorithm)

	var responseHeader http.Header
	err = client.SetBucketEncryption(bucketName, encryptionRule, GetResponseHeader(&responseHeader))
	c.Assert(err, IsNil)
	requestId := GetRequestId(responseHeader)
	c.Assert(len(requestId) > 0, Equals, true)

	// GetBucketEncryption
	getResult, err := client.GetBucketEncryption(bucketName, GetResponseHeader(&responseHeader))
	c.Assert(err, IsNil)
	requestId = GetRequestId(responseHeader)
	c.Assert(len(requestId) > 0, Equals, true)

	// check encryption value
	c.Assert(getResult.SSEDefault.SSEAlgorithm, Equals, string(SM4Algorithm))
	c.Assert(getResult.SSEDefault.KMSMasterKeyID, Equals, "")
	c.Assert(getResult.SSEDefault.KMSDataEncryption, Equals, "")

	// Get default bucket info
	bucketResult, err := client.GetBucketInfo(bucketName)
	c.Assert(err, IsNil)

	c.Assert(bucketResult.BucketInfo.SseRule.SSEAlgorithm, Equals, string(SM4Algorithm))
	c.Assert(bucketResult.BucketInfo.SseRule.KMSMasterKeyID, Equals, "")
	c.Assert(bucketResult.BucketInfo.SseRule.KMSDataEncryption, Equals, "")

	err = client.DeleteBucket(bucketName)
	c.Assert(err, IsNil)
}

func (s *OssClientSuite) TestBucketEncryptionWithKmsSm4(c *C) {
	client, err := New(endpoint, accessID, accessKey)
	c.Assert(err, IsNil)

	bucketName := bucketNamePrefix + RandLowStr(5)
	err = client.CreateBucket(bucketName)
	c.Assert(err, IsNil)

	// SetBucketEncryption:SM4 ,""
	encryptionRule := ServerEncryptionRule{}
	encryptionRule.SSEDefault.SSEAlgorithm = string(KMSAlgorithm)
	encryptionRule.SSEDefault.KMSDataEncryption = string(SM4Algorithm)

	var responseHeader http.Header
	err = client.SetBucketEncryption(bucketName, encryptionRule, GetResponseHeader(&responseHeader))
	c.Assert(err, IsNil)
	requestId := GetRequestId(responseHeader)
	c.Assert(len(requestId) > 0, Equals, true)

	// GetBucketEncryption
	getResult, err := client.GetBucketEncryption(bucketName, GetResponseHeader(&responseHeader))
	c.Assert(err, IsNil)
	requestId = GetRequestId(responseHeader)
	c.Assert(len(requestId) > 0, Equals, true)

	// check encryption value
	c.Assert(getResult.SSEDefault.SSEAlgorithm, Equals, string(KMSAlgorithm))
	c.Assert(getResult.SSEDefault.KMSMasterKeyID, Equals, "")
	c.Assert(getResult.SSEDefault.KMSDataEncryption, Equals, string(SM4Algorithm))

	// Get default bucket info
	bucketResult, err := client.GetBucketInfo(bucketName)
	c.Assert(err, IsNil)

	c.Assert(bucketResult.BucketInfo.SseRule.SSEAlgorithm, Equals, string(KMSAlgorithm))
	c.Assert(bucketResult.BucketInfo.SseRule.KMSMasterKeyID, Equals, "")
	c.Assert(bucketResult.BucketInfo.SseRule.KMSDataEncryption, Equals, string(SM4Algorithm))

	err = client.DeleteBucket(bucketName)
	c.Assert(err, IsNil)
}

func (s *OssClientSuite) TestBucketEncyptionPutObjectSuccess(c *C) {
	client, err := New(endpoint, accessID, accessKey)
	c.Assert(err, IsNil)

	bucketName := bucketNamePrefix + RandLowStr(5)
	err = client.CreateBucket(bucketName)
	c.Assert(err, IsNil)

	// SetBucketEncryption:KMS ,""
	encryptionRule := ServerEncryptionRule{}
	encryptionRule.SSEDefault.SSEAlgorithm = string(KMSAlgorithm)

	var responseHeader http.Header
	err = client.SetBucketEncryption(bucketName, encryptionRule, GetResponseHeader(&responseHeader))
	c.Assert(err, IsNil)
	requestId := GetRequestId(responseHeader)
	c.Assert(len(requestId) > 0, Equals, true)

	// GetBucketEncryption
	getResult, err := client.GetBucketEncryption(bucketName, GetResponseHeader(&responseHeader))
	c.Assert(err, IsNil)
	requestId = GetRequestId(responseHeader)
	c.Assert(len(requestId) > 0, Equals, true)

	// check encryption value
	c.Assert(encryptionRule.SSEDefault.SSEAlgorithm, Equals, getResult.SSEDefault.SSEAlgorithm)
	c.Assert(encryptionRule.SSEDefault.KMSMasterKeyID, Equals, getResult.SSEDefault.KMSMasterKeyID)

	// Get default bucket info
	bucketResult, err := client.GetBucketInfo(bucketName)
	c.Assert(err, IsNil)

	c.Assert(bucketResult.BucketInfo.SseRule.SSEAlgorithm, Equals, "KMS")
	c.Assert(bucketResult.BucketInfo.SseRule.KMSMasterKeyID, Equals, "")
	c.Assert(bucketResult.BucketInfo.Versioning, Equals, "")

	err = client.DeleteBucket(bucketName)
	c.Assert(err, IsNil)
}

func (s *OssClientSuite) TestBucketEncyptionPutObjectError(c *C) {
	client, err := New(endpoint, accessID, accessKey)
	c.Assert(err, IsNil)

	bucketName := bucketNamePrefix + RandLowStr(5)
	err = client.CreateBucket(bucketName)
	c.Assert(err, IsNil)

	// SetBucketEncryption:KMS ,""
	encryptionRule := ServerEncryptionRule{}
	encryptionRule.SSEDefault.SSEAlgorithm = string(KMSAlgorithm)
	kmsId := "123"
	encryptionRule.SSEDefault.KMSMasterKeyID = kmsId

	var responseHeader http.Header
	err = client.SetBucketEncryption(bucketName, encryptionRule, GetResponseHeader(&responseHeader))
	c.Assert(err, IsNil)
	requestId := GetRequestId(responseHeader)
	c.Assert(len(requestId) > 0, Equals, true)

	// GetBucketEncryption
	getResult, err := client.GetBucketEncryption(bucketName, GetResponseHeader(&responseHeader))
	c.Assert(err, IsNil)
	requestId = GetRequestId(responseHeader)
	c.Assert(len(requestId) > 0, Equals, true)

	// check encryption value
	c.Assert(encryptionRule.SSEDefault.SSEAlgorithm, Equals, getResult.SSEDefault.SSEAlgorithm)
	c.Assert(encryptionRule.SSEDefault.KMSMasterKeyID, Equals, getResult.SSEDefault.KMSMasterKeyID)

	// Get default bucket info
	bucketResult, err := client.GetBucketInfo(bucketName)
	c.Assert(err, IsNil)

	c.Assert(bucketResult.BucketInfo.SseRule.SSEAlgorithm, Equals, "KMS")
	c.Assert(bucketResult.BucketInfo.SseRule.KMSMasterKeyID, Equals, kmsId)
	c.Assert(bucketResult.BucketInfo.Versioning, Equals, "")

	// put and get object failure
	bucket, err := client.Bucket(bucketName)
	c.Assert(err, IsNil)

	// put object failure
	objectName := objectNamePrefix + RandStr(8)
	context := RandStr(100)
	err = bucket.PutObject(objectName, strings.NewReader(context))
	c.Assert(err, NotNil)

	err = client.DeleteBucket(bucketName)
	c.Assert(err, IsNil)
}

func (s *OssClientSuite) TestMaxConnsPutObjectSuccess(c *C) {
	client, err := New(endpoint, accessID, accessKey)
	c.Assert(err, IsNil)

	bucketName := bucketNamePrefix + RandLowStr(5)
	err = client.CreateBucket(bucketName)
	c.Assert(err, IsNil)

	bucket, err := client.Bucket(bucketName)
	c.Assert(err, IsNil)

	c.Assert(bucket.GetConfig().HTTPMaxConns.MaxIdleConns, Equals, 100)
	c.Assert(bucket.GetConfig().HTTPMaxConns.MaxIdleConnsPerHost, Equals, 100)
	c.Assert(bucket.GetConfig().HTTPMaxConns.MaxConnsPerHost, Equals, 0)

	// put object
	objectName := objectNamePrefix + RandLowStr(5)
	err = bucket.PutObject(objectName, strings.NewReader(RandStr(10)))
	c.Assert(err, IsNil)

	bucket.DeleteObject(objectName)
	err = bucket.PutObject(objectName, strings.NewReader(RandStr(10)))
	c.Assert(err, IsNil)
	bucket.DeleteObject(objectName)

	client.DeleteBucket(bucketName)
}

func (s *OssClientSuite) TestBucketTaggingOperation(c *C) {
	client, err := New(endpoint, accessID, accessKey)
	c.Assert(err, IsNil)

	bucketName := bucketNamePrefix + RandLowStr(5)
	err = client.CreateBucket(bucketName)
	c.Assert(err, IsNil)

	var respHeader http.Header

	// Bucket Tagging
	var tagging Tagging
	tagging.Tags = []Tag{{Key: "testkey2", Value: "testvalue2"}}
	err = client.SetBucketTagging(bucketName, tagging, GetResponseHeader(&respHeader))
	c.Assert(err, IsNil)
	c.Assert(GetRequestId(respHeader) != "", Equals, true)

	getResult, err := client.GetBucketTagging(bucketName)
	c.Assert(err, IsNil)
	c.Assert(getResult.Tags[0].Key, Equals, tagging.Tags[0].Key)
	c.Assert(getResult.Tags[0].Value, Equals, tagging.Tags[0].Value)

	// delete BucketTagging
	err = client.DeleteBucketTagging(bucketName, GetResponseHeader(&respHeader))
	c.Assert(err, IsNil)
	c.Assert(GetRequestId(respHeader) != "", Equals, true)

	getResult, err = client.GetBucketTagging(bucketName, GetResponseHeader(&respHeader))
	c.Assert(err, IsNil)
	c.Assert(GetRequestId(respHeader) != "", Equals, true)
	c.Assert(len(getResult.Tags), Equals, 0)

	err = client.DeleteBucket(bucketName)
	c.Assert(err, IsNil)
}

func (s *OssClientSuite) TestBucketTaggingWithKey(c *C) {
	client, err := New(endpoint, accessID, accessKey)
	c.Assert(err, IsNil)

	bucketName := bucketNamePrefix + RandLowStr(5)
	err = client.CreateBucket(bucketName)
	c.Assert(err, IsNil)
	// Bucket Tagging
	var tagging Tagging
	tagging.Tags = []Tag{{Key: "k1", Value: "v1"}, {Key: "k2", Value: "v2"}, {Key: "k3", Value: "v3"}, {Key: "k4", Value: "v4"}, {Key: "k5", Value: "v5"}, {Key: "k6", Value: "v6"}}
	err = client.SetBucketTagging(bucketName, tagging)
	c.Assert(err, IsNil)

	getResult, err := client.GetBucketTagging(bucketName)
	c.Assert(err, IsNil)
	c.Assert(getResult.Tags[0].Key, Equals, tagging.Tags[0].Key)
	c.Assert(getResult.Tags[0].Value, Equals, tagging.Tags[0].Value)

	// delete BucketTagging
	err = client.DeleteBucketTagging(bucketName, addParam("tagging", "k1"))
	c.Assert(err, IsNil)

	getResult, err = client.GetBucketTagging(bucketName)
	c.Assert(err, IsNil)
	c.Assert(len(getResult.Tags), Equals, 5)

	err = client.DeleteBucketTagging(bucketName, addParam("tagging", "k2,k3,k4"))
	c.Assert(err, IsNil)

	getResult, err = client.GetBucketTagging(bucketName)
	c.Assert(err, IsNil)
	c.Assert(len(getResult.Tags), Equals, 2)

	err = client.DeleteBucketTagging(bucketName)
	c.Assert(err, IsNil)

	getResult, err = client.GetBucketTagging(bucketName)
	c.Assert(err, IsNil)
	c.Assert(len(getResult.Tags), Equals, 0)

	err = client.DeleteBucket(bucketName)
	c.Assert(err, IsNil)
}

func (s *OssClientSuite) TestListBucketsTagging(c *C) {
	client, err := New(endpoint, accessID, accessKey)
	c.Assert(err, IsNil)

	bucketName1 := bucketNamePrefix + RandLowStr(5)
	err = client.CreateBucket(bucketName1)
	c.Assert(err, IsNil)

	bucketName2 := bucketNamePrefix + RandLowStr(5)
	err = client.CreateBucket(bucketName2)
	c.Assert(err, IsNil)

	// Bucket Tagging
	var tagging Tagging
	tagging.Tags = []Tag{{Key: "testkey", Value: "testvalue"}}
	err = client.SetBucketTagging(bucketName1, tagging)
	c.Assert(err, IsNil)

	// list bucket
	listResult, err := client.ListBuckets(TagKey("testkey"))
	c.Assert(err, IsNil)
	c.Assert(len(listResult.Buckets), Equals, 1)
	c.Assert(listResult.Buckets[0].Name, Equals, bucketName1)

	client.DeleteBucket(bucketName1)
	client.DeleteBucket(bucketName2)
}

func (s *OssClientSuite) TestGetBucketStat(c *C) {
	client, err := New(endpoint, accessID, accessKey)
	c.Assert(err, IsNil)

	bucketName := bucketNamePrefix + RandLowStr(5)
	err = client.CreateBucket(bucketName)
	c.Assert(err, IsNil)

	bucket, err := client.Bucket(bucketName)
	c.Assert(err, IsNil)

	// put object
	objectName := objectNamePrefix + RandLowStr(5)
	err = bucket.PutObject(objectName, strings.NewReader(RandStr(10)))
	c.Assert(err, IsNil)

	bucket.DeleteObject(objectName)
	err = bucket.PutObject(objectName, strings.NewReader(RandStr(10)))
	c.Assert(err, IsNil)
	bucket.DeleteObject(objectName)

	_, err = client.GetBucketStat(bucketName)
	c.Assert(err, IsNil)

	client.DeleteBucket(bucketName)
}

func (s *OssBucketSuite) TestGetBucketVersioning(c *C) {
	// create a bucket with default proprety
	client, err := New(endpoint, accessID, accessKey)
	c.Assert(err, IsNil)

	bucketName := bucketNamePrefix + RandLowStr(6)

	var respHeader http.Header
	err = client.CreateBucket(bucketName, GetResponseHeader(&respHeader))
	c.Assert(err, IsNil)
	c.Assert(GetRequestId(respHeader) != "", Equals, true)

	// put bucket version:enabled
	var versioningConfig VersioningConfig
	versioningConfig.Status = string(VersionEnabled)
	err = client.SetBucketVersioning(bucketName, versioningConfig)
	c.Assert(err, IsNil)
	time.Sleep(timeoutInOperation)

	// get bucket version success
	versioningResult, err := client.GetBucketVersioning(bucketName, GetResponseHeader(&respHeader))
	c.Assert(versioningResult.Status, Equals, "Enabled")
	c.Assert(GetRequestId(respHeader) != "", Equals, true)

	ForceDeleteBucket(client, bucketName, c)
}

func (s *OssClientSuite) TestBucketPolicy(c *C) {
	client, err := New(endpoint, accessID, accessKey)
	c.Assert(err, IsNil)

	bucketName := bucketNamePrefix + RandLowStr(5)
	err = client.CreateBucket(bucketName)
	c.Assert(err, IsNil)

	var responseHeader http.Header
	ret, err := client.GetBucketPolicy(bucketName, GetResponseHeader(&responseHeader))
	c.Assert(err, NotNil)
	requestId := GetRequestId(responseHeader)
	c.Assert(len(requestId) > 0, Equals, true)

	policyInfo := `
	{
		"Version":"1",
		"Statement":[
			{
				"Action":[
					"oss:GetObject",
					"oss:PutObject"
				],
				"Effect":"Deny",
				"Principal":"[123456790]",
				"Resource":["acs:oss:*:1234567890:*/*"]
			}
		]
	}`

	err = client.SetBucketPolicy(bucketName, policyInfo, GetResponseHeader(&responseHeader))
	c.Assert(err, IsNil)
	requestId = GetRequestId(responseHeader)
	c.Assert(len(requestId) > 0, Equals, true)

	ret, err = client.GetBucketPolicy(bucketName, GetResponseHeader(&responseHeader))
	c.Assert(err, IsNil)
	testLogger.Println("policy:", ret)
	c.Assert(ret, Equals, policyInfo)
	requestId = GetRequestId(responseHeader)
	c.Assert(len(requestId) > 0, Equals, true)

	err = client.DeleteBucketPolicy(bucketName, GetResponseHeader(&responseHeader))
	c.Assert(err, IsNil)
	requestId = GetRequestId(responseHeader)
	c.Assert(len(requestId) > 0, Equals, true)
	client.DeleteBucket(bucketName)
}

func (s *OssClientSuite) TestBucketPolicyNegative(c *C) {
	client, err := New(endpoint, accessID, accessKey)
	c.Assert(err, IsNil)

	bucketName := bucketNamePrefix + RandLowStr(5)
	err = client.CreateBucket(bucketName)
	c.Assert(err, IsNil)

	var responseHeader http.Header
	_, err = client.GetBucketPolicy(bucketName, GetResponseHeader(&responseHeader))
	c.Assert(err, NotNil)
	requestId := GetRequestId(responseHeader)
	c.Assert(len(requestId) > 0, Equals, true)

	// Setting the Version is 2, this is error policy
	errPolicy := `
	{
		"Version":"2",
		"Statement":[
			{
				"Action":[
					"oss:GetObject",
					"oss:PutObject"
				],
				"Effect":"Deny",
				"Principal":"[123456790]",
				"Resource":["acs:oss:*:1234567890:*/*"]
			}
		]
	}`
	err = client.SetBucketPolicy(bucketName, errPolicy, GetResponseHeader(&responseHeader))
	c.Assert(err, NotNil)
	testLogger.Println("err:", err)
	requestId = GetRequestId(responseHeader)
	c.Assert(len(requestId) > 0, Equals, true)

	err = client.DeleteBucketPolicy(bucketName, GetResponseHeader(&responseHeader))
	c.Assert(err, IsNil)

	bucketNameEmpty := bucketNamePrefix + RandLowStr(5)
	client.DeleteBucket(bucketNameEmpty)

	err = client.DeleteBucketPolicy(bucketNameEmpty, GetResponseHeader(&responseHeader))
	c.Assert(err, NotNil)
	requestId = GetRequestId(responseHeader)
	c.Assert(len(requestId) > 0, Equals, true)

	client.DeleteBucket(bucketName)
}

func (s *OssClientSuite) TestSetBucketRequestPayment(c *C) {
	client, err := New(endpoint, accessID, accessKey)
	c.Assert(err, IsNil)

	bucketName := bucketNamePrefix + RandLowStr(5)
	err = client.CreateBucket(bucketName)
	c.Assert(err, IsNil)

	reqPayConf := RequestPaymentConfiguration{
		Payer: "Requester",
	}
	err = client.SetBucketRequestPayment(bucketName, reqPayConf)
	c.Assert(err, IsNil)

	ret, err := client.GetBucketRequestPayment(bucketName)
	c.Assert(err, IsNil)
	c.Assert(ret.Payer, Equals, "Requester")

	client.DeleteBucket(bucketName)
	c.Assert(err, IsNil)
}

func (s *OssClientSuite) TestSetBucketRequestPaymentNegative(c *C) {
	client, err := New(endpoint, accessID, accessKey)
	c.Assert(err, IsNil)

	bucketName := bucketNamePrefix + RandLowStr(5)
	err = client.CreateBucket(bucketName)
	c.Assert(err, IsNil)

	reqPayConf := RequestPaymentConfiguration{
		Payer: "Requesterttttt", // this is a error configuration
	}
	err = client.SetBucketRequestPayment(bucketName, reqPayConf)
	c.Assert(err, NotNil)

	ret, err := client.GetBucketRequestPayment(bucketName)
	c.Assert(err, IsNil)
	c.Assert(ret.Payer, Equals, "BucketOwner")

	client.DeleteBucket(bucketName)
	c.Assert(err, IsNil)
}

func (s *OssClientSuite) TestBucketQos(c *C) {
	client, err := New(endpoint, accessID, accessKey)
	c.Assert(err, IsNil)

	ret, err := client.GetUserQoSInfo()
	c.Assert(err, IsNil)
	testLogger.Println("QosInfo:", ret)

	bucketName := bucketNamePrefix + RandLowStr(5)
	_ = client.DeleteBucket(bucketName)

	err = client.CreateBucket(bucketName)
	c.Assert(err, IsNil)

	_, err = client.GetBucketQosInfo(bucketName)
	c.Assert(err, NotNil)

	// case 1 set BucketQoSConfiguration every member
	five := 5
	four := 4
	totalQps := 200
	qosConf := BucketQoSConfiguration{
		TotalUploadBandwidth:      &five,
		IntranetUploadBandwidth:   &four,
		ExtranetUploadBandwidth:   &four,
		TotalDownloadBandwidth:    &four,
		IntranetDownloadBandwidth: &four,
		ExtranetDownloadBandwidth: &four,
		TotalQPS:                  &totalQps,
		IntranetQPS:               &totalQps,
		ExtranetQPS:               &totalQps,
	}
	var responseHeader http.Header
	err = client.SetBucketQoSInfo(bucketName, qosConf, GetResponseHeader(&responseHeader))
	c.Assert(err, IsNil)
	requestId := GetRequestId(responseHeader)
	c.Assert(len(requestId) > 0, Equals, true)

	// wait a moment for configuration effect
	time.Sleep(timeoutInOperation)

	retQos, err := client.GetBucketQosInfo(bucketName)
	c.Assert(err, IsNil)

	// set qosConf default value
	qosConf.XMLName.Local = "QoSConfiguration"
	c.Assert(struct2string(retQos, c), Equals, struct2string(qosConf, c))

	// case 2 set BucketQoSConfiguration not every member
	qosConfNo := BucketQoSConfiguration{
		TotalUploadBandwidth:      &five,
		IntranetUploadBandwidth:   &four,
		ExtranetUploadBandwidth:   &four,
		TotalDownloadBandwidth:    &four,
		IntranetDownloadBandwidth: &four,
		ExtranetDownloadBandwidth: &four,
		TotalQPS:                  &totalQps,
	}
	err = client.SetBucketQoSInfo(bucketName, qosConfNo)
	c.Assert(err, IsNil)

	// wait a moment for configuration effect
	time.Sleep(timeoutInOperation)

	retQos, err = client.GetBucketQosInfo(bucketName)
	c.Assert(err, IsNil)

	// set qosConfNo default value
	qosConfNo.XMLName.Local = "QoSConfiguration"
	defNum := -1
	qosConfNo.IntranetQPS = &defNum
	qosConfNo.ExtranetQPS = &defNum
	c.Assert(struct2string(retQos, c), Equals, struct2string(qosConfNo, c))

	err = client.DeleteBucketQosInfo(bucketName)
	c.Assert(err, IsNil)

	// wait a moment for configuration effect
	time.Sleep(timeoutInOperation)

	_, err = client.GetBucketQosInfo(bucketName)
	c.Assert(err, NotNil)

	// this is a error qos configuration
	to := *ret.TotalUploadBandwidth + 2
	qosErrConf := BucketQoSConfiguration{
		TotalUploadBandwidth:      &to, // this exceed user TotalUploadBandwidth
		IntranetUploadBandwidth:   &four,
		ExtranetUploadBandwidth:   &four,
		TotalDownloadBandwidth:    &four,
		IntranetDownloadBandwidth: &four,
		ExtranetDownloadBandwidth: &four,
		TotalQPS:                  &totalQps,
		IntranetQPS:               &totalQps,
		ExtranetQPS:               &totalQps,
	}
	err = client.SetBucketQoSInfo(bucketName, qosErrConf)
	c.Assert(err, NotNil)

	err = client.DeleteBucketQosInfo(bucketName)
	c.Assert(err, IsNil)

	err = client.DeleteBucket(bucketName)
	c.Assert(err, IsNil)
}

// struct to string
func struct2string(obj interface{}, c *C) string {
	t := reflect.TypeOf(obj)
	v := reflect.ValueOf(obj)

	var data = make(map[string]interface{})
	for i := 0; i < t.NumField(); i++ {
		data[t.Field(i).Name] = v.Field(i).Interface()
	}
	str, err := json.Marshal(data)
	c.Assert(err, IsNil)
	return string(str)
}

type TestCredentials struct {
}

func (testCreInf *TestCredentials) GetAccessKeyID() string {
	return os.Getenv("OSS_TEST_ACCESS_KEY_ID")
}

func (testCreInf *TestCredentials) GetAccessKeySecret() string {
	return os.Getenv("OSS_TEST_ACCESS_KEY_SECRET")
}

func (testCreInf *TestCredentials) GetSecurityToken() string {
	return ""
}

type TestCredentialsProvider struct {
}

func (testInfBuild *TestCredentialsProvider) GetCredentials() Credentials {
	return &TestCredentials{}
}

func (s *OssClientSuite) TestClientCredentialInfBuild(c *C) {
	var bucketNameTest = bucketNamePrefix + RandLowStr(6)
	var defaultBuild TestCredentialsProvider
	client, err := New(endpoint, "", "", SetCredentialsProvider(&defaultBuild))
	c.Assert(err, IsNil)
	err = client.CreateBucket(bucketNameTest)
	c.Assert(err, IsNil)
	err = client.DeleteBucket(bucketNameTest)
	c.Assert(err, IsNil)
}

func (s *OssClientSuite) TestClientNewEnvironmentVariableCredentialsProvider(c *C) {
	provider, err := ProviderWrongKeyId()
	c.Assert(err, NotNil)
	c.Assert(err.Error(), Equals, "access key id is empty!")
	provider, err = ProviderWrongSecret()
	c.Assert(err, NotNil)
	c.Assert(err.Error(), Equals, "access key secret is empty!")

	os.Setenv("OSS_ACCESS_KEY_ID", accessID)
	os.Setenv("OSS_ACCESS_KEY_SECRET", accessKey)
	provider, err = NewEnvironmentVariableCredentialsProvider()
	c.Assert(err, IsNil)
	var bucketNameTest = bucketNamePrefix + RandLowStr(6)
	client, err := New(endpoint, "", "", SetCredentialsProvider(&provider))
	c.Assert(err, IsNil)
	err = client.CreateBucket(bucketNameTest)
	c.Assert(err, IsNil)
	err = client.DeleteBucket(bucketNameTest)
	c.Assert(err, IsNil)
}

func ProviderWrongKeyId() (EnvironmentVariableCredentialsProvider, error) {
	var provider EnvironmentVariableCredentialsProvider
	keyId := os.Getenv("OSS_ACCESS_KEY_ID_d")
	if keyId == "" {
		return provider, fmt.Errorf("access key id is empty!")
	}
	secret := os.Getenv("OSS_ACCESS_KEY_SECRET_d")
	if accessKey == "" {
		return provider, fmt.Errorf("access key secret is empty!")
	}
	token := os.Getenv("OSS_SESSION_TOKEN")

	envCredential := &envCredentials{
		AccessKeyId:     keyId,
		AccessKeySecret: secret,
		SecurityToken:   token,
	}
	return EnvironmentVariableCredentialsProvider{
		cred: envCredential,
	}, nil
}

func ProviderWrongSecret() (EnvironmentVariableCredentialsProvider, error) {
	var provider EnvironmentVariableCredentialsProvider
	keyId := os.Getenv("OSS_ACCESS_KEY_ID")
	if keyId == "" {
		return provider, fmt.Errorf("access key id is empty!")
	}
	secret := os.Getenv("OSS_ACCESS_KEY_SECRET_NOT_EXIST")
	if secret == "" {
		return provider, fmt.Errorf("access key secret is empty!")
	}
	token := os.Getenv("OSS_SESSION_TOKEN")

	envCredential := &envCredentials{
		AccessKeyId:     keyId,
		AccessKeySecret: secret,
		SecurityToken:   token,
	}
	return EnvironmentVariableCredentialsProvider{
		cred: envCredential,
	}, nil
}

func (s *OssClientSuite) TestClientSetLocalIpError(c *C) {
	// create client and bucket
	ipAddr, err := net.ResolveIPAddr("ip", "127.0.0.1")
	c.Assert(err, IsNil)
	localTCPAddr := &(net.TCPAddr{IP: ipAddr.IP})
	client, err := New(endpoint, accessID, accessKey, SetLocalAddr(localTCPAddr))
	c.Assert(err, IsNil)

	var bucketNameTest = bucketNamePrefix + RandLowStr(6)
	err = client.CreateBucket(bucketNameTest)
	c.Assert(err, NotNil)
}

func (s *OssClientSuite) TestClientSetLocalIpSuccess(c *C) {
	//get local ip
	conn, err := net.Dial("udp", "8.8.8.8:80")
	c.Assert(err, IsNil)
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	localIp := localAddr.IP.String()
	conn.Close()

	ipAddr, err := net.ResolveIPAddr("ip", localIp)
	c.Assert(err, IsNil)
	localTCPAddr := &(net.TCPAddr{IP: ipAddr.IP})
	client, err := New(endpoint, accessID, accessKey, SetLocalAddr(localTCPAddr))
	c.Assert(err, IsNil)

	var bucketNameTest = bucketNamePrefix + RandLowStr(6)
	err = client.CreateBucket(bucketNameTest)
	c.Assert(err, IsNil)
	err = client.DeleteBucket(bucketNameTest)
	c.Assert(err, IsNil)
}

// TestCreateBucketInvalidName
func (s *OssClientSuite) TestCreateBucketInvalidName(c *C) {
	var bucketNameTest = "-" + bucketNamePrefix + RandLowStr(6)
	client, err := New(endpoint, accessID, accessKey)
	c.Assert(err, IsNil)
	// Create
	err = client.CreateBucket(bucketNameTest)
	c.Assert(err, NotNil)
}

// TestClientProcessEndpointSuccess
func (s *OssClientSuite) TestClientProcessEndpointSuccess(c *C) {
	var bucketNameTest = bucketNamePrefix + RandLowStr(6)

	testEndpoint := endpoint + "/" + "sina.com" + "?" + "para=abc"

	client, err := New(testEndpoint, accessID, accessKey)
	c.Assert(err, IsNil)

	// Create
	err = client.CreateBucket(bucketNameTest)
	c.Assert(err, IsNil)

	// delete
	err = client.DeleteBucket(bucketNameTest)
	c.Assert(err, IsNil)
}

// TestClientProcessEndpointSuccess
func (s *OssClientSuite) TestClientProcessEndpointError(c *C) {
	var bucketNameTest = bucketNamePrefix + RandLowStr(6)

	testEndpoint := "https://127.0.0.1/" + endpoint

	client, err := New(testEndpoint, accessID, accessKey)
	c.Assert(err, IsNil)

	// Create
	err = client.CreateBucket(bucketNameTest)
	c.Assert(err, NotNil)
}

// TestClientBucketError
func (s *OssClientSuite) TestClientBucketError(c *C) {
	client, err := New(endpoint, accessID, accessKey)
	c.Assert(err, IsNil)

	bucketName := "-" + RandLowStr(5)
	_, err = client.Bucket(bucketName)
	c.Assert(err, NotNil)
}

func (s *OssClientSuite) TestSetBucketInventory(c *C) {
	client, err := New(endpoint, accessID, accessKey)
	c.Assert(err, IsNil)

	bucketName := bucketNamePrefix + RandLowStr(5)
	err = client.CreateBucket(bucketName)
	c.Assert(err, IsNil)

	bucket, err := client.Bucket(bucketName)
	c.Assert(err, IsNil)
	err = bucket.PutObject("key", strings.NewReader(""), ServerSideEncryption("AES256"))

	pros, err := bucket.GetObjectDetailedMeta("key")

	bucket.DeleteObject("key")

	// encryption config
	var invSseOss InvSseOss
	invSseKms := InvSseKms{
		KmsId: pros.Get("x-oss-server-side-encryption-key-id"),
	}
	var invEncryption InvEncryption

	bl := true
	// not any encryption
	invConfig := InventoryConfiguration{
		Id:        "report1",
		IsEnabled: &bl,
		Prefix:    "filterPrefix/",
		OSSBucketDestination: OSSBucketDestination{
			Format:    "CSV",
			AccountId: accountID,
			RoleArn:   stsARN,
			Bucket:    "acs:oss:::" + bucketName,
			Prefix:    "prefix1",
		},
		Frequency:              "Daily",
		IncludedObjectVersions: "All",
		OptionalFields: OptionalFields{
			Field: []string{
				"Size", "LastModifiedDate", "ETag", "StorageClass", "IsMultipartUploaded", "EncryptionStatus",
			},
		},
	}

	// case 1: not any encryption
	err = client.SetBucketInventory(bucketName, invConfig)
	c.Assert(err, IsNil)

	// case 2: use kms encryption
	invConfig.Id = "report2"
	invEncryption.SseKms = &invSseKms
	invEncryption.SseOss = nil
	invConfig.OSSBucketDestination.Encryption = &invEncryption
	err = client.SetBucketInventory(bucketName, invConfig)
	c.Assert(err, IsNil)

	// case 3: use SseOss encryption
	invConfig.Id = "report3"
	invEncryption.SseKms = nil
	invEncryption.SseOss = &invSseOss
	invConfig.OSSBucketDestination.Encryption = &invEncryption
	err = client.SetBucketInventory(bucketName, invConfig)
	c.Assert(err, IsNil)

	//case 4: use two type encryption
	invConfig.Id = "report4"
	invEncryption.SseKms = &invSseKms
	invEncryption.SseOss = &invSseOss
	invConfig.OSSBucketDestination.Encryption = &invEncryption
	err = client.SetBucketInventory(bucketName, invConfig)
	c.Assert(err, NotNil)

	err = client.DeleteBucket(bucketName)
	c.Assert(err, IsNil)
}

func (s *OssClientSuite) TestBucketInventory(c *C) {
	client, err := New(endpoint, accessID, accessKey)
	c.Assert(err, IsNil)

	bucketName := bucketNamePrefix + RandLowStr(5)
	err = client.CreateBucket(bucketName)
	c.Assert(err, IsNil)

	bl := true
	invConfig := InventoryConfiguration{
		Id:        "report1",
		IsEnabled: &bl,
		Prefix:    "filterPrefix/",
		OSSBucketDestination: OSSBucketDestination{
			Format:    "CSV",
			AccountId: accountID,
			RoleArn:   stsARN,
			Bucket:    "acs:oss:::" + bucketName,
			Prefix:    "prefix1",
		},
		Frequency:              "Daily",
		IncludedObjectVersions: "All",
		OptionalFields: OptionalFields{
			Field: []string{
				"Size", "LastModifiedDate", "ETag", "StorageClass", "IsMultipartUploaded", "EncryptionStatus",
			},
		},
	}

	// case 1: test SetBucketInventory
	err = client.SetBucketInventory(bucketName, invConfig)
	c.Assert(err, IsNil)

	// case 2: test GetBucketInventory
	out, err := client.GetBucketInventory(bucketName, "report1")
	c.Assert(err, IsNil)
	invConfig.XMLName.Local = "InventoryConfiguration"
	invConfig.OSSBucketDestination.XMLName.Local = "OSSBucketDestination"
	invConfig.OptionalFields.XMLName.Local = "OptionalFields"
	c.Assert(struct2string(invConfig, c), Equals, struct2string(out, c))

	// case 3: test ListBucketInventory
	invConfig2 := InventoryConfiguration{
		Id:        "report2",
		IsEnabled: &bl,
		Prefix:    "filterPrefix/",
		OSSBucketDestination: OSSBucketDestination{
			Format:    "CSV",
			AccountId: accountID,
			RoleArn:   stsARN,
			Bucket:    "acs:oss:::" + bucketName,
			Prefix:    "prefix1",
		},
		Frequency:              "Daily",
		IncludedObjectVersions: "All",
		OptionalFields: OptionalFields{
			Field: []string{
				"Size", "LastModifiedDate", "ETag", "StorageClass", "IsMultipartUploaded", "EncryptionStatus",
			},
		},
	}
	invConfig2.XMLName.Local = "InventoryConfiguration"
	invConfig2.OSSBucketDestination.XMLName.Local = "OSSBucketDestination"
	invConfig2.OptionalFields.XMLName.Local = "OptionalFields"

	err = client.SetBucketInventory(bucketName, invConfig2)
	c.Assert(err, IsNil)

	listInvConf, err := client.ListBucketInventory(bucketName, "", Marker("report1"), MaxKeys(2))
	c.Assert(err, IsNil)
	var listInvLocal ListInventoryConfigurationsResult
	listInvLocal.InventoryConfiguration = []InventoryConfiguration{
		invConfig,
		invConfig2,
	}
	bo := false
	listInvLocal.IsTruncated = &bo
	listInvLocal.XMLName.Local = "ListInventoryConfigurationsResult"
	c.Assert(struct2string(listInvLocal, c), Equals, struct2string(listInvConf, c))

	for i := 3; i < 109; i++ {
		invConfig2 := InventoryConfiguration{
			Id:        "report" + strconv.Itoa(i),
			IsEnabled: &bl,
			Prefix:    "filterPrefix/",
			OSSBucketDestination: OSSBucketDestination{
				Format:    "CSV",
				AccountId: accountID,
				RoleArn:   stsARN,
				Bucket:    "acs:oss:::" + bucketName,
				Prefix:    "prefix1",
			},
			Frequency:              "Daily",
			IncludedObjectVersions: "All",
			OptionalFields: OptionalFields{
				Field: []string{
					"Size", "LastModifiedDate", "ETag", "StorageClass", "IsMultipartUploaded", "EncryptionStatus",
				},
			},
		}
		err = client.SetBucketInventory(bucketName, invConfig2)
		c.Assert(err, IsNil)
	}
	token := ""
	for {
		listInvConf1, err := client.ListBucketInventory(bucketName, token)
		c.Assert(err, IsNil)
		token = listInvConf1.NextContinuationToken
		testLogger.Println(listInvConf1.NextContinuationToken, *listInvConf1.IsTruncated, token)
		if *listInvConf1.IsTruncated == false {
			break
		} else {
			c.Assert(listInvConf1.NextContinuationToken, Equals, "report91")
		}
	}

	// case 4: test DeleteBucketInventory
	for i := 1; i < 109; i++ {
		err = client.DeleteBucketInventory(bucketName, "report"+strconv.Itoa(i))
		c.Assert(err, IsNil)
	}

	err = client.DeleteBucket(bucketName)
	c.Assert(err, IsNil)
}

func (s *OssClientSuite) TestBucketInventoryNegative(c *C) {
	client, err := New(endpoint, accessID, accessKey)
	c.Assert(err, IsNil)

	bucketName := bucketNamePrefix + RandLowStr(5)
	err = client.CreateBucket(bucketName)
	c.Assert(err, IsNil)

	bl := true
	invConfigErr := InventoryConfiguration{
		Id:        "report1",
		IsEnabled: &bl,
		Prefix:    "filterPrefix/",
		OSSBucketDestination: OSSBucketDestination{
			Format:    "CSV",
			AccountId: accountID,
			RoleArn:   stsARN,
			Bucket:    "test",
			Prefix:    "prefix1",
		},
		Frequency:              "Daily",
		IncludedObjectVersions: "All",
		OptionalFields: OptionalFields{
			Field: []string{
				"Size", "LastModifiedDate", "ETag", "StorageClass", "IsMultipartUploaded", "EncryptionStatus",
			},
		},
	}
	// case 1: test SetBucketInventory
	err = client.SetBucketInventory(bucketName, invConfigErr)
	c.Assert(err, NotNil)

	// case 2: test GetBucketInventory
	_, err = client.GetBucketInventory(bucketName, "report1")
	c.Assert(err, NotNil)

	// case 3: test ListBucketInventory
	_, err = client.ListBucketInventory(bucketName, "", Marker("report1"), MaxKeys(2))
	c.Assert(err, NotNil)

	// case 4: test DeleteBucketInventory
	err = client.DeleteBucketInventory(bucketName, "report1")
	c.Assert(err, IsNil)

	err = client.DeleteBucket(bucketName)
	c.Assert(err, IsNil)
}

func (s *OssClientSuite) TestBucketInventoryXmlSuccess(c *C) {
	client, err := New(endpoint, accessID, accessKey)
	c.Assert(err, IsNil)

	bucketName := bucketNamePrefix + RandLowStr(5)
	err = client.CreateBucket(bucketName)
	c.Assert(err, IsNil)

	bl := true
	invConfig := InventoryConfiguration{
		Id:        "report1",
		IsEnabled: &bl,
		Prefix:    "filterPrefix/",
		OSSBucketDestination: OSSBucketDestination{
			Format:    "CSV",
			AccountId: accountID,
			RoleArn:   stsARN,
			Bucket:    "acs:oss:::" + bucketName,
			Prefix:    "prefix1",
		},
		Frequency:              "Daily",
		IncludedObjectVersions: "All",
		OptionalFields: OptionalFields{
			Field: []string{
				"Size", "LastModifiedDate", "ETag", "StorageClass", "IsMultipartUploaded", "EncryptionStatus",
			},
		},
	}

	tagBucket := "acs:oss:::" + bucketName

	var xmlBody []byte
	xmlBody, _ = xml.Marshal(invConfig)

	// set enventory
	err = client.SetBucketInventoryXml(bucketName, string(xmlBody))
	c.Assert(err, IsNil)

	// get enventory
	xmlGet, err := client.GetBucketInventoryXml(bucketName, "report1")
	c.Assert(err, IsNil)
	c.Assert(strings.Contains(xmlGet, tagBucket), Equals, true)

	// list enventory
	xmlList, err := client.ListBucketInventoryXml(bucketName, "", Marker("report1"))
	c.Assert(err, IsNil)
	c.Assert(strings.Contains(xmlList, tagBucket), Equals, true)

	err = client.DeleteBucket(bucketName)
	c.Assert(err, IsNil)
}

func (s *OssClientSuite) TestBucketInventoryXmlError(c *C) {
	client, err := New(endpoint, accessID, accessKey)
	c.Assert(err, IsNil)

	bucketName := bucketNamePrefix + RandLowStr(5)
	err = client.CreateBucket(bucketName)
	c.Assert(err, IsNil)

	// not xml format
	xmlBody := `
	<InventoryConfiguration>
	<Id>report1</Id>
	<IsEnabled>true</IsEnabled>
	`
	err = client.SetBucketInventoryXml(bucketName, xmlBody)
	c.Assert(err, NotNil)

	// no id
	xmlBody = `
	<InventoryConfiguration>
	<IsEnabled>true</IsEnabled>
	<Filter>
	   <Prefix>filterPrefix/</Prefix>
		  <LastModifyBeginTimeStamp>1637883649</LastModifyBeginTimeStamp>
		  <LastModifyEndTimeStamp>1638347592</LastModifyEndTimeStamp>
		  <LowerSizeBound>1024</LowerSizeBound>
		  <UpperSizeBound>1048576</UpperSizeBound>
	   <StorageClass>Standard,IA</StorageClass>
	</Filter>
 </InventoryConfiguration>
	`
	err = client.SetBucketInventoryXml(bucketName, xmlBody)
	c.Assert(err, NotNil)

	err = client.DeleteBucket(bucketName)
	c.Assert(err, IsNil)
}

func (s *OssClientSuite) TestBucketAsyncTask(c *C) {
	client, err := New(endpoint, accessID, accessKey)
	c.Assert(err, IsNil)

	bucketName := bucketNamePrefix + RandLowStr(5)
	err = client.CreateBucket(bucketName)
	c.Assert(err, IsNil)

	objectName := objectNamePrefix + RandLowStr(6)

	// set asyn task,IgnoreSameKey is false
	asynConf := AsyncFetchTaskConfiguration{
		Url:           "http://www.baidu.com",
		Object:        objectName,
		Host:          "",
		ContentMD5:    "",
		Callback:      "",
		StorageClass:  "",
		IgnoreSameKey: false,
	}

	asynResult, err := client.SetBucketAsyncTask(bucketName, asynConf)
	c.Assert(err, IsNil)
	c.Assert(len(asynResult.TaskId) > 0, Equals, true)

	// get asyn task
	asynTask, err := client.GetBucketAsyncTask(bucketName, asynResult.TaskId)
	c.Assert(err, IsNil)
	c.Assert(asynResult.TaskId, Equals, asynTask.TaskId)
	c.Assert(len(asynTask.State) > 0, Equals, true)
	c.Assert(asynConf.Url, Equals, asynTask.TaskInfo.Url)
	c.Assert(asynConf.Object, Equals, asynTask.TaskInfo.Object)
	c.Assert(asynConf.Callback, Equals, asynTask.TaskInfo.Callback)
	c.Assert(asynConf.IgnoreSameKey, Equals, asynTask.TaskInfo.IgnoreSameKey)

	// test again,IgnoreSameKey is true
	asynConf.IgnoreSameKey = true
	asynResult, err = client.SetBucketAsyncTask(bucketName, asynConf)
	c.Assert(err, IsNil)
	c.Assert(len(asynResult.TaskId) > 0, Equals, true)

	asynTask, err = client.GetBucketAsyncTask(bucketName, asynResult.TaskId)
	c.Assert(asynConf.IgnoreSameKey, Equals, asynTask.TaskInfo.IgnoreSameKey)

	err = client.DeleteBucket(bucketName)
	c.Assert(err, IsNil)
}

func (s *OssClientSuite) TestClientOptionHeader(c *C) {
	// create a bucket with default proprety
	client, err := New(endpoint, accessID, accessKey)
	c.Assert(err, IsNil)

	bucketName := bucketNamePrefix + RandLowStr(6)

	var respHeader http.Header
	err = client.CreateBucket(bucketName, GetResponseHeader(&respHeader))
	c.Assert(err, IsNil)
	c.Assert(GetRequestId(respHeader) != "", Equals, true)

	// put bucket version:enabled
	var versioningConfig VersioningConfig
	versioningConfig.Status = string(VersionEnabled)
	err = client.SetBucketVersioning(bucketName, versioningConfig)
	c.Assert(err, IsNil)
	time.Sleep(timeoutInOperation)

	// get bucket version success,use payer
	options := []Option{RequestPayer(BucketOwner), GetResponseHeader(&respHeader)}
	versioningResult, err := client.GetBucketVersioning(bucketName, options...)
	c.Assert(versioningResult.Status, Equals, "Enabled")
	c.Assert(GetRequestId(respHeader) != "", Equals, true)

	//list buckets,use payer
	_, err = client.ListBuckets(options...)
	c.Assert(err, IsNil)

	ForceDeleteBucket(client, bucketName, c)
}

// compare with go1.7
func compareVersion(goVersion string) bool {
	nowVersion := runtime.Version()
	nowVersion = strings.Replace(nowVersion, "go", "", -1)
	pSlice1 := strings.Split(goVersion, ".")
	pSlice2 := strings.Split(nowVersion, ".")
	for k, v := range pSlice2 {
		n2, _ := strconv.Atoi(string(v))
		n1, _ := strconv.Atoi(string(pSlice1[k]))
		if n2 > n1 {
			return true
		}
		if n2 < n1 {
			return false
		}
	}
	return true
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/redirectTo", http.StatusFound)
}
func targetHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "You have been redirected here!")
}

func (s *OssClientSuite) TestClientRedirect(c *C) {
	// must go1.7.0 onward
	if !compareVersion("1.7.0") {
		return
	}

	// get port
	rand.Seed(time.Now().Unix())
	port := 10000 + rand.Intn(10000)

	// start http server
	httpAddr := fmt.Sprintf("127.0.0.1:%d", port)
	mux := http.NewServeMux()
	mux.HandleFunc("/redirectTo", targetHandler)
	mux.HandleFunc("/", homeHandler)
	svr := &http.Server{
		Addr:           httpAddr,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
		Handler:        mux,
	}

	go func() {
		svr.ListenAndServe()
	}()

	time.Sleep(3 * time.Second)

	url := "http://" + httpAddr

	// create client 1,redirect disable
	client1, err := New(endpoint, accessID, accessKey, RedirectEnabled(false))
	resp, err := client1.Conn.client.Get(url)
	c.Assert(err, IsNil)
	c.Assert(resp.StatusCode, Equals, http.StatusFound)
	resp.Body.Close()

	// create client2, redirect enabled
	client2, err := New(endpoint, accessID, accessKey, RedirectEnabled(true))
	resp, err = client2.Conn.client.Get(url)
	c.Assert(err, IsNil)
	c.Assert(resp.StatusCode, Equals, 200)
	data, err := ioutil.ReadAll(resp.Body)
	c.Assert(string(data), Equals, "You have been redirected here!")
	resp.Body.Close()

	svr.Close()
}

func verifyCertificatehandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("verifyCertificatehandler"))
}

func (s *OssClientSuite) TestClientSkipVerifyCertificateTestServer(c *C) {
	// get port
	rand.Seed(time.Now().Unix())
	port := 10000 + rand.Intn(10000)

	// start https server
	httpAddr := fmt.Sprintf("127.0.0.1:%d", port)
	mux := http.NewServeMux()
	mux.HandleFunc("/", verifyCertificatehandler)
	svr := &http.Server{
		Addr:           httpAddr,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
		Handler:        mux,
	}

	go func() {
		svr.ListenAndServeTLS("../sample/test_cert.pem", "../sample/test_key.pem")
	}()

	// wait http server started
	time.Sleep(3 * time.Second)

	url := "https://" + httpAddr

	// create client 1,not verify certificate
	client1, err := New(endpoint, accessID, accessKey, InsecureSkipVerify(true))
	resp, err := client1.Conn.client.Get(url)
	c.Assert(err, IsNil)
	c.Assert(resp.StatusCode, Equals, 200)
	data, err := ioutil.ReadAll(resp.Body)
	c.Assert(string(data), Equals, "verifyCertificatehandler")
	resp.Body.Close()

	// create client2, verify certificate
	client2, err := New(endpoint, accessID, accessKey, InsecureSkipVerify(false))
	resp, err = client2.Conn.client.Get(url)
	c.Assert(err, NotNil)
	fmt.Println(err)
}

func (s *OssClientSuite) TestClientSkipVerifyCertificateOssServer(c *C) {
	// create a bucket with default proprety
	client, err := New(endpoint, accessID, accessKey, InsecureSkipVerify(true))
	c.Assert(err, IsNil)

	bucketName := bucketNamePrefix + RandLowStr(6)
	err = client.CreateBucket(bucketName)
	c.Assert(err, IsNil)
	bucket, err := client.Bucket(bucketName)

	objectName := objectNamePrefix + RandStr(8)
	objectLen := 1000
	objectValue := RandStr(objectLen)

	// Put
	err = bucket.PutObject(objectName, strings.NewReader(objectValue))
	c.Assert(err, IsNil)

	//
	resp, err := bucket.GetObject(objectName)
	c.Assert(err, IsNil)
	str, err := readBody(resp)
	c.Assert(err, IsNil)
	c.Assert(str, Equals, objectValue)

	ForceDeleteBucket(client, bucketName, c)

}

// TestInitiateBucketWormSuccess
func (s *OssClientSuite) TestInitiateBucketWormSuccess(c *C) {
	var bucketNameTest = bucketNamePrefix + RandLowStr(6)
	client, err := New(endpoint, accessID, accessKey)
	c.Assert(err, IsNil)

	err = client.CreateBucket(bucketNameTest)
	c.Assert(err, IsNil)

	// InitiateBucketWorm
	wormId, err := client.InitiateBucketWorm(bucketNameTest, 10)
	c.Assert(err, IsNil)
	c.Assert(len(wormId) > 0, Equals, true)

	// GetBucketWorm
	wormConfig, err := client.GetBucketWorm(bucketNameTest)
	c.Assert(err, IsNil)
	c.Assert(wormConfig.WormId, Equals, wormId)
	c.Assert(wormConfig.State, Equals, "InProgress")
	c.Assert(wormConfig.RetentionPeriodInDays, Equals, 10)

	err = client.DeleteBucket(bucketNameTest)
	c.Assert(err, IsNil)
}

// TestInitiateBucketWormFailure
func (s *OssClientSuite) TestInitiateBucketWormFailure(c *C) {
	var bucketNameTest = bucketNamePrefix + RandLowStr(6)
	client, err := New(endpoint, accessID, accessKey)
	c.Assert(err, IsNil)

	// bucket not exist
	wormId, err := client.InitiateBucketWorm(bucketNameTest, 10)
	c.Assert(err, NotNil)
	c.Assert(len(wormId), Equals, 0)
}

// TestAbortBucketWorm
func (s *OssClientSuite) TestAbortBucketWorm(c *C) {
	var bucketNameTest = bucketNamePrefix + RandLowStr(6)
	client, err := New(endpoint, accessID, accessKey)
	c.Assert(err, IsNil)

	err = client.CreateBucket(bucketNameTest)
	c.Assert(err, IsNil)

	// InitiateBucketWorm
	wormId, err := client.InitiateBucketWorm(bucketNameTest, 10)
	c.Assert(err, IsNil)
	c.Assert(len(wormId) > 0, Equals, true)

	// GetBucketWorm success
	wormConfig, err := client.GetBucketWorm(bucketNameTest)
	c.Assert(err, IsNil)
	c.Assert(wormConfig.WormId, Equals, wormId)
	c.Assert(wormConfig.State, Equals, "InProgress")
	c.Assert(wormConfig.RetentionPeriodInDays, Equals, 10)

	// abort worm
	err = client.AbortBucketWorm(bucketNameTest)
	c.Assert(err, IsNil)

	// GetBucketWorm failure
	_, err = client.GetBucketWorm(bucketNameTest)
	c.Assert(err, NotNil)

	err = client.DeleteBucket(bucketNameTest)
	c.Assert(err, IsNil)
}

// TestCompleteBucketWorm
func (s *OssClientSuite) TestCompleteBucketWorm(c *C) {
	var bucketNameTest = bucketNamePrefix + "-worm-" + RandLowStr(6)
	client, err := New(endpoint, accessID, accessKey)
	c.Assert(err, IsNil)

	err = client.CreateBucket(bucketNameTest)
	c.Assert(err, IsNil)

	// InitiateBucketWorm
	wormId, err := client.InitiateBucketWorm(bucketNameTest, 1)
	c.Assert(err, IsNil)
	c.Assert(len(wormId) > 0, Equals, true)

	// GetBucketWorm
	wormConfig, err := client.GetBucketWorm(bucketNameTest)
	c.Assert(err, IsNil)
	c.Assert(wormConfig.WormId, Equals, wormId)
	c.Assert(wormConfig.State, Equals, "InProgress")
	c.Assert(wormConfig.RetentionPeriodInDays, Equals, 1)

	// CompleteBucketWorm
	err = client.CompleteBucketWorm(bucketNameTest, wormId)
	c.Assert(err, IsNil)

	// GetBucketWorm again
	wormConfig, err = client.GetBucketWorm(bucketNameTest)
	c.Assert(err, IsNil)
	c.Assert(wormConfig.WormId, Equals, wormId)
	c.Assert(wormConfig.State, Equals, "Locked")
	c.Assert(wormConfig.RetentionPeriodInDays, Equals, 1)

	err = client.DeleteBucket(bucketNameTest)
	c.Assert(err, IsNil)
}

// TestExtendBucketWorm
func (s *OssClientSuite) TestExtendBucketWorm(c *C) {
	var bucketNameTest = bucketNamePrefix + "-worm-" + RandLowStr(6)
	client, err := New(endpoint, accessID, accessKey)
	c.Assert(err, IsNil)

	err = client.CreateBucket(bucketNameTest)
	c.Assert(err, IsNil)

	// InitiateBucketWorm
	wormId, err := client.InitiateBucketWorm(bucketNameTest, 1)
	c.Assert(err, IsNil)
	c.Assert(len(wormId) > 0, Equals, true)

	// CompleteBucketWorm
	err = client.CompleteBucketWorm(bucketNameTest, wormId)
	c.Assert(err, IsNil)

	// GetBucketWorm
	wormConfig, err := client.GetBucketWorm(bucketNameTest)
	c.Assert(err, IsNil)
	c.Assert(wormConfig.WormId, Equals, wormId)
	c.Assert(wormConfig.State, Equals, "Locked")
	c.Assert(wormConfig.RetentionPeriodInDays, Equals, 1)

	// CompleteBucketWorm
	err = client.ExtendBucketWorm(bucketNameTest, 2, wormId)
	c.Assert(err, IsNil)

	// GetBucketWorm again
	wormConfig, err = client.GetBucketWorm(bucketNameTest)
	c.Assert(err, IsNil)
	c.Assert(wormConfig.WormId, Equals, wormId)
	c.Assert(wormConfig.State, Equals, "Locked")
	c.Assert(wormConfig.RetentionPeriodInDays, Equals, 2)

	err = client.DeleteBucket(bucketNameTest)
	c.Assert(err, IsNil)
}

// TestBucketTransferAcc
func (s *OssClientSuite) TestBucketTransferAcc(c *C) {
	var bucketNameTest = bucketNamePrefix + "-acc-" + RandLowStr(6)
	client, err := New(endpoint, accessID, accessKey)
	c.Assert(err, IsNil)

	err = client.CreateBucket(bucketNameTest)
	c.Assert(err, IsNil)

	accConfig := TransferAccConfiguration{}
	accConfig.Enabled = true

	// SetBucketTransferAcc true
	err = client.SetBucketTransferAcc(bucketNameTest, accConfig)
	c.Assert(err, IsNil)

	// GetBucketTransferAcc
	accConfigRes, err := client.GetBucketTransferAcc(bucketNameTest)
	c.Assert(err, IsNil)
	c.Assert(accConfigRes.Enabled, Equals, true)

	// SetBucketTransferAcc false
	accConfig.Enabled = false
	err = client.SetBucketTransferAcc(bucketNameTest, accConfig)
	c.Assert(err, IsNil)

	// GetBucketTransferAcc
	accConfigRes, err = client.GetBucketTransferAcc(bucketNameTest)
	c.Assert(err, IsNil)
	c.Assert(accConfigRes.Enabled, Equals, false)

	// DeleteBucketTransferAcc
	err = client.DeleteBucketTransferAcc(bucketNameTest)
	c.Assert(err, IsNil)

	// GetBucketTransferAcc
	_, err = client.GetBucketTransferAcc(bucketNameTest)
	c.Assert(err, NotNil)

	err = client.DeleteBucket(bucketNameTest)
	c.Assert(err, IsNil)
}

// TestBucketReplicationPutAndGet
func (s *OssClientSuite) TestBucketReplicationPutAndGet(c *C) {
	var sourceBucketNameTest = bucketNamePrefix + "-replication-" + RandLowStr(6)

	sourceRegion := "hangzhou"
	sourceEndpoint := "oss-cn-" + sourceRegion + ".aliyuncs.com"

	client1, err := New(sourceEndpoint, accessID, accessKey)
	c.Assert(err, IsNil)

	err = client1.CreateBucket(sourceBucketNameTest)
	c.Assert(err, IsNil)

	var destinationBucketNameTest = bucketNamePrefix + "-replication-" + RandLowStr(6)

	destinationRegion := "beijing"
	destinationEndpoint := "oss-cn-" + destinationRegion + ".aliyuncs.com"

	client2, err := New(destinationEndpoint, accessID, accessKey)
	c.Assert(err, IsNil)

	err = client2.CreateBucket(destinationBucketNameTest)
	c.Assert(err, IsNil)

	putXml := `<?xml version="1.0" encoding="UTF-8"?>
	<ReplicationConfiguration>
	  <Rule>
		<Action>PUT</Action>
		<Destination>
		  <Bucket>` + destinationBucketNameTest + `</Bucket>
		  <Location>oss-cn-` + destinationRegion + `</Location>
		</Destination>
		<HistoricalObjectReplication>enabled</HistoricalObjectReplication>
	  </Rule>
	</ReplicationConfiguration>`

	// replication command and put method test
	err = client1.PutBucketReplication(sourceBucketNameTest, putXml)
	c.Assert(err, IsNil)

	time.Sleep(5 * time.Second)

	data, err := client1.GetBucketReplication(sourceBucketNameTest)
	c.Assert(err, IsNil)

	var result GetResult
	err = xml.Unmarshal([]byte(data), &result)
	c.Assert(err, IsNil)

	c.Assert(result.Rules[0].Status, Equals, "starting")
	c.Assert(result.Rules[0].Destination.Location, Equals, "oss-cn-"+destinationRegion)
	c.Assert(result.Rules[0].Destination.Bucket, Equals, destinationBucketNameTest)
	c.Assert(result.Rules[0].HistoricalObjectReplication, Equals, "enabled")

	err = client1.DeleteBucket(sourceBucketNameTest)
	c.Assert(err, IsNil)

	err = client2.DeleteBucket(destinationBucketNameTest)
	c.Assert(err, IsNil)

}

// TestBucketReplicationDelete
func (s *OssClientSuite) TestBucketReplicationDeleteSuccess(c *C) {
	var sourceBucketNameTest = bucketNamePrefix + "-replication-" + RandLowStr(6)

	sourceRegion := "hangzhou"
	sourceEndpoint := "oss-cn-" + sourceRegion + ".aliyuncs.com"

	client1, err := New(sourceEndpoint, accessID, accessKey)
	c.Assert(err, IsNil)

	err = client1.CreateBucket(sourceBucketNameTest)
	c.Assert(err, IsNil)

	var destinationBucketNameTest = bucketNamePrefix + "-replication-" + RandLowStr(6)

	destinationRegion := "beijing"
	destinationEndpoint := "oss-cn-" + destinationRegion + ".aliyuncs.com"

	client2, err := New(destinationEndpoint, accessID, accessKey)
	c.Assert(err, IsNil)

	err = client2.CreateBucket(destinationBucketNameTest)
	c.Assert(err, IsNil)

	putXml := `<?xml version="1.0" encoding="UTF-8"?>
	<ReplicationConfiguration>
	  <Rule>
		<Action>PUT</Action>
		<Destination>
		  <Bucket>` + destinationBucketNameTest + `</Bucket>
		  <Location>oss-cn-` + destinationRegion + `</Location>
		</Destination>
		<HistoricalObjectReplication>enabled</HistoricalObjectReplication>
	  </Rule>
	</ReplicationConfiguration>`

	// replication command and put method test
	err = client1.PutBucketReplication(sourceBucketNameTest, putXml)
	c.Assert(err, IsNil)

	time.Sleep(5 * time.Second)

	data, err := client1.GetBucketReplication(sourceBucketNameTest)
	c.Assert(err, IsNil)

	var result GetResult
	err = xml.Unmarshal([]byte(data), &result)
	c.Assert(err, IsNil)

	c.Assert(result.Rules[0].Status, Equals, "starting")
	c.Assert(result.Rules[0].Destination.Location, Equals, "oss-cn-"+destinationRegion)
	c.Assert(result.Rules[0].Destination.Bucket, Equals, destinationBucketNameTest)
	c.Assert(result.Rules[0].HistoricalObjectReplication, Equals, "enabled")

	ruleID := result.Rules[0].ID

	err = client1.DeleteBucketReplication(sourceBucketNameTest, ruleID)
	c.Assert(err, IsNil)

	// get again
	afterDeleteData, err := client1.GetBucketReplication(sourceBucketNameTest)
	c.Assert(err, IsNil)

	var afterDeleteResult GetResult
	err = xml.Unmarshal([]byte(afterDeleteData), &afterDeleteResult)
	c.Assert(err, IsNil)
	c.Assert(afterDeleteResult.Rules[0].Status, Equals, "closing")

	err = client1.DeleteBucket(sourceBucketNameTest)
	c.Assert(err, IsNil)

	err = client2.DeleteBucket(destinationBucketNameTest)
	c.Assert(err, IsNil)
}

// TestBucketReplicationDelete
func (s *OssClientSuite) TestBucketReplicationDeleteWithEmptyRuleID(c *C) {
	var sourceBucketNameTest = bucketNamePrefix + "-replication-" + RandLowStr(6)

	sourceRegion := "hangzhou"
	sourceEndpoint := "oss-cn-" + sourceRegion + ".aliyuncs.com"

	client1, err := New(sourceEndpoint, accessID, accessKey)
	c.Assert(err, IsNil)

	err = client1.CreateBucket(sourceBucketNameTest)
	c.Assert(err, IsNil)

	var destinationBucketNameTest = bucketNamePrefix + "-replication-" + RandLowStr(6)

	destinationRegion := "beijing"
	destinationEndpoint := "oss-cn-" + destinationRegion + ".aliyuncs.com"

	client2, err := New(destinationEndpoint, accessID, accessKey)
	c.Assert(err, IsNil)

	err = client2.CreateBucket(destinationBucketNameTest)
	c.Assert(err, IsNil)

	putXml := `<?xml version="1.0" encoding="UTF-8"?>
	<ReplicationConfiguration>
	  <Rule>
		<Action>PUT</Action>
		<Destination>
		  <Bucket>` + destinationBucketNameTest + `</Bucket>
		  <Location>oss-cn-` + destinationRegion + `</Location>
		</Destination>
		<HistoricalObjectReplication>enabled</HistoricalObjectReplication>
	  </Rule>
	</ReplicationConfiguration>`

	// replication command and put method test
	err = client1.PutBucketReplication(sourceBucketNameTest, putXml)
	c.Assert(err, IsNil)

	time.Sleep(5 * time.Second)

	data, err := client1.GetBucketReplication(sourceBucketNameTest)
	c.Assert(err, IsNil)

	var result GetBucketReplicationResult
	err = xml.Unmarshal([]byte(data), &result)
	c.Assert(err, IsNil)

	c.Assert(result.Rule[0].Status, Equals, "starting")
	c.Assert(result.Rule[0].Destination.Location, Equals, "oss-cn-"+destinationRegion)
	c.Assert(result.Rule[0].Destination.Bucket, Equals, destinationBucketNameTest)
	c.Assert(result.Rule[0].HistoricalObjectReplication, Equals, "enabled")

	ruleID := ""

	err = client1.DeleteBucketReplication(sourceBucketNameTest, ruleID)
	c.Assert(err, NotNil)

	// get again
	afterDeleteData, err := client1.GetBucketReplication(sourceBucketNameTest)
	c.Assert(err, IsNil)

	var afterDeleteResult GetResult
	err = xml.Unmarshal([]byte(afterDeleteData), &afterDeleteResult)
	c.Assert(err, IsNil)
	c.Assert(afterDeleteResult.Rules[0].Status, Equals, "starting")

	err = client1.DeleteBucket(sourceBucketNameTest)
	c.Assert(err, IsNil)

	err = client2.DeleteBucket(destinationBucketNameTest)
	c.Assert(err, IsNil)
}

// TestBucketReplicationGetLocation
func (s *OssClientSuite) TestBucketReplicationGetLocation(c *C) {
	var sourceBucketNameTest = bucketNamePrefix + "-replication-" + RandLowStr(6)

	sourceRegion := "hangzhou"
	sourceEndpoint := "oss-cn-" + sourceRegion + ".aliyuncs.com"

	client1, err := New(sourceEndpoint, accessID, accessKey)
	c.Assert(err, IsNil)

	err = client1.CreateBucket(sourceBucketNameTest)
	c.Assert(err, IsNil)

	data, err := client1.GetBucketReplicationLocation(sourceBucketNameTest)
	c.Assert(err, IsNil)

	c.Assert(strings.Contains(data, "<ReplicationLocation>"), Equals, true)

	stringData, err := client1.GetBucketReplicationLocation(sourceBucketNameTest)
	var repResult GetBucketReplicationLocationResult
	err = xml.Unmarshal([]byte(stringData), &repResult)
	c.Assert(err, IsNil)
	c.Assert(repResult.Location[0], Equals, "oss-ap-northeast-1")
	c.Assert(repResult.Location[9], Equals, "oss-cn-beijing")
	c.Assert(repResult.LocationTransferType[1].Location, Equals, "oss-eu-central-1")
	c.Assert(repResult.LocationTransferType[1].TransferTypes, Equals, "oss_acc")
	c.Assert(repResult.RTCLocation[2], Equals, "oss-cn-shanghai")

	err = client1.DeleteBucket(sourceBucketNameTest)
	c.Assert(err, IsNil)
}

func (s *OssClientSuite) TestBucketReplicationGetProgressWithRuleID(c *C) {
	var sourceBucketNameTest = bucketNamePrefix + "-replication-" + RandLowStr(6)

	sourceRegion := "hangzhou"
	sourceEndpoint := "oss-cn-" + sourceRegion + ".aliyuncs.com"

	client1, err := New(sourceEndpoint, accessID, accessKey)
	c.Assert(err, IsNil)

	err = client1.CreateBucket(sourceBucketNameTest)
	c.Assert(err, IsNil)

	var firstDestinationBucketNameTest = bucketNamePrefix + "-replication-" + RandLowStr(6)

	firstDestinationRegion := "beijing"
	firstDestinationEndpoint := "oss-cn-" + firstDestinationRegion + ".aliyuncs.com"

	client2, err := New(firstDestinationEndpoint, accessID, accessKey)
	c.Assert(err, IsNil)

	err = client2.CreateBucket(firstDestinationBucketNameTest)
	c.Assert(err, IsNil)

	var secondDestinationBucketNameTest = bucketNamePrefix + "-replication-" + RandLowStr(6)

	secondDestinationRegion := "shenzhen"
	secondDestinationEndpoint := "oss-cn-" + secondDestinationRegion + ".aliyuncs.com"

	client3, err := New(secondDestinationEndpoint, accessID, accessKey)
	c.Assert(err, IsNil)

	err = client3.CreateBucket(secondDestinationBucketNameTest)
	c.Assert(err, IsNil)

	firstPutXml := `<?xml version="1.0" encoding="UTF-8"?>
	<ReplicationConfiguration>
	  <Rule>
		<Action>PUT</Action>
		<Destination>
		  <Bucket>` + firstDestinationBucketNameTest + `</Bucket>
		  <Location>oss-cn-` + firstDestinationRegion + `</Location>
		</Destination>
		<HistoricalObjectReplication>enabled</HistoricalObjectReplication>
	  </Rule>
	</ReplicationConfiguration>`

	// replication command and put method test
	err = client1.PutBucketReplication(sourceBucketNameTest, firstPutXml)
	c.Assert(err, IsNil)

	secondPutXml := `<?xml version="1.0" encoding="UTF-8"?>
	<ReplicationConfiguration>
	  <Rule>
		<Action>PUT</Action>
		<Destination>
		  <Bucket>` + secondDestinationBucketNameTest + `</Bucket>
		  <Location>oss-cn-` + secondDestinationRegion + `</Location>
		</Destination>
		<HistoricalObjectReplication>enabled</HistoricalObjectReplication>
	  </Rule>
	</ReplicationConfiguration>`

	// replication command and put method test
	err = client1.PutBucketReplication(sourceBucketNameTest, secondPutXml)
	c.Assert(err, IsNil)

	time.Sleep(5 * time.Second)

	data, err := client1.GetBucketReplication(sourceBucketNameTest)
	c.Assert(err, IsNil)

	var result GetResult
	err = xml.Unmarshal([]byte(data), &result)
	c.Assert(err, IsNil)

	var index int
	for i := 0; i <= 1; i++ {
		if result.Rules[i].Destination.Location == "oss-cn-"+secondDestinationRegion {
			index = i
			break
		}
	}

	ruleID := result.Rules[index].ID

	progressData, err := client1.GetBucketReplicationProgress(sourceBucketNameTest, ruleID)
	c.Assert(err, IsNil)

	var progressResult GetResult
	err = xml.Unmarshal([]byte(progressData), &progressResult)
	c.Assert(err, IsNil)

	c.Assert(progressResult.Rules[0].ID, Equals, ruleID)
	c.Assert(progressResult.Rules[0].Status, Equals, "starting")
	c.Assert(progressResult.Rules[0].Destination.Location, Equals, "oss-cn-"+secondDestinationRegion)
	c.Assert(progressResult.Rules[0].Destination.Bucket, Equals, secondDestinationBucketNameTest)
	c.Assert(progressResult.Rules[0].HistoricalObjectReplication, Equals, "enabled")

	stringData, err := client1.GetBucketReplicationProgress(sourceBucketNameTest, ruleID)

	var reqProgress GetBucketReplicationProgressResult
	err = xml.Unmarshal([]byte(stringData), &reqProgress)
	c.Assert(err, IsNil)
	c.Assert(reqProgress.Rule[0].ID, Equals, ruleID)
	c.Assert(reqProgress.Rule[0].Status, Equals, "starting")
	c.Assert(reqProgress.Rule[0].Destination.Location, Equals, "oss-cn-"+secondDestinationRegion)
	c.Assert(reqProgress.Rule[0].Destination.Bucket, Equals, secondDestinationBucketNameTest)
	c.Assert(reqProgress.Rule[0].HistoricalObjectReplication, Equals, "enabled")

	err = client1.DeleteBucket(sourceBucketNameTest)
	c.Assert(err, IsNil)

	err = client2.DeleteBucket(firstDestinationBucketNameTest)
	c.Assert(err, IsNil)

	err = client3.DeleteBucket(secondDestinationBucketNameTest)
	c.Assert(err, IsNil)
}

func (s *OssClientSuite) TestBucketReplicationGetProgressWithEmptyRuleID(c *C) {
	var sourceBucketNameTest = bucketNamePrefix + "-replication-" + RandLowStr(6)

	sourceRegion := "hangzhou"
	sourceEndpoint := "oss-cn-" + sourceRegion + ".aliyuncs.com"

	client1, err := New(sourceEndpoint, accessID, accessKey)
	c.Assert(err, IsNil)

	err = client1.CreateBucket(sourceBucketNameTest)
	c.Assert(err, IsNil)

	var firstDestinationBucketNameTest = bucketNamePrefix + "-replication-" + RandLowStr(6)

	firstDestinationRegion := "beijing"
	firstDestinationEndpoint := "oss-cn-" + firstDestinationRegion + ".aliyuncs.com"

	client2, err := New(firstDestinationEndpoint, accessID, accessKey)
	c.Assert(err, IsNil)

	err = client2.CreateBucket(firstDestinationBucketNameTest)
	c.Assert(err, IsNil)

	var secondDestinationBucketNameTest = bucketNamePrefix + "-replication-" + RandLowStr(6)

	secondDestinationRegion := "shenzhen"
	secondDestinationEndpoint := "oss-cn-" + secondDestinationRegion + ".aliyuncs.com"

	client3, err := New(secondDestinationEndpoint, accessID, accessKey)
	c.Assert(err, IsNil)

	err = client3.CreateBucket(secondDestinationBucketNameTest)
	c.Assert(err, IsNil)

	firstPutXml := `<?xml version="1.0" encoding="UTF-8"?>
	<ReplicationConfiguration>
	  <Rule>
		<Action>PUT</Action>
		<Destination>
		  <Bucket>` + firstDestinationBucketNameTest + `</Bucket>
		  <Location>oss-cn-` + firstDestinationRegion + `</Location>
		</Destination>
		<HistoricalObjectReplication>enabled</HistoricalObjectReplication>
	  </Rule>
	</ReplicationConfiguration>`

	// replication command and put method test
	err = client1.PutBucketReplication(sourceBucketNameTest, firstPutXml)
	c.Assert(err, IsNil)

	secondPutXml := `<?xml version="1.0" encoding="UTF-8"?>
	<ReplicationConfiguration>
	  <Rule>
		<Action>PUT</Action>
		<Destination>
		  <Bucket>` + secondDestinationBucketNameTest + `</Bucket>
		  <Location>oss-cn-` + secondDestinationRegion + `</Location>
		</Destination>
		<HistoricalObjectReplication>enabled</HistoricalObjectReplication>
	  </Rule>
	</ReplicationConfiguration>`

	// replication command and put method test
	err = client1.PutBucketReplication(sourceBucketNameTest, secondPutXml)
	c.Assert(err, IsNil)

	time.Sleep(5 * time.Second)

	data, err := client1.GetBucketReplicationProgress(sourceBucketNameTest, "")
	c.Assert(err, IsNil)

	var result GetResult
	err = xml.Unmarshal([]byte(data), &result)
	c.Assert(err, IsNil)

	var firstIndex int
	for i := 0; i <= 1; i++ {
		if result.Rules[i].Destination.Location == ("oss-cn-" + firstDestinationRegion) {
			firstIndex = i
			break
		}
	}
	secondIndex := 1 - firstIndex

	c.Assert(result.Rules[firstIndex].Status, Equals, "starting")
	c.Assert(result.Rules[firstIndex].Destination.Location, Equals, "oss-cn-"+firstDestinationRegion)
	c.Assert(result.Rules[firstIndex].Destination.Bucket, Equals, firstDestinationBucketNameTest)
	c.Assert(result.Rules[firstIndex].HistoricalObjectReplication, Equals, "enabled")

	c.Assert(result.Rules[secondIndex].Status, Equals, "starting")
	c.Assert(result.Rules[secondIndex].Destination.Location, Equals, "oss-cn-"+secondDestinationRegion)
	c.Assert(result.Rules[secondIndex].Destination.Bucket, Equals, secondDestinationBucketNameTest)
	c.Assert(result.Rules[secondIndex].HistoricalObjectReplication, Equals, "enabled")

	err = client1.DeleteBucket(sourceBucketNameTest)
	c.Assert(err, IsNil)

	err = client2.DeleteBucket(firstDestinationBucketNameTest)
	c.Assert(err, IsNil)

	err = client3.DeleteBucket(secondDestinationBucketNameTest)
	c.Assert(err, IsNil)
}

func (s *OssClientSuite) TestBucketCName(c *C) {
	client, err := New(endpoint, accessID, accessKey)
	c.Assert(err, IsNil)

	bucketName := bucketNamePrefix + RandLowStr(5)
	err = client.CreateBucket(bucketName)
	c.Assert(err, IsNil)

	cbResult, err := client.CreateBucketCnameToken(bucketName, "www.example.com")
	c.Assert(err, IsNil)
	c.Assert(cbResult.Bucket, Equals, bucketName)
	c.Assert(cbResult.Cname, Equals, "www.example.com")

	gbResult, err := client.GetBucketCnameToken(bucketName, "www.example.com")
	c.Assert(err, IsNil)
	c.Assert(gbResult.Bucket, Equals, bucketName)
	c.Assert(gbResult.Cname, Equals, "www.example.com")

	err = client.PutBucketCname(bucketName, "www.example.com")
	serviceErr, isSuc := err.(ServiceError)
	c.Assert(isSuc, Equals, true)
	c.Assert(serviceErr.Code, Equals, "NeedVerifyDomainOwnership")

	var bindCnameConfig PutBucketCname
	var bindCertificateConfig CertificateConfiguration
	bindCnameConfig.Cname = "www.example.com"
	bindCertificate := "-----BEGIN CERTIFICATE-----MIIGeDCCBOCgAwIBAgIRAPj4FWpW5XN6kwgU7*******-----END CERTIFICATE-----"
	privateKey := "-----BEGIN CERTIFICATE-----MIIFBzCCA++gT2H2hT6Wb3nwxjpLIfXmSVcV*****-----END CERTIFICATE-----"
	bindCertificateConfig.CertId = "92******-cn-hangzhou"
	bindCertificateConfig.Certificate = bindCertificate
	bindCertificateConfig.PrivateKey = privateKey
	bindCertificateConfig.Force = true
	bindCnameConfig.CertificateConfiguration = &bindCertificateConfig
	err = client.PutBucketCnameWithCertificate(bucketName, bindCnameConfig)
	serviceErr, isSuc = err.(ServiceError)
	c.Assert(isSuc, Equals, true)
	c.Assert(serviceErr.Code, Equals, "NeedVerifyDomainOwnership")

	xmlBody, err := client.GetBucketCname(bucketName)
	c.Assert(err, IsNil)
	c.Assert(strings.Contains(xmlBody, bucketName), Equals, true)

	cnResult, err := client.ListBucketCname(bucketName)
	c.Assert(err, IsNil)
	c.Assert(cnResult.Bucket, Equals, bucketName)
	c.Assert(cnResult.Owner != "", Equals, true)
	c.Assert(len(cnResult.Cname) == 0, Equals, true)

	err = client.DeleteBucketCname(bucketName, "www.example.com")
	c.Assert(err, IsNil)

	client.DeleteBucket(bucketName)
	c.Assert(err, IsNil)
}

func (s *OssClientSuite) TestCreateBucketXml(c *C) {
	client, err := New(endpoint, accessID, accessKey)
	c.Assert(err, IsNil)

	bucketName := bucketNamePrefix + RandLowStr(5)
	xmlBody := `
        <?xml version="1.0" encoding="UTF-8"?>
        <CreateBucketConfiguration>
            <StorageClass>IA</StorageClass>
        </CreateBucketConfiguration>
        `
	err = client.CreateBucketXml(bucketName, xmlBody)
	c.Assert(err, IsNil)

	//check
	bucketInfo, _ := client.GetBucketInfo(bucketName)
	c.Assert(bucketInfo.BucketInfo.StorageClass, Equals, "IA")
	err = client.DeleteBucket(bucketName)
	c.Assert(err, IsNil)
}

func emptyBodytargetHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set(HTTPHeaderOssRequestID, "123456")
	w.WriteHeader(309)
	fmt.Fprintf(w, "")
}

func serviceErrorBodytargetHandler(w http.ResponseWriter, r *http.Request) {
	err := ServiceError{
		Code:       "510",
		Message:    "service response error",
		RequestID:  "ABCDEF",
		HostID:     "127.0.0.1",
		Endpoint:   "127.0.0.1",
		StatusCode: 510,
	}
	data, _ := xml.MarshalIndent(&err, "", " ")
	w.Header().Set(HTTPHeaderOssRequestID, "ABCDEF")
	w.WriteHeader(510)
	fmt.Fprintf(w, string(data))
}
func unkownErrorBodytargetHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(509)
	fmt.Fprintf(w, "unkown response error")
}

func (s *OssClientSuite) TestExtendHttpResponseStatusCode(c *C) {
	// get port
	rand.Seed(time.Now().Unix())
	port := 10000 + rand.Intn(10000)

	// start http server
	httpAddr := fmt.Sprintf("127.0.0.1:%d", port)
	mux := http.NewServeMux()
	mux.HandleFunc("/empty-body/empty-body", emptyBodytargetHandler)
	mux.HandleFunc("/service-error/service-error", serviceErrorBodytargetHandler)
	mux.HandleFunc("/unkown-error/unkown-error", unkownErrorBodytargetHandler)
	svr := &http.Server{
		Addr:           httpAddr,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
		Handler:        mux,
	}

	go func() {
		svr.ListenAndServe()
	}()

	time.Sleep(5 * time.Second)

	testEndpoint := httpAddr
	client, err := New(testEndpoint, accessID, accessKey)

	//emptyBodytargetHandler
	bucket, err := client.Bucket("empty-body")
	_, err = bucket.GetObject("empty-body")
	fmt.Println(err)
	serviceErr, isSuc := err.(ServiceError)
	c.Assert(isSuc, Equals, true)
	c.Assert(serviceErr.StatusCode, Equals, 309)
	c.Assert(serviceErr.RequestID, Equals, "123456")

	//serviceErrorBodytargetHandler
	bucket, err = client.Bucket("service-error")
	_, err = bucket.GetObject("service-error")
	serviceErr, isSuc = (err).(ServiceError)
	c.Assert(isSuc, Equals, true)
	c.Assert(serviceErr.StatusCode, Equals, 510)
	c.Assert(serviceErr.RequestID, Equals, "ABCDEF")

	//unkownErrorBodytargetHandler
	bucket, err = client.Bucket("unkown-error")
	_, err = bucket.GetObject("unkown-error")
	serviceErr, isSuc = err.(ServiceError)
	c.Assert(isSuc, Equals, false)
	prefix := "unknown response body, status = 509"
	ok := strings.Contains(err.Error(), prefix)
	c.Assert(ok, Equals, true)
	svr.Close()
}

func emptyBodyEcHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set(HTTPHeaderOssRequestID, "123456")
	w.Header().Set(HTTPHeaderOssEc, "0001-00000309")
	w.WriteHeader(309)
	fmt.Fprintf(w, "")
}

func emptyBodyHeaderHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set(HTTPHeaderOssRequestID, "64195A905C006935335FE181")
	w.Header().Set(HTTPHeaderOssEc, "0026-00000001")
	w.Header().Set(HTTPHeaderOssErr, "PD94bWwgdmVyc2lvbj0iMS4wIiBlbmNvZGluZz0iVVRGLTgiPz4KPEVycm9yPgogIDxDb2RlPk5vU3VjaEtleTwvQ29kZT4KICA8TWVzc2FnZT5UaGUgc3BlY2lmaWVkIGtleSBkb2VzIG5vdCBleGlzdC48L01lc3NhZ2U+CiAgPFJlcXVlc3RJZD42NDE5NUE5MDVDMDA2OTM1MzM1RkUxODE8L1JlcXVlc3RJZD4KICA8SG9zdElkPmRlbW8td2Fsa2VyLTY5NjEub3NzLWNuLWhhbmd6aG91LmFsaXl1bmNzLmNvbTwvSG9zdElkPgogIDxLZXk+ZGVtby0xMTEudXh1PC9LZXk+CiAgPEVDPjAwMjYtMDAwMDAwMDE8L0VDPgo8L0Vycm9yPgo=")
	w.WriteHeader(404)
	fmt.Fprintf(w, "")
}

func serviceErrorBodyEcHandler(w http.ResponseWriter, r *http.Request) {
	err := ServiceError{
		Code:       "510",
		Message:    "service response error",
		RequestID:  "ABCDEF",
		HostID:     "127.0.0.1",
		Endpoint:   "127.0.0.1",
		StatusCode: 510,
		Ec:         "0001-00000510",
	}
	data, _ := xml.MarshalIndent(&err, "", " ")
	w.Header().Set(HTTPHeaderOssRequestID, "ABCDEF")
	w.WriteHeader(510)
	fmt.Fprintf(w, string(data))
}
func serviceErrorBodyEmptyEndpiointEcHandler(w http.ResponseWriter, r *http.Request) {
	err := ServiceError{
		Code:       "510",
		Message:    "service response error",
		RequestID:  "ABCDEF",
		HostID:     "127.0.0.1",
		Endpoint:   "",
		StatusCode: 510,
		Ec:         "0001-00000510",
	}
	data, _ := xml.MarshalIndent(&err, "", " ")
	w.Header().Set(HTTPHeaderOssRequestID, "ABCDEF")
	w.WriteHeader(510)
	fmt.Fprintf(w, string(data))
}

func unknownErrorBodyEcHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(509)
	w.Header().Set(HTTPHeaderOssEc, "0001-00000509")
	fmt.Fprintf(w, "unkown response error")
}

func (s *OssClientSuite) TestExtendHttpResponseStatusCodeWithEc(c *C) {
	// get port
	rand.Seed(time.Now().Unix())
	port := 10000 + rand.Intn(10000)

	// start http server
	httpAddr := fmt.Sprintf("127.0.0.1:%d", port)
	mux := http.NewServeMux()
	mux.HandleFunc("/empty-body/empty-body", emptyBodyEcHandler)
	mux.HandleFunc("/service-error/service-error", serviceErrorBodyEcHandler)
	mux.HandleFunc("/service-error/service-error-point", serviceErrorBodyEmptyEndpiointEcHandler)
	mux.HandleFunc("/unknown-error/unknown-error", unknownErrorBodyEcHandler)

	mux.HandleFunc("/empty-body/read-err-from-header", emptyBodyHeaderHandler)
	svr := &http.Server{
		Addr:           httpAddr,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
		Handler:        mux,
	}

	go func() {
		svr.ListenAndServe()
	}()

	time.Sleep(5 * time.Second)

	testEndpoint := httpAddr
	client, err := New(testEndpoint, accessID, accessKey)

	bucket, err := client.Bucket("empty-body")
	_, err = bucket.GetObject("empty-body")
	serviceErr, isSuc := err.(ServiceError)
	testLogger.Println(serviceErr)
	c.Assert(isSuc, Equals, true)
	c.Assert(serviceErr.StatusCode, Equals, 309)
	c.Assert(serviceErr.RequestID, Equals, "123456")
	c.Assert(serviceErr.Ec, Equals, "0001-00000309")
	prefix := "oss: service returned error: StatusCode=309, ErrorCode=, ErrorMessage=\"\", RequestId=123456, Ec=0001-00000309"
	ok := strings.Contains(err.Error(), prefix)
	c.Assert(ok, Equals, true)

	bucket, err = client.Bucket("empty-body")
	_, err = bucket.GetObject("read-err-from-header")
	serviceErr, isSuc = err.(ServiceError)
	c.Assert(isSuc, Equals, true)
	c.Assert(serviceErr.StatusCode, Equals, 404)
	c.Assert(serviceErr.RequestID, Equals, "64195A905C006935335FE181")
	c.Assert(serviceErr.Ec, Equals, "0026-00000001")

	prefix = "oss: service returned error: StatusCode=404, ErrorCode=NoSuchKey, ErrorMessage=\"The specified key does not exist.\", RequestId=64195A905C006935335FE181, Ec=0026-00000001"
	c.Assert(err.Error(), Equals, prefix)

	bucket, err = client.Bucket("service-error")
	_, err = bucket.GetObject("service-error")
	serviceErr, isSuc = (err).(ServiceError)
	c.Assert(isSuc, Equals, true)
	c.Assert(serviceErr.StatusCode, Equals, 510)
	c.Assert(serviceErr.RequestID, Equals, "ABCDEF")
	c.Assert(serviceErr.Ec, Equals, "0001-00000510")
	prefix = "oss: service returned error: StatusCode=510, ErrorCode=510, ErrorMessage=\"service response error\", RequestId=ABCDEF, Endpoint=127.0.0.1, Ec=0001-00000510"
	ok = strings.Contains(err.Error(), prefix)
	c.Assert(ok, Equals, true)

	bucket, err = client.Bucket("service-error")
	_, err = bucket.GetObject("service-error-point")
	serviceErr, isSuc = (err).(ServiceError)
	c.Assert(isSuc, Equals, true)
	c.Assert(serviceErr.StatusCode, Equals, 510)
	c.Assert(serviceErr.RequestID, Equals, "ABCDEF")
	c.Assert(serviceErr.Ec, Equals, "0001-00000510")
	prefix = "oss: service returned error: StatusCode=510, ErrorCode=510, ErrorMessage=\"service response error\", RequestId=ABCDEF, Ec=0001-00000510"
	ok = strings.Contains(err.Error(), prefix)
	c.Assert(ok, Equals, true)

	bucket, err = client.Bucket("unknown-error")
	_, err = bucket.GetObject("unknown-error")
	serviceErr, isSuc = err.(ServiceError)
	c.Assert(serviceErr.Ec, Equals, "")
	testLogger.Println(err.Error())
	c.Assert(isSuc, Equals, false)
	prefix = "unknown response body, status = 509 status code 509, RequestId = "
	c.Assert(err.Error(), Equals, prefix)
	svr.Close()
}

func (s *OssClientSuite) TestCloudBoxCreateAndDeleteBucketV1(c *C) {

	c.Assert(len(cloudboxControlEndpoint) > 0, Equals, true)

	var bucketNameTest = bucketNamePrefix + "cloudbox-" + RandLowStr(6)

	client, err := New(cloudboxControlEndpoint, accessID, accessKey)
	c.Assert(err, IsNil)

	// Create
	client.DeleteBucket(bucketNameTest)
	err = client.CreateBucket(bucketNameTest)
	c.Assert(err, IsNil)

	//sleep 3 seconds after create bucket
	time.Sleep(timeoutInOperation)

	// verify bucket is exist
	found, err := client.IsBucketExist(bucketNameTest)
	c.Assert(err, IsNil)
	c.Assert(found, Equals, true)

	err = client.DeleteBucket(bucketNameTest)
	c.Assert(err, IsNil)
	time.Sleep(timeoutInOperation)

	_, err = client.GetBucketACL(bucketNameTest)
	c.Assert(err, NotNil)
}

func (s *OssClientSuite) TestCloudBoxCreateAndDeleteBucketV4(c *C) {

	c.Assert(len(cloudboxControlEndpoint) > 0, Equals, true)

	var bucketNameTest = bucketNamePrefix + "cloudbox-" + RandLowStr(6)

	// set oss v4 signatrue
	options := []ClientOption{
		Region(envRegion),
		AuthVersion(AuthV4),
		CloudBoxId(cloudBoxID),
	}

	client, err := New(cloudboxControlEndpoint, accessID, accessKey, options...)

	c.Assert(err, IsNil)

	// Create
	client.DeleteBucket(bucketNameTest)
	err = client.CreateBucket(bucketNameTest)
	c.Assert(err, IsNil)

	//sleep 3 seconds after create bucket
	time.Sleep(timeoutInOperation)

	// verify bucket is exist
	found, err := client.IsBucketExist(bucketNameTest)
	c.Assert(err, IsNil)
	c.Assert(found, Equals, true)

	err = client.DeleteBucket(bucketNameTest)
	c.Assert(err, IsNil)
	time.Sleep(timeoutInOperation)

	_, err = client.GetBucketACL(bucketNameTest)
	c.Assert(err, NotNil)
}

func (s *OssClientSuite) TestCloudBoxListCloudBoxV1(c *C) {
	client, err := New(cloudboxControlEndpoint, accessID, accessKey)
	c.Assert(err, IsNil)

	boxRes, err := client.ListCloudBoxes()
	c.Assert(err, IsNil)
	c.Assert(len(boxRes.CloudBoxes) > 0, Equals, true)
}

func (s *OssClientSuite) TestCloudBoxListCloudBoxV4(c *C) {
	// set oss v4 signatrue
	options := []ClientOption{
		Region(envRegion),
		AuthVersion(AuthV4),
		CloudBoxId(cloudBoxID),
	}

	client, err := New(cloudboxControlEndpoint, accessID, accessKey, options...)
	c.Assert(err, IsNil)

	boxRes, err := client.ListCloudBoxes()
	c.Assert(err, IsNil)
	c.Assert(len(boxRes.CloudBoxes) > 0, Equals, true)
}

// TestBucketMetaQuery
func (s *OssClientSuite) TestBucketMetaQuery(c *C) {
	var bucketNameTest = bucketNamePrefix + "-acc-" + RandLowStr(6)
	client, err := New(endpoint, accessID, accessKey)
	c.Assert(err, IsNil)

	err = client.CreateBucket(bucketNameTest)
	c.Assert(err, IsNil)

	// Open meta query
	err = client.OpenMetaQuery(bucketNameTest)
	c.Assert(err, IsNil)

	// get meta query status
	result, err := client.GetMetaQueryStatus(bucketNameTest)
	c.Assert(err, IsNil)
	c.Assert(result.State != "", Equals, true)
	c.Assert(result.CreateTime != "", Equals, true)
	c.Assert(result.UpdateTime != "", Equals, true)

	// do meta query
	bucket, err := client.Bucket(bucketNameTest)
	c.Assert(err, IsNil)
	objectName := ""
	objectValue := ""
	for i := 0; i < 15; i++ {
		objectName = objectNamePrefix + RandLowStr(6) + ".txt"
		objectValue = RandLowStr(1000)
		err = bucket.PutObject(objectName, strings.NewReader(objectValue))
		c.Assert(err, IsNil)
	}
	time.Sleep(240 * time.Second)
	query := MetaQuery{
		NextToken:  "",
		MaxResults: 20,
		Query:      `{"Field": "Size","Value": "888","Operation": "gt"}`,
		Sort:       "Size",
		Order:      "asc",
	}
	queryResult, err := client.DoMetaQuery(bucketNameTest, query)
	c.Assert(err, IsNil)
	c.Assert(queryResult.NextToken, Equals, "")
	c.Assert(len(queryResult.Files), Equals, 15)
	// do meta query use aggregate
	queryTwo := MetaQuery{
		NextToken:  "",
		MaxResults: 100,
		Query:      `{"Field": "Size","Value": "888","Operation": "gt"}`,
		Sort:       "Size",
		Order:      "asc",
		Aggregations: []MetaQueryAggregationRequest{
			{
				Field:     "Size",
				Operation: "sum",
			},
			{
				Field:     "Size",
				Operation: "max",
			},
			{
				Field:     "Size",
				Operation: "group",
			},
		},
	}
	queryTwoResult, err := client.DoMetaQuery(bucketNameTest, queryTwo)
	c.Assert(err, IsNil)
	c.Assert(queryResult.NextToken, Equals, "")
	c.Assert(len(queryTwoResult.Files), Equals, 0)
	c.Assert(len(queryTwoResult.Aggregations) > 0, Equals, true)
	// Close meta query
	err = client.CloseMetaQuery(bucketNameTest)
	c.Assert(err, IsNil)
}

// TestBucketAccessMonitor
func (s *OssClientSuite) TestBucketAccessMonitor(c *C) {
	var bucketNameTest = bucketNamePrefix + "-acc-" + RandLowStr(6)
	client, err := New(endpoint, accessID, accessKey)
	c.Assert(err, IsNil)

	err = client.CreateBucket(bucketNameTest)
	c.Assert(err, IsNil)
	time.Sleep(3 * time.Second)

	res, err := client.GetBucketInfo(bucketNameTest)
	c.Assert(err, IsNil)
	c.Assert(res.BucketInfo.AccessMonitor, Equals, "Disabled")

	result, err := client.GetBucketAccessMonitor(bucketNameTest)
	c.Assert(err, IsNil)
	c.Assert(result.Status, Equals, "Disabled")

	// Put Bucket Access Monitor
	access := PutBucketAccessMonitor{
		Status: "Enabled",
	}
	err = client.PutBucketAccessMonitor(bucketNameTest, access)
	c.Assert(err, IsNil)
	time.Sleep(3 * time.Second)

	// Put Bucket Access Monitor twice
	access = PutBucketAccessMonitor{
		Status: "Enabled",
	}
	err = client.PutBucketAccessMonitor(bucketNameTest, access)
	c.Assert(err, IsNil)
	time.Sleep(3 * time.Second)

	// get bucket info
	res, err = client.GetBucketInfo(bucketNameTest)
	c.Assert(err, IsNil)
	c.Assert(res.BucketInfo.AccessMonitor, Equals, "Enabled")

	// get bucket access monitor
	result, err = client.GetBucketAccessMonitor(bucketNameTest)
	c.Assert(err, IsNil)
	c.Assert(result.Status, Equals, "Enabled")

	// set bucket life cycle with access time
	isTrue := true
	isFalse := false
	rule1 := LifecycleRule{
		ID:     "mtime transition1",
		Prefix: "logs1",
		Status: "Enabled",
		Transitions: []LifecycleTransition{
			{
				Days:         30,
				StorageClass: StorageIA,
			},
		},
	}
	rule2 := LifecycleRule{
		ID:     "mtime transition2",
		Prefix: "logs2",
		Status: "Enabled",
		Transitions: []LifecycleTransition{
			{
				Days:         30,
				StorageClass: StorageIA,
				IsAccessTime: &isFalse,
			},
		},
	}
	rule3 := LifecycleRule{
		ID:     "amtime transition1",
		Prefix: "logs3",
		Status: "Enabled",
		Transitions: []LifecycleTransition{
			{
				Days:                 30,
				StorageClass:         StorageIA,
				IsAccessTime:         &isTrue,
				ReturnToStdWhenVisit: &isFalse,
			},
		},
	}
	rule4 := LifecycleRule{
		ID:     "amtime transition2",
		Prefix: "logs4",
		Status: "Enabled",
		Transitions: []LifecycleTransition{
			{
				Days:                 30,
				StorageClass:         StorageIA,
				IsAccessTime:         &isTrue,
				ReturnToStdWhenVisit: &isTrue,
				AllowSmallFile:       &isTrue,
			},
		},
	}
	rule5 := LifecycleRule{
		ID:     "amtime transition3",
		Prefix: "logs5",
		Status: "Enabled",
		NonVersionTransitions: []LifecycleVersionTransition{
			{
				NoncurrentDays:       10,
				StorageClass:         StorageIA,
				IsAccessTime:         &isTrue,
				ReturnToStdWhenVisit: &isFalse,
				AllowSmallFile:       &isFalse,
			},
		},
	}
	var rules = []LifecycleRule{rule1, rule2, rule3, rule4, rule5}
	err = client.SetBucketLifecycle(bucketNameTest, rules)
	c.Assert(err, IsNil)
	time.Sleep(3 * time.Second)

	// Get bucket's cycle
	lc, err := client.GetBucketLifecycle(bucketNameTest)
	c.Assert(err, IsNil)
	c.Assert(lc.Rules[0].Transitions[0].Days, Equals, 30)
	c.Assert(lc.Rules[0].Transitions[0].StorageClass, Equals, StorageIA)

	c.Assert(*lc.Rules[1].Transitions[0].IsAccessTime, Equals, false)

	c.Assert(*lc.Rules[2].Transitions[0].IsAccessTime, Equals, true)
	c.Assert(*lc.Rules[2].Transitions[0].ReturnToStdWhenVisit, Equals, false)

	c.Assert(*lc.Rules[3].Transitions[0].IsAccessTime, Equals, true)
	c.Assert(*lc.Rules[3].Transitions[0].ReturnToStdWhenVisit, Equals, true)
	c.Assert(*lc.Rules[3].Transitions[0].AllowSmallFile, Equals, true)

	c.Assert(lc.Rules[4].NonVersionTransitions[0].NoncurrentDays, Equals, 10)
	c.Assert(lc.Rules[4].NonVersionTransitions[0].StorageClass, Equals, StorageIA)
	c.Assert(*lc.Rules[4].NonVersionTransitions[0].IsAccessTime, Equals, true)
	c.Assert(*lc.Rules[4].NonVersionTransitions[0].ReturnToStdWhenVisit, Equals, false)
	c.Assert(*lc.Rules[4].NonVersionTransitions[0].AllowSmallFile, Equals, false)

	// can't disable Access Monitor
	access = PutBucketAccessMonitor{
		Status: "Disabled",
	}
	err = client.PutBucketAccessMonitor(bucketNameTest, access)
	c.Assert(err, NotNil)

	// delete bucket's cycle
	err = client.DeleteBucketLifecycle(bucketNameTest)
	c.Assert(err, IsNil)

	// Put Bucket Access Monitor
	access = PutBucketAccessMonitor{
		Status: "Disabled",
	}
	err = client.PutBucketAccessMonitor(bucketNameTest, access)
	c.Assert(err, IsNil)
	time.Sleep(3 * time.Second)

	// get bucket info
	res, err = client.GetBucketInfo(bucketNameTest)
	c.Assert(err, IsNil)
	c.Assert(res.BucketInfo.AccessMonitor, Equals, "Disabled")

	// get bucket access monitor
	result, err = client.GetBucketAccessMonitor(bucketNameTest)
	c.Assert(err, IsNil)
	c.Assert(result.Status, Equals, "Disabled")
}

// TestBucketResourceGroup
func (s *OssClientSuite) TestBucketResourceGroup(c *C) {
	var bucketNameTest = bucketNamePrefix + "-acc-" + RandLowStr(6)
	var bucketNameTest2 = bucketNamePrefix + "-acc-" + RandLowStr(8)
	client, err := New(endpoint, accessID, accessKey)
	c.Assert(err, IsNil)

	err = client.CreateBucket(bucketNameTest)
	c.Assert(err, IsNil)
	err = client.CreateBucket(bucketNameTest2)
	c.Assert(err, IsNil)

	time.Sleep(3 * time.Second)

	res, err := client.GetBucketResourceGroup(bucketNameTest)
	c.Assert(err, IsNil)

	id := res.ResourceGroupId

	// Put Bucket Resource Group
	resource := PutBucketResourceGroup{
		ResourceGroupId: id,
	}
	err = client.PutBucketResourceGroup(bucketNameTest2, resource)
	c.Assert(err, IsNil)
	time.Sleep(3 * time.Second)

	// Get Bucket Resource Group
	res, err = client.GetBucketResourceGroup(bucketNameTest2)
	c.Assert(err, IsNil)
	c.Assert(res.ResourceGroupId, Equals, id)

	// Put Bucket Resource Group With Empty Resource GroupId
	resource = PutBucketResourceGroup{
		ResourceGroupId: "",
	}
	err = client.PutBucketResourceGroup(bucketNameTest, resource)
	c.Assert(err, IsNil)
	time.Sleep(3 * time.Second)

	// Get Bucket Resource Group
	res, err = client.GetBucketResourceGroup(bucketNameTest)
	c.Assert(err, IsNil)
	c.Assert(res.ResourceGroupId, Equals, id)

	// Put Bucket Resource Group With Empty Resource GroupId
	resource = PutBucketResourceGroup{
		ResourceGroupId: "no-exist-id",
	}
	err = client.PutBucketResourceGroup(bucketNameTest, resource)
	c.Assert(err, NotNil)

	resource = PutBucketResourceGroup{
		ResourceGroupId: "no-exist-id",
	}
	err = client.PutBucketResourceGroup(bucketNameTest2, resource)
	c.Assert(err, NotNil)

	ForceDeleteBucket(client, bucketNameTest, c)
	ForceDeleteBucket(client, bucketNameTest2, c)
}

// TestBucketStyle
func (s *OssClientSuite) TestBucketStyle(c *C) {
	var bucketNameTest = bucketNamePrefix + "-acc-" + RandLowStr(6)
	client, err := New(endpoint, accessID, accessKey)
	c.Assert(err, IsNil)

	err = client.CreateBucket(bucketNameTest)
	c.Assert(err, IsNil)
	time.Sleep(3 * time.Second)

	// Put Bucket Style
	style := "image/resize,p_50"
	styleName := "image-" + RandLowStr(6)
	err = client.PutBucketStyle(bucketNameTest, styleName, style)
	c.Assert(err, IsNil)
	time.Sleep(1 * time.Second)

	// get bucket style
	res, err := client.GetBucketStyle(bucketNameTest, styleName)
	c.Assert(err, IsNil)
	c.Assert(res.Name, Equals, styleName)
	c.Assert(res.Content, Equals, "image/resize,p_50")
	c.Assert(res.CreateTime != "", Equals, true)
	c.Assert(res.LastModifyTime != "", Equals, true)

	_, err = client.GetBucketStyle(bucketNameTest, "no-exist-style")
	c.Assert(err, NotNil)
	c.Assert(err.(ServiceError).StatusCode, Equals, 404)
	c.Assert(err.(ServiceError).Code, Equals, "NoSuchStyle")

	style1 := "image/resize,w_200"
	styleName1 := "image-" + RandLowStr(6)
	err = client.PutBucketStyle(bucketNameTest, styleName1, style1)
	c.Assert(err, IsNil)
	time.Sleep(1 * time.Second)

	style2 := "image/resize,w_300"
	styleName2 := "image-" + RandLowStr(6)
	err = client.PutBucketStyle(bucketNameTest, styleName2, style2)
	c.Assert(err, IsNil)
	time.Sleep(1 * time.Second)

	// list bucket style
	list, err := client.ListBucketStyle(bucketNameTest)
	c.Assert(err, IsNil)
	c.Assert(len(list.Style), Equals, 3)

	c.Assert(list.Style[1].Name, Equals, styleName1)
	c.Assert(list.Style[1].Content, Equals, "image/resize,w_200")
	c.Assert(list.Style[1].CreateTime != "", Equals, true)
	c.Assert(list.Style[1].LastModifyTime != "", Equals, true)
	c.Assert(list.Style[2].Name, Equals, styleName2)
	c.Assert(list.Style[2].Content, Equals, "image/resize,w_300")
	c.Assert(list.Style[2].CreateTime != "", Equals, true)
	c.Assert(list.Style[2].LastModifyTime != "", Equals, true)

	// delete bucket style
	err = client.DeleteBucketStyle(bucketNameTest, styleName)
	c.Assert(err, IsNil)

	err = client.DeleteBucketStyle(bucketNameTest, styleName1)
	c.Assert(err, IsNil)

	err = client.DeleteBucketStyle(bucketNameTest, styleName2)
	c.Assert(err, IsNil)

	err = client.DeleteBucketStyle(bucketNameTest, "no-exist-style")
	c.Assert(err, IsNil)

	list, err = client.ListBucketStyle(bucketNameTest)
	c.Assert(err, IsNil)
	c.Assert(len(list.Style), Equals, 0)

}

// TestDescribeRegions
func (s *OssClientSuite) TestDescribeRegions(c *C) {
	client, err := New(endpoint, accessID, accessKey)
	c.Assert(err, IsNil)

	list, err := client.DescribeRegions()
	c.Assert(err, IsNil)
	c.Assert(len(list.Regions) > 0, Equals, true)

	info, err := client.DescribeRegions(AddParam("regions", "oss-cn-hangzhou"))
	c.Assert(err, IsNil)
	c.Assert(len(info.Regions), Equals, 1)
	c.Assert(info.Regions[0].Region, Equals, "oss-cn-hangzhou")
	c.Assert(info.Regions[0].InternetEndpoint, Equals, "oss-cn-hangzhou.aliyuncs.com")
	c.Assert(info.Regions[0].InternalEndpoint, Equals, "oss-cn-hangzhou-internal.aliyuncs.com")
	c.Assert(info.Regions[0].AccelerateEndpoint, Equals, "oss-accelerate.aliyuncs.com")
}

// TestDescribeRegions
func (s *OssClientSuite) TestBucketResponseHeader(c *C) {
	client, err := New("oss-ap-southeast-2.aliyuncs.com", accessID, accessKey)
	c.Assert(err, IsNil)

	bucketName := bucketNamePrefix + "-resp-" + RandLowStr(6)
	err = client.CreateBucket(bucketName)
	c.Assert(err, IsNil)
	time.Sleep(3 * time.Second)

	reqHeader := PutBucketResponseHeader{
		Rule: []ResponseHeaderRule{
			{
				Name: "name1",
				Filters: ResponseHeaderRuleFilters{
					[]string{
						"Put", "GetObject",
					},
				},
				HideHeaders: ResponseHeaderRuleHeaders{
					[]string{
						"Last-Modified",
					},
				},
			},
			{
				Name: "name2",
				Filters: ResponseHeaderRuleFilters{
					[]string{
						"*",
					},
				},
				HideHeaders: ResponseHeaderRuleHeaders{
					[]string{
						"Last-Modified",
					},
				},
			},
		},
	}
	err = client.PutBucketResponseHeader(bucketName, reqHeader)
	c.Assert(err, IsNil)

	rule, err := client.GetBucketResponseHeader(bucketName)
	c.Assert(err, IsNil)
	c.Assert(len(rule.Rule), Equals, 2)
	c.Assert(rule.Rule[0].Name, Equals, "name1")
	c.Assert(rule.Rule[1].Name, Equals, "name2")
	c.Assert(rule.Rule[0].Filters.Operation[0], Equals, "Put")
	c.Assert(rule.Rule[0].Filters.Operation[1], Equals, "GetObject")
	err = client.DeleteBucketResponseHeader(bucketName)
	c.Assert(err, IsNil)
	client.DeleteBucket(bucketName)
}
