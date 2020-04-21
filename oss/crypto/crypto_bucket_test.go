package osscrypto

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	math_rand "math/rand"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	kms "github.com/aliyun/alibaba-cloud-sdk-go/services/kms"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	. "gopkg.in/check.v1"
)

func Test(t *testing.T) {
	TestingT(t)
}

type OssCryptoBucketSuite struct {
}

var _ = Suite(&OssCryptoBucketSuite{})

var (
	matDesc = make(map[string]string)

	rsaPublicKey string = `-----BEGIN PUBLIC KEY-----
MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQCokfiAVXXf5ImFzKDw+XO/UByW
6mse2QsIgz3ZwBtMNu59fR5zttSx+8fB7vR4CN3bTztrP9A6bjoN0FFnhlQ3vNJC
5MFO1PByrE/MNd5AAfSVba93I6sx8NSk5MzUCA4NJzAUqYOEWGtGBcom6kEF6MmR
1EKib1Id8hpooY5xaQIDAQAB
-----END PUBLIC KEY-----`

	rsaPrivateKey string = `-----BEGIN PRIVATE KEY-----
MIICdQIBADANBgkqhkiG9w0BAQEFAASCAl8wggJbAgEAAoGBAKiR+IBVdd/kiYXM
oPD5c79QHJbqax7ZCwiDPdnAG0w27n19HnO21LH7x8Hu9HgI3dtPO2s/0DpuOg3Q
UWeGVDe80kLkwU7U8HKsT8w13kAB9JVtr3cjqzHw1KTkzNQIDg0nMBSpg4RYa0YF
yibqQQXoyZHUQqJvUh3yGmihjnFpAgMBAAECgYA49RmCQ14QyKevDfVTdvYlLmx6
kbqgMbYIqk+7w611kxoCTMR9VMmJWgmk/Zic9mIAOEVbd7RkCdqT0E+xKzJJFpI2
ZHjrlwb21uqlcUqH1Gn+wI+jgmrafrnKih0kGucavr/GFi81rXixDrGON9KBE0FJ
cPVdc0XiQAvCBnIIAQJBANXu3htPH0VsSznfqcDE+w8zpoAJdo6S/p30tcjsDQnx
l/jYV4FXpErSrtAbmI013VYkdJcghNSLNUXppfk2e8UCQQDJt5c07BS9i2SDEXiz
byzqCfXVzkdnDj9ry9mba1dcr9B9NCslVelXDGZKvQUBqNYCVxg398aRfWlYDTjU
IoVVAkAbTyjPN6R4SkC4HJMg5oReBmvkwFCAFsemBk0GXwuzD0IlJAjXnAZ+/rIO
ItewfwXIL1Mqz53lO/gK+q6TR585AkB304KUIoWzjyF3JqLP3IQOxzns92u9EV6l
V2P+CkbMPXiZV6sls6I4XppJXX2i3bu7iidN3/dqJ9izQK94fMU9AkBZvgsIPCot
y1/POIbv9LtnviDKrmpkXgVQSU4BmTPvXwTJm8APC7P/horSh3SVf1zgmnsyjm9D
hO92gGc+4ajL
-----END PRIVATE KEY-----`

	rsaPublicKeyPks1 string = `-----BEGIN RSA PUBLIC KEY-----
MIGJAoGBAKiR+IBVdd/kiYXMoPD5c79QHJbqax7ZCwiDPdnAG0w27n19HnO21LH7
x8Hu9HgI3dtPO2s/0DpuOg3QUWeGVDe80kLkwU7U8HKsT8w13kAB9JVtr3cjqzHw
1KTkzNQIDg0nMBSpg4RYa0YFyibqQQXoyZHUQqJvUh3yGmihjnFpAgMBAAE=
-----END RSA PUBLIC KEY-----`

	rsaPrivateKeyPks1 string = `-----BEGIN RSA PRIVATE KEY-----
MIICWwIBAAKBgQCokfiAVXXf5ImFzKDw+XO/UByW6mse2QsIgz3ZwBtMNu59fR5z
ttSx+8fB7vR4CN3bTztrP9A6bjoN0FFnhlQ3vNJC5MFO1PByrE/MNd5AAfSVba93
I6sx8NSk5MzUCA4NJzAUqYOEWGtGBcom6kEF6MmR1EKib1Id8hpooY5xaQIDAQAB
AoGAOPUZgkNeEMinrw31U3b2JS5sepG6oDG2CKpPu8OtdZMaAkzEfVTJiVoJpP2Y
nPZiADhFW3e0ZAnak9BPsSsySRaSNmR465cG9tbqpXFKh9Rp/sCPo4Jq2n65yood
JBrnGr6/xhYvNa14sQ6xjjfSgRNBSXD1XXNF4kALwgZyCAECQQDV7t4bTx9FbEs5
36nAxPsPM6aACXaOkv6d9LXI7A0J8Zf42FeBV6RK0q7QG5iNNd1WJHSXIITUizVF
6aX5NnvFAkEAybeXNOwUvYtkgxF4s28s6gn11c5HZw4/a8vZm2tXXK/QfTQrJVXp
VwxmSr0FAajWAlcYN/fGkX1pWA041CKFVQJAG08ozzekeEpAuByTIOaEXgZr5MBQ
gBbHpgZNBl8Lsw9CJSQI15wGfv6yDiLXsH8FyC9TKs+d5Tv4Cvquk0efOQJAd9OC
lCKFs48hdyaiz9yEDsc57PdrvRFepVdj/gpGzD14mVerJbOiOF6aSV19ot27u4on
Td/3aifYs0CveHzFPQJAWb4LCDwqLctfzziG7/S7Z74gyq5qZF4FUElOAZkz718E
yZvADwuz/4aK0od0lX9c4Jp7Mo5vQ4TvdoBnPuGoyw==
-----END RSA PRIVATE KEY-----`
)

var (
	// Endpoint/ID/Key
	endpoint         = os.Getenv("OSS_TEST_ENDPOINT")
	accessID         = os.Getenv("OSS_TEST_ACCESS_KEY_ID")
	accessKey        = os.Getenv("OSS_TEST_ACCESS_KEY_SECRET")
	kmsID            = os.Getenv("OSS_TEST_KMS_ID")
	kmsRegion        = os.Getenv("OSS_TEST_KMS_REGION")
	kmsAccessID      = accessID
	kmsAccessKey     = accessKey
	bucketNamePrefix = "go-sdk-test-bucket-"
	objectNamePrefix = "go-sdk-test-object-"
)

var (
	logPath            = "go_sdk_test_" + time.Now().Format("20060102_150405") + ".log"
	testLogFile, _     = os.OpenFile(logPath, os.O_RDWR|os.O_CREATE, 0664)
	testLogger         = log.New(testLogFile, "", log.Ldate|log.Ltime|log.Lshortfile)
	letters            = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	timeoutInOperation = 3 * time.Second
)

func RandStr(n int) string {
	b := make([]rune, n)
	randMarker := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := range b {
		b[i] = letters[randMarker.Intn(len(letters))]
	}
	return string(b)
}

func RandLowStr(n int) string {
	return strings.ToLower(RandStr(n))
}

func GetFileMD5(filePath string) (string, error) {
	fd, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer fd.Close()

	md5 := md5.New()
	_, err = io.Copy(md5, fd)
	if err != nil {
		return "", fmt.Errorf("buff copy error")
	}
	md5Str := hex.EncodeToString(md5.Sum(nil))
	return md5Str, nil
}

func GetStringMd5(s string) string {
	md5 := md5.New()
	md5.Write([]byte(s))
	md5Str := hex.EncodeToString(md5.Sum(nil))
	return md5Str
}

func ForceDeleteBucket(client *oss.Client, bucketName string, c *C) {
	bucket, err := client.Bucket(bucketName)
	c.Assert(err, IsNil)

	// Delete Object
	marker := oss.Marker("")
	for {
		lor, err := bucket.ListObjects(marker)
		c.Assert(err, IsNil)
		for _, object := range lor.Objects {
			err = bucket.DeleteObject(object.Key)
			c.Assert(err, IsNil)
		}
		marker = oss.Marker(lor.NextMarker)
		if !lor.IsTruncated {
			break
		}
	}

	// Delete Object Versions and DeleteMarks
	keyMarker := oss.KeyMarker("")
	versionIdMarker := oss.VersionIdMarker("")
	options := []oss.Option{keyMarker, versionIdMarker}
	for {
		lor, err := bucket.ListObjectVersions(options...)
		if err != nil {
			break
		}

		for _, object := range lor.ObjectDeleteMarkers {
			err = bucket.DeleteObject(object.Key, oss.VersionId(object.VersionId))
			c.Assert(err, IsNil)
		}

		for _, object := range lor.ObjectVersions {
			err = bucket.DeleteObject(object.Key, oss.VersionId(object.VersionId))
			c.Assert(err, IsNil)
		}

		keyMarker = oss.KeyMarker(lor.NextKeyMarker)
		versionIdMarker := oss.VersionIdMarker(lor.NextVersionIdMarker)
		options = []oss.Option{keyMarker, versionIdMarker}

		if !lor.IsTruncated {
			break
		}
	}

	// Delete Part
	keyMarker = oss.KeyMarker("")
	uploadIDMarker := oss.UploadIDMarker("")
	for {
		lmur, err := bucket.ListMultipartUploads(keyMarker, uploadIDMarker)
		c.Assert(err, IsNil)
		for _, upload := range lmur.Uploads {
			var imur = oss.InitiateMultipartUploadResult{Bucket: bucketName,
				Key: upload.Key, UploadID: upload.UploadID}
			err = bucket.AbortMultipartUpload(imur)
			c.Assert(err, IsNil)
		}
		keyMarker = oss.KeyMarker(lmur.NextKeyMarker)
		uploadIDMarker = oss.UploadIDMarker(lmur.NextUploadIDMarker)
		if !lmur.IsTruncated {
			break
		}
	}

	// delete live channel
	strMarker := ""
	for {
		result, err := bucket.ListLiveChannel(oss.Marker(strMarker))
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

func ReadBody(body io.ReadCloser) (string, error) {
	data, err := ioutil.ReadAll(body)
	body.Close()
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// SetUpSuite runs once when the suite starts running
func (s *OssCryptoBucketSuite) SetUpSuite(c *C) {
}

// TearDownSuite runs before each test or benchmark starts running
func (s *OssCryptoBucketSuite) TearDownSuite(c *C) {
}

// SetUpTest runs after each test or benchmark runs
func (s *OssCryptoBucketSuite) SetUpTest(c *C) {
}

// TearDownTest runs once after all tests or benchmarks have finished running
func (s *OssCryptoBucketSuite) TearDownTest(c *C) {

}

func (s *OssCryptoBucketSuite) TestPutObjectNormalPks8(c *C) {
	// create a bucket with default proprety
	client, err := oss.New(endpoint, accessID, accessKey)
	c.Assert(err, IsNil)

	bucketName := bucketNamePrefix + RandLowStr(6)
	err = client.CreateBucket(bucketName)
	c.Assert(err, IsNil)

	// crypto bucket
	testMatDesc := make(map[string]string)
	testMatDesc["desc"] = "test rsa key"
	masterRsaCipher, _ := CreateMasterRsa(testMatDesc, rsaPublicKey, rsaPrivateKey)
	contentProvider := CreateAesCtrCipher(masterRsaCipher)
	bucket, err := GetCryptoBucket(client, bucketName, contentProvider)

	objectName := objectNamePrefix + RandStr(8)
	objectValue := RandStr(1023)

	// Put string
	var respHeader http.Header
	err = bucket.PutObject(objectName, strings.NewReader(objectValue), oss.GetResponseHeader(&respHeader))
	c.Assert(err, IsNil)

	// Check
	body, err := bucket.GetObject(objectName)
	c.Assert(err, IsNil)
	text, err := ReadBody(body)
	c.Assert(text, Equals, objectValue)

	// non-crypto bucket download
	normalBucket, err := client.Bucket(bucketName)
	c.Assert(err, IsNil)
	body, err = normalBucket.GetObject(objectName)
	c.Assert(err, IsNil)
	encryptText, err := ReadBody(body)
	c.Assert(encryptText != objectValue, Equals, true)

	// acl
	acl, err := bucket.GetObjectACL(objectName)
	c.Assert(err, IsNil)
	c.Assert(acl.ACL, Equals, "default")

	err = bucket.DeleteObject(objectName)
	c.Assert(err, IsNil)

	// put with meta
	options := []oss.Option{
		oss.ObjectACL(oss.ACLPublicRead),
		oss.Meta("myprop", "mypropval"),
	}
	err = bucket.PutObject(objectName, strings.NewReader(objectValue), options...)
	c.Assert(err, IsNil)

	// Check
	body, err = bucket.GetObject(objectName)
	c.Assert(err, IsNil)
	text, err = ReadBody(body)
	c.Assert(err, IsNil)
	c.Assert(text, Equals, objectValue)

	acl, err = bucket.GetObjectACL(objectName)
	c.Assert(err, IsNil)
	c.Assert(acl.ACL, Equals, string(oss.ACLPublicRead))

	meta, err := bucket.GetObjectDetailedMeta(objectName)
	c.Assert(err, IsNil)
	c.Assert(meta.Get("X-Oss-Meta-Myprop"), Equals, "mypropval")

	ForceDeleteBucket(client, bucketName, c)
}

func (s *OssCryptoBucketSuite) TestPutObjectNormalPks1(c *C) {
	// create a bucket with default proprety
	client, err := oss.New(endpoint, accessID, accessKey)
	c.Assert(err, IsNil)

	bucketName := bucketNamePrefix + RandLowStr(6)
	err = client.CreateBucket(bucketName)
	c.Assert(err, IsNil)

	// crypto bucket
	masterRsaCipher, _ := CreateMasterRsa(matDesc, rsaPublicKeyPks1, rsaPrivateKeyPks1)
	contentProvider := CreateAesCtrCipher(masterRsaCipher)
	bucket, err := GetCryptoBucket(client, bucketName, contentProvider)

	objectName := objectNamePrefix + RandStr(8)
	objectValue := RandStr(1023)

	// Put string
	var respHeader http.Header
	err = bucket.PutObject(objectName, strings.NewReader(objectValue), oss.GetResponseHeader(&respHeader))
	c.Assert(err, IsNil)

	// Check
	body, err := bucket.GetObject(objectName)
	c.Assert(err, IsNil)
	text, err := ReadBody(body)
	c.Assert(text, Equals, objectValue)

	// non-crypto bucket download
	normalBucket, err := client.Bucket(bucketName)
	c.Assert(err, IsNil)
	body, err = normalBucket.GetObject(objectName)
	c.Assert(err, IsNil)
	encryptText, err := ReadBody(body)
	c.Assert(encryptText != objectValue, Equals, true)

	// acl
	acl, err := bucket.GetObjectACL(objectName)
	c.Assert(err, IsNil)
	c.Assert(acl.ACL, Equals, "default")

	err = bucket.DeleteObject(objectName)
	c.Assert(err, IsNil)

	// put with meta
	options := []oss.Option{
		oss.ObjectACL(oss.ACLPublicRead),
		oss.Meta("myprop", "mypropval"),
	}
	err = bucket.PutObject(objectName, strings.NewReader(objectValue), options...)
	c.Assert(err, IsNil)

	// Check
	body, err = bucket.GetObject(objectName)
	c.Assert(err, IsNil)
	text, err = ReadBody(body)
	c.Assert(err, IsNil)
	c.Assert(text, Equals, objectValue)

	acl, err = bucket.GetObjectACL(objectName)
	c.Assert(err, IsNil)
	c.Assert(acl.ACL, Equals, string(oss.ACLPublicRead))

	meta, err := bucket.GetObjectDetailedMeta(objectName)
	c.Assert(err, IsNil)
	c.Assert(meta.Get("X-Oss-Meta-Myprop"), Equals, "mypropval")

	ForceDeleteBucket(client, bucketName, c)
}

func (s *OssCryptoBucketSuite) TestPutObjectEmptyPks1(c *C) {
	// create a bucket with default proprety
	client, err := oss.New(endpoint, accessID, accessKey)
	c.Assert(err, IsNil)

	bucketName := bucketNamePrefix + RandLowStr(6)
	err = client.CreateBucket(bucketName)
	c.Assert(err, IsNil)

	// crypto bucket
	masterRsaCipher, _ := CreateMasterRsa(matDesc, rsaPublicKeyPks1, rsaPrivateKeyPks1)
	contentProvider := CreateAesCtrCipher(masterRsaCipher)
	bucket, err := GetCryptoBucket(client, bucketName, contentProvider)

	objectName := objectNamePrefix + RandStr(8)
	objectValue := ""

	// Put empty string
	var respHeader http.Header
	err = bucket.PutObject(objectName, strings.NewReader(objectValue), oss.GetResponseHeader(&respHeader))
	c.Assert(err, IsNil)

	// Check
	body, err := bucket.GetObject(objectName)
	c.Assert(err, IsNil)
	text, err := ReadBody(body)
	c.Assert(text, Equals, objectValue)

	// non-crypto bucket download
	normalBucket, err := client.Bucket(bucketName)
	c.Assert(err, IsNil)
	body, err = normalBucket.GetObject(objectName)
	c.Assert(err, IsNil)
	encryptText, err := ReadBody(body)
	c.Assert(encryptText == objectValue, Equals, true)
	ForceDeleteBucket(client, bucketName, c)
}

func (s *OssCryptoBucketSuite) TestPutObjectSmallSizePks1(c *C) {
	// create a bucket with default proprety
	client, err := oss.New(endpoint, accessID, accessKey)
	c.Assert(err, IsNil)

	bucketName := bucketNamePrefix + RandLowStr(6)
	err = client.CreateBucket(bucketName)
	c.Assert(err, IsNil)

	// crypto bucket
	masterRsaCipher, _ := CreateMasterRsa(matDesc, rsaPublicKeyPks1, rsaPrivateKeyPks1)
	contentProvider := CreateAesCtrCipher(masterRsaCipher)
	bucket, err := GetCryptoBucket(client, bucketName, contentProvider)

	objectName := objectNamePrefix + RandStr(8)
	objectValue := "123"

	var respHeader http.Header
	err = bucket.PutObject(objectName, strings.NewReader(objectValue), oss.GetResponseHeader(&respHeader))
	c.Assert(err, IsNil)

	// Check
	body, err := bucket.GetObject(objectName)
	c.Assert(err, IsNil)
	text, err := ReadBody(body)
	c.Assert(text, Equals, objectValue)

	// non-crypto bucket download
	normalBucket, err := client.Bucket(bucketName)
	c.Assert(err, IsNil)
	body, err = normalBucket.GetObject(objectName)
	c.Assert(err, IsNil)
	encryptText, err := ReadBody(body)
	c.Assert(encryptText != objectValue, Equals, true)
	ForceDeleteBucket(client, bucketName, c)
}

func (s *OssCryptoBucketSuite) TestPutObjectEmptyFilePks1(c *C) {
	// create a bucket with default proprety
	client, err := oss.New(endpoint, accessID, accessKey)
	c.Assert(err, IsNil)

	bucketName := bucketNamePrefix + RandLowStr(6)
	err = client.CreateBucket(bucketName)
	c.Assert(err, IsNil)

	// crypto bucket
	masterRsaCipher, _ := CreateMasterRsa(matDesc, rsaPublicKeyPks1, rsaPrivateKeyPks1)
	contentProvider := CreateAesCtrCipher(masterRsaCipher)
	bucket, err := GetCryptoBucket(client, bucketName, contentProvider)

	fileName := "oss-go-sdk-test-file-" + RandStr(5)
	fo, err := os.Create(fileName)
	c.Assert(err, IsNil)
	_, err = fo.Write([]byte(""))
	c.Assert(err, IsNil)
	fo.Close()

	objectName := objectNamePrefix + RandStr(8)

	// file not exist
	err = bucket.PutObjectFromFile(objectName, "/root1/abc.txt")
	c.Assert(err, NotNil)

	err = bucket.PutObjectFromFile(objectName, fileName)
	c.Assert(err, IsNil)

	downFileName := fileName + "-down"

	// Check
	err = bucket.GetObjectToFile(objectName, downFileName)
	c.Assert(err, IsNil)

	b1, err := ioutil.ReadFile(fileName)
	b2, err := ioutil.ReadFile(downFileName)
	c.Assert(len(b1), Equals, 0)
	c.Assert(string(b1), Equals, string(b2))

	os.Remove(downFileName)
	os.Remove(fileName)

	ForceDeleteBucket(client, bucketName, c)
}

func (s *OssCryptoBucketSuite) TestKmsPutObjectNormal(c *C) {
	// create a bucket with default proprety
	client, err := oss.New(endpoint, accessID, accessKey)
	c.Assert(err, IsNil)

	bucketName := bucketNamePrefix + RandLowStr(6)
	err = client.CreateBucket(bucketName)
	c.Assert(err, IsNil)

	kmsClient, err := kms.NewClientWithAccessKey(kmsRegion, kmsAccessID, kmsAccessKey)
	c.Assert(err, IsNil)

	// crypto bucket
	masterKmsCipher, _ := CreateMasterAliKms(matDesc, kmsID, kmsClient)
	contentProvider := CreateAesCtrCipher(masterKmsCipher)
	bucket, err := GetCryptoBucket(client, bucketName, contentProvider)

	objectName := objectNamePrefix + RandStr(8)
	objectValue := RandStr(1023)

	// Put string
	var respHeader http.Header
	err = bucket.PutObject(objectName, strings.NewReader(objectValue), oss.GetResponseHeader(&respHeader))
	c.Assert(err, IsNil)

	// Check
	body, err := bucket.GetObject(objectName)
	c.Assert(err, IsNil)
	text, err := ReadBody(body)
	c.Assert(text, Equals, objectValue)

	// non-crypto bucket download
	normalBucket, err := client.Bucket(bucketName)
	c.Assert(err, IsNil)
	body, err = normalBucket.GetObject(objectName)
	c.Assert(err, IsNil)
	encryptText, err := ReadBody(body)
	c.Assert(encryptText != objectValue, Equals, true)

	// acl
	acl, err := bucket.GetObjectACL(objectName)
	c.Assert(err, IsNil)
	c.Assert(acl.ACL, Equals, "default")

	err = bucket.DeleteObject(objectName)
	c.Assert(err, IsNil)

	// put with meta
	options := []oss.Option{
		oss.ObjectACL(oss.ACLPublicRead),
		oss.Meta("myprop", "mypropval"),
	}
	err = bucket.PutObject(objectName, strings.NewReader(objectValue), options...)
	c.Assert(err, IsNil)

	// Check
	body, err = bucket.GetObject(objectName)
	c.Assert(err, IsNil)
	text, err = ReadBody(body)
	c.Assert(err, IsNil)
	c.Assert(text, Equals, objectValue)

	acl, err = bucket.GetObjectACL(objectName)
	c.Assert(err, IsNil)
	c.Assert(acl.ACL, Equals, string(oss.ACLPublicRead))

	meta, err := bucket.GetObjectDetailedMeta(objectName)
	c.Assert(err, IsNil)
	c.Assert(meta.Get("X-Oss-Meta-Myprop"), Equals, "mypropval")

	// put object error,bucket not exist
	bucket.BucketName = bucket.BucketName + "-not-exist"
	err = bucket.PutObject(objectName, strings.NewReader(objectValue), options...)
	c.Assert(err, NotNil)

	ForceDeleteBucket(client, bucketName, c)
}

type MockKmsManager struct {
}

func (mg *MockKmsManager) GetMasterKey(matDesc map[string]string) ([]string, error) {
	if len(matDesc) == 0 {
		return nil, fmt.Errorf("not found")
	}

	keyList := []string{kmsID}
	return keyList, nil
}

func (s *OssCryptoBucketSuite) TestRsaBucketDecrptObjectWithKmsSuccess(c *C) {
	// create a bucket with default proprety
	client, err := oss.New(endpoint, accessID, accessKey)
	c.Assert(err, IsNil)

	bucketName := bucketNamePrefix + RandLowStr(6)
	err = client.CreateBucket(bucketName)
	c.Assert(err, IsNil)

	kmsClient, err := kms.NewClientWithAccessKey(kmsRegion, kmsAccessID, kmsAccessKey)
	c.Assert(err, IsNil)

	// crypto bucket with kms
	testMatDesc := make(map[string]string)
	testMatDesc["desc"] = "test kms wrap"
	masterKmsCipher, _ := CreateMasterAliKms(testMatDesc, kmsID, kmsClient)
	contentProvider := CreateAesCtrCipher(masterKmsCipher)
	bucket, err := GetCryptoBucket(client, bucketName, contentProvider)

	objectName := objectNamePrefix + RandStr(8)
	objectValue := RandStr(1023)

	// Put string
	var respHeader http.Header
	err = bucket.PutObject(objectName, strings.NewReader(objectValue), oss.GetResponseHeader(&respHeader))
	c.Assert(err, IsNil)

	// crypto bucket with rsa
	var masterManager MockKmsManager
	var options []CryptoBucketOption
	options = append(options, SetAliKmsClient(kmsClient))
	options = append(options, SetMasterCipherManager(&masterManager))

	masterRsaCipher, _ := CreateMasterRsa(matDesc, rsaPublicKey, rsaPrivateKey)
	rsaProvider := CreateAesCtrCipher(masterRsaCipher)
	rsaBucket, err := GetCryptoBucket(client, bucketName, rsaProvider, options...)

	// Check
	body, err := rsaBucket.GetObject(objectName)
	c.Assert(err, IsNil)
	text, err := ReadBody(body)
	c.Assert(text, Equals, objectValue)

	// non-crypto bucket download
	normalBucket, err := client.Bucket(bucketName)
	c.Assert(err, IsNil)
	body, err = normalBucket.GetObject(objectName)
	c.Assert(err, IsNil)
	encryptText, err := ReadBody(body)
	c.Assert(encryptText != objectValue, Equals, true)
	ForceDeleteBucket(client, bucketName, c)
}

func (s *OssCryptoBucketSuite) TestRsaBucketDecrptObjectWithKmsError(c *C) {
	// create a bucket with default proprety
	client, err := oss.New(endpoint, accessID, accessKey)
	c.Assert(err, IsNil)

	bucketName := bucketNamePrefix + RandLowStr(6)
	err = client.CreateBucket(bucketName)
	c.Assert(err, IsNil)

	kmsClient, err := kms.NewClientWithAccessKey(kmsRegion, kmsAccessID, kmsAccessKey)
	c.Assert(err, IsNil)

	// crypto bucket with kms
	testMatDesc := make(map[string]string)
	testMatDesc["desc"] = "test kms wrap"
	masterKmsCipher, _ := CreateMasterAliKms(testMatDesc, kmsID, kmsClient)
	contentProvider := CreateAesCtrCipher(masterKmsCipher)
	bucket, err := GetCryptoBucket(client, bucketName, contentProvider)

	objectName := objectNamePrefix + RandStr(8)
	objectValue := RandStr(1023)

	// Put string
	var respHeader http.Header
	err = bucket.PutObject(objectName, strings.NewReader(objectValue), oss.GetResponseHeader(&respHeader))
	c.Assert(err, IsNil)

	// crypto bucket with rsa
	var masterManager MockKmsManager
	var options []CryptoBucketOption

	// kms client is nil
	//options = append(options, SetAliKmsClient(kmsClient))

	options = append(options, SetMasterCipherManager(&masterManager))

	masterRsaCipher, _ := CreateMasterRsa(matDesc, rsaPublicKey, rsaPrivateKey)
	rsaProvider := CreateAesCtrCipher(masterRsaCipher)
	rsaBucket, err := GetCryptoBucket(client, bucketName, rsaProvider, options...)

	// Check
	_, err = rsaBucket.GetObject(objectName)
	c.Assert(err, NotNil)
	ForceDeleteBucket(client, bucketName, c)
}

func (s *OssCryptoBucketSuite) TestRangeGetObject(c *C) {
	// create a bucket with default proprety
	client, err := oss.New(endpoint, accessID, accessKey)
	c.Assert(err, IsNil)

	bucketName := bucketNamePrefix + RandLowStr(6)
	err = client.CreateBucket(bucketName)
	c.Assert(err, IsNil)

	// crypto bucket
	masterRsaCipher, _ := CreateMasterRsa(matDesc, rsaPublicKey, rsaPrivateKey)
	contentProvider := CreateAesCtrCipher(masterRsaCipher)
	bucket, err := GetCryptoBucket(client, bucketName, contentProvider)

	objectName := objectNamePrefix + RandStr(8)
	contentLen := 1024 * 1024
	content := RandStr(contentLen)
	err = bucket.PutObject(objectName, strings.NewReader(content))
	c.Assert(err, IsNil)

	// range get
	for i := 0; i < 20; i++ {
		math_rand.Seed(time.Now().UnixNano())
		rangeStart := rand.Intn(contentLen)
		rangeEnd := rangeStart + rand.Intn(contentLen-rangeStart)
		if rangeEnd == rangeStart || rangeStart >= contentLen-1 {
			continue
		}

		body, err := bucket.GetObject(objectName, oss.Range(int64(rangeStart), int64(rangeEnd)))
		c.Assert(err, IsNil)
		downText, err := ReadBody(body)
		c.Assert(len(downText) > 0, Equals, true)
		downMd5 := GetStringMd5(downText)

		srcText := content[rangeStart : rangeEnd+1]
		srcMd5 := GetStringMd5(srcText)

		c.Assert(len(downText), Equals, len(srcText))
		c.Assert(downMd5, Equals, srcMd5)
	}
	ForceDeleteBucket(client, bucketName, c)
}

func (s *OssCryptoBucketSuite) TestGetNormalObject(c *C) {
	// create a bucket with default proprety
	client, err := oss.New(endpoint, accessID, accessKey)
	c.Assert(err, IsNil)

	bucketName := bucketNamePrefix + RandLowStr(6)
	err = client.CreateBucket(bucketName)
	c.Assert(err, IsNil)

	// crypto bucket
	masterRsaCipher, _ := CreateMasterRsa(matDesc, rsaPublicKey, rsaPrivateKey)
	contentProvider := CreateAesCtrCipher(masterRsaCipher)
	bucket, err := GetCryptoBucket(client, bucketName, contentProvider)

	// normal bucket
	normalBucket, _ := client.Bucket(bucketName)

	objectName := objectNamePrefix + RandStr(8)
	objectValue := RandStr(1023)

	// Put string
	err = normalBucket.PutObject(objectName, strings.NewReader(objectValue))
	c.Assert(err, IsNil)

	// Check
	body, err := bucket.GetObject(objectName)
	c.Assert(err, IsNil)
	text, err := ReadBody(body)
	c.Assert(text, Equals, objectValue)

	// delete object
	err = bucket.DeleteObject(objectName)
	c.Assert(err, IsNil)

	// get object again
	body, err = bucket.GetObject(objectName)
	c.Assert(err, NotNil)

	ForceDeleteBucket(client, bucketName, c)
}

func (s *OssCryptoBucketSuite) TestGetCryptoBucketNotSupport(c *C) {
	// create a bucket with default proprety
	client, err := oss.New(endpoint, accessID, accessKey)
	c.Assert(err, IsNil)

	bucketName := bucketNamePrefix + RandLowStr(6)

	// crypto bucket
	masterRsaCipher, _ := CreateMasterRsa(matDesc, rsaPublicKey, rsaPrivateKey)
	contentProvider := CreateAesCtrCipher(masterRsaCipher)
	bucket, err := GetCryptoBucket(client, bucketName, contentProvider)

	objectName := objectNamePrefix + RandStr(8)
	objectValue := RandStr(1023)

	// AppendObject
	_, err = bucket.AppendObject(objectName, strings.NewReader(objectValue), 0)
	c.Assert(err, NotNil)

	// DoAppendObject
	var request oss.AppendObjectRequest
	var options []oss.Option
	_, err = bucket.DoAppendObject(&request, options)
	c.Assert(err, NotNil)

	// PutObjectWithURL
	err = bucket.PutObjectWithURL("oss://bucket/object", strings.NewReader(objectValue))
	c.Assert(err, NotNil)

	// PutObjectFromFileWithURL
	err = bucket.PutObjectFromFileWithURL("oss://bucket/object", "file.txt")
	c.Assert(err, NotNil)

	// DoPutObjectWithURL
	_, err = bucket.DoPutObjectWithURL("oss://bucket/object", strings.NewReader(objectValue), options)
	c.Assert(err, NotNil)

	// GetObjectWithURL
	_, err = bucket.GetObjectWithURL("oss://bucket/object")
	c.Assert(err, NotNil)

	// GetObjectToFileWithURL
	err = bucket.GetObjectToFileWithURL("oss://bucket/object", "file.txt")
	c.Assert(err, NotNil)

	// DoGetObjectWithURL
	_, err = bucket.DoGetObjectWithURL("oss://bucket/object", options)
	c.Assert(err, NotNil)

	// ProcessObject
	_, err = bucket.ProcessObject("oss://bucket/object", "")
	c.Assert(err, NotNil)

	// DownloadFile
	err = bucket.DownloadFile(objectName, "file.txt", 1024)
	c.Assert(err, NotNil)

	// CopyFile
	err = bucket.CopyFile("src-bucket", "src-object", "dest-object", 1024)
	c.Assert(err, NotNil)

	// UploadFile
	err = bucket.UploadFile(objectName, "file.txt", 1024)
	c.Assert(err, NotNil)
}

type MockRsaManager struct {
}

func (mg *MockRsaManager) GetMasterKey(matDesc map[string]string) ([]string, error) {
	if len(matDesc) == 0 {
		return nil, fmt.Errorf("not found")
	}

	keyList := []string{rsaPublicKey, rsaPrivateKey}
	return keyList, nil
}

func (s *OssCryptoBucketSuite) TestGetMasterKey(c *C) {
	// create a bucket with default proprety
	client, err := oss.New(endpoint, accessID, accessKey)
	c.Assert(err, IsNil)

	bucketName := bucketNamePrefix + RandLowStr(6)
	err = client.CreateBucket(bucketName)
	c.Assert(err, IsNil)

	// crypto bucket
	testMatDesc := make(map[string]string)
	testMatDesc["desc"] = "test rsa key"
	masterRsaCipher, _ := CreateMasterRsa(testMatDesc, rsaPublicKey, rsaPrivateKey)
	contentProvider := CreateAesCtrCipher(masterRsaCipher)
	bucket, err := GetCryptoBucket(client, bucketName, contentProvider)

	objectName := objectNamePrefix + RandStr(8)

	fileName := "../../sample/BingWallpaper-2015-11-07.jpg"
	srcMD5, err := GetFileMD5(fileName)
	c.Assert(err, IsNil)

	err = bucket.PutObjectFromFile(objectName, fileName)
	c.Assert(err, IsNil)

	// other crypto bucket
	var rsaManager MockRsaManager
	masterRsaCipherOther, _ := CreateMasterRsa(matDesc, rsaPublicKey, rsaPrivateKey)
	contentProviderOther := CreateAesCtrCipher(masterRsaCipherOther)
	bucketOther, err := GetCryptoBucket(client, bucketName, contentProviderOther, SetMasterCipherManager(&rsaManager))

	//  download
	downfileName := "test-go-sdk-file-" + RandLowStr(5) + ".jpg"
	err = bucketOther.GetObjectToFile(objectName, downfileName)
	c.Assert(err, IsNil)
	downFileMD5, err := GetFileMD5(downfileName)
	c.Assert(err, IsNil)
	c.Assert(downFileMD5, Equals, srcMD5)

	// GetObjectToFile error
	err = bucketOther.GetObjectToFile(objectName, "/root1/"+downfileName)
	c.Assert(err, NotNil)

	os.Remove(downfileName)
	ForceDeleteBucket(client, bucketName, c)
}

type MockReader struct {
	Reader io.Reader
}

func (r *MockReader) Read(b []byte) (int, error) {
	return r.Reader.Read(b)
}

func (s *OssCryptoBucketSuite) TestPutObjectUnkownReaderLen(c *C) {
	// create a bucket with default proprety
	client, err := oss.New(endpoint, accessID, accessKey)
	c.Assert(err, IsNil)

	bucketName := bucketNamePrefix + RandLowStr(6)
	err = client.CreateBucket(bucketName)
	c.Assert(err, IsNil)

	// crypto bucket
	masterRsaCipher, _ := CreateMasterRsa(matDesc, rsaPublicKey, rsaPrivateKey)
	contentProvider := CreateAesCtrCipher(masterRsaCipher)
	bucket, err := GetCryptoBucket(client, bucketName, contentProvider)

	objectName := objectNamePrefix + RandStr(8)
	objectValue := RandStr(1023)

	srcMD5 := GetStringMd5(objectValue)
	options := []oss.Option{oss.ContentMD5(srcMD5), oss.ContentLength(1023)}

	// Put string
	mockReader := &MockReader{strings.NewReader(objectValue)}
	err = bucket.PutObject(objectName, mockReader, options...)
	c.Assert(err, IsNil)

	// Check
	body, err := bucket.GetObject(objectName)
	c.Assert(err, IsNil)
	text, err := ReadBody(body)
	c.Assert(text, Equals, objectValue)

	ForceDeleteBucket(client, bucketName, c)
}

func (s *OssCryptoBucketSuite) TestGetDecryptCipher(c *C) {
	// create a bucket with default proprety
	client, err := oss.New(endpoint, accessID, accessKey)
	c.Assert(err, IsNil)

	bucketName := bucketNamePrefix + RandLowStr(6)
	err = client.CreateBucket(bucketName)
	c.Assert(err, IsNil)

	// crypto bucket
	var rsaManager MockRsaManager
	masterRsaCipher, _ := CreateMasterRsa(matDesc, rsaPublicKey, rsaPrivateKey)
	contentProvider := CreateAesCtrCipher(masterRsaCipher)
	bucket, err := GetCryptoBucket(client, bucketName, contentProvider, SetMasterCipherManager(&rsaManager))

	objectName := objectNamePrefix + RandStr(8)
	objectValue := RandStr(1023)

	// Put string
	var respHeader http.Header
	err = bucket.PutObject(objectName, strings.NewReader(objectValue), oss.GetResponseHeader(&respHeader))
	c.Assert(err, IsNil)

	// first,we must head object
	metaInfo, err := bucket.GetObjectDetailedMeta(objectName)
	c.Assert(err, IsNil)

	envelope, _ := getEnvelopeFromHeader(metaInfo)

	// test for getEnvelopeFromHeader
	metaInfo.Set(oss.HTTPHeaderOssMetaPrefix+OssClientSideEncryptionKey, string([]byte{200, 200, 200}))
	_, err = getEnvelopeFromHeader(metaInfo)
	c.Assert(err, NotNil)
	metaInfo.Set(oss.HTTPHeaderOssMetaPrefix+OssClientSideEncryptionKey, envelope.CipherKey)

	metaInfo.Set(oss.HTTPHeaderOssMetaPrefix+OssClientSideEncryptionStart, string([]byte{200, 200, 200}))
	_, err = getEnvelopeFromHeader(metaInfo)
	c.Assert(err, NotNil)
	metaInfo.Set(oss.HTTPHeaderOssMetaPrefix+OssClientSideEncryptionKey, envelope.IV)

	// test for getDecryptCipher
	CEKAlg := envelope.CEKAlg
	envelope.CEKAlg = ""
	_, err = bucket.ExtraCipherBuilder.GetDecryptCipher(envelope, bucket.MasterCipherManager)
	c.Assert(err, NotNil)
	envelope.CEKAlg = CEKAlg

	// matDesc is emtpy
	bucket.MasterCipherManager = &MockRsaManager{}
	_, err = bucket.ExtraCipherBuilder.GetDecryptCipher(envelope, bucket.MasterCipherManager)
	c.Assert(err, NotNil)

	// MasterCipherManager is nil
	bucket.MasterCipherManager = nil
	_, err = bucket.ExtraCipherBuilder.GetDecryptCipher(envelope, bucket.MasterCipherManager)
	c.Assert(err, NotNil)

	WrapAlg := envelope.WrapAlg
	envelope.WrapAlg = "test"
	_, err = bucket.ExtraCipherBuilder.GetDecryptCipher(envelope, bucket.MasterCipherManager)
	c.Assert(err, NotNil)
	envelope.WrapAlg = WrapAlg

	envelope.WrapAlg = KmsAliCryptoWrap
	_, err = bucket.ExtraCipherBuilder.GetDecryptCipher(envelope, bucket.MasterCipherManager)
	c.Assert(err, NotNil)
	ForceDeleteBucket(client, bucketName, c)
}

func (s *OssCryptoBucketSuite) TestGetObjectEncryptedByCppRsa(c *C) {
	// create a bucket with default proprety
	client, err := oss.New(endpoint, accessID, accessKey)
	c.Assert(err, IsNil)

	bucketName := bucketNamePrefix + RandLowStr(6)
	err = client.CreateBucket(bucketName)
	c.Assert(err, IsNil)

	// put object encrypted by cpp
	bucket, err := client.Bucket(bucketName)
	c.Assert(err, IsNil)

	objectName := objectNamePrefix + RandStr(8)
	srcJpgFile := "../../sample/test-client-encryption-src.jpg"
	fileEncryptedByCpp := "../../sample/test-client-encryption-crypto-cpp-rsa.jpg"

	opts := []oss.Option{}
	opts = append(opts, oss.Meta(OssClientSideEncryptionKey, "nyXOp7delQ/MQLjKQMhHLaT0w7u2yQoDLkSnK8MFg/MwYdh4na4/LS8LLbLcM18m8I/ObWUHU775I50sJCpdv+f4e0jLeVRRiDFWe+uo7Puc9j4xHj8YB3QlcIOFQiTxHIB6q+C+RA6lGwqqYVa+n3aV5uWhygyv1MWmESurppg="))
	opts = append(opts, oss.Meta(OssClientSideEncryptionStart, "De/S3T8wFjx7QPxAAFl7h7TeI2EsZlfCwox4WhLGng5DK2vNXxULmulMUUpYkdc9umqmDilgSy5Z3Foafw+v4JJThfw68T/9G2gxZLrQTbAlvFPFfPM9Ehk6cY4+8WpY32uN8w5vrHyoSZGr343NxCUGIp6fQ9sSuOLMoJg7hNw="))
	opts = append(opts, oss.Meta(OssClientSideEncryptionWrapAlg, "RSA/NONE/PKCS1Padding"))
	opts = append(opts, oss.Meta(OssClientSideEncryptionCekAlg, "AES/CTR/NoPadding"))
	err = bucket.PutObjectFromFile(objectName, fileEncryptedByCpp, opts...)
	c.Assert(err, IsNil)

	// download with crypto bucket
	masterRsaCipher, _ := CreateMasterRsa(matDesc, rsaPublicKey, rsaPrivateKey)
	contentProvider := CreateAesCtrCipher(masterRsaCipher)
	cryptoBucket, err := GetCryptoBucket(client, bucketName, contentProvider)

	downFileName := "oss-go-sdk-test-file-" + RandStr(5)
	err = cryptoBucket.GetObjectToFile(objectName, downFileName)
	c.Assert(err, IsNil)

	downMd5, _ := GetFileMD5(downFileName)
	srcJpgMd5, _ := GetFileMD5(srcJpgFile)
	c.Assert(downMd5, Equals, srcJpgMd5)
	os.Remove(downFileName)

	ForceDeleteBucket(client, bucketName, c)
}

func (s *OssCryptoBucketSuite) TestGetObjectEncryptedByPythonRsa(c *C) {
	// create a bucket with default proprety
	client, err := oss.New(endpoint, accessID, accessKey)
	c.Assert(err, IsNil)

	bucketName := bucketNamePrefix + RandLowStr(6)
	err = client.CreateBucket(bucketName)
	c.Assert(err, IsNil)

	// put object encrypted by python
	bucket, err := client.Bucket(bucketName)
	c.Assert(err, IsNil)

	objectName := objectNamePrefix + RandStr(8)
	srcJpgFile := "../../sample/test-client-encryption-src.jpg"
	fileEncryptedByCpp := "../../sample/test-client-encryption-crypto-python-rsa.jpg"

	opts := []oss.Option{}
	opts = append(opts, oss.Meta(OssClientSideEncryptionKey, "ZNQM4g+JykUfOBMkfL8kbvChD3R23UH53sRyTg42h9H2ph8ZJJlo2tSP5Oi3nR5gJAwA/OTrruNq02M2Zt4N7zVWdbFArKbY/CkHpihVYOqsSU4Z8RmrNBm4QfC5om2WElRHNt8hlqhnvzhdorGDB5OoMQ8KvQqXDC53aM5OY64="))
	opts = append(opts, oss.Meta(OssClientSideEncryptionStart, "mZ6kts6kaMm++0akhQQZl+tj8gPWznZ+giHciCQTIzriwBzZZO4d85YZeBStuUPshdnO3QHK63/NH9QFL6pwpLiXI9UZxkGygkp82oB4jaF4HKoQ4ujd670pXLxpljBLnp0sCxiCIaf5Fzp4jgNCurXycY10/5DN7yPPtdw7dkk="))
	opts = append(opts, oss.Meta(OssClientSideEncryptionWrapAlg, "RSA/NONE/PKCS1Padding"))
	opts = append(opts, oss.Meta(OssClientSideEncryptionCekAlg, "AES/CTR/NoPadding"))
	err = bucket.PutObjectFromFile(objectName, fileEncryptedByCpp, opts...)
	c.Assert(err, IsNil)

	// download with crypto bucket
	masterRsaCipher, _ := CreateMasterRsa(matDesc, rsaPublicKey, rsaPrivateKey)
	contentProvider := CreateAesCtrCipher(masterRsaCipher)
	cryptoBucket, err := GetCryptoBucket(client, bucketName, contentProvider)

	downFileName := "oss-go-sdk-test-file-" + RandStr(5)
	err = cryptoBucket.GetObjectToFile(objectName, downFileName)
	c.Assert(err, IsNil)

	downMd5, _ := GetFileMD5(downFileName)
	srcJpgMd5, _ := GetFileMD5(srcJpgFile)
	c.Assert(downMd5, Equals, srcJpgMd5)
	os.Remove(downFileName)

	ForceDeleteBucket(client, bucketName, c)
}

func (s *OssCryptoBucketSuite) TestRepeatedPutObjectFromFile(c *C) {
	// create a bucket with default proprety
	client, err := oss.New(endpoint, accessID, accessKey)
	c.Assert(err, IsNil)

	bucketName := bucketNamePrefix + RandLowStr(6)
	err = client.CreateBucket(bucketName)
	c.Assert(err, IsNil)

	objectName := objectNamePrefix + RandStr(8)
	srcJpgFile := "../../sample/test-client-encryption-src.jpg"

	// put object from file
	masterRsaCipher, _ := CreateMasterRsa(matDesc, rsaPublicKey, rsaPrivateKey)
	contentProvider := CreateAesCtrCipher(masterRsaCipher)
	cryptoBucket, err := GetCryptoBucket(client, bucketName, contentProvider)

	err = cryptoBucket.PutObjectFromFile(objectName, srcJpgFile)
	c.Assert(err, IsNil)

	downFileName := "oss-go-sdk-test-file-" + RandStr(5)
	err = cryptoBucket.GetObjectToFile(objectName, downFileName)
	c.Assert(err, IsNil)

	srcJpgMd5, _ := GetFileMD5(srcJpgFile)
	downMd5, _ := GetFileMD5(downFileName)
	c.Assert(len(srcJpgMd5) > 0, Equals, true)
	c.Assert(len(downMd5) > 0, Equals, true)
	c.Assert(downMd5, Equals, srcJpgMd5)
	os.Remove(downFileName)

	err = cryptoBucket.PutObjectFromFile(objectName+"-other", srcJpgFile)
	c.Assert(err, IsNil)
	err = cryptoBucket.GetObjectToFile(objectName, downFileName)
	c.Assert(err, IsNil)
	downMd5, _ = GetFileMD5(downFileName)
	c.Assert(downMd5, Equals, srcJpgMd5)

	os.Remove(downFileName)
	ForceDeleteBucket(client, bucketName, c)
}

func (s *OssCryptoBucketSuite) TestPutObjectEncryptionUserAgent(c *C) {
	logName := "." + string(os.PathSeparator) + "test-go-sdk.log" + RandStr(5)
	f, err := os.OpenFile(logName, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0660)
	c.Assert(err, IsNil)

	// create a bucket with default proprety
	client, err := oss.New(endpoint, accessID, accessKey)
	c.Assert(err, IsNil)
	client.Config.LogLevel = oss.Debug
	client.Config.Logger = log.New(f, "", log.LstdFlags)

	bucketName := bucketNamePrefix + RandLowStr(6)
	err = client.CreateBucket(bucketName)
	c.Assert(err, IsNil)

	objectName := objectNamePrefix + RandStr(8)
	srcJpgFile := "../../sample/test-client-encryption-src.jpg"

	// put object from file
	masterRsaCipher, _ := CreateMasterRsa(matDesc, rsaPublicKey, rsaPrivateKey)
	contentProvider := CreateAesCtrCipher(masterRsaCipher)
	cryptoBucket, err := GetCryptoBucket(client, bucketName, contentProvider)

	err = cryptoBucket.PutObjectFromFile(objectName, srcJpgFile)
	c.Assert(err, IsNil)

	// read log file,get http info
	contents, err := ioutil.ReadFile(logName)
	c.Assert(err, IsNil)

	httpContent := string(contents)
	c.Assert(strings.Contains(httpContent, EncryptionUaSuffix), Equals, true)

	f.Close()
	os.Remove(logName)
	ForceDeleteBucket(client, bucketName, c)
}

func (s *OssCryptoBucketSuite) TestPutObjectNormalUserAgent(c *C) {
	logName := "." + string(os.PathSeparator) + "test-go-sdk.log" + RandStr(5)
	f, err := os.OpenFile(logName, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0660)
	c.Assert(err, IsNil)

	// create a bucket with default proprety
	client, err := oss.New(endpoint, accessID, accessKey)
	c.Assert(err, IsNil)
	client.Config.LogLevel = oss.Debug
	client.Config.Logger = log.New(f, "", log.LstdFlags)

	bucketName := bucketNamePrefix + RandLowStr(6)
	err = client.CreateBucket(bucketName)
	c.Assert(err, IsNil)

	objectName := objectNamePrefix + RandStr(8)
	srcJpgFile := "../../sample/test-client-encryption-src.jpg"

	bucket, err := client.Bucket(bucketName)

	err = bucket.PutObjectFromFile(objectName, srcJpgFile)
	c.Assert(err, IsNil)

	// read log file,get http info
	contents, err := ioutil.ReadFile(logName)
	c.Assert(err, IsNil)

	httpContent := string(contents)
	c.Assert(strings.Contains(httpContent, EncryptionUaSuffix), Equals, false)

	f.Close()
	os.Remove(logName)
	ForceDeleteBucket(client, bucketName, c)
}
