package oss

import (
	"context"
	"encoding/xml"
	"time"
)

func (c *Client) ListBuckets(ctx context.Context, request *ListBucketsRequest, optFns ...func(*Options)) (result *ListBucketsResult, err error) {
	if request == nil {
		request = &ListBucketsRequest{}
	}
	input := &OperationInput{
		OperationName: "ListBuckets",
		Method:        "GET",
		Headers:       request.Headers,
		Parameters:    request.Parameters,
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
	result = &ListBucketsResult{
		ResultCommon: ResultCommon{
			Status:     output.Status,
			StatusCode: output.StatusCode,
			Headers:    output.Headers,
			Metadata:   output.Metadata,
		},
	}
	err = xml.NewDecoder(output.Body).Decode(result)
	return result, err
}

type ListBucketsRequest struct {
	RequestCommon
	Marker          string `input:"query,marker"`   // Marker is where you want server to start listing from
	MaxKeys         int32  `input:"query,max-keys"` // Sets the maximum number of keys returned in the response
	Prefix          string `input:"query,prefix"`   // Limits the response to keys that begin with the specified prefix
	ResourceGroupId string `input:"header,x-oss-resource-group-id"`
}

type ListBucketsResult struct {
	ResultCommon
	XMLName     xml.Name           `xml:"ListAllMyBucketsResult"`
	Prefix      string             `xml:"Prefix"`          // The object prefix
	Marker      string             `xml:"Marker"`          // The marker filter.
	MaxKeys     int                `xml:"MaxKeys"`         // Max keys to return
	IsTruncated bool               `xml:"IsTruncated"`     // Flag indicates if all results are returned (when it's false)
	NextMarker  string             `xml:"NextMarker"`      // The start point of the next query
	Owner       Owner              `xml:"Owner,omitempty"` // Object owner information
	Buckets     []BucketProperties `xml:"Buckets>Bucket"`  // You can think of commonprefixes as "folders" whose names end with the delimiter
}

type BucketProperties struct {
	//XMLName          xml.Name  `xml:"Bucket"`
	Name             string    `xml:"Name"`             // Bucket name
	Location         string    `xml:"Location"`         // Bucket datacenter
	CreationDate     time.Time `xml:"CreationDate"`     // Bucket create time
	StorageClass     string    `xml:"StorageClass"`     // Bucket storage class
	Region           string    `xml:"Region"`           // Bucket storage class
	ExtranetEndpoint string    `xml:"ExtranetEndpoint"` // Bucket name
	IntranetEndpoint string    `xml:"IntranetEndpoint"` // Bucket name
}
