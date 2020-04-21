// multipart test

package osscrypto

import (
	"io/ioutil"
	"os"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	. "gopkg.in/check.v1"
)

func (s *OssCryptoBucketSuite) TestMultipartUpload(c *C) {
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
	fileName := "../../sample/BingWallpaper-2015-11-07.jpg"

	srcMD5, err := GetFileMD5(fileName)
	c.Assert(err, IsNil)

	fileInfo, err := os.Stat(fileName)
	dataSize := fileInfo.Size()
	c.Assert(err, IsNil)

	options := []oss.Option{oss.Meta("my", "myprop")}
	var cryptoContext PartCryptoContext
	cryptoContext.DataSize = dataSize
	cryptoContext.PartSize = (dataSize / 16 / 3) * 16
	imur, err := bucket.InitiateMultipartUpload(objectName, &cryptoContext, options...)
	c.Assert(err, IsNil)

	chunks, err := oss.SplitFileByPartSize(fileName, cryptoContext.PartSize)
	c.Assert(err, IsNil)

	fd, err := os.Open(fileName)
	c.Assert(err, IsNil)
	defer fd.Close()

	var parts []oss.UploadPart
	for _, chunk := range chunks {
		fd.Seek(chunk.Offset, os.SEEK_SET)
		part, err := bucket.UploadPart(imur, fd, chunk.Size, chunk.Number, cryptoContext)
		c.Assert(err, IsNil)
		parts = append(parts, part)
	}

	_, err = bucket.CompleteMultipartUpload(imur, parts)
	c.Assert(err, IsNil)

	meta, err := bucket.GetObjectDetailedMeta(objectName)
	c.Assert(err, IsNil)

	c.Assert(meta.Get("X-Oss-Meta-My"), Equals, "myprop")
	c.Assert(meta.Get("X-Oss-Object-Type"), Equals, "Multipart")

	downfileName := "test-go-sdk-file-" + RandLowStr(5)
	err = bucket.GetObjectToFile(objectName, downfileName)
	c.Assert(err, IsNil)

	downFileMD5, err := GetFileMD5(downfileName)
	c.Assert(err, IsNil)
	c.Assert(downFileMD5, Equals, srcMD5)

	os.Remove(downfileName)
	ForceDeleteBucket(client, bucketName, c)
}

func (s *OssCryptoBucketSuite) TestMultipartUploadFromFile(c *C) {
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
	fileName := "../../sample/BingWallpaper-2015-11-07.jpg"

	srcMD5, err := GetFileMD5(fileName)
	c.Assert(err, IsNil)

	fileInfo, err := os.Stat(fileName)
	dataSize := fileInfo.Size()
	c.Assert(err, IsNil)

	options := []oss.Option{oss.Meta("my", "myprop")}
	var cryptoContext PartCryptoContext
	cryptoContext.DataSize = dataSize
	cryptoContext.PartSize = (dataSize / 16 / 3) * 16
	imur, err := bucket.InitiateMultipartUpload(objectName, &cryptoContext, options...)
	c.Assert(err, IsNil)

	chunks, err := oss.SplitFileByPartSize(fileName, cryptoContext.PartSize)
	c.Assert(err, IsNil)

	var parts []oss.UploadPart
	for _, chunk := range chunks {
		part, err := bucket.UploadPartFromFile(imur, fileName, chunk.Offset, chunk.Size, chunk.Number, cryptoContext)
		c.Assert(err, IsNil)
		parts = append(parts, part)
	}

	_, err = bucket.CompleteMultipartUpload(imur, parts)
	c.Assert(err, IsNil)

	meta, err := bucket.GetObjectDetailedMeta(objectName)
	c.Assert(err, IsNil)

	c.Assert(meta.Get("X-Oss-Meta-My"), Equals, "myprop")
	c.Assert(meta.Get("X-Oss-Object-Type"), Equals, "Multipart")

	downfileName := "test-go-sdk-file-" + RandLowStr(5) + ".jpg"
	err = bucket.GetObjectToFile(objectName, downfileName)
	c.Assert(err, IsNil)

	downFileMD5, err := GetFileMD5(downfileName)
	c.Assert(err, IsNil)
	c.Assert(downFileMD5, Equals, srcMD5)

	os.Remove(downfileName)
	ForceDeleteBucket(client, bucketName, c)
}

func (s *OssCryptoBucketSuite) TestMultipartUploadFromSmallSizeFile(c *C) {
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
	fileName := "oss-go-sdk-test-file-" + RandStr(5)
	fo, err := os.Create(fileName)
	c.Assert(err, IsNil)
	_, err = fo.Write([]byte("123"))
	c.Assert(err, IsNil)
	fo.Close()

	srcMD5, err := GetFileMD5(fileName)
	c.Assert(err, IsNil)

	fileInfo, err := os.Stat(fileName)
	dataSize := fileInfo.Size()
	c.Assert(err, IsNil)

	options := []oss.Option{oss.Meta("my", "myprop")}
	var cryptoContext PartCryptoContext
	cryptoContext.DataSize = dataSize
	cryptoContext.PartSize = 16
	imur, err := bucket.InitiateMultipartUpload(objectName, &cryptoContext, options...)
	c.Assert(err, IsNil)

	chunks, err := oss.SplitFileByPartSize(fileName, cryptoContext.PartSize)
	c.Assert(err, IsNil)

	var parts []oss.UploadPart
	for _, chunk := range chunks {
		part, err := bucket.UploadPartFromFile(imur, fileName, chunk.Offset, chunk.Size, chunk.Number, cryptoContext)
		c.Assert(err, IsNil)
		parts = append(parts, part)
	}

	_, err = bucket.CompleteMultipartUpload(imur, parts)
	c.Assert(err, IsNil)

	meta, err := bucket.GetObjectDetailedMeta(objectName)
	c.Assert(err, IsNil)

	c.Assert(meta.Get("X-Oss-Meta-My"), Equals, "myprop")
	c.Assert(meta.Get("X-Oss-Object-Type"), Equals, "Multipart")

	downfileName := "test-go-sdk-file-" + RandLowStr(5) + ".jpg"
	err = bucket.GetObjectToFile(objectName, downfileName)
	c.Assert(err, IsNil)

	downFileMD5, err := GetFileMD5(downfileName)
	c.Assert(err, IsNil)
	c.Assert(downFileMD5, Equals, srcMD5)

	os.Remove(fileName)
	os.Remove(downfileName)
	ForceDeleteBucket(client, bucketName, c)
}

func (s *OssCryptoBucketSuite) TestListUploadedPartsNormal(c *C) {
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
	fileName := "../../sample/BingWallpaper-2015-11-07.jpg"

	srcMD5, err := GetFileMD5(fileName)
	c.Assert(err, IsNil)

	fileInfo, err := os.Stat(fileName)
	dataSize := fileInfo.Size()
	c.Assert(err, IsNil)

	// Upload
	var cryptoContext PartCryptoContext
	cryptoContext.DataSize = dataSize
	cryptoContext.PartSize = (dataSize / 16 / 3) * 16
	imurUpload, err := bucket.InitiateMultipartUpload(objectName, &cryptoContext)
	chunks, err := oss.SplitFileByPartSize(fileName, cryptoContext.PartSize)
	c.Assert(err, IsNil)

	var partsUpload []oss.UploadPart
	for _, chunk := range chunks {
		part, err := bucket.UploadPartFromFile(imurUpload, fileName, chunk.Offset, chunk.Size, (int)(chunk.Number), cryptoContext)
		c.Assert(err, IsNil)
		partsUpload = append(partsUpload, part)
	}

	// List
	lupr, err := bucket.ListUploadedParts(imurUpload)
	c.Assert(err, IsNil)
	c.Assert(len(lupr.UploadedParts), Equals, len(chunks))

	// Complete
	_, err = bucket.CompleteMultipartUpload(imurUpload, partsUpload)
	c.Assert(err, IsNil)

	// Download
	downfileName := "test-go-sdk-file-" + RandLowStr(5) + ".jpg"
	err = bucket.GetObjectToFile(objectName, downfileName)

	downFileMD5, err := GetFileMD5(downfileName)
	c.Assert(err, IsNil)
	c.Assert(downFileMD5, Equals, srcMD5)

	os.Remove(downfileName)
	ForceDeleteBucket(client, bucketName, c)
}

func (s *OssCryptoBucketSuite) TestListUploadedPartsComplete(c *C) {
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
	fileName := "../../sample/BingWallpaper-2015-11-07.jpg"

	srcMD5, err := GetFileMD5(fileName)
	c.Assert(err, IsNil)

	fileInfo, err := os.Stat(fileName)
	dataSize := fileInfo.Size()
	c.Assert(err, IsNil)

	// Upload
	var cryptoContext PartCryptoContext
	cryptoContext.DataSize = dataSize
	cryptoContext.PartSize = (dataSize / 16 / 3) * 16
	imurUpload, err := bucket.InitiateMultipartUpload(objectName, &cryptoContext)
	chunks, err := oss.SplitFileByPartSize(fileName, cryptoContext.PartSize)
	c.Assert(err, IsNil)

	var partsUpload []oss.UploadPart
	i := 0
	// upload excepted the last part
	for ; i < len(chunks)-1; i++ {
		part, err := bucket.UploadPartFromFile(imurUpload, fileName, chunks[i].Offset, chunks[i].Size, (int)(chunks[i].Number), cryptoContext)
		c.Assert(err, IsNil)
		partsUpload = append(partsUpload, part)
	}

	// List
	lupr, err := bucket.ListUploadedParts(imurUpload)
	c.Assert(err, IsNil)
	c.Assert(len(lupr.UploadedParts), Equals, len(chunks)-1)

	lmur, err := bucket.ListMultipartUploads()
	c.Assert(err, IsNil)
	c.Assert(len(lmur.Uploads), Equals, 1)

	// upload the last part with list part result
	part, err := bucket.UploadPartFromFile(imurUpload, fileName, chunks[i].Offset, chunks[i].Size, (int)(chunks[i].Number), cryptoContext)
	c.Assert(err, IsNil)
	partsUpload = append(partsUpload, part)

	// Complete
	_, err = bucket.CompleteMultipartUpload(imurUpload, partsUpload)
	c.Assert(err, IsNil)

	// Download
	downfileName := "test-go-sdk-file-" + RandLowStr(5) + ".jpg"
	err = bucket.GetObjectToFile(objectName, downfileName)

	downFileMD5, err := GetFileMD5(downfileName)
	c.Assert(err, IsNil)
	c.Assert(downFileMD5, Equals, srcMD5)

	os.Remove(downfileName)
	ForceDeleteBucket(client, bucketName, c)
}

func (s *OssCryptoBucketSuite) TestListUploadedPartsAbortUseInit(c *C) {
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
	fileName := "../../sample/BingWallpaper-2015-11-07.jpg"

	fileInfo, err := os.Stat(fileName)
	dataSize := fileInfo.Size()
	c.Assert(err, IsNil)

	// Upload
	var cryptoContext PartCryptoContext
	cryptoContext.DataSize = dataSize
	cryptoContext.PartSize = (dataSize / 16 / 3) * 16
	imurUpload, err := bucket.InitiateMultipartUpload(objectName, &cryptoContext)
	chunks, err := oss.SplitFileByPartSize(fileName, cryptoContext.PartSize)
	c.Assert(err, IsNil)
	c.Assert(cryptoContext.Valid(), Equals, true)

	var partsUpload []oss.UploadPart
	i := 0
	// upload excepted the last part
	for ; i < len(chunks)-1; i++ {
		part, err := bucket.UploadPartFromFile(imurUpload, fileName, chunks[i].Offset, chunks[i].Size, (int)(chunks[i].Number), cryptoContext)
		c.Assert(err, IsNil)
		partsUpload = append(partsUpload, part)
	}

	// List
	lupr, err := bucket.ListUploadedParts(imurUpload)
	c.Assert(err, IsNil)
	c.Assert(len(lupr.UploadedParts), Equals, len(chunks)-1)

	lmur, err := bucket.ListMultipartUploads()
	c.Assert(err, IsNil)
	c.Assert(len(lmur.Uploads), Equals, 1)

	// abort upload
	err = bucket.AbortMultipartUpload(imurUpload)
	c.Assert(err, IsNil)

	// list again
	lupr, err = bucket.ListUploadedParts(imurUpload)
	c.Assert(err, NotNil)

	ForceDeleteBucket(client, bucketName, c)
}

func (s *OssCryptoBucketSuite) TestListUploadedPartsAbortUseList(c *C) {
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
	fileName := "../../sample/BingWallpaper-2015-11-07.jpg"

	fileInfo, err := os.Stat(fileName)
	dataSize := fileInfo.Size()
	c.Assert(err, IsNil)

	// Upload
	var cryptoContext PartCryptoContext
	cryptoContext.DataSize = dataSize
	cryptoContext.PartSize = (dataSize / 16 / 3) * 16
	imurUpload, err := bucket.InitiateMultipartUpload(objectName, &cryptoContext)
	chunks, err := oss.SplitFileByPartSize(fileName, cryptoContext.PartSize)
	c.Assert(err, IsNil)

	var partsUpload []oss.UploadPart
	i := 0
	// upload excepted the last part
	for ; i < len(chunks)-1; i++ {
		part, err := bucket.UploadPartFromFile(imurUpload, fileName, chunks[i].Offset, chunks[i].Size, (int)(chunks[i].Number), cryptoContext)
		c.Assert(err, IsNil)
		partsUpload = append(partsUpload, part)
	}

	// List
	lupr, err := bucket.ListUploadedParts(imurUpload)
	c.Assert(err, IsNil)
	c.Assert(len(lupr.UploadedParts), Equals, len(chunks)-1)

	// abort upload
	err = bucket.AbortMultipartUpload(imurUpload)
	c.Assert(err, IsNil)

	// list again
	lupr, err = bucket.ListUploadedParts(imurUpload)
	c.Assert(err, NotNil)

	ForceDeleteBucket(client, bucketName, c)
}

func (s *OssCryptoBucketSuite) TestInitiateMultipartUpload(c *C) {
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
	context := RandStr(ivSize * 1024 * 10)
	fileName := "test-go-sdk-file-" + RandStr(5)

	err = ioutil.WriteFile(fileName, []byte(context), 0666)
	c.Assert(err, IsNil)

	fileInfo, err := os.Stat(fileName)
	dataSize := fileInfo.Size()
	c.Assert(err, IsNil)

	var cryptoContext PartCryptoContext
	cryptoContext.DataSize = dataSize
	cryptoContext.PartSize = ivSize * 1024
	imurUpload, err := bucket.InitiateMultipartUpload(objectName, &cryptoContext)
	c.Assert(err, IsNil)

	err = bucket.AbortMultipartUpload(imurUpload)
	c.Assert(err, IsNil)

	cryptoContext.DataSize = dataSize
	cryptoContext.PartSize = ivSize / 2
	imurUpload, err = bucket.InitiateMultipartUpload(objectName, &cryptoContext)
	c.Assert(err, NotNil)

	os.Remove(fileName)

	ForceDeleteBucket(client, bucketName, c)
}

func (s *OssCryptoBucketSuite) TestUploadPartError(c *C) {
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
	context := RandStr(ivSize * 1024 * 10)
	fileName := "test-go-sdk-file-" + RandStr(5)

	err = ioutil.WriteFile(fileName, []byte(context), 0666)
	c.Assert(err, IsNil)

	fileInfo, err := os.Stat(fileName)
	dataSize := fileInfo.Size()
	c.Assert(err, IsNil)

	var cryptoContext PartCryptoContext
	cryptoContext.DataSize = dataSize
	cryptoContext.PartSize = ivSize * 1024
	imur, err := bucket.InitiateMultipartUpload(objectName, &cryptoContext)
	c.Assert(err, IsNil)

	chunks, err := oss.SplitFileByPartSize(fileName, cryptoContext.PartSize)
	c.Assert(err, IsNil)

	fd, err := os.Open(fileName)
	c.Assert(err, IsNil)
	for _, chunk := range chunks {
		fd.Seek(chunk.Offset, os.SEEK_SET)
		_, err := bucket.UploadPart(imur, fd, chunk.Size+1, chunk.Number, cryptoContext)
		c.Assert(err, NotNil)
	}

	for _, chunk := range chunks {
		fd.Seek(chunk.Offset, os.SEEK_SET)
		_, err := bucket.UploadPart(imur, fd, chunk.Size, 0, cryptoContext)
		c.Assert(err, NotNil)
	}
	fd.Close()

	err = bucket.AbortMultipartUpload(imur)
	c.Assert(err, IsNil)
	os.Remove(fileName)
	ForceDeleteBucket(client, bucketName, c)
}

func (s *OssCryptoBucketSuite) TestUploadPartFromFileError(c *C) {
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
	context := RandStr(ivSize * 1024 * 10)
	fileName := "test-go-sdk-file-" + RandStr(5)

	err = ioutil.WriteFile(fileName, []byte(context), 0666)
	c.Assert(err, IsNil)

	fileInfo, err := os.Stat(fileName)
	dataSize := fileInfo.Size()
	c.Assert(err, IsNil)

	var cryptoContext PartCryptoContext
	cryptoContext.DataSize = dataSize
	cryptoContext.PartSize = ivSize * 1024
	imur, err := bucket.InitiateMultipartUpload(objectName, &cryptoContext)
	c.Assert(err, IsNil)

	chunks, err := oss.SplitFileByPartSize(fileName, cryptoContext.PartSize)
	c.Assert(err, IsNil)

	for _, chunk := range chunks {
		_, err := bucket.UploadPartFromFile(imur, fileName+".test", chunk.Offset, chunk.Size, chunk.Number, cryptoContext)
		c.Assert(err, NotNil)
	}

	_, err = bucket.UploadPartFromFile(imur, fileName+".test", chunks[0].Offset, chunks[0].Size+1, chunks[0].Number, cryptoContext)
	c.Assert(err, NotNil)

	_, err = bucket.UploadPartFromFile(imur, fileName+".test", chunks[0].Offset, chunks[0].Size, 0, cryptoContext)
	c.Assert(err, NotNil)

	err = bucket.AbortMultipartUpload(imur)
	c.Assert(err, IsNil)
	os.Remove(fileName)
	ForceDeleteBucket(client, bucketName, c)
}

func (s *OssCryptoBucketSuite) TestUploadPartCopyError(c *C) {
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
	context := RandStr(ivSize * 1024 * 10)
	fileName := "test-go-sdk-file-" + RandStr(5)

	err = ioutil.WriteFile(fileName, []byte(context), 0666)
	c.Assert(err, IsNil)

	fileInfo, err := os.Stat(fileName)
	dataSize := fileInfo.Size()
	c.Assert(err, IsNil)

	var cryptoContext PartCryptoContext
	cryptoContext.DataSize = dataSize
	cryptoContext.PartSize = ivSize * 1024
	imur, err := bucket.InitiateMultipartUpload(objectName, &cryptoContext)
	c.Assert(err, IsNil)

	chunks, err := oss.SplitFileByPartSize(fileName, cryptoContext.PartSize)
	c.Assert(err, IsNil)

	_, err = bucket.UploadPartCopy(imur, bucketName, objectName, 0, chunks[0].Size, 1, cryptoContext)
	c.Assert(err, NotNil)

	err = bucket.AbortMultipartUpload(imur)
	c.Assert(err, IsNil)
	os.Remove(fileName)
	ForceDeleteBucket(client, bucketName, c)
}

func (s *OssCryptoBucketSuite) TestMultipartUploadFromFileError(c *C) {
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
	fileName := "../../sample/BingWallpaper-2015-11-07.jpg"

	fileInfo, err := os.Stat(fileName)
	dataSize := fileInfo.Size()
	c.Assert(err, IsNil)

	options := []oss.Option{oss.Meta("my", "myprop")}
	var cryptoContext PartCryptoContext
	cryptoContext.DataSize = dataSize
	cryptoContext.PartSize = -1
	_, err = bucket.InitiateMultipartUpload(objectName, nil, options...)
	c.Assert(err, NotNil)

	_, err = bucket.InitiateMultipartUpload(objectName, &cryptoContext, options...)
	c.Assert(err, NotNil)

	cryptoContext.PartSize = (dataSize / 16 / 3) * 16
	imur, err := bucket.InitiateMultipartUpload(objectName, &cryptoContext, options...)
	c.Assert(err, IsNil)

	chunks, err := oss.SplitFileByPartSize(fileName, cryptoContext.PartSize)
	c.Assert(err, IsNil)

	bakCC := cryptoContext.ContentCipher
	cryptoContext.ContentCipher = nil

	i := 0
	// upload excepted the last part
	for ; i < len(chunks); i++ {
		_, err = bucket.UploadPartFromFile(imur, fileName, chunks[i].Offset, chunks[i].Size, (int)(chunks[i].Number), cryptoContext)
		c.Assert(err, NotNil)

	}

	i = 0
	cryptoContext.ContentCipher = bakCC
	cryptoContext.PartSize -= 1
	for ; i < len(chunks); i++ {
		_, err = bucket.UploadPartFromFile(imur, fileName, chunks[i].Offset, chunks[i].Size, (int)(chunks[i].Number), cryptoContext)
		c.Assert(err, NotNil)
	}

	ForceDeleteBucket(client, bucketName, c)
}

func (s *OssCryptoBucketSuite) TestMultipartUploadPartError(c *C) {
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
	fileName := "../../sample/BingWallpaper-2015-11-07.jpg"

	fileInfo, err := os.Stat(fileName)
	dataSize := fileInfo.Size()
	c.Assert(err, IsNil)

	options := []oss.Option{oss.Meta("my", "myprop")}
	var cryptoContext PartCryptoContext
	cryptoContext.DataSize = dataSize
	cryptoContext.PartSize = (dataSize / 16 / 3) * 16
	imur, err := bucket.InitiateMultipartUpload(objectName, &cryptoContext, options...)
	c.Assert(err, IsNil)

	chunks, err := oss.SplitFileByPartSize(fileName, cryptoContext.PartSize)
	c.Assert(err, IsNil)

	bakCC := cryptoContext.ContentCipher
	cryptoContext.ContentCipher = nil

	fd, err := os.Open(fileName)
	c.Assert(err, IsNil)

	for _, chunk := range chunks {
		fd.Seek(chunk.Offset, os.SEEK_SET)
		_, err = bucket.UploadPart(imur, fd, chunk.Size, chunk.Number, cryptoContext)
		c.Assert(err, NotNil)
	}

	cryptoContext.ContentCipher = bakCC
	cryptoContext.PartSize -= 1
	for _, chunk := range chunks {
		_, err = bucket.UploadPart(imur, fd, chunk.Size, chunk.Number, cryptoContext)
		c.Assert(err, NotNil)
	}

	ForceDeleteBucket(client, bucketName, c)
}
