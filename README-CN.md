# Aliyun OSS SDK for Go

[![GitHub version](https://badge.fury.io/gh/aliyun%2Faliyun-oss-go-sdk.svg)](https://badge.fury.io/gh/aliyun%2Faliyun-oss-go-sdk)
[![Build Status](https://travis-ci.org/aliyun/aliyun-oss-go-sdk.svg?branch=master)](https://travis-ci.org/aliyun/aliyun-oss-go-sdk)
[![Coverage Status](https://coveralls.io/repos/github/aliyun/aliyun-oss-go-sdk/badge.svg?branch=master)](https://coveralls.io/github/aliyun/aliyun-oss-go-sdk?branch=master)

## [README of English](https://github.com/aliyun/aliyun-oss-go-sdk/blob/master/README.md)

## 关于
> - 此Go SDK基于[阿里云对象存储服务](http://www.aliyun.com/product/oss/)官方API构建。
> - 阿里云对象存储（Object Storage Service，简称OSS），是阿里云对外提供的海量，安全，低成本，高可靠的云存储服务。
> - OSS适合存放任意文件类型，适合各种网站、开发企业及开发者使用。
> - 使用此SDK，用户可以方便地在任何应用、任何时间、任何地点上传，下载和管理数据。

## 版本
> - 当前版本：1.8.0

## 运行环境
> - Go 1.5及以上。

## 安装方法
### GitHub安装
> - 执行命令`go get github.com/aliyun/aliyun-oss-go-sdk/oss`获取远程代码包。
> - 在您的代码中使用`import "github.com/aliyun/aliyun-oss-go-sdk/oss"`引入OSS Go SDK的包。

## 快速使用
#### 获取存储空间列表（List Bucket）
```go
    client, err := oss.New("Endpoint", "AccessKeyId", "AccessKeySecret")
    if err != nil {
        // HandleError(err)
    }
    
    lsRes, err := client.ListBuckets()
    if err != nil {
        // HandleError(err)
    }
    
    for _, bucket := range lsRes.Buckets {
        fmt.Println("Buckets:", bucket.Name)
    }
```

#### 创建存储空间（Create Bucket）
```go
    client, err := oss.New("Endpoint", "AccessKeyId", "AccessKeySecret")
    if err != nil {
        // HandleError(err)
    }
    
    err = client.CreateBucket("my-bucket")
    if err != nil {
        // HandleError(err)
    }
```
    
#### 删除存储空间（Delete Bucket）
```go
    client, err := oss.New("Endpoint", "AccessKeyId", "AccessKeySecret")
    if err != nil {
        // HandleError(err)
    }
    
    err = client.DeleteBucket("my-bucket")
    if err != nil {
        // HandleError(err)
    }
```

#### 上传文件（Put Object）
```go
    client, err := oss.New("Endpoint", "AccessKeyId", "AccessKeySecret")
    if err != nil {
        // HandleError(err)
    }
    
    bucket, err := client.Bucket("my-bucket")
    if err != nil {
        // HandleError(err)
    }
    
    err = bucket.PutObjectFromFile("my-object", "LocalFile")
    if err != nil {
        // HandleError(err)
    }
```

#### 下载文件 (Get Object)
```go
    client, err := oss.New("Endpoint", "AccessKeyId", "AccessKeySecret")
    if err != nil {
        // HandleError(err)
    }
    
    bucket, err := client.Bucket("my-bucket")
    if err != nil {
        // HandleError(err)
    }
    
    err = bucket.GetObjectToFile("my-object", "LocalFile")
    if err != nil {
        // HandleError(err)
    }
```

#### 获取文件列表（List Objects）
```go
    client, err := oss.New("Endpoint", "AccessKeyId", "AccessKeySecret")
    if err != nil {
        // HandleError(err)
    }
    
    bucket, err := client.Bucket("my-bucket")
    if err != nil {
        // HandleError(err)
    }
    
    lsRes, err := bucket.ListObjects()
    if err != nil {
        // HandleError(err)
    }
    
    for _, object := range lsRes.Objects {
        fmt.Println("Objects:", object.Key)
    }
```
    
#### 删除文件(Delete Object)
```go
    client, err := oss.New("Endpoint", "AccessKeyId", "AccessKeySecret")
    if err != nil {
        // HandleError(err)
    }
    
    bucket, err := client.Bucket("my-bucket")
    if err != nil {
        // HandleError(err)
    }
    
    err = bucket.DeleteObject("my-object")
    if err != nil {
        // HandleError(err)
    }
```

#### 创建livechannel(Put Livechannel)
```go
    client, err := oss.New("Endpoint", "AccessKeyId", "AccessKeySecret")
    if err != nil {
        // HandleError(err)
    }
    
    bucket, err := client.Bucket("my-bucket")
    if err != nil {
        // HandleError(err)
    }
    
    config := oss.LiveChannelConfiguration{
		Description: "sample-for-livechannel", //livechannel的描述信息，最长128字节，非必须
		Status:      "enabled",                //指定livechannel的状态，非必须（默认值："enabled", 值只能是"enabled", "disabled"）
		Target: oss.LiveChannelTarget{
			Type:         "HLS",                          //指定转储的类型，只支持HLS, 必须
			FragDuration: 10,                             //指定每个ts文件的时长（单位：秒），取值范围为【1，100】的整数, 非必须（默认值：5）
			FragCount:    4,                              //指定m3u8文件中包含ts文件的个数，取值范围为【1，100】的整数，非必须（默认值：3）
			PlaylistName: "test-get-channel-status.m3u8", //当Type为HLS时，指定生成m3u8文件的名称，必须以“.m3u8”结尾，长度范围为【6，128】，非必须（默认值：playlist.m3u8）
		},
	}

	result, err := bucket.CreateLiveChannel(channelName, config)
	if err != nil {
	    //HandleError(err)
	}

	playURL := result.PlayUrls[0]
	publishURL := result.PublishUrls[0]
	fmt.Printf("create livechannel:%s  with config respones: playURL:%s, publishURL: %s\n", channelName, playURL, publishURL)
```

#### 生成签名rtmp地址（Sign rtmp URL）
```go
    client, err := oss.New("Endpoint", "AccessKeyId", "AccessKeySecret")
    if err != nil {
        // HandleError(err)
    }
    
    bucket, err := client.Bucket("my-bucket")
    if err != nil {
        // HandleError(err)
    }
    
    signedRtmpURL, err := bucket.SignRtmpURL(channelName, playlistName, 3600)
    if err != nil {
        // HandleError(err)
    }
    fmt.Printf("livechannel:%s, sinedRtmpURL: %s\n", channelName, signedRtmpURL)
```

#### 删除livechannel(Delete Livechannel)
```go
    client, err := oss.New("Endpoint", "AccessKeyId", "AccessKeySecret")
    if err != nil {
        // HandleError(err)
    }
    
    bucket, err := client.Bucket("my-bucket")
    if err != nil {
        // HandleError(err)
    }
    
    // 只支持”enalbed“,"disabled"两种状态
    err = bucket.DeleteLiveChannel(channelName)
    if err != nil {
        // HandleError(err)
    }
```

#### 设置livechannel状态(Put Livechannel Status)
```go
    client, err := oss.New("Endpoint", "AccessKeyId", "AccessKeySecret")
    if err != nil {
        // HandleError(err)
    }
    
    bucket, err := client.Bucket("my-bucket")
    if err != nil {
        // HandleError(err)
    }
    
    // 只支持”enalbed“,"disabled"两种状态
    err = bucket.PutLiveChannelStatus(channelName, "disabled")
    if err != nil {
        // HandleError(err)
    }
```

#### 获取livechannel当前推流的状态(Get Livechannel Stat)
```go
    client, err := oss.New("Endpoint", "AccessKeyId", "AccessKeySecret")
    if err != nil {
        // HandleError(err)
    }
    
    bucket, err := client.Bucket("my-bucket")
    if err != nil {
        // HandleError(err)
    }
    
    stat, err := bucket.GetLiveChannelStat(channelName)
    if err != nil {
        // HandleError(err)
    }

    status := stat.Status
    connectedTime := stat.ConnectedTime
    remoteAddr := stat.RemoteAddr

    audioBW := stat.Audio.Bandwidth
    audioCodec := stat.Audio.Codec
    audioSampleRate := stat.Audio.SampleRate

    videoBW := stat.Video.Bandwidth
    videoFrameRate := stat.Video.FrameRate
    videoHeight := stat.Video.Height
    videoWidth := stat.Video.Width

    fmt.Printf("get channel stat:（%v, %v，%v, %v）, audio(%v, %v, %v), video(%v, %v, %v, %v)\n", channelName, status, connectedTime, remoteAddr, audioBW, audioCodec, audioSampleRate, videoBW, videoFrameRate, videoHeight, videoWidth)
```

#### 获取livechannel的配置信息(Get Livechannel Info)
```go
    client, err := oss.New("Endpoint", "AccessKeyId", "AccessKeySecret")
    if err != nil {
        // HandleError(err)
    }
    
    bucket, err := client.Bucket("my-bucket")
    if err != nil {
        // HandleError(err)
    }
    
    info, err := bucket.GetLiveChannelInfo(channelName)
    if err != nil {
        // HandleError(err)
    }

    desc := info.Description
    status := info.Status
    fragCount := info.Target.FragCount
    fragDuation := info.Target.FragDuration
    playlistName := info.Target.PlaylistName
    targetType := info.Target.Type

    fmt.Printf("get channel stat:（%v,%v, %v）, target(%v, %v, %v, %v)\n", channelName, desc, status, fragCount, fragDuation, playlistName, targetType)
```

#### 获取livechannel的历史推流记录列表(Get Livechannel History)
```go
    client, err := oss.New("Endpoint", "AccessKeyId", "AccessKeySecret")
    if err != nil {
        // HandleError(err)
    }
    
    bucket, err := client.Bucket("my-bucket")
    if err != nil {
        // HandleError(err)
    }
    
    //目前最多会返回指定LiveChannel最近的10次推流记录
    history, err := bucket.GetLiveChannelHistory(channelName)
    for _, record := range history.Record {
        remoteAddr := record.RemoteAddr
        startTime := record.StartTime
        endTime := record.EndTime
        fmt.Printf("get channel:%s history:(%v, %v, %v)\n", channelName, remoteAddr, startTime, endTime)
	}
```

#### 获取当前bucket下livechannel的信息列表(List Livechannel)
```go
    client, err := oss.New("Endpoint", "AccessKeyId", "AccessKeySecret")
    if err != nil {
        // HandleError(err)
    }
    
    bucket, err := client.Bucket("my-bucket")
    if err != nil {
        // HandleError(err)
    }
    
    // 场景1：列出当前bucket下所有的livechannel
	marker := ""
	for {
		//设定marker值，第一次执行时为”“，后续执行时以返回结果的nextmarker的值作为marker的值
		result, err := bucket.ListLiveChannel(oss.Marker(marker))
		if err != nil {
			// HandleError(err)
		}

		for _, channel := range result.LiveChannel {
			fmt.Printf("list livechannel: (%v, %v, %v, %v, %v, %v)\n", channel.Name, channel.Status, channel.Description, channel.LastModified, channel.PlayUrls[0], channel.PublishUrls[0])
		}

		if result.IsTruncated {
			marker = result.NextMarker
		} else {
			break
		}
	}

    // 场景2：使用参数”max-keys“指定返回记录的最大个数, 但max-keys的值不能超过1000
	result, err := bucket.ListLiveChannel(oss.MaxKeys(10))
	if err != nil {
		HandleError(err)
	}
	for _, channel := range result.LiveChannel {
		fmt.Printf("list livechannel: (%v, %v, %v, %v, %v, %v)\n", channel.Name, channel.Status, channel.Description, channel.LastModified, channel.PlayUrls[0], channel.PublishUrls[0])
	}

	// 场景3；使用参数”prefix“过滤只列出包含”prefix“的值作为前缀的livechannel
	// max-keys, prefix, maker参数可以组合使用
	result, err = bucket.ListLiveChannel(oss.MaxKeys(10), oss.Prefix("list-"))
	if err != nil {
        // HandleError(err)
	}
	for _, channel := range result.LiveChannel {
		fmt.Printf("list livechannel: (%v, %v, %v, %v, %v, %v)\n", channel.Name, channel.Status, channel.Description, channel.LastModified, channel.PlayUrls[0], channel.PublishUrls[0])
	}
```

#### 生成livechannel的点播列表(Post Vod Playlist)
```go
    client, err := oss.New("Endpoint", "AccessKeyId", "AccessKeySecret")
    if err != nil {
        // HandleError(err)
    }
    
    bucket, err := client.Bucket("my-bucket")
    if err != nil {
        // HandleError(err)
    }
    
    endTime := time.Now().Add(-1 * time.Minute)
    startTime := endTime.Add(-60 * time.Minute)
    err = bucket.PostVodPlaylist(channelName, playlistName, startTime, endTime)
    if err != nil {
        // HandleError(err)
    }
```

#### 其它
更多的示例程序，请参看OSS Go SDK安装路径（即GOPATH变量中的第一个路径）下的`src\github.com\aliyun\aliyun-oss-go-sdk\sample`，该目录下为示例程序，
或者参看`https://github.com/aliyun/aliyun-oss-go-sdk`下sample目录中的示例文件。

## 注意事项
### 运行sample
> - 拷贝示例文件。到OSS Go SDK的安装路径（即GOPATH变量中的第一个路径），进入OSS Go SDK的代码目录`src\github.com\aliyun\aliyun-oss-go-sdk`，
把其下的sample目录和sample.go复制到您的测试工程src目录下。
> - 修改sample/config.go里的endpoint、AccessKeyId、AccessKeySecret、BucketName等配置。
> - 请在您的工程目录下执行`go run src/sample.go`。

## 联系我们
> - [阿里云OSS官方网站](http://oss.aliyun.com)
> - [阿里云OSS官方论坛](http://bbs.aliyun.com)
> - [阿里云OSS官方文档中心](http://www.aliyun.com/product/oss#Docs)
> - 阿里云官方技术支持：[提交工单](https://workorder.console.aliyun.com/#/ticket/createIndex)

## 作者
> - Yubin Bai
> - Hǎiliàng Wáng

## License
> - Apache License 2.0
