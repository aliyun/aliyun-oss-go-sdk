# ChangeLog - Aliyun OSS SDK for Go

## 版本号：1.0.0 日期：2016-06-24
### 变更内容
 - 增加断点分片复制接口Bucket.CopyFile
 - 增加Bucket间复制接口Bucket.CopyObjectTo、Bucket.CopyObjectFrom
 - 增加Client.GetBucketInfo接口
 - Bucket.UploadPartCopy支持Bucket间复制
 - 修复断点上传、断点下载出错后，协程不退出的Bug
 - 去除接口Bucket.CopyObjectToBucket
