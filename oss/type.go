package oss

import (
	"encoding/xml"
	"net/url"
	"time"
)

// ListBucketsResult The result object from ListBuckets request
type ListBucketsResult struct {
	XMLName     xml.Name           `xml:"ListAllMyBucketsResult"`
	Prefix      string             `xml:"Prefix"`         // the prefix in this query
	Marker      string             `xml:"Marker"`         // the marker filter.
	MaxKeys     int                `xml:"MaxKeys"`        // the max entry count to return. This information is returned when IsTruncated is true.
	IsTruncated bool               `xml:"IsTruncated"`    // flag True means there's remaining buckets to return.
	NextMarker  string             `xml:"NextMarker"`     // the marker filter for the next list call
	Owner       Owner              `xml:"Owner"`          // owner information
	Buckets     []BucketProperties `xml:"Buckets>Bucket"` // Bucket list
}

// BucketProperties Bucket properties
type BucketProperties struct {
	XMLName      xml.Name  `xml:"Bucket"`
	Name         string    `xml:"Name"`         // Bucket name
	Location     string    `xml:"Location"`     // Bucket datacenter
	CreationDate time.Time `xml:"CreationDate"` // Bucket create time
	StorageClass string    `xml:"StorageClass"` // Bucket storage class
}

// GetBucketACLResult GetBucketACL request's result
type GetBucketACLResult struct {
	XMLName xml.Name `xml:"AccessControlPolicy"`
	ACL     string   `xml:"AccessControlList>Grant"` // Bucket ACL
	Owner   Owner    `xml:"Owner"`                   // Bucket owner
}

// LifecycleConfiguration Bucket Lifecycle configuration
type LifecycleConfiguration struct {
	XMLName xml.Name        `xml:"LifecycleConfiguration"`
	Rules   []LifecycleRule `xml:"Rule"`
}

// LifecycleRule Lifecycle rules
type LifecycleRule struct {
	XMLName    xml.Name            `xml:"Rule"`
	ID         string              `xml:"ID"`         // Rule Id
	Prefix     string              `xml:"Prefix"`     // object key prefix
	Status     string              `xml:"Status"`     // the rule status (enabled or not)
	Expiration LifecycleExpiration `xml:"Expiration"` // the expiration property
}

// LifecycleExpiration the rule's expiration property
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

// BuildLifecycleRuleByDays Builds a lifecycle rule with specified expiration days
func BuildLifecycleRuleByDays(id, prefix string, status bool, days int) LifecycleRule {
	var statusStr = "Enabled"
	if !status {
		statusStr = "Disabled"
	}
	return LifecycleRule{ID: id, Prefix: prefix, Status: statusStr,
		Expiration: LifecycleExpiration{Days: days}}
}

// BuildLifecycleRuleByDate Builds a lifecycle rule with specified expiration time.
func BuildLifecycleRuleByDate(id, prefix string, status bool, year, month, day int) LifecycleRule {
	var statusStr = "Enabled"
	if !status {
		statusStr = "Disabled"
	}
	date := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	return LifecycleRule{ID: id, Prefix: prefix, Status: statusStr,
		Expiration: LifecycleExpiration{Date: date}}
}

// GetBucketLifecycleResult GetBucketLifecycle's result object
type GetBucketLifecycleResult LifecycleConfiguration

// RefererXML Referer config
type RefererXML struct {
	XMLName           xml.Name `xml:"RefererConfiguration"`
	AllowEmptyReferer bool     `xml:"AllowEmptyReferer"`   // Allow empty referrer
	RefererList       []string `xml:"RefererList>Referer"` // referer whitelist
}

// GetBucketRefererResult result object for GetBucketReferer request
type GetBucketRefererResult RefererXML

// LoggingXML Logging config
type LoggingXML struct {
	XMLName        xml.Name       `xml:"BucketLoggingStatus"`
	LoggingEnabled LoggingEnabled `xml:"LoggingEnabled"` // The logging  config information
}

type loggingXMLEmpty struct {
	XMLName xml.Name `xml:"BucketLoggingStatus"`
}

// LoggingEnabled the logging config information
type LoggingEnabled struct {
	XMLName      xml.Name `xml:"LoggingEnabled"`
	TargetBucket string   `xml:"TargetBucket"` // The bucket name for storing the log files
	TargetPrefix string   `xml:"TargetPrefix"` // The log file prefix
}

// GetBucketLoggingResult The result from GetBucketLogging request
type GetBucketLoggingResult LoggingXML

// WebsiteXML Website configuration
type WebsiteXML struct {
	XMLName       xml.Name      `xml:"WebsiteConfiguration"`
	IndexDocument IndexDocument `xml:"IndexDocument"` // the index page
	ErrorDocument ErrorDocument `xml:"ErrorDocument"` // the error page
}

// IndexDocument The index page info
type IndexDocument struct {
	XMLName xml.Name `xml:"IndexDocument"`
	Suffix  string   `xml:"Suffix"` // the file name for the index page
}

// ErrorDocument the 404 error page info
type ErrorDocument struct {
	XMLName xml.Name `xml:"ErrorDocument"`
	Key     string   `xml:"Key"` // 404 error file name
}

// GetBucketWebsiteResult The result from GetBucketWebsite request.
type GetBucketWebsiteResult WebsiteXML

// CORSXML CORS configuration
type CORSXML struct {
	XMLName   xml.Name   `xml:"CORSConfiguration"`
	CORSRules []CORSRule `xml:"CORSRule"` // CORS rules
}

// CORSRule CORS rules
type CORSRule struct {
	XMLName       xml.Name `xml:"CORSRule"`
	AllowedOrigin []string `xml:"AllowedOrigin"` // allowed origins. By default it's wildcard '*'
	AllowedMethod []string `xml:"AllowedMethod"` // allowed methods
	AllowedHeader []string `xml:"AllowedHeader"` // allowed headers
	ExposeHeader  []string `xml:"ExposeHeader"`  // allowed response headers
	MaxAgeSeconds int      `xml:"MaxAgeSeconds"` // max cache ages in seconds
}

// GetBucketCORSResult The result from GetBucketCORS request
type GetBucketCORSResult CORSXML

// GetBucketInfoResult The result from GetBucketInfo request.
type GetBucketInfoResult struct {
	XMLName    xml.Name   `xml:"BucketInfo"`
	BucketInfo BucketInfo `xml:"Bucket"`
}

// BucketInfo Bucket information
type BucketInfo struct {
	XMLName          xml.Name  `xml:"Bucket"`
	Name             string    `xml:"Name"`                    // Bucket name
	Location         string    `xml:"Location"`                // Bucket datacenter
	CreationDate     time.Time `xml:"CreationDate"`            // Bucket creation time
	ExtranetEndpoint string    `xml:"ExtranetEndpoint"`        // Bucket external endpoint
	IntranetEndpoint string    `xml:"IntranetEndpoint"`        // Bucket internal endpoint
	ACL              string    `xml:"AccessControlList>Grant"` // Bucket ACL
	Owner            Owner     `xml:"Owner"`                   // Bucket Owner
	StorageClass     string    `xml:"StorageClass"`            // Bucket storage class
}

// ListObjectsResult the result from ListObjects request
type ListObjectsResult struct {
	XMLName        xml.Name           `xml:"ListBucketResult"`
	Prefix         string             `xml:"Prefix"`                // The object prefix
	Marker         string             `xml:"Marker"`                // The marker filter.
	MaxKeys        int                `xml:"MaxKeys"`               // max keys to return
	Delimiter      string             `xml:"Delimiter"`             // the delimiter for grouping objects' name
	IsTruncated    bool               `xml:"IsTruncated"`           // flag indicates if all results are returned (when it's false)
	NextMarker     string             `xml:"NextMarker"`            // the start point of the next query
	Objects        []ObjectProperties `xml:"Contents"`              // Object list
	CommonPrefixes []string           `xml:"CommonPrefixes>Prefix"` // you can think of commonprefixes as "folders" whose names end with the delimiter
}

// ObjectProperties Objecct properties
type ObjectProperties struct {
	XMLName      xml.Name  `xml:"Contents"`
	Key          string    `xml:"Key"`          // Object Key
	Type         string    `xml:"Type"`         // Object Type
	Size         int64     `xml:"Size"`         // Object size
	ETag         string    `xml:"ETag"`         // Object ETag
	Owner        Owner     `xml:"Owner"`        // Object owner information
	LastModified time.Time `xml:"LastModified"` // Object last modified time
	StorageClass string    `xml:"StorageClass"` // Object storage class (Standard, IA, Archive)
}

// Owner Bucket/Object's owner
type Owner struct {
	XMLName     xml.Name `xml:"Owner"`
	ID          string   `xml:"ID"`          // owner Id
	DisplayName string   `xml:"DisplayName"` // Owner's display name
}

// CopyObjectResult result object of CopyObject
type CopyObjectResult struct {
	XMLName      xml.Name  `xml:"CopyObjectResult"`
	LastModified time.Time `xml:"LastModified"` // new Object's last modified time.
	ETag         string    `xml:"ETag"`         // new Object's ETag
}

// GetObjectACLResult result of GetObjectACL request
type GetObjectACLResult GetBucketACLResult

type deleteXML struct {
	XMLName xml.Name       `xml:"Delete"`
	Objects []DeleteObject `xml:"Object"` // objects to delete
	Quiet   bool           `xml:"Quiet"`  // flag of quiet mode.
}

// DeleteObject the struct for deleting object
type DeleteObject struct {
	XMLName xml.Name `xml:"Object"`
	Key     string   `xml:"Key"` // Object name
}

// DeleteObjectsResult result of DeleteObjects request
type DeleteObjectsResult struct {
	XMLName        xml.Name `xml:"DeleteResult"`
	DeletedObjects []string `xml:"Deleted>Key"` // deleted object list
}

// InitiateMultipartUploadResult result of InitiateMultipartUpload request
type InitiateMultipartUploadResult struct {
	XMLName  xml.Name `xml:"InitiateMultipartUploadResult"`
	Bucket   string   `xml:"Bucket"`   // Bucket name
	Key      string   `xml:"Key"`      // Object name to upload
	UploadID string   `xml:"UploadId"` // generated UploadId
}

// UploadPart the upload/copy part
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

// UploadPartCopyResult result object of multipart copy request.
type UploadPartCopyResult struct {
	XMLName      xml.Name  `xml:"CopyPartResult"`
	LastModified time.Time `xml:"LastModified"` // last modified time
	ETag         string    `xml:"ETag"`         // ETag
}

type completeMultipartUploadXML struct {
	XMLName xml.Name     `xml:"CompleteMultipartUpload"`
	Part    []UploadPart `xml:"Part"`
}

// CompleteMultipartUploadResult result object of CompleteMultipartUploadRequest
type CompleteMultipartUploadResult struct {
	XMLName  xml.Name `xml:"CompleteMultipartUploadResult"`
	Location string   `xml:"Location"` // Object URL
	Bucket   string   `xml:"Bucket"`   // Bucket name
	ETag     string   `xml:"ETag"`     // Object ETag
	Key      string   `xml:"Key"`      // Object name
}

// ListUploadedPartsResult result object of ListUploadedParts
type ListUploadedPartsResult struct {
	XMLName              xml.Name       `xml:"ListPartsResult"`
	Bucket               string         `xml:"Bucket"`               // Bucket name
	Key                  string         `xml:"Key"`                  // Object name
	UploadID             string         `xml:"UploadId"`             // upload Id
	NextPartNumberMarker string         `xml:"NextPartNumberMarker"` // Next Part number
	MaxParts             int            `xml:"MaxParts"`             // max parts count
	IsTruncated          bool           `xml:"IsTruncated"`          // flag indicates all entries returned.false: all entries returned.
	UploadedParts        []UploadedPart `xml:"Part"`                 // uploaded parts
}

// UploadedPart uploaded part
type UploadedPart struct {
	XMLName      xml.Name  `xml:"Part"`
	PartNumber   int       `xml:"PartNumber"`   // Part number
	LastModified time.Time `xml:"LastModified"` // last modified time
	ETag         string    `xml:"ETag"`         // ETag cache
	Size         int       `xml:"Size"`         // Part size
}

// ListMultipartUploadResult result object of ListMultipartUpload
type ListMultipartUploadResult struct {
	XMLName            xml.Name            `xml:"ListMultipartUploadsResult"`
	Bucket             string              `xml:"Bucket"`                // Bucket name
	Delimiter          string              `xml:"Delimiter"`             // Delimiter for grouping object.
	Prefix             string              `xml:"Prefix"`                // object prefix
	KeyMarker          string              `xml:"KeyMarker"`             // object key marker
	UploadIDMarker     string              `xml:"UploadIdMarker"`        // uploadId marker
	NextKeyMarker      string              `xml:"NextKeyMarker"`         // next key marker, if not all entries returned.
	NextUploadIDMarker string              `xml:"NextUploadIdMarker"`    // next uploadId marker, if not all entries returned.
	MaxUploads         int                 `xml:"MaxUploads"`            // max uploads to return
	IsTruncated        bool                `xml:"IsTruncated"`           // flag indicates all entries are returned.
	Uploads            []UncompletedUpload `xml:"Upload"`                // ongoing uploads (not completed, not aborted)
	CommonPrefixes     []string            `xml:"CommonPrefixes>Prefix"` // common prefixes list.
}

// UncompletedUpload structure wraps an uncompleted Upload task
type UncompletedUpload struct {
	XMLName   xml.Name  `xml:"Upload"`
	Key       string    `xml:"Key"`       // Object name
	UploadID  string    `xml:"UploadId"`  // the UploadId
	Initiated time.Time `xml:"Initiated"` // initialization time in the format such as 2012-02-23T04:18:23.000Z
}

// decode deleting objects result in URL encoding
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

// decode list objects result in URL encoding
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

// decode list multipart upload result in URL encoding
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

// createBucketConfiguration the configuration for creating a bucket.
type createBucketConfiguration struct {
	XMLName      xml.Name         `xml:"CreateBucketConfiguration"`
	StorageClass StorageClassType `xml:"StorageClass,omitempty"`
}
