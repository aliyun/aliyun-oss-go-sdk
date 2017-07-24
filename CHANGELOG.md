# ChangeLog - Aliyun OSS SDK for Go

## 版本号：1.4.0 日期：2017-05-23
### 变更内容
 - 增加：支持符号链接symlink
 - 增加：支持RestoreObject
 - 增加：CreateBucket支持StorageClass
 - 增加：支持范围读NormalizedRange
 - 修复：IsObjectExist使用GetObjectMeta实现

## 版本号：1.3.0 日期：2017-01-13
### 变更内容
 - 增加：上传下载支持进度条功能

## 版本号：1.2.3 日期：2016-12-28
### 变更内容
 - 修复：每次请求使用一个http.Client修改为共用http.Client

## 版本号：1.2.2 日期：2016-12-10
### 变更内容
 - 修复：GetObjectToFile/DownloadFile使用临时文件下载，成功后重命名成下载文件
 - 修复：新建的下载文件权限修改为0664

## 版本号：1.2.1 日期：2016-11-11
### 变更内容
 - 修复：只有当OSS返回x-oss-hash-crc64ecma头部时，才对上传的文件进行CRC64完整性校验

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
