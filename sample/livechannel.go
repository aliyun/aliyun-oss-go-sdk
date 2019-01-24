package sample

import (
	"fmt"
	"io/ioutil"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

// CreateLiveChannelSample Samples for create a live-channel
func CreateLiveChannelSample() {
	channelName := "create-livechannel"
	//create bucket
	bucket, err := GetTestBucket(bucketName)
	if err != nil {
		HandleError(err)
	}

	// Case 1 - Create live-channel with Completely configure
	config := oss.LiveChannelConfiguration{
		Description: "sample-for-livechannel", //description information, up to 128 bytes
		Status:      "enabled",                //enabled or disabled
		Target: oss.LiveChannelTarget{
			Type:         "HLS",                          //the type of object, only supports HLS, required
			FragDuration: 10,                             //the length of each ts object (in seconds), in the range [1,100], default: 5
			FragCount:    4,                              //the number of ts objects in the m3u8 object, in the range of [1,100], default: 3
			PlaylistName: "test-get-channel-status.m3u8", //the name of m3u8 object, which must end with ".m3u8" and the length range is [6,128]，default: playlist.m3u8
		},
	}

	result, err := bucket.CreateLiveChannel(channelName, config)
	if err != nil {
		HandleError(err)
	}

	playURL := result.PlayUrls[0]
	publishURL := result.PublishUrls[0]
	fmt.Printf("create livechannel:%s  with config respones: playURL:%s, publishURL: %s\n", channelName, playURL, publishURL)

	// Case 2 - Create live-channel only specified type of target which is required
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

// PutLiveChannelStatusSample Set the status of the live-channel sample: enabled/disabled
func PutLiveChannelStatusSample() {
	channelName := "put-livechannel-status"
	bucket, err := GetTestBucket(bucketName)
	if err != nil {
		HandleError(err)
	}

	config := oss.LiveChannelConfiguration{
		Target: oss.LiveChannelTarget{
			Type: "HLS", //the type of object, only supports HLS, required
		},
	}

	_, err = bucket.CreateLiveChannel(channelName, config)
	if err != nil {
		HandleError(err)
	}

	// Case 1 - Set the status of live-channel to disabled
	err = bucket.PutLiveChannelStatus(channelName, "disabled")
	if err != nil {
		HandleError(err)
	}

	// Case 2 - Set the status of live-channel to enabled
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

// PostVodPlayListSample Sample for generate playlist
func PostVodPlayListSample() {
	channelName := "post-vod-playlist"
	playlistName := "playlist.m3u8"
	bucket, err := GetTestBucket(bucketName)
	if err != nil {
		HandleError(err)
	}

	config := oss.LiveChannelConfiguration{
		Target: oss.LiveChannelTarget{
			Type:         "HLS", //the type of object, only supports HLS, required
			PlaylistName: "playlist.m3u8",
		},
	}

	_, err = bucket.CreateLiveChannel(channelName, config)
	if err != nil {
		HandleError(err)
	}

	//This stage you can push live stream, and after that you could generator playlist

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

// GetVodPlayListSample Sample for generate playlist and return the content of the playlist
func GetVodPlayListSample() {
	channelName := "get-vod-playlist"
	bucket, err := GetTestBucket(bucketName)
	if err != nil {
		HandleError(err)
	}

	config := oss.LiveChannelConfiguration{
		Target: oss.LiveChannelTarget{
			Type:         "HLS", //the type of object, only supports HLS, required
			PlaylistName: "playlist.m3u8",
		},
	}

	_, err = bucket.CreateLiveChannel(channelName, config)
	if err != nil {
		HandleError(err)
	}

	//This stage you can push live stream, and after that you could generator playlist

	endTime := time.Now().Add(-1 * time.Minute)
	startTime := endTime.Add(-60 * time.Minute)
	body, err := bucket.GetVodPlaylist(channelName, startTime, endTime)
	if err != nil {
		HandleError(err)
	}
	defer body.Close()

	data, err := ioutil.ReadAll(body)
	if err != nil {
		HandleError(err)
	}
	fmt.Printf("content of playlist is:%v\n", string(data))

	err = DeleteTestBucketAndLiveChannel(bucketName)
	if err != nil {
		HandleError(err)
	}

	fmt.Println("PostVodPlayListSampleSample completed")
}

// GetLiveChannelStatSample Sample for get the state of live-channel
func GetLiveChannelStatSample() {
	channelName := "get-livechannel-stat"
	bucket, err := GetTestBucket(bucketName)
	if err != nil {
		HandleError(err)
	}

	config := oss.LiveChannelConfiguration{
		Target: oss.LiveChannelTarget{
			Type: "HLS", //the type of object, only supports HLS, required
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

// GetLiveChannelInfoSample Sample for get the configuration infomation of live-channel
func GetLiveChannelInfoSample() {
	channelName := "get-livechannel-info"
	bucket, err := GetTestBucket(bucketName)
	if err != nil {
		HandleError(err)
	}

	config := oss.LiveChannelConfiguration{
		Target: oss.LiveChannelTarget{
			Type: "HLS", //the type of object, only supports HLS, required
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

// GetLiveChannelHistorySample  Sample for get push records of live-channel
func GetLiveChannelHistorySample() {
	channelName := "get-livechannel-info"
	bucket, err := GetTestBucket(bucketName)
	if err != nil {
		HandleError(err)
	}

	config := oss.LiveChannelConfiguration{
		Target: oss.LiveChannelTarget{
			Type: "HLS", //the type of object, only supports HLS, required
		},
	}

	_, err = bucket.CreateLiveChannel(channelName, config)
	if err != nil {
		HandleError(err)
	}

	//at most return up to lastest 10 push records
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

// ListLiveChannelSample Samples for list live-channels with specified bucket name
func ListLiveChannelSample() {
	channelName := "list-livechannel"
	bucket, err := GetTestBucket(bucketName)
	if err != nil {
		HandleError(err)
	}

	config := oss.LiveChannelConfiguration{
		Target: oss.LiveChannelTarget{
			Type: "HLS", //the type of object, only supports HLS, required
		},
	}

	_, err = bucket.CreateLiveChannel(channelName, config)
	if err != nil {
		HandleError(err)
	}

	// Case 1: list all the live-channels
	marker := ""
	for {
		// Set the marker value, the first time is "", the value of NextMarker that returned should as the marker in the next time
		// At most return up to lastest 100 live-channels if "max-keys" is not specified
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

	// Case 2: Use the parameter "max-keys" to specify the maximum number of records returned, the value of max-keys cannot exceed 1000
	// if "max-keys" the default value is 100
	result, err := bucket.ListLiveChannel(oss.MaxKeys(10))
	if err != nil {
		HandleError(err)
	}
	for _, channel := range result.LiveChannel {
		fmt.Printf("list livechannel: (%v, %v, %v, %v, %v, %v)\n", channel.Name, channel.Status, channel.Description, channel.LastModified, channel.PlayUrls[0], channel.PublishUrls[0])
	}

	// Case 3: Only list the live-channels with the value of parameter "prefix" as prefix
	// max-keys, prefix, maker parameters can be combined
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

// DeleteLiveChannelSample Sample for delete live-channel
func DeleteLiveChannelSample() {
	channelName := "delete-livechannel"
	bucket, err := GetTestBucket(bucketName)
	if err != nil {
		HandleError(err)
	}

	config := oss.LiveChannelConfiguration{
		Target: oss.LiveChannelTarget{
			Type: "HLS", //the type of object, only supports HLS, required
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

// SignRtmpURLSample Sample for generate a RTMP push-stream signature URL for the trusted user to push the RTMP stream to the live channel.
func SignRtmpURLSample() {
	channelName := "sign-rtmp-url"
	playlistName := "playlist.m3u8"
	bucket, err := GetTestBucket(bucketName)
	if err != nil {
		HandleError(err)
	}

	config := oss.LiveChannelConfiguration{
		Target: oss.LiveChannelTarget{
			Type:         "HLS", //the type of object, only supports HLS, required
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

	fmt.Println("SignRtmpURLSample completed")
}
