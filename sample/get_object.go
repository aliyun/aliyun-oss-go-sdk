package sample

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strconv"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

// GetObjectSample Get Object Sample
func GetObjectSample() {
	// 创建Bucket
	bucket, err := GetTestBucket(bucketName)
	if err != nil {
		HandleError(err)
	}

	// 上传对象
	err = bucket.PutObjectFromFile(objectKey, localFile)
	if err != nil {
		HandleError(err)
	}

	// 场景1：下载object存储到ReadCloser，注意需要Close
	body, err := bucket.GetObject(objectKey)
	if err != nil {
		HandleError(err)
	}
	data, err := ioutil.ReadAll(body)
	body.Close()
	if err != nil {
		HandleError(err)
	}
	data = data // use data

	// 场景2：下载object存储到bytes数组，适合小对象
	buf := new(bytes.Buffer)
	body, err = bucket.GetObject(objectKey)
	if err != nil {
		HandleError(err)
	}
	io.Copy(buf, body)
	body.Close()

	// 场景3：下载object存储到本地文件，用户打开文件
	fd, err := os.OpenFile("mynewfile-1.jpg", os.O_WRONLY|os.O_CREATE, 0660)
	if err != nil {
		HandleError(err)
	}
	defer fd.Close()

	body, err = bucket.GetObject(objectKey)
	if err != nil {
		HandleError(err)
	}
	io.Copy(fd, body)
	body.Close()

	// 场景4：下载object存储到本地文件
	err = bucket.GetObjectToFile(objectKey, "mynewfile-2.jpg")
	if err != nil {
		HandleError(err)
	}

	// 场景5：满足约束条件下载，否则返回错误。GetObjectToFile具有相同功能。
	// 修改时间，约束条件满足，执行下载
	body, err = bucket.GetObject(objectKey, oss.IfModifiedSince(pastDate))
	if err != nil {
		HandleError(err)
	}
	body.Close()
	// 修改时间，约束条件不满足，不执行下载
	_, err = bucket.GetObject(objectKey, oss.IfUnmodifiedSince(pastDate))
	if err == nil {
		HandleError(err)
	}

	meta, err := bucket.GetObjectDetailedMeta(objectKey)
	if err != nil {
		HandleError(err)
	}
	md5 := meta.Get(oss.HTTPHeaderEtag)
	// 校验内容，约束条件满足，执行下载
	body, err = bucket.GetObject(objectKey, oss.IfMatch(md5))
	if err != nil {
		HandleError(err)
	}
	body.Close()

	// 校验内容，约束条件不满足，不执行下载
	body, err = bucket.GetObject(objectKey, oss.IfNoneMatch(md5))
	if err == nil {
		HandleError(err)
	}

	// 场景6：指定value的开始结束位置下载object，可以实现断点下载。GetObjectToFile具有相同功能。
	meta, err = bucket.GetObjectDetailedMeta(objectKey)
	if err != nil {
		HandleError(err)
	}
	fmt.Println("Object Meta:", meta[oss.HTTPHeaderContentLength])

	var partSize int64 = 100 * 1024
	objectSize, err := strconv.ParseInt(meta.Get(oss.HTTPHeaderContentLength), 10, 0)
	fd, err = os.OpenFile("myfile.jpg", os.O_WRONLY|os.O_CREATE, 0660)
	if err != nil {
		HandleError(err)
	}
	defer fd.Close()

	for i := int64(0); i < objectSize; i += partSize {
		option := oss.Range(i, oss.GetPartEnd(i, objectSize, partSize))
		body, err := bucket.GetObject(objectKey, option)
		if err != nil {
			HandleError(err)
		}
		io.Copy(fd, body)
		body.Close()
	}

	// 场景7：内容进行 GZIP压缩传输的用户。GetObject/GetObjectToWriter具有相同功能。
	err = bucket.PutObjectFromFile(objectKey, htmlLocalFile)
	if err != nil {
		HandleError(err)
	}

	err = bucket.GetObjectToFile(objectKey, "myhtml.gzip", oss.AcceptEncoding("gzip"))
	if err != nil {
		HandleError(err)
	}

	// 删除object和bucket
	err = DeleteTestBucketAndObject(bucketName)
	if err != nil {
		HandleError(err)
	}

	fmt.Println("GetObjectSample completed")
}
