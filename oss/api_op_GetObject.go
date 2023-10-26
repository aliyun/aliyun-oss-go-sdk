package oss

import (
	"context"
	"io"
)

func (c *Client) GetObject(ctx context.Context, request *GetObjectRequest, optFns ...func(*Options)) (result *GetObjectResult, err error) {
	if request == nil {
		request = &GetObjectRequest{}
	}
	input := &OperationInput{
		OperationName: "GetObject",
		Method:        "GET",
		Bucket:        request.Bucket,
		Key:           request.Key,
		Headers:       request.Headers,
		Parameters:    request.Parameters,
	}
	if err = validateObjectOpsInput(input); err != nil {
		return nil, err
	}
	if err = c.MarshalInput(request, input); err != nil {
		return nil, err
	}
	output, err := c.invokeOperation(ctx, input, optFns)
	if err != nil {
		return nil, err
	}
	result = &GetObjectResult{
		ResultCommon: ResultCommon{
			Status:     output.Status,
			StatusCode: output.StatusCode,
			Headers:    output.Headers,
			Metadata:   output.Metadata,
		},
		Body: output.Body,
	}
	return result, err
}

type GetObjectRequest struct {
	RequestCommon
	Bucket string `input:"host,bucket,required"` // The name of the bucket containing the objects
	Key    string `input:"path,key,required"`    // The name of the object
}

type GetObjectResult struct {
	ResultCommon
	Body io.ReadCloser
}
