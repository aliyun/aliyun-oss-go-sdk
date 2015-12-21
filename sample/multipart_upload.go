package sample

import (
	"fmt"
	"os"
	"sync"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

// MultipartUploadSample Multipart Upload Sample
func MultipartUploadSample() {
	// 创建Bucket
	bucket, err := GetTestBucket(bucketName)
	if err != nil {
		HandleError(err)
	}

	// 场景1：大文件分片上传，按照文件片大小分片
	chunks, err := oss.SplitFileByPartNum(localFile, 3)
	if err != nil {
		HandleError(err)
	}

	imur, err := bucket.InitiateMultipartUpload(objectKey)
	if err != nil {
		HandleError(err)
	}

	parts := []oss.UploadPart{}
	for _, chunk := range chunks {
		part, err := bucket.UploadPartFromFile(imur, localFile, chunk.Offset,
			chunk.Size, chunk.Number)
		if err != nil {
			HandleError(err)
		}
		parts = append(parts, part)
	}

	_, err = bucket.CompleteMultipartUpload(imur, parts)
	if err != nil {
		HandleError(err)
	}

	err = bucket.DeleteObject(objectKey)
	if err != nil {
		HandleError(err)
	}

	// 场景2：大文件分片上传，按照指定文件片数
	chunks, err = oss.SplitFileByPartSize(localFile, 1024*100)
	if err != nil {
		HandleError(err)
	}

	imur, err = bucket.InitiateMultipartUpload(objectKey)
	if err != nil {
		HandleError(err)
	}

	parts = []oss.UploadPart{}
	for _, chunk := range chunks {
		part, err := bucket.UploadPartFromFile(imur, localFile, chunk.Offset,
			chunk.Size, chunk.Number)
		if err != nil {
			HandleError(err)
		}
		parts = append(parts, part)
	}

	_, err = bucket.CompleteMultipartUpload(imur, parts)
	if err != nil {
		HandleError(err)
	}

	err = bucket.DeleteObject(objectKey)
	if err != nil {
		HandleError(err)
	}

	chunks = []oss.FileChunk{
		{Number: 1, Offset: 0 * 1024 * 1024, Size: 1024 * 1024},
		{Number: 2, Offset: 1 * 1024 * 1024, Size: 1024 * 1024},
		{Number: 3, Offset: 2 * 1024 * 1024, Size: 1024 * 1024},
	}

	// 创建3：大文件上传，您自己打开文件，传入句柄
	chunks, err = oss.SplitFileByPartNum(localFile, 3)
	if err != nil {
		HandleError(err)
	}

	fd, err := os.Open(localFile)
	if err != nil {
		HandleError(err)
	}
	defer fd.Close()

	imur, err = bucket.InitiateMultipartUpload(objectKey)
	if err != nil {
		HandleError(err)
	}

	parts = []oss.UploadPart{}
	for _, chunk := range chunks {
		fd.Seek(chunk.Offset, os.SEEK_SET)
		part, err := bucket.UploadPart(imur, fd, chunk.Size, chunk.Number)
		if err != nil {
			HandleError(err)
		}
		parts = append(parts, part)
	}

	_, err = bucket.CompleteMultipartUpload(imur, parts)
	if err != nil {
		HandleError(err)
	}

	err = bucket.DeleteObject(objectKey)
	if err != nil {
		HandleError(err)
	}

	// 场景4：大文件分片上传，初始化时指定对象属性
	chunks, err = oss.SplitFileByPartNum(localFile, 3)
	if err != nil {
		HandleError(err)
	}

	imur, err = bucket.InitiateMultipartUpload(objectKey, oss.Meta("myprop", "mypropval"))
	if err != nil {
		HandleError(err)
	}

	parts = []oss.UploadPart{}
	for _, chunk := range chunks {
		part, err := bucket.UploadPartFromFile(imur, localFile, chunk.Offset,
			chunk.Size, chunk.Number)
		if err != nil {
			HandleError(err)
		}
		parts = append(parts, part)
	}

	_, err = bucket.CompleteMultipartUpload(imur, parts)
	if err != nil {
		HandleError(err)
	}

	err = bucket.DeleteObject(objectKey)
	if err != nil {
		HandleError(err)
	}

	// 场景5：大文件分片上传，每个分片可以有线程/进程/机器独立完成，下面示例是每个线程上传一个分片
	partNum := 4
	chunks, err = oss.SplitFileByPartNum(localFile, partNum)
	if err != nil {
		HandleError(err)
	}

	imur, err = bucket.InitiateMultipartUpload(objectKey)
	if err != nil {
		HandleError(err)
	}

	// 并发上传分片
	var waitgroup sync.WaitGroup
	var ps = make([]oss.UploadPart, partNum)
	for _, chunk := range chunks {
		waitgroup.Add(1)
		go func(chunk oss.FileChunk) {
			part, err := bucket.UploadPartFromFile(imur, localFile, chunk.Offset,
				chunk.Size, chunk.Number)
			if err != nil {
				HandleError(err)
			}
			ps[chunk.Number-1] = part
			waitgroup.Done()
		}(chunk)
	}

	// 等待上传完成
	waitgroup.Wait()

	// 通知完成
	_, err = bucket.CompleteMultipartUpload(imur, ps)
	if err != nil {
		HandleError(err)
	}

	err = bucket.DeleteObject(objectKey)
	if err != nil {
		HandleError(err)
	}

	// 场景6：大文件分片上传一部分后，中止上传，上传的数据将丢弃，UploadId也将无效
	chunks, err = oss.SplitFileByPartNum(localFile, 3)
	if err != nil {
		HandleError(err)
	}

	imur, err = bucket.InitiateMultipartUpload(objectKey)
	if err != nil {
		HandleError(err)
	}

	parts = []oss.UploadPart{}
	for _, chunk := range chunks {
		part, err := bucket.UploadPartFromFile(imur, localFile, chunk.Offset,
			chunk.Size, chunk.Number)
		if err != nil {
			HandleError(err)
		}
		parts = append(parts, part)
	}

	err = bucket.AbortMultipartUpload(imur)
	if err != nil {
		HandleError(err)
	}

	// 删除object和bucket
	err = DeleteTestBucketAndObject(bucketName)
	if err != nil {
		HandleError(err)
	}

	fmt.Println("MultipartUploadSample completed")
}
