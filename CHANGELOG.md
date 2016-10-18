# ChangeLog - Aliyun OSS SDK for Go

## 版本号：1.2.0 日期：2016-10-18
### 变更内容
 - 支持CRC64校验
 - 修复计算MD5占用内存大的问题
 - 修复CopyObject时Object名称没有URL编码的问题
 - 支持指定Useragent

## 版本号：1.1.0 日期：2016-08-09
### 变更内容
 - 支持代理服务器
 

## 版本号：1.0.0 日期：2016-06-24
### 变更内容
 - 增加断点分片复制接口Bucket.CopyFile
 - 增加Bucket间复制接口Bucket.CopyObjectTo、Bucket.CopyObjectFrom
 - 增加Client.GetBucketInfo接口
 - Bucket.UploadPartCopy支持Bucket间复制
 - 修复断点上传、断点下载出错后，协程不退出的Bug
 - 去除接口Bucket.CopyObjectToBucket
