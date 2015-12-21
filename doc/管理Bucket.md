# 管理存储空间（Bucket）

存储空间（Bucket）是OSS上的命名空间，也是计费、权限控制、日志记录等高级功能的管理实体。

## 查看所有Bucket

使用`Client.ListBuckets`接口列出当前用户下的所有Bucket，用户还可以指
定`Prefix`等参数，列出Bucket名字为特定前缀的所有Bucket：

> 提示：
> 
> - ListBuckets的示例代码在`sample/list_buckets.go`。
>

```go
    import (
        "fmt"
        "github.com/aliyun/aliyun-oss-go-sdk/oss"
    )
  
    client, err := oss.New("Endpoint", "AccessKeyId", "AccessKeySecret")
    if err != nil {
        // HandleError(err)
    }
    
    // 列出Bucket，默认100条。
    lsRes, err := client.ListBuckets()
    if err != nil {
        // HandleError(err)
    }
    fmt.Println("buckets:", lsRes.Buckets)
    
    // 指定前缀筛选
    lsRes, err = client.ListBuckets(oss.Prefix("my-bucket"))
    if err != nil {
        // HandleError(err)
    }
    fmt.Println("buckets:", lsRes.Buckets)
```

## 创建Bucket

> 提示：
> 
> - CreateBucket的示例代码在`sample/create_bucket.go`。
>

使用`Client.CreateBucket`接口创建一个Bucket，用户需要指定Bucket的名字：
```go
    import "github.com/aliyun/aliyun-oss-go-sdk/oss"
    
    client, err := oss.New("Endpoint", "AccessKeyId", "AccessKeySecret")
    if err != nil {
        // HandleError(err)
    }
    
    err = client.CreateBucket("my-bucket")
    if err != nil {
        // HandleError(err)
    }
```
创建Bucket时不指定权限，使用默认权限oss.ACLPrivate。创建时用户可以指定Bucket的权限：
```go
    err = client.CreateBucket("my-bucket", oss.ACL(oss.ACLPublicRead))
    if err != nil {
        // HandleError(err)
    }
```

> 注意：
> 
> - Bucket的命名规范请查看[创建Bucket]({{doc/[2]Get-Started/快速开始.md}})
> - 由于存储空间的名字是全局唯一的，所以必须保证您的Bucket名字不与别人的重复

## 删除Bucket

使用`Client.DeleteBucket`接口删除一个Bucket，用户需要指定Bucket的名字：
```go
    import "github.com/aliyun/aliyun-oss-go-sdk/oss"
    
    client, err := oss.New("Endpoint", "AccessKeyId", "AccessKeySecret")
    if err != nil {
        // HandleError(err)
    }
    
    err = client.DeleteBucket("my-bucket")
    if err != nil {
        // HandleError(err)
    }
```

> 注意：
> 
> - 如果该Bucket下还有文件存在，则需要先删除所有文件才能删除Bucket
> - 如果该Bucket下还有未完成的上传请求，则需要通过`Bucket.ListMultipartUploads`和
>   `Bucket.AbortMultipartUpload`先取消那些请求才能删除Bucket。用法请参考
>   [API文档][sdk-api]

## 查看Bucket是否存在

用户可以通过`Client.IsBucketExist`接口查看当前用户的某个Bucket是否存在：
```go
    import "github.com/aliyun/aliyun-oss-go-sdk/oss"
    
    client, err := oss.New("Endpoint", "AccessKeyId", "AccessKeySecret")
    if err != nil {
        // HandleError(err)
    }
    
    isExist, err := client.IsBucketExist("my-bucket")
    if err != nil {
        // HandleError(err)
    }
```

## Bucket访问权限

用户可以设置Bucket的访问权限，允许或者禁止匿名用户对其内容进行读写。更
多关于访问权限的内容请参考[访问权限][bucket-acl]

> 提示：
> 
> - Bucket访问权限的示例代码`sample/bucket_acl.go`。
>

### 查看Bucket的访问权限

通过`Client.GetBucketACL`查看Bucket的ACL：
```go
    import "fmt"
    import "github.com/aliyun/aliyun-oss-go-sdk/oss"
    
    client, err := oss.New("Endpoint", "AccessKeyId", "AccessKeySecret")
    if err != nil {
        // HandleError(err)
    }

    aclRes, err := client.GetBucketACL("my-bucket")
    if err != nil {
        // HandleError(err)
    }
    fmt.Println("Bucket ACL:", aclRes.ACL)
```

### 设置Bucket的访问权限（ACL）

通过`Client.SetBucketACL`设置Bucket的ACL：
```go
    import "github.com/aliyun/aliyun-oss-go-sdk/oss"
    
    client, err := oss.New("Endpoint", "AccessKeyId", "AccessKeySecret")
    if err != nil {
        // HandleError(err)
    }
    
    err = client.SetBucketACL("my-bucket", oss.ACLPublicRead)
    if err != nil {
        // HandleError(err)
    }
```

> 提示：
> 
> - Bucket有三种权限私有读写、公共读私有写、公共读写，分布对应Go sdk的常量ACLPrivate、ACLPublicRead、ACLPublicReadWrite。
>

[sdk-api]: http://www.rubydoc.info/gems/aliyun-sdk/0.1.6
[oss-regions]: http://help.aliyun.com/document_detail/oss/user_guide/oss_concept/endpoint.html
[bucket-acl]: http://help.aliyun.com/document_detail/oss/user_guide/security_management/access_control.html
