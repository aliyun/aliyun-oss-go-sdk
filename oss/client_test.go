// client test
// use gocheck, install gocheck to execute "go get gopkg.in/check.v1",
// see https://labix.org/gocheck

package oss

import (
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"runtime"
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

	// Proxy
	proxyHost   = os.Getenv("OSS_TEST_PROXY_HOST")
	proxyUser   = os.Getenv("OSS_TEST_PROXY_USER")
	proxyPasswd = os.Getenv("OSS_TEST_PROXY_PASSWORD")

	// STS
	stsaccessID  = os.Getenv("OSS_TEST_STS_ID")
	stsaccessKey = os.Getenv("OSS_TEST_STS_KEY")
	stsARN       = os.Getenv("OSS_TEST_STS_ARN")
)

var (
	// prefix of bucket name for bucket ops test
	bucketNamePrefix = "go-sdk-test-bucket-"
	// bucket name for object ops test
	bucketName        = bucketNamePrefix + randLowStr(6)
	archiveBucketName = bucketNamePrefix + "arch-" + randLowStr(6)
	// object name for object ops test
	objectNamePrefix = "go-sdk-test-object-"
	// sts region is one and only hangzhou
	stsRegion = "cn-hangzhou"
)

var (
	logPath            = "go_sdk_test_" + time.Now().Format("20060102_150405") + ".log"
	testLogFile, _     = os.OpenFile(logPath, os.O_RDWR|os.O_CREATE, 0664)
	testLogger         = log.New(testLogFile, "", log.Ldate|log.Ltime|log.Lshortfile)
	letters            = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	timeoutInOperation = 3 * time.Second
)

func randStr(n int) string {
	b := make([]rune, n)
	randMarker := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := range b {
		b[i] = letters[randMarker.Intn(len(letters))]
	}
	return string(b)
}

func createFile(fileName, content string, c *C) {
	fout, err := os.Create(fileName)
	defer fout.Close()
	c.Assert(err, IsNil)
	_, err = fout.WriteString(content)
	c.Assert(err, IsNil)
}

func randLowStr(n int) string {
	return strings.ToLower(randStr(n))
}

func forceDeleteBucket(client *Client, bucketName string, c *C) {
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
		s.deleteBucket(client, bucket.Name, c)
	}

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

	testLogger.Println("test client completed")
}

func (s *OssClientSuite) deleteBucket(client *Client, bucketName string, c *C) {
    forceDeleteBucket(client,bucketName,c)
}

// SetUpTest runs after each test or benchmark runs
func (s *OssClientSuite) SetUpTest(c *C) {
}

// TearDownTest runs once after all tests or benchmarks have finished running
func (s *OssClientSuite) TearDownTest(c *C) {
}

// TestCreateBucket
func (s *OssClientSuite) TestCreateBucket(c *C) {
	var bucketNameTest = bucketNamePrefix + randLowStr(6)

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
	for _, storage := range []StorageClassType{StorageStandard, StorageIA, StorageArchive} {
		bucketNameTest := bucketNamePrefix + randLowStr(6)
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
	for _, storage := range []StorageClassType{StorageStandard, StorageIA, StorageArchive} {
		bucketNameTest := bucketNamePrefix + randLowStr(6)
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
	err = client.CreateBucket(bucketNamePrefix+randLowStr(6), ACL("InvaldAcl"))
	c.Assert(err, NotNil)
	testLogger.Println(err)
}

// TestDeleteBucket
func (s *OssClientSuite) TestDeleteBucket(c *C) {
	var bucketNameTest = bucketNamePrefix + randLowStr(6)

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
	var bucketNameTest = bucketNamePrefix + randLowStr(6)

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
	var prefix = bucketNamePrefix + randLowStr(6)
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
	lbr, err := client.ListBuckets(Prefix(prefix), MaxKeys(2))
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
	var prefix = bucketNamePrefix + randLowStr(6)
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
	var bucketNameTest = bucketNamePrefix + randLowStr(6)

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

	res, err = client.GetBucketACL(bucketNameTest)
	c.Assert(err, IsNil)
	c.Assert(res.ACL, Equals, string(ACLPrivate))

	err = client.DeleteBucket(bucketNameTest)
	c.Assert(err, IsNil)
}

// TestSetBucketAclNegative
func (s *OssClientSuite) TestBucketAclNegative(c *C) {
	var bucketNameTest = bucketNamePrefix + randLowStr(6)

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
	var bucketNameTest = bucketNamePrefix + randLowStr(6)

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
	var bucketNameTest = bucketNamePrefix + randLowStr(6)

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
	var bucketNameTest = bucketNamePrefix + randLowStr(6)

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
	var bucketNameTest = bucketNamePrefix + randLowStr(6)
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
	var bucketNameTest = bucketNamePrefix + randLowStr(6)

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
		CreatedBeforeDate: randStr(10),
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

// TestDeleteBucketLifecycle
func (s *OssClientSuite) TestDeleteBucketLifecycle(c *C) {
	var bucketNameTest = bucketNamePrefix + randLowStr(6)

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
	var bucketNameTest = bucketNamePrefix + randLowStr(6)
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

// TestSetBucketReferer
func (s *OssClientSuite) TestSetBucketReferer(c *C) {
	var bucketNameTest = bucketNamePrefix + randLowStr(6)
	var referers = []string{"http://www.aliyun.com", "https://www.aliyun.com"}

	client, err := New(endpoint, accessID, accessKey)
	c.Assert(err, IsNil)

	err = client.CreateBucket(bucketNameTest)
	c.Assert(err, IsNil)
	time.Sleep(timeoutInOperation)

	res, err := client.GetBucketReferer(bucketNameTest)
	c.Assert(res.AllowEmptyReferer, Equals, true)
	c.Assert(len(res.RefererList), Equals, 0)

	// Set referers
	err = client.SetBucketReferer(bucketNameTest, referers, false)
	c.Assert(err, IsNil)
	time.Sleep(timeoutInOperation)

	res, err = client.GetBucketReferer(bucketNameTest)
	c.Assert(res.AllowEmptyReferer, Equals, false)
	c.Assert(len(res.RefererList), Equals, 2)
	c.Assert(res.RefererList[0], Equals, "http://www.aliyun.com")
	c.Assert(res.RefererList[1], Equals, "https://www.aliyun.com")

	// Reset referer, referers empty
	referers = []string{""}
	err = client.SetBucketReferer(bucketNameTest, referers, true)
	c.Assert(err, IsNil)

	referers = []string{}
	err = client.SetBucketReferer(bucketNameTest, referers, true)
	c.Assert(err, IsNil)

	res, err = client.GetBucketReferer(bucketNameTest)
	c.Assert(res.AllowEmptyReferer, Equals, true)
	c.Assert(len(res.RefererList), Equals, 0)

	err = client.DeleteBucket(bucketNameTest)
	c.Assert(err, IsNil)
}

// TestSetBucketRefererNegative
func (s *OssClientSuite) TestBucketRefererNegative(c *C) {
	var bucketNameTest = bucketNamePrefix + randLowStr(6)
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

// TestSetBucketLogging
func (s *OssClientSuite) TestSetBucketLogging(c *C) {
	var bucketNameTest = bucketNamePrefix + randLowStr(6)
	var bucketNameTarget = bucketNamePrefix + randLowStr(6)

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
	var bucketNameTest = bucketNamePrefix + randLowStr(6)
	var bucketNameTarget = bucketNamePrefix + randLowStr(6)

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
	var bucketNameTest = bucketNamePrefix + randLowStr(6)
	var bucketNameTarget = bucketNamePrefix + randLowStr(6)

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
	var bucketNameTest = bucketNamePrefix + randLowStr(6)
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
	var bucketNameTest = bucketNamePrefix + randLowStr(6)
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
	var bucketNameTest = bucketNamePrefix + randLowStr(6)
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

// TestSetBucketWebsite
func (s *OssClientSuite) TestSetBucketCORS(c *C) {
	var bucketNameTest = bucketNamePrefix + randLowStr(6)
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

	err = client.DeleteBucketCORS(bucketNameTest)
	c.Assert(err, IsNil)

	err = client.DeleteBucket(bucketNameTest)
	c.Assert(err, IsNil)
}

// TestDeleteBucketCORS
func (s *OssClientSuite) TestDeleteBucketCORS(c *C) {
	var bucketNameTest = bucketNamePrefix + randLowStr(6)
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
	var bucketNameTest = bucketNamePrefix + randLowStr(6)
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
	var bucketNameTest = bucketNamePrefix + randLowStr(6)

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

	err = client.DeleteBucket(bucketNameTest)
	c.Assert(err, IsNil)
}

// TestGetBucketInfoNegative
func (s *OssClientSuite) TestGetBucketInfoNegative(c *C) {
	var bucketNameTest = bucketNamePrefix + randLowStr(6)

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
	var bucketNameTest = bucketNamePrefix + randLowStr(6)

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
	var bucketNameTest = bucketNamePrefix + randLowStr(6)

	client, err := New(endpoint, accessID, accessKey, UseCname(true),
		Timeout(11, 12), SecurityToken("token"), Proxy(proxyHost))
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
	c.Assert(client.Config.ProxyHost, Equals, proxyHost)

	client, err = New(endpoint, accessID, accessKey, AuthProxy(proxyHost, proxyUser, proxyPasswd))

	c.Assert(client.Conn.config.IsUseProxy, Equals, true)
	c.Assert(client.Config.ProxyHost, Equals, proxyHost)
	c.Assert(client.Conn.config.IsAuthProxy, Equals, true)
	c.Assert(client.Conn.config.ProxyUser, Equals, proxyUser)
	c.Assert(client.Conn.config.ProxyPassword, Equals, proxyPasswd)

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
func (s *OssClientSuite) TestProxy(c *C) {
	bucketNameTest := bucketNamePrefix + randLowStr(6)
	objectName := "体育/奥运/首金"
	objectValue := "大江东去，浪淘尽，千古风流人物。 故垒西边，人道是、三国周郎赤壁。"

	client, err := New(endpoint, accessID, accessKey, AuthProxy(proxyHost, proxyUser, proxyPasswd))

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
	c.Assert(strings.Contains(str, HTTPParamExpires+"="), Equals, true)
	c.Assert(strings.Contains(str, HTTPParamAccessKeyID+"="), Equals, true)
	c.Assert(strings.Contains(str, HTTPParamSignature+"="), Equals, true)

	// Put object with URL
	err = bucket.PutObjectWithURL(str, strings.NewReader(objectValue))
	c.Assert(err, IsNil)

	// Sign URL for get object
	str, err = bucket.SignURL(objectName, HTTPGet, 60)
	c.Assert(err, IsNil)
	c.Assert(strings.Contains(str, HTTPParamExpires+"="), Equals, true)
	c.Assert(strings.Contains(str, HTTPParamAccessKeyID+"="), Equals, true)
	c.Assert(strings.Contains(str, HTTPParamSignature+"="), Equals, true)

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
	logName := "." + string(os.PathSeparator) + "test-go-sdk-httpdebug.log" + randStr(5)
	f, err := os.OpenFile(logName, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0660)
	c.Assert(err, IsNil)

	client, err := New(endpoint, accessID, accessKey)
	client.Config.LogLevel = Debug

	client.Config.Logger = log.New(f, "", log.LstdFlags)

	var testBucketName = bucketNamePrefix + randLowStr(6)

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

func (s *OssClientSuite) TestHttpLogSignUrl(c *C) {
	logName := "." + string(os.PathSeparator) + "test-go-sdk-httpdebug-signurl.log" + randStr(5)
	f, err := os.OpenFile(logName, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0660)
	c.Assert(err, IsNil)

	client, err := New(endpoint, accessID, accessKey)
	client.Config.LogLevel = Debug
	client.Config.Logger = log.New(f, "", log.LstdFlags)

	var testBucketName = bucketNamePrefix + randLowStr(6)

	// CreateBucket
	err = client.CreateBucket(testBucketName)
	f.Close()

	// clear log
	f, err = os.OpenFile(logName, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0660)
	client.Config.Logger = log.New(f, "", log.LstdFlags)

	bucket, _ := client.Bucket(testBucketName)
	objectName := objectNamePrefix + randStr(8)
	objectValue := randStr(20)

	// Sign URL for put
	str, err := bucket.SignURL(objectName, HTTPPut, 60)
	c.Assert(err, IsNil)
	c.Assert(strings.Contains(str, HTTPParamExpires+"="), Equals, true)
	c.Assert(strings.Contains(str, HTTPParamAccessKeyID+"="), Equals, true)
	c.Assert(strings.Contains(str, HTTPParamSignature+"="), Equals, true)

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
		} else if pSlice[0] == "go1" && pSlice[1] >= "7" {
			c.Assert(err, IsNil)
		} else {
			c.Assert(err, NotNil)
		}
	} else {
		c.Assert(err, NotNil)
	}
}

func (s *OssClientSuite) TestBucketEncyptionError(c *C) {
	client, err := New(endpoint, accessID, accessKey)
	c.Assert(err, IsNil)

	bucketName := bucketNamePrefix + randLowStr(5)
	err = client.CreateBucket(bucketName)
	c.Assert(err, IsNil)

	// SetBucketEncryption:AES256 ,"123"
	encryptionRule := ServerEncryptionRule{}
	encryptionRule.SSEDefault.SSEAlgorithm = "AES256"
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

func (s *OssClientSuite) TestBucketEncyptionPutAndGetAndDelete(c *C) {
	client, err := New(endpoint, accessID, accessKey)
	c.Assert(err, IsNil)

	bucketName := bucketNamePrefix + randLowStr(5)
	err = client.CreateBucket(bucketName)
	c.Assert(err, IsNil)

	// SetBucketEncryption:KMS ,""
	encryptionRule := ServerEncryptionRule{}
	encryptionRule.SSEDefault.SSEAlgorithm = "KMS"

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
	err = client.DeleteBucketEncryption(bucketName)
	c.Assert(err, IsNil)

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

func (s *OssClientSuite) TestBucketEncyptionPutObjectSuccess(c *C) {
	client, err := New(endpoint, accessID, accessKey)
	c.Assert(err, IsNil)

	bucketName := bucketNamePrefix + randLowStr(5)
	err = client.CreateBucket(bucketName)
	c.Assert(err, IsNil)

	// SetBucketEncryption:KMS ,""
	encryptionRule := ServerEncryptionRule{}
	encryptionRule.SSEDefault.SSEAlgorithm = "KMS"

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

	// put and get object success
	//bucket, err := client.Bucket(bucketName)
	//c.Assert(err, IsNil)

	// put object success
	//objectName := objectNamePrefix + randStr(8)
	//context := randStr(100)
	//err = bucket.PutObject(objectName, strings.NewReader(context))
	//c.Assert(err, IsNil)

	// get object success
	//body, err := bucket.GetObject(objectName)
	//c.Assert(err, IsNil)
	//str, err := readBody(body)
	//c.Assert(err, IsNil)
	//body.Close()
	//c.Assert(str, Equals, context)

	//bucket.DeleteObject(objectName)
	err = client.DeleteBucket(bucketName)
	c.Assert(err, IsNil)
}

func (s *OssClientSuite) TestBucketEncyptionPutObjectError(c *C) {
	client, err := New(endpoint, accessID, accessKey)
	c.Assert(err, IsNil)

	bucketName := bucketNamePrefix + randLowStr(5)
	err = client.CreateBucket(bucketName)
	c.Assert(err, IsNil)

	// SetBucketEncryption:KMS ,""
	encryptionRule := ServerEncryptionRule{}
	encryptionRule.SSEDefault.SSEAlgorithm = "KMS"
	encryptionRule.SSEDefault.KMSMasterKeyID = "123"

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
	c.Assert(bucketResult.BucketInfo.SseRule.KMSMasterKeyID, Equals, "123")
	c.Assert(bucketResult.BucketInfo.Versioning, Equals, "")

	// put and get object failure
	bucket, err := client.Bucket(bucketName)
	c.Assert(err, IsNil)

	// put object failure
	objectName := objectNamePrefix + randStr(8)
	context := randStr(100)
	err = bucket.PutObject(objectName, strings.NewReader(context))
	c.Assert(err, NotNil)

	err = client.DeleteBucket(bucketName)
	c.Assert(err, IsNil)
}

func (s *OssClientSuite) TestBucketTaggingOperation(c *C) {
	client, err := New(endpoint, accessID, accessKey)
	c.Assert(err, IsNil)

	bucketName := bucketNamePrefix + randLowStr(5)
	err = client.CreateBucket(bucketName)
	c.Assert(err, IsNil)

	// Bucket Tagging
	var tagging Tagging
	tagging.Tags = []Tag{Tag{Key: "testkey2", Value: "testvalue2"}}
	err = client.SetBucketTagging(bucketName, tagging)
	c.Assert(err, IsNil)

	getResult, err := client.GetBucketTagging(bucketName)
	c.Assert(err, IsNil)
	c.Assert(getResult.Tags[0].Key, Equals, tagging.Tags[0].Key)
	c.Assert(getResult.Tags[0].Value, Equals, tagging.Tags[0].Value)

	// delete BucketTagging
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

	bucketName1 := bucketNamePrefix + randLowStr(5)
	err = client.CreateBucket(bucketName1)
	c.Assert(err, IsNil)

	bucketName2 := bucketNamePrefix + randLowStr(5)
	err = client.CreateBucket(bucketName2)
	c.Assert(err, IsNil)

	// Bucket Tagging
	var tagging Tagging
	tagging.Tags = []Tag{Tag{Key: "testkey", Value: "testvalue"}}
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
