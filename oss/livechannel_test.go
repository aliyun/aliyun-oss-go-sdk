package oss

import (
	"fmt"
	"strings"
	"time"

	. "gopkg.in/check.v1"
)

type OssBucketLiveChannelSuite struct {
	cloudBoxControlClient *Client
	client                *Client
	bucket                *Bucket
}

var _ = Suite(&OssBucketLiveChannelSuite{})

// SetUpSuite Run once when the suite starts running
func (s *OssBucketLiveChannelSuite) SetUpSuite(c *C) {
	bucketName := bucketNamePrefix + RandLowStr(6)
	if cloudboxControlEndpoint == "" {
		client, err := New(endpoint, accessID, accessKey)
		c.Assert(err, IsNil)
		s.client = client

		err = s.client.CreateBucket(bucketName)
		c.Assert(err, IsNil)
		time.Sleep(5 * time.Second)

		bucket, err := s.client.Bucket(bucketName)
		c.Assert(err, IsNil)
		s.bucket = bucket
	} else {
		client, err := New(cloudboxEndpoint, accessID, accessKey)
		c.Assert(err, IsNil)
		s.client = client

		controlClient, err := New(cloudboxControlEndpoint, accessID, accessKey)
		c.Assert(err, IsNil)
		s.cloudBoxControlClient = controlClient
		controlClient.CreateBucket(bucketName)

		bucket, err := s.client.Bucket(bucketName)
		c.Assert(err, IsNil)
		s.bucket = bucket
	}

	testLogger.Println("test livechannel started...")
}

// TearDownSuite Run once after all tests or benchmarks
func (s *OssBucketLiveChannelSuite) TearDownSuite(c *C) {
	marker := ""
	for {
		result, err := s.bucket.ListLiveChannel(Marker(marker))
		c.Assert(err, IsNil)

		for _, channel := range result.LiveChannel {
			err := s.bucket.DeleteLiveChannel(channel.Name)
			c.Assert(err, IsNil)
		}

		if result.IsTruncated {
			marker = result.NextMarker
		} else {
			break
		}
	}

	// Delete bucket
	if s.cloudBoxControlClient != nil {
		err := s.cloudBoxControlClient.DeleteBucket(s.bucket.BucketName)
		c.Assert(err, IsNil)
	} else {
		err := s.client.DeleteBucket(s.bucket.BucketName)
		c.Assert(err, IsNil)
	}

	testLogger.Println("test livechannel done...")
}

// SetUpTest Run before each test or benchmark starts
func (s *OssBucketLiveChannelSuite) SetUpTest(c *C) {

}

// TearDownTest	Run after each test or benchmark runs.
func (s *OssBucketLiveChannelSuite) TearDownTest(c *C) {

}

// TestCreateLiveChannel
func (s *OssBucketLiveChannelSuite) TestCreateLiveChannel(c *C) {
	channelName := "test-create-channel"
	playlistName := "test-create-channel.m3u8"

	target := LiveChannelTarget{
		PlaylistName: playlistName,
		Type:         "HLS",
	}

	config := LiveChannelConfiguration{
		Description: "livechannel for test",
		Status:      "enabled",
		Target:      target,
	}

	result, err := s.bucket.CreateLiveChannel(channelName, config)
	c.Assert(err, IsNil)

	playURL := getPlayURL(s.bucket.BucketName, channelName, playlistName)
	publishURL := getPublishURL(s.bucket.BucketName, channelName)
	c.Assert(result.PlayUrls[0], Equals, playURL)
	c.Assert(result.PublishUrls[0], Equals, publishURL)

	err = s.bucket.DeleteLiveChannel(channelName)
	c.Assert(err, IsNil)

	invalidType := RandStr(4)
	invalidTarget := LiveChannelTarget{
		PlaylistName: playlistName,
		Type:         invalidType,
	}

	invalidConfig := LiveChannelConfiguration{
		Description: "livechannel for test",
		Status:      "enabled",
		Target:      invalidTarget,
	}

	_, err = s.bucket.CreateLiveChannel(channelName, invalidConfig)
	c.Assert(err, NotNil)
}

// TestDeleteLiveChannel
func (s *OssBucketLiveChannelSuite) TestDeleteLiveChannel(c *C) {
	channelName := "test-delete-channel"

	target := LiveChannelTarget{
		Type: "HLS",
	}

	config := LiveChannelConfiguration{
		Target: target,
	}

	_, err := s.bucket.CreateLiveChannel(channelName, config)
	c.Assert(err, IsNil)

	err = s.bucket.DeleteLiveChannel(channelName)
	c.Assert(err, IsNil)

	emptyChannelName := ""
	err = s.bucket.DeleteLiveChannel(emptyChannelName)
	c.Assert(err, NotNil)

	longChannelName := RandStr(65)
	err = s.bucket.DeleteLiveChannel(longChannelName)
	c.Assert(err, NotNil)

	config, err = s.bucket.GetLiveChannelInfo(channelName)
	c.Assert(err, NotNil)
}

// TestGetLiveChannelInfo
func (s *OssBucketLiveChannelSuite) TestGetLiveChannelInfo(c *C) {
	channelName := "test-get-channel-status"

	_, err := s.bucket.GetLiveChannelInfo(channelName)
	c.Assert(err, NotNil)

	createCfg := LiveChannelConfiguration{
		Target: LiveChannelTarget{
			Type:         "HLS",
			FragDuration: 10,
			FragCount:    4,
			PlaylistName: "test-get-channel-status.m3u8",
		},
	}

	_, err = s.bucket.CreateLiveChannel(channelName, createCfg)
	c.Assert(err, IsNil)

	getCfg, err := s.bucket.GetLiveChannelInfo(channelName)
	c.Assert(err, IsNil)
	c.Assert("enabled", Equals, getCfg.Status) //The default value is enabled
	c.Assert("HLS", Equals, getCfg.Target.Type)
	c.Assert(10, Equals, getCfg.Target.FragDuration)
	c.Assert(4, Equals, getCfg.Target.FragCount)
	c.Assert("test-get-channel-status.m3u8", Equals, getCfg.Target.PlaylistName)

	err = s.bucket.DeleteLiveChannel(channelName)
	c.Assert(err, IsNil)
}

// TestPutLiveChannelStatus
func (s *OssBucketLiveChannelSuite) TestPutLiveChannelStatus(c *C) {
	channelName := "test-put-channel-status"

	config := LiveChannelConfiguration{
		Status: "disabled",
		Target: LiveChannelTarget{
			Type: "HLS",
		},
	}

	_, err := s.bucket.CreateLiveChannel(channelName, config)
	c.Assert(err, IsNil)
	getCfg, err := s.bucket.GetLiveChannelInfo(channelName)
	c.Assert(err, IsNil)
	c.Assert("disabled", Equals, getCfg.Status)

	err = s.bucket.PutLiveChannelStatus(channelName, "enabled")
	c.Assert(err, IsNil)
	getCfg, err = s.bucket.GetLiveChannelInfo(channelName)
	c.Assert(err, IsNil)
	c.Assert("enabled", Equals, getCfg.Status)

	err = s.bucket.PutLiveChannelStatus(channelName, "disabled")
	c.Assert(err, IsNil)
	getCfg, err = s.bucket.GetLiveChannelInfo(channelName)
	c.Assert(err, IsNil)
	c.Assert("disabled", Equals, getCfg.Status)

	invalidStatus := RandLowStr(9)
	err = s.bucket.PutLiveChannelStatus(channelName, invalidStatus)
	c.Assert(err, NotNil)

	err = s.bucket.DeleteLiveChannel(channelName)
	c.Assert(err, IsNil)
}

// TestGetLiveChannelHistory
func (s *OssBucketLiveChannelSuite) TestGetLiveChannelHistory(c *C) {
	channelName := "test-get-channel-history"

	_, err := s.bucket.GetLiveChannelHistory(channelName)
	c.Assert(err, NotNil)

	config := LiveChannelConfiguration{
		Target: LiveChannelTarget{
			Type: "HLS",
		},
	}

	_, err = s.bucket.CreateLiveChannel(channelName, config)
	c.Assert(err, IsNil)

	history, err := s.bucket.GetLiveChannelHistory(channelName)
	c.Assert(err, IsNil)
	c.Assert(len(history.Record), Equals, 0)

	err = s.bucket.DeleteLiveChannel(channelName)
	c.Assert(err, IsNil)
}

// TestGetLiveChannelStat
func (s *OssBucketLiveChannelSuite) TestGetLiveChannelStat(c *C) {
	channelName := "test-get-channel-stat"

	_, err := s.bucket.GetLiveChannelStat(channelName)
	c.Assert(err, NotNil)

	config := LiveChannelConfiguration{
		Target: LiveChannelTarget{
			Type: "HLS",
		},
	}

	_, err = s.bucket.CreateLiveChannel(channelName, config)
	c.Assert(err, IsNil)

	stat, err := s.bucket.GetLiveChannelStat(channelName)
	c.Assert(err, IsNil)
	c.Assert(stat.Status, Equals, "Idle")

	err = s.bucket.DeleteLiveChannel(channelName)
	c.Assert(err, IsNil)
}

// TestPostVodPlaylist
func (s *OssBucketLiveChannelSuite) TestPostVodPlaylist(c *C) {
	channelName := "test-post-vod-playlist"
	playlistName := "test-post-vod-playlist.m3u8"

	config := LiveChannelConfiguration{
		Target: LiveChannelTarget{
			Type: "HLS",
		},
	}

	_, err := s.bucket.CreateLiveChannel(channelName, config)
	c.Assert(err, IsNil)

	endTime := time.Now().Add(-1 * time.Minute)
	startTime := endTime.Add(-60 * time.Minute)

	err = s.bucket.PostVodPlaylist(channelName, playlistName, startTime, endTime)
	c.Assert(err, NotNil)

	err = s.bucket.DeleteLiveChannel(channelName)
	c.Assert(err, IsNil)
}

// TestPostVodPlaylist
func (s *OssBucketLiveChannelSuite) TestGetVodPlaylist(c *C) {
	channelName := "test-get-vod-playlist"

	config := LiveChannelConfiguration{
		Target: LiveChannelTarget{
			Type: "HLS",
		},
	}

	_, err := s.bucket.CreateLiveChannel(channelName, config)
	c.Assert(err, IsNil)

	endTime := time.Now().Add(-1 * time.Minute)
	startTime := endTime.Add(-60 * time.Minute)

	_, err = s.bucket.GetVodPlaylist(channelName, startTime, endTime)
	c.Assert(err, NotNil)

	err = s.bucket.DeleteLiveChannel(channelName)
	c.Assert(err, IsNil)
}

// TestListLiveChannel
func (s *OssBucketLiveChannelSuite) TestListLiveChannel(c *C) {
	result, err := s.bucket.ListLiveChannel()
	c.Assert(err, IsNil)
	ok := compareListResult(result, "", "", "", 100, false, 0)
	c.Assert(ok, Equals, true)

	prefix := "test-list-channel"
	for i := 0; i < 200; i++ {
		channelName := fmt.Sprintf("%s-%03d", prefix, i)

		config := LiveChannelConfiguration{
			Target: LiveChannelTarget{
				Type: "HLS",
			},
		}

		_, err := s.bucket.CreateLiveChannel(channelName, config)
		c.Assert(err, IsNil)
	}

	result, err = s.bucket.ListLiveChannel()
	c.Assert(err, IsNil)
	nextMarker := fmt.Sprintf("%s-099", prefix)
	ok = compareListResult(result, "", "", nextMarker, 100, true, 100)
	c.Assert(ok, Equals, true)

	randPrefix := RandStr(5)
	result, err = s.bucket.ListLiveChannel(Prefix(randPrefix))
	c.Assert(err, IsNil)
	ok = compareListResult(result, randPrefix, "", "", 100, false, 0)
	c.Assert(ok, Equals, true)

	marker := fmt.Sprintf("%s-100", prefix)
	result, err = s.bucket.ListLiveChannel(Prefix(prefix), Marker(marker))
	c.Assert(err, IsNil)
	ok = compareListResult(result, prefix, marker, "", 100, false, 99)
	c.Assert(ok, Equals, true)

	maxKeys := 1000
	result, err = s.bucket.ListLiveChannel(MaxKeys(maxKeys))
	c.Assert(err, IsNil)
	ok = compareListResult(result, "", "", "", maxKeys, false, 200)

	invalidMaxKeys := -1
	result, err = s.bucket.ListLiveChannel(MaxKeys(invalidMaxKeys))
	c.Assert(err, NotNil)

	for i := 0; i < 200; i++ {
		channelName := fmt.Sprintf("%s-%03d", prefix, i)
		err := s.bucket.DeleteLiveChannel(channelName)
		c.Assert(err, IsNil)
	}
}

// TestSignRtmpURL
func (s *OssBucketLiveChannelSuite) TestSignRtmpURL(c *C) {
	channelName := "test-sign-rtmp-url"
	playlistName := "test-sign-rtmp-url.m3u8"

	config := LiveChannelConfiguration{
		Target: LiveChannelTarget{
			Type:         "HLS",
			PlaylistName: playlistName,
		},
	}

	_, err := s.bucket.CreateLiveChannel(channelName, config)
	c.Assert(err, IsNil)

	expires := int64(3600)
	signedRtmpURL, err := s.bucket.SignRtmpURL(channelName, playlistName, expires)
	c.Assert(err, IsNil)
	playURL := getPublishURL(s.bucket.BucketName, channelName)
	hasPrefix := strings.HasPrefix(signedRtmpURL, playURL)
	c.Assert(hasPrefix, Equals, true)

	invalidExpires := int64(-1)
	signedRtmpURL, err = s.bucket.SignRtmpURL(channelName, playlistName, invalidExpires)
	c.Assert(err, NotNil)

	err = s.bucket.DeleteLiveChannel(channelName)
	c.Assert(err, IsNil)
}

// private
// getPlayURL Get the play url of the live channel
func getPlayURL(bucketName, channelName, playlistName string) string {
	host := ""
	useHTTPS := false
	if strings.Contains(endpoint, "https://") {
		host = endpoint[8:]
		useHTTPS = true
	} else if strings.Contains(endpoint, "http://") {
		host = endpoint[7:]
	} else {
		host = endpoint
	}

	if useHTTPS {
		return fmt.Sprintf("https://%s.%s/%s/%s", bucketName, host, channelName, playlistName)
	}
	return fmt.Sprintf("http://%s.%s/%s/%s", bucketName, host, channelName, playlistName)
}

// getPublistURL Get the push url of the live stream channel
func getPublishURL(bucketName, channelName string) string {
	host := ""
	if strings.Contains(endpoint, "https://") {
		host = endpoint[8:]
	} else if strings.Contains(endpoint, "http://") {
		host = endpoint[7:]
	} else {
		host = endpoint
	}

	return fmt.Sprintf("rtmp://%s.%s/live/%s", bucketName, host, channelName)
}

func compareListResult(result ListLiveChannelResult, prefix, marker, nextMarker string, maxKey int, isTruncated bool, count int) bool {
	if result.Prefix != prefix || result.Marker != marker || result.NextMarker != nextMarker || result.MaxKeys != maxKey || result.IsTruncated != isTruncated || len(result.LiveChannel) != count {
		return false
	}

	return true
}
