package sample

import (
	"fmt"
	"sync"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

// MultipartCopySample Multipart Copy Sample
func MultipartCopySample() {
	var objectSrc = "my-object-src"
	var objectDesc = "my-object-desc"

	// 创建Bucket
	bucket, err := GetTestBucket(bucketName)
	if err != nil {
		HandleError(err)
	}

	err = bucket.PutObjectFromFile(objectSrc, localFile)
	if err != nil {
		HandleError(err)
	}

	// 场景1：大文件分片拷贝，按照文件片大小分片
	chunks, err := oss.SplitFileByPartNum(localFile, 3)
	if err != nil {
		HandleError(err)
	}

	imur, err := bucket.InitiateMultipartUpload(objectDesc)
	if err != nil {
		HandleError(err)
	}

	parts := []oss.UploadPart{}
	for _, chunk := range chunks {
		part, err := bucket.UploadPartCopy(imur, objectSrc, chunk.Offset, chunk.Size,
			chunk.Number)
		if err != nil {
			HandleError(err)
		}
		parts = append(parts, part)
	}

	_, err = bucket.CompleteMultipartUpload(imur, parts)
	if err != nil {
		HandleError(err)
	}

	err = bucket.DeleteObject(objectDesc)
	if err != nil {
		HandleError(err)
	}

	// 场景2：大文件分片拷贝，按照指定文件片数
	chunks, err = oss.SplitFileByPartSize(localFile, 1024*100)
	if err != nil {
		HandleError(err)
	}

	imur, err = bucket.InitiateMultipartUpload(objectDesc)
	if err != nil {
		HandleError(err)
	}

	parts = []oss.UploadPart{}
	for _, chunk := range chunks {
		part, err := bucket.UploadPartCopy(imur, objectSrc, chunk.Offset, chunk.Size,
			chunk.Number)
		if err != nil {
			HandleError(err)
		}
		parts = append(parts, part)
	}

	_, err = bucket.CompleteMultipartUpload(imur, parts)
	if err != nil {
		HandleError(err)
	}

	err = bucket.DeleteObject(objectDesc)
	if err != nil {
		HandleError(err)
	}

	// 场景3：大文件分片拷贝，初始化时指定对象属性
	chunks, err = oss.SplitFileByPartNum(localFile, 3)
	if err != nil {
		HandleError(err)
	}

	imur, err = bucket.InitiateMultipartUpload(objectDesc, oss.Meta("myprop", "mypropval"))
	if err != nil {
		HandleError(err)
	}

	parts = []oss.UploadPart{}
	for _, chunk := range chunks {
		part, err := bucket.UploadPartCopy(imur, objectSrc, chunk.Offset, chunk.Size,
			chunk.Number)
		if err != nil {
			HandleError(err)
		}
		parts = append(parts, part)
	}

	_, err = bucket.CompleteMultipartUpload(imur, parts)
	if err != nil {
		HandleError(err)
	}

	err = bucket.DeleteObject(objectDesc)
	if err != nil {
		HandleError(err)
	}

	// 场景4：大文件分片拷贝，每个分片可以有线程/进程/机器独立完成，下面示例是每个线程拷贝一个分片
	partNum := 4
	chunks, err = oss.SplitFileByPartNum(localFile, partNum)
	if err != nil {
		HandleError(err)
	}

	imur, err = bucket.InitiateMultipartUpload(objectDesc)
	if err != nil {
		HandleError(err)
	}

	// 并发拷贝分片上传
	var waitgroup sync.WaitGroup
	var ps = make([]oss.UploadPart, partNum)
	for _, chunk := range chunks {
		waitgroup.Add(1)
		go func(chunk oss.FileChunk) {
			part, err := bucket.UploadPartCopy(imur, objectSrc, chunk.Offset, chunk.Size,
				chunk.Number)
			if err != nil {
				HandleError(err)
			}
			ps[chunk.Number-1] = part
			waitgroup.Done()
		}(chunk)
	}

	// 等待拷贝完成
	waitgroup.Wait()

	// 通知完成
	_, err = bucket.CompleteMultipartUpload(imur, ps)
	if err != nil {
		HandleError(err)
	}

	err = bucket.DeleteObject(objectDesc)
	if err != nil {
		HandleError(err)
	}

	// 场景5：大文件分片拷贝，对拷贝有约束条件，满足时候拷贝，不满足时报错
	chunks, err = oss.SplitFileByPartNum(localFile, 3)
	if err != nil {
		HandleError(err)
	}

	imur, err = bucket.InitiateMultipartUpload(objectDesc)
	if err != nil {
		HandleError(err)
	}

	parts = []oss.UploadPart{}
	for _, chunk := range chunks {
		constraint := oss.CopySourceIfMatch("InvalidETag")
		_, err := bucket.UploadPartCopy(imur, objectSrc, chunk.Offset, chunk.Size,
			chunk.Number, constraint)
		fmt.Println(err)
	}

	err = bucket.AbortMultipartUpload(imur)
	if err != nil {
		HandleError(err)
	}

	err = bucket.DeleteObject(objectDesc)
	if err != nil {
		HandleError(err)
	}

	// 场景6：大文件分片拷贝一部分后，中止上传，上传的数据将丢弃，UploadId也将无效
	chunks, err = oss.SplitFileByPartNum(localFile, 3)
	if err != nil {
		HandleError(err)
	}

	imur, err = bucket.InitiateMultipartUpload(objectDesc)
	if err != nil {
		HandleError(err)
	}

	parts = []oss.UploadPart{}
	for _, chunk := range chunks {
		part, err := bucket.UploadPartCopy(imur, objectSrc, chunk.Offset, chunk.Size,
			chunk.Number)
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

	fmt.Println("MultipartCopySample completed")
}
