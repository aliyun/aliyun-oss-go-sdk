package oss

import (
	"encoding/xml"
	"net/url"
	"time"
)

// ListBucketsResult defines the result object from ListBuckets request
type ListBucketsResult struct {
	XMLName     xml.Name           `xml:"ListAllMyBucketsResult"`
	Prefix      string             `xml:"Prefix"`         // The prefix in this query
	Marker      string             `xml:"Marker"`         // The marker filter
	MaxKeys     int                `xml:"MaxKeys"`        // The max entry count to return. This information is returned when IsTruncated is true.
	IsTruncated bool               `xml:"IsTruncated"`    // Flag true means there's remaining buckets to return.
	NextMarker  string             `xml:"NextMarker"`     // The marker filter for the next list call
	Owner       Owner              `xml:"Owner"`          // The owner information
	Buckets     []BucketProperties `xml:"Buckets>Bucket"` // The bucket list
}

// BucketProperties defines bucket properties
type BucketProperties struct {
	XMLName      xml.Name  `xml:"Bucket"`
	Name         string    `xml:"Name"`         // Bucket name
	Location     string    `xml:"Location"`     // Bucket datacenter
	CreationDate time.Time `xml:"CreationDate"` // Bucket create time
	StorageClass string    `xml:"StorageClass"` // Bucket storage class
}

// GetBucketACLResult defines GetBucketACL request's result
type GetBucketACLResult struct {
	XMLName xml.Name `xml:"AccessControlPolicy"`
	ACL     string   `xml:"AccessControlList>Grant"` // Bucket ACL
	Owner   Owner    `xml:"Owner"`                   // Bucket owner
}

// LifecycleConfiguration is the Bucket Lifecycle configuration
type LifecycleConfiguration struct {
	XMLName xml.Name        `xml:"LifecycleConfiguration"`
	Rules   []LifecycleRule `xml:"Rule"`
}

// LifecycleRule defines Lifecycle rules
type LifecycleRule struct {
	XMLName    xml.Name            `xml:"Rule"`
	ID         string              `xml:"ID"`         // The rule ID
	Prefix     string              `xml:"Prefix"`     // The object key prefix
	Status     string              `xml:"Status"`     // The rule status (enabled or not)
	Expiration LifecycleExpiration `xml:"Expiration"` // The expiration property
}

// LifecycleExpiration defines the rule's expiration property
type LifecycleExpiration struct {
	XMLName xml.Name  `xml:"Expiration"`
	Days    int       `xml:"Days,omitempty"` // Relative expiration time: The expiration time in days after the last modified time
	Date    time.Time `xml:"Date,omitempty"` // Absolute expiration time: The expiration time in date.
}

type lifecycleXML struct {
	XMLName xml.Name        `xml:"LifecycleConfiguration"`
	Rules   []lifecycleRule `xml:"Rule"`
}

type lifecycleRule struct {
	XMLName    xml.Name            `xml:"Rule"`
	ID         string              `xml:"ID"`
	Prefix     string              `xml:"Prefix"`
	Status     string              `xml:"Status"`
	Expiration lifecycleExpiration `xml:"Expiration"`
}

type lifecycleExpiration struct {
	XMLName xml.Name `xml:"Expiration"`
	Days    int      `xml:"Days,omitempty"`
	Date    string   `xml:"Date,omitempty"`
}

const expirationDateFormat = "2006-01-02T15:04:05.000Z"

func convLifecycleRule(rules []LifecycleRule) []lifecycleRule {
	rs := []lifecycleRule{}
	for _, rule := range rules {
		r := lifecycleRule{}
		r.ID = rule.ID
		r.Prefix = rule.Prefix
		r.Status = rule.Status
		if rule.Expiration.Date.IsZero() {
			r.Expiration.Days = rule.Expiration.Days
		} else {
			r.Expiration.Date = rule.Expiration.Date.Format(expirationDateFormat)
		}
		rs = append(rs, r)
	}
	return rs
}

// BuildLifecycleRuleByDays builds a lifecycle rule with specified expiration days
func BuildLifecycleRuleByDays(id, prefix string, status bool, days int) LifecycleRule {
	var statusStr = "Enabled"
	if !status {
		statusStr = "Disabled"
	}
	return LifecycleRule{ID: id, Prefix: prefix, Status: statusStr,
		Expiration: LifecycleExpiration{Days: days}}
}

// BuildLifecycleRuleByDate builds a lifecycle rule with specified expiration time.
func BuildLifecycleRuleByDate(id, prefix string, status bool, year, month, day int) LifecycleRule {
	var statusStr = "Enabled"
	if !status {
		statusStr = "Disabled"
	}
	date := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	return LifecycleRule{ID: id, Prefix: prefix, Status: statusStr,
		Expiration: LifecycleExpiration{Date: date}}
}

// GetBucketLifecycleResult defines GetBucketLifecycle's result object
type GetBucketLifecycleResult LifecycleConfiguration

// RefererXML defines Referer configuration
type RefererXML struct {
	XMLName           xml.Name `xml:"RefererConfiguration"`
	AllowEmptyReferer bool     `xml:"AllowEmptyReferer"`   // Allow empty referrer
	RefererList       []string `xml:"RefererList>Referer"` // Referer whitelist
}

// GetBucketRefererResult defines result object for GetBucketReferer request
type GetBucketRefererResult RefererXML

// LoggingXML defines logging configuration
type LoggingXML struct {
	XMLName        xml.Name       `xml:"BucketLoggingStatus"`
	LoggingEnabled LoggingEnabled `xml:"LoggingEnabled"` // The logging configuration information
}

type loggingXMLEmpty struct {
	XMLName xml.Name `xml:"BucketLoggingStatus"`
}

// LoggingEnabled defines the logging configuration information
type LoggingEnabled struct {
	XMLName      xml.Name `xml:"LoggingEnabled"`
	TargetBucket string   `xml:"TargetBucket"` // The bucket name for storing the log files
	TargetPrefix string   `xml:"TargetPrefix"` // The log file prefix
}

// GetBucketLoggingResult defines the result from GetBucketLogging request
type GetBucketLoggingResult LoggingXML

// WebsiteXML defines Website configuration
type WebsiteXML struct {
	XMLName       xml.Name      `xml:"WebsiteConfiguration"`
	IndexDocument IndexDocument `xml:"IndexDocument"` // The index page
	ErrorDocument ErrorDocument `xml:"ErrorDocument"` // The error page
}

// IndexDocument defines the index page info
type IndexDocument struct {
	XMLName xml.Name `xml:"IndexDocument"`
	Suffix  string   `xml:"Suffix"` // The file name for the index page
}

// ErrorDocument defines the 404 error page info
type ErrorDocument struct {
	XMLName xml.Name `xml:"ErrorDocument"`
	Key     string   `xml:"Key"` // 404 error file name
}

// GetBucketWebsiteResult defines the result from GetBucketWebsite request.
type GetBucketWebsiteResult WebsiteXML

// CORSXML defines CORS configuration
type CORSXML struct {
	XMLName   xml.Name   `xml:"CORSConfiguration"`
	CORSRules []CORSRule `xml:"CORSRule"` // CORS rules
}

// CORSRule defines CORS rules
type CORSRule struct {
	XMLName       xml.Name `xml:"CORSRule"`
	AllowedOrigin []string `xml:"AllowedOrigin"` // Allowed origins. By default it's wildcard '*'
	AllowedMethod []string `xml:"AllowedMethod"` // Allowed methods
	AllowedHeader []string `xml:"AllowedHeader"` // Allowed headers
	ExposeHeader  []string `xml:"ExposeHeader"`  // Allowed response headers
	MaxAgeSeconds int      `xml:"MaxAgeSeconds"` // Max cache ages in seconds
}

// GetBucketCORSResult defines the result from GetBucketCORS request.
type GetBucketCORSResult CORSXML

// GetBucketInfoResult defines the result from GetBucketInfo request.
type GetBucketInfoResult struct {
	XMLName    xml.Name   `xml:"BucketInfo"`
	BucketInfo BucketInfo `xml:"Bucket"`
}

// BucketInfo defines Bucket information
type BucketInfo struct {
	XMLName          xml.Name  `xml:"Bucket"`
	Name             string    `xml:"Name"`                    // Bucket name
	Location         string    `xml:"Location"`                // Bucket datacenter
	CreationDate     time.Time `xml:"CreationDate"`            // Bucket creation time
	ExtranetEndpoint string    `xml:"ExtranetEndpoint"`        // Bucket external endpoint
	IntranetEndpoint string    `xml:"IntranetEndpoint"`        // Bucket internal endpoint
	ACL              string    `xml:"AccessControlList>Grant"` // Bucket ACL
	Owner            Owner     `xml:"Owner"`                   // Bucket owner
	StorageClass     string    `xml:"StorageClass"`            // Bucket storage class
}

// ListObjectsResult defines the result from ListObjects request
type ListObjectsResult struct {
	XMLName        xml.Name           `xml:"ListBucketResult"`
	Prefix         string             `xml:"Prefix"`                // The object prefix
	Marker         string             `xml:"Marker"`                // The marker filter.
	MaxKeys        int                `xml:"MaxKeys"`               // Max keys to return
	Delimiter      string             `xml:"Delimiter"`             // The delimiter for grouping objects' name
	IsTruncated    bool               `xml:"IsTruncated"`           // Flag indicates if all results are returned (when it's false)
	NextMarker     string             `xml:"NextMarker"`            // The start point of the next query
	Objects        []ObjectProperties `xml:"Contents"`              // Object list
	CommonPrefixes []string           `xml:"CommonPrefixes>Prefix"` // You can think of commonprefixes as "folders" whose names end with the delimiter
}

// ObjectProperties defines Objecct properties
type ObjectProperties struct {
	XMLName      xml.Name  `xml:"Contents"`
	Key          string    `xml:"Key"`          // Object key
	Type         string    `xml:"Type"`         // Object type
	Size         int64     `xml:"Size"`         // Object size
	ETag         string    `xml:"ETag"`         // Object ETag
	Owner        Owner     `xml:"Owner"`        // Object owner information
	LastModified time.Time `xml:"LastModified"` // Object last modified time
	StorageClass string    `xml:"StorageClass"` // Object storage class (Standard, IA, Archive)
}

// Owner defines Bucket/Object's owner
type Owner struct {
	XMLName     xml.Name `xml:"Owner"`
	ID          string   `xml:"ID"`          // Owner ID
	DisplayName string   `xml:"DisplayName"` // Owner's display name
}

// CopyObjectResult defines result object of CopyObject
type CopyObjectResult struct {
	XMLName      xml.Name  `xml:"CopyObjectResult"`
	LastModified time.Time `xml:"LastModified"` // New object's last modified time.
	ETag         string    `xml:"ETag"`         // New object's ETag
}

// GetObjectACLResult defines result of GetObjectACL request
type GetObjectACLResult GetBucketACLResult

type deleteXML struct {
	XMLName xml.Name       `xml:"Delete"`
	Objects []DeleteObject `xml:"Object"` // Objects to delete
	Quiet   bool           `xml:"Quiet"`  // Flag of quiet mode.
}

// DeleteObject defines the struct for deleting object
type DeleteObject struct {
	XMLName xml.Name `xml:"Object"`
	Key     string   `xml:"Key"` // Object name
}

// DeleteObjectsResult defines result of DeleteObjects request
type DeleteObjectsResult struct {
	XMLName        xml.Name `xml:"DeleteResult"`
	DeletedObjects []string `xml:"Deleted>Key"` // Deleted object list
}

// InitiateMultipartUploadResult defines result of InitiateMultipartUpload request
type InitiateMultipartUploadResult struct {
	XMLName  xml.Name `xml:"InitiateMultipartUploadResult"`
	Bucket   string   `xml:"Bucket"`   // Bucket name
	Key      string   `xml:"Key"`      // Object name to upload
	UploadID string   `xml:"UploadId"` // Generated UploadId
}

// UploadPart defines the upload/copy part
type UploadPart struct {
	XMLName    xml.Name `xml:"Part"`
	PartNumber int      `xml:"PartNumber"` // Part number
	ETag       string   `xml:"ETag"`       // ETag value of the part's data
}

type uploadParts []UploadPart

func (slice uploadParts) Len() int {
	return len(slice)
}

func (slice uploadParts) Less(i, j int) bool {
	return slice[i].PartNumber < slice[j].PartNumber
}

func (slice uploadParts) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}

// UploadPartCopyResult defines result object of multipart copy request.
type UploadPartCopyResult struct {
	XMLName      xml.Name  `xml:"CopyPartResult"`
	LastModified time.Time `xml:"LastModified"` // Last modified time
	ETag         string    `xml:"ETag"`         // ETag
}

type completeMultipartUploadXML struct {
	XMLName xml.Name     `xml:"CompleteMultipartUpload"`
	Part    []UploadPart `xml:"Part"`
}

// CompleteMultipartUploadResult defines result object of CompleteMultipartUploadRequest
type CompleteMultipartUploadResult struct {
	XMLName  xml.Name `xml:"CompleteMultipartUploadResult"`
	Location string   `xml:"Location"` // Object URL
	Bucket   string   `xml:"Bucket"`   // Bucket name
	ETag     string   `xml:"ETag"`     // Object ETag
	Key      string   `xml:"Key"`      // Object name
}

// ListUploadedPartsResult defines result object of ListUploadedParts
type ListUploadedPartsResult struct {
	XMLName              xml.Name       `xml:"ListPartsResult"`
	Bucket               string         `xml:"Bucket"`               // Bucket name
	Key                  string         `xml:"Key"`                  // Object name
	UploadID             string         `xml:"UploadId"`             // Upload ID
	NextPartNumberMarker string         `xml:"NextPartNumberMarker"` // Next part number
	MaxParts             int            `xml:"MaxParts"`             // Max parts count
	IsTruncated          bool           `xml:"IsTruncated"`          // Flag indicates all entries returned.false: all entries returned.
	UploadedParts        []UploadedPart `xml:"Part"`                 // Uploaded parts
}

// UploadedPart defines uploaded part
type UploadedPart struct {
	XMLName      xml.Name  `xml:"Part"`
	PartNumber   int       `xml:"PartNumber"`   // Part number
	LastModified time.Time `xml:"LastModified"` // Last modified time
	ETag         string    `xml:"ETag"`         // ETag cache
	Size         int       `xml:"Size"`         // Part size
}

// ListMultipartUploadResult defines result object of ListMultipartUpload
type ListMultipartUploadResult struct {
	XMLName            xml.Name            `xml:"ListMultipartUploadsResult"`
	Bucket             string              `xml:"Bucket"`                // Bucket name
	Delimiter          string              `xml:"Delimiter"`             // Delimiter for grouping object.
	Prefix             string              `xml:"Prefix"`                // Object prefix
	KeyMarker          string              `xml:"KeyMarker"`             // Object key marker
	UploadIDMarker     string              `xml:"UploadIdMarker"`        // UploadId marker
	NextKeyMarker      string              `xml:"NextKeyMarker"`         // Next key marker, if not all entries returned.
	NextUploadIDMarker string              `xml:"NextUploadIdMarker"`    // Next uploadId marker, if not all entries returned.
	MaxUploads         int                 `xml:"MaxUploads"`            // Max uploads to return
	IsTruncated        bool                `xml:"IsTruncated"`           // Flag indicates all entries are returned.
	Uploads            []UncompletedUpload `xml:"Upload"`                // Ongoing uploads (not completed, not aborted)
	CommonPrefixes     []string            `xml:"CommonPrefixes>Prefix"` // Common prefixes list.
}

// UncompletedUpload structure wraps an uncompleted upload task
type UncompletedUpload struct {
	XMLName   xml.Name  `xml:"Upload"`
	Key       string    `xml:"Key"`       // Object name
	UploadID  string    `xml:"UploadId"`  // The UploadId
	Initiated time.Time `xml:"Initiated"` // Initialization time in the format such as 2012-02-23T04:18:23.000Z
}

// decodeDeleteObjectsResult decodes deleting objects result in URL encoding
func decodeDeleteObjectsResult(result *DeleteObjectsResult) error {
	var err error
	for i := 0; i < len(result.DeletedObjects); i++ {
		result.DeletedObjects[i], err = url.QueryUnescape(result.DeletedObjects[i])
		if err != nil {
			return err
		}
	}
	return nil
}

// decodeListObjectsResult decodes list objects result in URL encoding
func decodeListObjectsResult(result *ListObjectsResult) error {
	var err error
	result.Prefix, err = url.QueryUnescape(result.Prefix)
	if err != nil {
		return err
	}
	result.Marker, err = url.QueryUnescape(result.Marker)
	if err != nil {
		return err
	}
	result.Delimiter, err = url.QueryUnescape(result.Delimiter)
	if err != nil {
		return err
	}
	result.NextMarker, err = url.QueryUnescape(result.NextMarker)
	if err != nil {
		return err
	}
	for i := 0; i < len(result.Objects); i++ {
		result.Objects[i].Key, err = url.QueryUnescape(result.Objects[i].Key)
		if err != nil {
			return err
		}
	}
	for i := 0; i < len(result.CommonPrefixes); i++ {
		result.CommonPrefixes[i], err = url.QueryUnescape(result.CommonPrefixes[i])
		if err != nil {
			return err
		}
	}
	return nil
}

// decodeListMultipartUploadResult decodes list multipart upload result in URL encoding
func decodeListMultipartUploadResult(result *ListMultipartUploadResult) error {
	var err error
	result.Prefix, err = url.QueryUnescape(result.Prefix)
	if err != nil {
		return err
	}
	result.Delimiter, err = url.QueryUnescape(result.Delimiter)
	if err != nil {
		return err
	}
	result.KeyMarker, err = url.QueryUnescape(result.KeyMarker)
	if err != nil {
		return err
	}
	result.NextKeyMarker, err = url.QueryUnescape(result.NextKeyMarker)
	if err != nil {
		return err
	}
	for i := 0; i < len(result.Uploads); i++ {
		result.Uploads[i].Key, err = url.QueryUnescape(result.Uploads[i].Key)
		if err != nil {
			return err
		}
	}
	for i := 0; i < len(result.CommonPrefixes); i++ {
		result.CommonPrefixes[i], err = url.QueryUnescape(result.CommonPrefixes[i])
		if err != nil {
			return err
		}
	}
	return nil
}

// createBucketConfiguration defines the configuration for creating a bucket.
type createBucketConfiguration struct {
	XMLName      xml.Name         `xml:"CreateBucketConfiguration"`
	StorageClass StorageClassType `xml:"StorageClass,omitempty"`
}

// LiveChannelConfiguration livechannel的配置信息
type LiveChannelConfiguration struct {
	XMLName     xml.Name          `xml:"LiveChannelConfiguration"`
	Description string            `xml:"Description,omitempty"` //livechannel的描述信息，最长128字节
	Status      string            `xml:"Status,omitempty"`      //指定livechannel的状态
	Target      LiveChannelTarget `xml:"Target"`                //保存转储配置的容器
	//Snapshot    LiveChannelSnapshot `xml:"Snapshot,omitempty"` //保存高频截图操作Snapshot选项的容器
}

// LiveChannelTarget livechannel的转储配置信息
type LiveChannelTarget struct {
	XMLName      xml.Name `xml:"Target"`
	Type         string   `xml:"Type"`                   //指定转储的类型，只支持HLS
	FragDuration int      `xml:"FragDuration,omitempty"` //当Type为HLS时，指定每个ts文件的时长（单位：秒），取值范围为【1，100】的整数
	FragCount    int      `xml:"FragCount,omitempty"`    //当Type为HLS时，指定m3u8文件中包含ts文件的个数，取值范围为【1，100】的整数
	PlaylistName string   `xml:"PlaylistName,omitempty"` //当Type为HLS时，指定生成m3u8文件的名称，必须以“.m3u8”结尾，长度范围为【6，128】
}

/*
// LiveChannelSnapshot livechannel关于高频截图操作snapshot的配置信息
type LiveChannelSnapshot struct {
	XMLName     xml.Name `xml:"Snapshot"`
	RoleName    string   `xml:"RoleName,omitempty"`    //高频截图操作的角色名称，要求有DestBucket的写权限和向NotifyTopic发消息的权限
	DestBucket  string   `xml:"DestBucket,omitempty"`  //保存高频截图的目标Bucket，要求与当前Bucket是同一个Owner
	NotifyTopic string   `xml:"NotifyTopic,omitempty"` //用于通知用户高频截图操作结果的MNS的Topic
	Interval    int      `xml:"Interval,omitempty"`    //高频截图的间隔长度，单位为s，如果该段间隔时间内没有关键帧（I帧），那么该间隔时间不截图
}
*/

// CreateLiveChannelResult 创建livechannel请求的返回结果
type CreateLiveChannelResult struct {
	XMLName     xml.Name `xml:"CreateLiveChannelResult"`
	PublishUrls []string `xml:"PublishUrls>Url"` //推流地址列表
	PlayUrls    []string `xml:"PlayUrls>Url"`    //播放地址的列表
}

// LiveChannelStat 获取livechannel状态请求的返回结果
type LiveChannelStat struct {
	XMLName       xml.Name         `xml:"LiveChannelStat"`
	Status        string           `xml:"Status"`        //livechannel当前推流的状态，Disabled, Live, Idle
	ConnectedTime time.Time        `xml:"ConnectedTime"` //当Status为Live时，表示当前客户开始推流的时间，使用ISO8601格式表示
	RemoteAddr    string           `xml:"RemoteAddr"`    //当Status为Live时，表示当前推流客户端的ip地址
	Video         LiveChannelVideo `xml:"Video"`         //当Status为Live时，表示视频流的信息
	Audio         LiveChannelAudio `xml:"Audio"`         //当Status为Live时, 表示音频流的信息
}

// LiveChannelVideo 当livechannel的状态为live时，livechannel视频流的信息
type LiveChannelVideo struct {
	XMLName   xml.Name `xml:"Video"`
	Width     int      `xml:"Width"`     //视频流的画面宽度（单位：像素）
	Height    int      `xml:"Height"`    //视频流的画面高度（单位：像素）
	FrameRate int      `xml:"FrameRate"` //视频流的帧率
	Bandwidth int      `xml:"Bandwidth"` //视频流的码率（单位：B/s）
}

// LiveChannelAudio 当livechannel的状态为live时，livechannel音频流的信息
type LiveChannelAudio struct {
	XMLName    xml.Name `xml:"Audio"`
	SampleRate int      `xml:"SampleRate"` //音频流的采样率
	Bandwidth  int      `xml:"Bandwidth"`  //音频流的码率（单位：B/s）
	Codec      string   `xml:"Codec"`      //音频流的编码格式
}

// LiveChannelHistory - livechannel的历史所有推流记录，目前最多会返回指定livechannel最近10次的推流记录
type LiveChannelHistory struct {
	XMLName xml.Name     `xml:"LiveChannelHistory"`
	Record  []LiveRecord `xml:"LiveRecord"` //单个推流记录
}

// LiveRecord - 单个推流记录
type LiveRecord struct {
	XMLName    xml.Name  `xml:"LiveRecord"`
	StartTime  time.Time `xml:"StartTime"`  //推流开始时间，使用ISO8601格式表示
	EndTime    time.Time `xml:"EndTime"`    //推流结束时间，使用ISO8601格式表示
	RemoteAddr string    `xml:"RemoteAddr"` //推流客户端的ip地址
}

// ListLiveChannelResult -
type ListLiveChannelResult struct {
	XMLName     xml.Name          `xml:"ListLiveChannelResult"`
	Prefix      string            `xml:"Prefix"`      //返回以prifix作为前缀的livechannel，注意使用prifix查询时，返回的key中仍会包含prifix
	Marker      string            `xml:"Marker"`      //以marker之后按字母排序的第一个livechanel开始返回
	MaxKeys     int               `xml:"MaxKeys"`     //返回livechannel的最大数，如果不设定，默认为100，max-key的取值不能大于1000
	IsTruncated bool              `xml:"IsTruncated"` //指明是否所有的结果都已经返回，“true”表示本次没有返回全部结果，“false”则表示已经返回了全部结果
	NextMarker  string            `xml:"NextMarker"`  //如果本次没有返回全部结果，NextMarker用于表明下一请求的Marker值
	LiveChannel []LiveChannelInfo `xml:"LiveChannel"` //livechannel的基本信息
}

// LiveChannelInfo -
type LiveChannelInfo struct {
	XMLName      xml.Name  `xml:"LiveChannel"`
	Name         string    `xml:"Name"`            //livechannel的名称
	Description  string    `xml:"Description"`     //livechannel的描述信息
	Status       string    `xml:"Status"`          //livechannel的状态，有效值：disabled, enabled
	LastModified time.Time `xml:"LastModified"`    //livechannel的最后修改时间，使用ISO8601格式表示
	PublishUrls  []string  `xml:"PublishUrls>Url"` //推流地址列表
	PlayUrls     []string  `xml:"PlayUrls>Url"`    //播放地址列表
}
