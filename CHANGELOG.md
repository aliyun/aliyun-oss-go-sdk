# ChangeLog - Aliyun OSS SDK for Go

## 版本号：v2.2.10 日期：2023-10-30
### 变更内容
- 增加：support return callback body.
- 增加：support response header api
- 增加：add region field in listBuckets
- 增加：add ResponseVary field in CORSXML

## 版本号：v2.2.9 日期：2023-08-25
### 变更内容
- 增加：support force path style option.
- 增加：support context.Context option.
- 修改：remove LifecycleFilterNot.Prefix omitempty attribute.


## 版本号：v2.2.8 日期：2023-07-31
### 变更内容
- 增加：support EnvironmentVariableCredentialsProvider
- 增加：support describe regions api.
- 增加：support create bucket with server encryption parameters.
- 增加：support referer black list.
- 增加：support aysnc process object api.
- 增加：support ObjectSizeGreaterThan and ObjectSizeLessThan in lifecycle rule.
- 增加：add DeepColdArchive storage class.
- 修复：fix bug.

## 版本号：v2.2.7 日期：2023-03-23
### 变更内容
- 增加：support get info form EC & x-oss-err.
- 增加：support bucket replication time control api.
- 增加：support bucket style api.
- 增加：support list bucket cname api.
- 增加：support bucket resource group api.
- 修复：do not use uname -* cmd to get platform information.
- 修复：call rand.Seed only once.

## 版本号：v2.2.6 日期：2022-11-16
### 变更内容
- 增加：the object name cannot be empty in object's apis.
- 增加：support access monitor api.
- 修复：fix GetBucketStat bug.
- 增加：lifecycle rule supports filter configuration.
- 增加：support deleting the specified bucket tags.
- 修复：can't delete objects where the keys contain special characters.

## 版本号：v2.2.5 日期：2022-08-19
### 变更内容
- 增加：add meta data indexing api
- 删除：remove github.com/baiyubin/aliyun-sts-go-sdk/sts deps.
- 修改：remove chartset info in text/* mime type.
- 增加：add restore info in listObjects/listObjectVersions
- 增加：add x-oss-ac-* into subresource list.
- 修改：fix select object bug.
- 增加：getBucketStat api returns more info
- 增加：support X-Oss-Notification header in CompleteMultipartUpload api.

## 版本号：v2.2.4 日期：2022-05-25
### 变更内容
- 增加：add cname api
- 增加：add inventory api for xml config


# ChangeLog - Aliyun OSS SDK for Go
## 版本号：v2.2.3 日期：2022-05-13
### 变更内容
- 增加：support cloud-box
- 增加：support v4 signature


# ChangeLog - Aliyun OSS SDK for Go
## 版本号：v2.2.2 日期：2022-03-24
### 变更内容
- 增加：add GetBucketCORSXml,SetBucketCORSXml,GetBucketLifecycleXml


# ChangeLog - Aliyun OSS SDK for Go
## 版本号：v2.2.1 日期：2022-02-18
### 变更内容
- 增加：对http response status code进行符合标准规范的判断


# ChangeLog - Aliyun OSS SDK for Go
## 版本号：v2.2.0 日期：2021-11-08
### 变更内容
- 增加：增加CreateBucketXml接口


# ChangeLog - Aliyun OSS SDK for Go
## 版本号：v2.1.11 日期：2021-08-26
### 变更内容
- 增加：增加cname查询接口
- 增加：增加SetBucketLifecycleXml接口

# ChangeLog - Aliyun OSS SDK for Go
## 版本号：v2.1.10 日期：2021-08-05
### 变更内容
- 增加：支持限速下载
- 增加：增加同步边管理接口

# ChangeLog - Aliyun OSS SDK for Go
## 版本号：v2.1.9 日期：2021-07-09
### 变更内容
- 增加：支持跳过服务端证书校验
- 增加：支持regionList参数

# ChangeLog - Aliyun OSS SDK for Go
## 版本号：v2.1.8 日期：2021-04-09
### 变更内容
- 增加：支持传输加速设置

# ChangeLog - Aliyun OSS SDK for Go
## 版本号：v2.1.7 日期：2021-03-24
### 变更内容
- 增加：并行上传part支持设置md5以及hash context
- 增加：GetBucketWebsiteXml

# ChangeLog - Aliyun OSS SDK for Go
## 版本号：v2.1.6 日期：2021-01-13
### 变更内容
- 增加：增加worm接口

# ChangeLog - Aliyun OSS SDK for Go
## 版本号：v2.1.5 日期：2020-11-19
### 变更内容
- 增加：增加ListObjectsV2接口
- 增加: 增加RestoreObjectXML接口

# ChangeLog - Aliyun OSS SDK for Go
## 版本号：v2.1.4 日期：2020-07-24
### 变更内容
- 修复：lifecycle配置支持输入LifecycleVersionTransition数组

# ChangeLog - Aliyun OSS SDK for Go
## 版本号：v2.1.3 日期：2020-07-10
### 变更内容
- 修复：lifecycle支持冷归档(ColdArchive)


# ChangeLog - Aliyun OSS SDK for Go
## 版本号：v2.1.2 日期：2020-06-19
### 变更内容
- 增加：支持禁止http跳转功能(go1.7.0版本及以上)

# ChangeLog - Aliyun OSS SDK for Go
## 版本号：v2.1.1 日期：2020-06-04
### 变更内容
- 增加：支持国密byok
- 增加：支持异步任务的设置和读取

# ChangeLog - Aliyun OSS SDK for Go
## 版本号：v2.1.0 日期：2020-04-21
### 变更内容
- 增加：支持客户端加密、清单、冷归档功能
- 增加：tcp连接增加keepalive心跳选项
- 优化: 分块上传事件通知优化


# ChangeLog - Aliyun OSS SDK for Go
## 版本号：v2.0.8 日期：2020-04-09
### 变更内容
- 增加：支持用户传入自定义的header和param参数
- 增加：增加对X-Oss-Range-Behavior支持

# ChangeLog - Aliyun OSS SDK for Go
## 版本号：v2.0.7 日期：2020-03-11
### 变更内容
- 增加：支持OSS V2 签名
- 增加：增加SetBucketWebsiteXml接口,支持直接传入xml文件内容

# ChangeLog - Aliyun OSS SDK for Go
## 版本号：v2.0.6 日期：2020-02-15
### 变更内容
- 修复：CopyFile接口需要支持服务端加密功能

# ChangeLog - Aliyun OSS SDK for Go
## 版本号：v2.0.5 日期：2020-01-03
### 变更内容
- 增加：增加禁止同名覆盖选项X-Oss-Forbid-Overwrite
- 增加：增加分块上传参数sequential, 支持分块上传返回md5校验值

## 版本号：v2.0.4 日期：2019-11-13
### 变更内容
- 增加：SSR 对bucket 和 endpoint 做合法性校验，不符合要求要直接提示错误。
- 增加：select object 功能merge
- 增加：断点续传文件支持多版本
- 增加：lifecycle 支持多版本
- 修复：断点续传文件中的时间比较方式优化
- 修复：修复断点上传不支持服务端加密的bug


# ChangeLog - Aliyun OSS SDK for Go
## 版本号：v2.0.3 日期：2019-09-17
### 变更内容
- 修复：不支持分块上传归档object
- 增加：增加绑定客户端ip地址
- 增加: 增加更多的mime type类型


# ChangeLog - Aliyun OSS SDK for Go
## 版本号：v2.0.2 日期：2019-08-06
### 变更内容
- 修复：proxy代理不支持https请求

# ChangeLog - Aliyun OSS SDK for Go
## 版本号：v2.0.1 日期：2019-07-11
### 变更内容
- 增加：增加qos相关api
- 增加：增加payment相关api
- 增加：增加自定义获取AccessKeyID、AccessKeySecret、SecurityToken
- 增加: 增加http请求限速option


# ChangeLog - Aliyun OSS SDK for Go
## 版本号：v2.0.0 日期：2019-06-18
### 变更内容
- 增加：增加各个接口对versioning的支持
- 增加：增加设置、查询、删除bucket policy接口
- 增加: 增加设置website详细配置接口: SetBucketWebsiteDetail
- 增加: 增加Bucket OptionsMethod 接口


# ChangeLog - Aliyun OSS SDK for Go
## 版本号：v1.9.8 日期：2019-05-25
### 变更内容
- 增加：增加设置、查询、删除bucket tagging接口

# ChangeLog - Aliyun OSS SDK for Go
## 版本号：v1.9.7 日期：2019-05-22
### 变更内容
- 增加：增加设置、查询、删除object tagging接口
- 增加：增加设置、查询、删除bucket encryption接口
- 增加：增加获取bucket统计信息接口

## 版本号：v1.9.6 日期：2019-04-15
### 变更内容
- 变更：扩展lifecycle功能，提供设置AbortMutipartUpload和Transitions两种规则的生命周期管理的处理
- 修复：测试用例BucketName使用固定前缀+随机的字符串
- 修复：测试用例ObjectName使用固定前缀+随机字符串
- 修复：测试用例有关bucket相关的异步操作，统一定义sleep时间
- 修复：测试集结束后，列出bucket内的所有对象并删除所有测试的对象
- 修复：测试集结束后，列出bucket内的所有未上传完成的分片并删除所有测试过程中产生的为上传完成的分片
- 修复：支持上传webp类型的对象时从对象的后缀名字自动解析对应的content-type并设置content-type
- 变更：增加在put/copy/append等接口时时设置对象的存储类型的sample
- 修复：sample示例中的配置项的值改为直接从环境变量读取

## 版本号：1.9.5 日期：2019-03-08
### 变更内容
- 变更：增加了限速上传功能

## 版本号：1.9.4 日期：2019-01-25
### 变更内容
- 修复：在开启日志后，如果接口返回错误readResponseBody函数被调用两次
- 变更：增加livechannel功能各个api接口

## 版本号：1.9.3 日期：2019-01-10
### 变更内容
- 修复：分片上传时传入partSize值不对是的提示信息不准确的问题
- 修复：仅仅在使用userAgent的时候初始化它的值
- 变更：添加ContentLanguage选项
- 变更：支持设置最大的空闲连接个数
- 变更：当配置的endpoint不对时，输出的错误信息将会打印出正确的endpoint
- 变更：支持ServerSideEncryptionKeyID选项，允许用户传入kms-id
- 变更：添加日志模块，支持设置日志级别

## 版本号：1.9.2 日期：2018-11-16
### 变更内容
- 变更：添加支持设置request Payer的option
- 变更：添加支持设置checkpoint目录的option
- 变更：getobjectmeta接口增加options参数，可以支持传入option选项
- 变更：listobjecs接口增加options参数，可以支持传入option选项
- 变更：listmultipartuploads接口增加options参数, 可以支持传入option选项
- 修复：解决调用接口返回出错时，且返回的http body为空时，打印错误消息不包含"request_id"的问题
- 变更：abortmultipartupload接口增加options参数, 可以支持传入option选项
- 变更：completemultipartupload接口增加options参数, 可以支持传入option选项

## 版本号：1.9.1 日期：2018-09-17
### 变更内容
 - 变更：支持ipv6
 - 变更：支持修改对象的存储类型
 - 修复：修改sample中GetBucketReferer方法名拼写错误
 - 修复：修复NopCloser在close的时候并不释放内存的内存泄漏问题
 - 变更：增加ProcessObject接口
 - 修复：修改图片处理接口参数拼写错误导致无法处理的bug
 - 修复：增加ListUploadedParts接口的options选项
 - 修复：增加Callback&CallbackVal选项，支持回调使用
 - 修复：GetObject接口返回Response，支持用户读取crc等返回值
 - 修复：当以压缩格式返回数据时，GetObject接口不校验crc

## 版本号：1.9.0 日期：2018-06-15
### 变更内容
 - 变更：国际化

## 版本号：1.8.0 日期：2017-12-12
### 变更内容
 - 变更：空闲链接关闭时间调整为50秒
 - 修复：修复临时账号使用SignURL的问题

## 版本号：1.7.0 日期：2017-09-25
### 变更内容
 - 增加：DownloadFile支持CRC校验
 - 增加：STS测试用例

## 版本号：1.6.0 日期：2017-09-01
### 变更内容
 - 修复：URL中特殊字符的编码问题
 - 变更：不再支持Golang 1.4
 
## 版本号：1.5.1 日期：2017-08-04
### 变更内容
 - 修复：SignURL中Key编码的问题
 - 修复：DownloadFile下载完成后rename失败的问题
 
## 版本号：1.5.0 日期：2017-07-25
### 变更内容
 - 增加：支持生成URL签名
 - 增加：GetObject支持ResponseContentType等选项
 - 修复：DownloadFile去除分片小于5GB的限制
 - 修复：AppendObject在appendPosition不正确时发生panic

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
 - 增加：支持CRC64校验
 - 增加：支持指定Useragent
 - 修复：计算MD5占用内存大的问题
 - 修复：CopyObject时Object名称没有URL编码的问题

## 版本号：1.1.0 日期：2016-08-09
### 变更内容
 - 增加：支持代理服务器

## 版本号：1.0.0 日期：2016-06-24
### 变更内容
 - 增加：断点分片复制接口Bucket.CopyFile
 - 增加：Bucket间复制接口Bucket.CopyObjectTo、Bucket.CopyObjectFrom
 - 增加：Client.GetBucketInfo接口
 - 增加：Bucket.UploadPartCopy支持Bucket间复制
 - 修复：断点上传、断点下载出错后，协程不退出的Bug
 - 删除：接口Bucket.CopyObjectToBucket
