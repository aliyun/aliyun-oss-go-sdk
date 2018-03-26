package sample

import (
	"fmt"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

// CreateLiveChannelSample - 创建livechannel的sample
func CreateLiveChannelSample() {
	channelName := "create-livechannel"
	//创建bucket
	bucket, err := GetTestBucket(bucketName)
	if err != nil {
		HandleError(err)
	}

	// 场景1 - 完整配置livechannel创建
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
		HandleError(err)
	}

	playURL := result.PlayUrls[0]
	publishURL := result.PublishUrls[0]
	fmt.Printf("create livechannel:%s  with config respones: playURL:%s, publishURL: %s\n", channelName, playURL, publishURL)

	// 场景2 - 简单配置livechannel创建
	simpleCfg := oss.LiveChannelConfiguration{
		Target: oss.LiveChannelTarget{
			Type: "HLS",
		},
	}
	result, err = bucket.CreateLiveChannel(channelName, simpleCfg)
	if err != nil {
		HandleError(err)
	}
	playURL = result.PlayUrls[0]
	publishURL = result.PublishUrls[0]
	fmt.Printf("create livechannel:%s  with  simple config respones: playURL:%s, publishURL: %s\n", channelName, playURL, publishURL)

	err = DeleteTestBucketAndLiveChannel(bucketName)
	if err != nil {
		HandleError(err)
	}

	fmt.Println("PutObjectSample completed")
}

// PutLiveChannelStatusSample - 设置直播频道的状态的sample，有两种状态可选：enabled和disabled
func PutLiveChannelStatusSample() {
	channelName := "put-livechannel-status"
	bucket, err := GetTestBucket(bucketName)
	if err != nil {
		HandleError(err)
	}

	config := oss.LiveChannelConfiguration{
		Target: oss.LiveChannelTarget{
			Type: "HLS", //指定转储的类型，只支持HLS, 必须
		},
	}

	_, err = bucket.CreateLiveChannel(channelName, config)
	if err != nil {
		HandleError(err)
	}

	// 场景1 - 设置livechannel的状态为disabled
	err = bucket.PutLiveChannelStatus(channelName, "disabled")
	if err != nil {
		HandleError(err)
	}

	// 场景2 - 设置livechannel的状态为enalbed
	err = bucket.PutLiveChannelStatus(channelName, "enabled")
	if err != nil {
		HandleError(err)
	}

	err = DeleteTestBucketAndLiveChannel(bucketName)
	if err != nil {
		HandleError(err)
	}

	fmt.Println("PutLiveChannelStatusSample completed")
}

// PostVodPlayListSample - 生成点播列表的sample
func PostVodPlayListSample() {
	channelName := "post-vod-playlist"
	playlistName := "playlist.m3u8"
	bucket, err := GetTestBucket(bucketName)
	if err != nil {
		HandleError(err)
	}

	config := oss.LiveChannelConfiguration{
		Target: oss.LiveChannelTarget{
			Type:         "HLS", //指定转储的类型，只支持HLS, 必须
			PlaylistName: "playlist.m3u8",
		},
	}

	_, err = bucket.CreateLiveChannel(channelName, config)
	if err != nil {
		HandleError(err)
	}

	//这个阶段可以推流了...

	endTime := time.Now().Add(-1 * time.Minute)
	startTime := endTime.Add(-60 * time.Minute)
	err = bucket.PostVodPlaylist(channelName, playlistName, startTime, endTime)
	if err != nil {
		HandleError(err)
	}

	err = DeleteTestBucketAndLiveChannel(bucketName)
	if err != nil {
		HandleError(err)
	}

	fmt.Println("PostVodPlayListSampleSample completed")
}

// GetLiveChannelStatSample - 获取指定直播流频道当前推流的状态的sample
func GetLiveChannelStatSample() {
	channelName := "get-livechannel-stat"
	bucket, err := GetTestBucket(bucketName)
	if err != nil {
		HandleError(err)
	}

	config := oss.LiveChannelConfiguration{
		Target: oss.LiveChannelTarget{
			Type: "HLS", //指定转储的类型，只支持HLS, 必须
		},
	}

	_, err = bucket.CreateLiveChannel(channelName, config)
	if err != nil {
		HandleError(err)
	}

	stat, err := bucket.GetLiveChannelStat(channelName)
	if err != nil {
		HandleError(err)
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

	err = DeleteTestBucketAndLiveChannel(bucketName)
	if err != nil {
		HandleError(err)
	}

	fmt.Println("GetLiveChannelStatSample completed")
}

// GetLiveChannelInfoSample - 获取直播流频道的配置信息的sample
func GetLiveChannelInfoSample() {
	channelName := "get-livechannel-info"
	bucket, err := GetTestBucket(bucketName)
	if err != nil {
		HandleError(err)
	}

	config := oss.LiveChannelConfiguration{
		Target: oss.LiveChannelTarget{
			Type: "HLS", //指定转储的类型，只支持HLS, 必须
		},
	}

	_, err = bucket.CreateLiveChannel(channelName, config)
	if err != nil {
		HandleError(err)
	}

	info, err := bucket.GetLiveChannelInfo(channelName)
	if err != nil {
		HandleError(err)
	}

	desc := info.Description
	status := info.Status
	fragCount := info.Target.FragCount
	fragDuation := info.Target.FragDuration
	playlistName := info.Target.PlaylistName
	targetType := info.Target.Type

	fmt.Printf("get channel stat:（%v,%v, %v）, target(%v, %v, %v, %v)\n", channelName, desc, status, fragCount, fragDuation, playlistName, targetType)

	err = DeleteTestBucketAndLiveChannel(bucketName)
	if err != nil {
		HandleError(err)
	}

	fmt.Println("GetLiveChannelInfoSample completed")
}

// GetLiveChannelHistorySample - 获取直播流频道的历史推流记录的sample
func GetLiveChannelHistorySample() {
	channelName := "get-livechannel-info"
	bucket, err := GetTestBucket(bucketName)
	if err != nil {
		HandleError(err)
	}

	config := oss.LiveChannelConfiguration{
		Target: oss.LiveChannelTarget{
			Type: "HLS", //指定转储的类型，只支持HLS, 必须
		},
	}

	_, err = bucket.CreateLiveChannel(channelName, config)
	if err != nil {
		HandleError(err)
	}

	//目前最多会返回指定LiveChannel最近的10次推流记录
	history, err := bucket.GetLiveChannelHistory(channelName)
	for _, record := range history.Record {
		remoteAddr := record.RemoteAddr
		startTime := record.StartTime
		endTime := record.EndTime
		fmt.Printf("get channel:%s history:(%v, %v, %v)\n", channelName, remoteAddr, startTime, endTime)
	}

	err = DeleteTestBucketAndLiveChannel(bucketName)
	if err != nil {
		HandleError(err)
	}

	fmt.Println("GetLiveChannelHistorySample completed")
}

// ListLiveChannelSample - 获取当前bucket下直播流频道的信息列表的sample
func ListLiveChannelSample() {
	channelName := "list-livechannel"
	bucket, err := GetTestBucket(bucketName)
	if err != nil {
		HandleError(err)
	}

	config := oss.LiveChannelConfiguration{
		Target: oss.LiveChannelTarget{
			Type: "HLS", //指定转储的类型，只支持HLS, 必须
		},
	}

	_, err = bucket.CreateLiveChannel(channelName, config)
	if err != nil {
		HandleError(err)
	}

	// 场景1：列出当前bucket下所有的livechannel
	marker := ""
	for {
		//设定marker值，第一次执行时为”“，后续执行时以返回结果的nextmarker的值作为marker的值
		result, err := bucket.ListLiveChannel(oss.Marker(marker))
		if err != nil {
			HandleError(err)
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
		HandleError(err)
	}
	for _, channel := range result.LiveChannel {
		fmt.Printf("list livechannel: (%v, %v, %v, %v, %v, %v)\n", channel.Name, channel.Status, channel.Description, channel.LastModified, channel.PlayUrls[0], channel.PublishUrls[0])
	}

	err = DeleteTestBucketAndLiveChannel(bucketName)
	if err != nil {
		HandleError(err)
	}

	fmt.Println("ListLiveChannelSample completed")
}

// DeleteLiveChannelSample - 删除直播频道的信息列表的sample
func DeleteLiveChannelSample() {
	channelName := "delete-livechannel"
	bucket, err := GetTestBucket(bucketName)
	if err != nil {
		HandleError(err)
	}

	config := oss.LiveChannelConfiguration{
		Target: oss.LiveChannelTarget{
			Type: "HLS", //指定转储的类型，只支持HLS, 必须
		},
	}

	_, err = bucket.CreateLiveChannel(channelName, config)
	if err != nil {
		HandleError(err)
	}

	err = bucket.DeleteLiveChannel(channelName)
	if err != nil {
		HandleError(err)
	}

	err = DeleteTestBucketAndLiveChannel(bucketName)
	if err != nil {
		HandleError(err)
	}

	fmt.Println("DeleteLiveChannelSample completed")
}

// SignRtmpURLSample - 创建签名推流地址的sample
func SignRtmpURLSample() {
	channelName := "sign-rtmp-url"
	playlistName := "playlist.m3u8"
	bucket, err := GetTestBucket(bucketName)
	if err != nil {
		HandleError(err)
	}

	config := oss.LiveChannelConfiguration{
		Target: oss.LiveChannelTarget{
			Type:         "HLS", //指定转储的类型，只支持HLS, 必须
			PlaylistName: "playlist.m3u8",
		},
	}

	result, err := bucket.CreateLiveChannel(channelName, config)
	if err != nil {
		HandleError(err)
	}

	playURL := result.PlayUrls[0]
	publishURL := result.PublishUrls[0]
	fmt.Printf("livechannel:%s, playURL:%s, publishURL: %s\n", channelName, playURL, publishURL)

	signedRtmpURL, err := bucket.SignRtmpURL(channelName, playlistName, 3600)
	if err != nil {
		HandleError(err)
	}
	fmt.Printf("livechannel:%s, sinedRtmpURL: %s\n", channelName, signedRtmpURL)

	err = DeleteTestBucketAndLiveChannel(bucketName)
	if err != nil {
		HandleError(err)
	}

	fmt.Println("PutObjectSample completed")
}
