package oss

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"net/http"
	"strconv"
)

//
// CreateLiveChannel 创建推流直播频道
//
// channelName 直播流频道名称
// config 直播流频的配置信息
//
// CreateLiveChannelResult 创建直播流频请求的返回结果
// error 操作无错误时返回nil，非nil为错误信息
//
func (bucket Bucket) CreateLiveChannel(channelName string, config LiveChannelConfiguration) (CreateLiveChannelResult, error) {
	var out CreateLiveChannelResult

	bs, err := xml.Marshal(config)
	if err != nil {
		return out, err
	}

	buffer := new(bytes.Buffer)
	buffer.Write(bs)

	params := map[string]interface{}{}
	params["live"] = nil
	resp, err := bucket.do("PUT", channelName, params, nil, buffer, nil)
	if err != nil {
		return out, err
	}
	defer resp.Body.Close()

	err = xmlUnmarshal(resp.Body, &out)
	return out, err
}

//
// PutLiveChannelStatus 设置直播频道的状态，有两种状态可选：enabled和disabled
//
// channelName 直播流频道的名称
// status 状态，有两种状态可选：enabled和disabled
//
// error 操作无错误时返回nil, 非nil为错误信息
//
func (bucket Bucket) PutLiveChannelStatus(channelName, status string) error {
	params := map[string]interface{}{}
	params["live"] = nil
	params["status"] = status

	resp, err := bucket.do("PUT", channelName, params, nil, nil, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return checkRespCode(resp.StatusCode, []int{http.StatusOK})
}

// PostVodPlaylist 根据指定的playlist name以及startTime和endTime生成一个点播的播放列表
//
// channelName 直播流频道的名称
// playlistName 指定生成的点播列表的名称，必须以”.m3u8“结尾
// startTime 指定查询ts文件的起始时间，格式为Unix timestamp
// endTime 指定查询ts文件的终止时间，格式为Unix timestamp
//
// error 操作无错误是返回nil, 非nil为错误信息
//
func (bucket Bucket) PostVodPlaylist(channelName, playlistName string, startTime, endTime int64) error {
	params := map[string]interface{}{}
	params["vod"] = nil
	params["startTime"] = strconv.FormatInt(startTime, 10)
	params["endTime"] = strconv.FormatInt(endTime, 10)

	key := fmt.Sprintf("%s/%s", channelName, playlistName)
	resp, err := bucket.do("POST", key, params, nil, nil, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return checkRespCode(resp.StatusCode, []int{http.StatusOK})
}

//
// GetLiveChannelStat 获取指定直播流频道当前推流的状态
//
// channelName 直播流频道的名称
//
// LiveChannelStat 直播流频道当前推流状态信息
// error 操作无错误是返回nil, 非nil为错误信息
//
func (bucket Bucket) GetLiveChannelStat(channelName string) (LiveChannelStat, error) {
	var out LiveChannelStat
	params := map[string]interface{}{}
	params["live"] = nil
	params["comp"] = "stat"

	resp, err := bucket.do("GET", channelName, params, nil, nil, nil)
	if err != nil {
		return out, err
	}
	defer resp.Body.Close()

	err = xmlUnmarshal(resp.Body, &out)
	return out, err
}

//
// GetLiveChannelInfo 获取直播流频道的配置信息
//
// channelName 直播流频道的名称
//
// LiveChannelConfiguration 直播流频道的配置信息
// error 操作无错误返回nil, 非nil为错误信息
//
func (bucket Bucket) GetLiveChannelInfo(channelName string) (LiveChannelConfiguration, error) {
	var out LiveChannelConfiguration
	params := map[string]interface{}{}
	params["live"] = nil

	resp, err := bucket.do("GET", channelName, params, nil, nil, nil)
	if err != nil {
		return out, err
	}
	defer resp.Body.Close()

	err = xmlUnmarshal(resp.Body, &out)
	return out, err
}

//
// GetLiveChannelHistory 获取直播流频道的历史推流记录
//
// channelName 直播流频道名称
//
// LiveChannelHistory 返回的直播流历史推流记录
// error 操作无错误返回nil, 非nil为错误信息
//
func (bucket Bucket) GetLiveChannelHistory(channelName string) (LiveChannelHistory, error) {
	var out LiveChannelHistory
	params := map[string]interface{}{}
	params["live"] = nil
	params["comp"] = "history"

	resp, err := bucket.do("GET", channelName, params, nil, nil, nil)
	if err != nil {
		return out, err
	}
	defer resp.Body.Close()

	err = xmlUnmarshal(resp.Body, &out)
	return out, err
}

//
// ListLiveChannel 获取直播流频道的信息列表
//
// options  筛选选项：Prefix指定的前缀、MaxKeys为返回的最大数目、Marker代表从哪个livechannel作为游标开始列表
//
// ListLiveChannelResult 返回的livechannel列表结果
// error 操作无结果返回nil, 非nil为错误信息
//
func (bucket Bucket) ListLiveChannel(options ...Option) (ListLiveChannelResult, error) {
	var out ListLiveChannelResult

	//options = append(options, EncodingType("url"))
	params, err := getRawParams(options)
	if err != nil {
		return out, err
	}

	params["live"] = nil

	resp, err := bucket.do("GET", "", params, nil, nil, nil)
	if err != nil {
		return out, err
	}
	defer resp.Body.Close()

	err = xmlUnmarshal(resp.Body, &out)
	return out, err
}

//
// DeleteLiveChannel 删除指定的livechannel，当有客户端正在想livechannel推流时，删除请求回失败，本接口志辉删除livechannel本身，不会删除推流生成的文件
//
// channelName 直播流的频道名称
//
// error 操作无错误返回nil, 非nil为错误信息
//
func (bucket Bucket) DeleteLiveChannel(channelName string) error {
	params := map[string]interface{}{}
	params["live"] = nil

	resp, err := bucket.do("DELETE", channelName, params, nil, nil, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return checkRespCode(resp.StatusCode, []int{http.StatusNoContent})
}
