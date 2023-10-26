package oss

import (
	"context"
	"encoding/xml"
	"net/url"
	"strings"
	"time"
)

func (c *Client) ListObjects(ctx context.Context, request *ListObjectsRequest, optFns ...func(*Options)) (result *ListObjectsResult, err error) {
	if request == nil {
		request = &ListObjectsRequest{}
	}
	input := &OperationInput{
		OperationName: "ListObjects",
		Bucket:        request.Bucket,
		Method:        "GET",
		Headers:       request.Headers,
		Parameters:    request.Parameters,
	}
	if err = validateBucketOpsInput(input); err != nil {
		return nil, err
	}
	if err = c.MarshalInput(request, input); err != nil {
		return nil, err
	}
	output, err := c.invokeOperation(ctx, input, optFns)
	if err != nil {
		return nil, err
	}
	if output.Body != nil {
		defer output.Body.Close()
	}
	result = &ListObjectsResult{
		ResultCommon: ResultCommon{
			Status:     output.Status,
			StatusCode: output.StatusCode,
			Headers:    output.Headers,
			Metadata:   output.Metadata,
		},
	}
	err = xml.NewDecoder(output.Body).Decode(result)
	if err == nil {
		err = postProcessResult(result)
	}
	return result, err
}

type ListObjectsRequest struct {
	RequestCommon
	Bucket       string `input:"host,bucket,required"` // The name of the bucket containing the objects
	Delimiter    string `input:"query,delimiter"`      // A delimiter is a character that you use to group keys
	EncodingType string `input:"query,encoding-type"`  // Requests Server to encode the object keys in the respons
	Marker       string `input:"query,marker"`         // Marker is where you want server to start listing from
	MaxKeys      int32  `input:"query,max-keys"`       // Sets the maximum number of keys returned in the response
	Prefix       string `input:"query,prefix"`         // Limits the response to keys that begin with the specified prefix
}

type ListObjectsResult struct {
	ResultCommon
	XMLName        xml.Name           `xml:"ListBucketResult"`
	Name           string             `xml:"Name"`           // The bucket name
	Prefix         string             `xml:"Prefix"`         // The object prefix
	Marker         string             `xml:"Marker"`         // The marker filter.
	MaxKeys        int                `xml:"MaxKeys"`        // Max keys to return
	Delimiter      string             `xml:"Delimiter"`      // The delimiter for grouping objects' name
	IsTruncated    bool               `xml:"IsTruncated"`    // Flag indicates if all results are returned (when it's false)
	NextMarker     string             `xml:"NextMarker"`     // The start point of the next query
	EncodingType   string             `xml:"EncodingType"`   // The encoding type of the content in the response
	Contents       []ObjectProperties `xml:"Contents"`       // Metadata about each object
	CommonPrefixes []CommonPrefix     `xml:"CommonPrefixes"` // You can think of commonprefixes as "folders" whose names end with the delimiter
}

type ObjectProperties struct {
	//XMLName      xml.Name  `xml:"Contents"`
	Key          string    `xml:"Key"`                   // Object key
	Type         string    `xml:"Type"`                  // Object type
	Size         int64     `xml:"Size"`                  // Object size
	ETag         string    `xml:"ETag"`                  // Object ETag
	Owner        Owner     `xml:"Owner,omitempty"`       // Object owner information
	LastModified time.Time `xml:"LastModified"`          // Object last modified time
	StorageClass string    `xml:"StorageClass"`          // Object storage class (Standard, IA, Archive)
	RestoreInfo  string    `xml:"RestoreInfo,omitempty"` // Object restoreInfo
}

type Owner struct {
	//XMLName     xml.Name `xml:"Owner"`
	ID          string `xml:"ID"`          // Owner ID
	DisplayName string `xml:"DisplayName"` // Owner's display name
}

type CommonPrefix struct {
	//XMLName xml.Name `xml:"CommonPrefix"`
	Prefix string `xml:"Prefix"` // The prefix contained in the returned object names.
}

func postProcessResult(result *ListObjectsResult) (err error) {
	if result == nil || !strings.EqualFold(result.EncodingType, "url") {
		return nil
	}

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
	for i := 0; i < len(result.Contents); i++ {
		result.Contents[i].Key, err = url.QueryUnescape(result.Contents[i].Key)
		if err != nil {
			return err
		}
	}
	for i := 0; i < len(result.CommonPrefixes); i++ {
		result.CommonPrefixes[i].Prefix, err = url.QueryUnescape(result.CommonPrefixes[i].Prefix)
		if err != nil {
			return err
		}
	}
	return nil
}
