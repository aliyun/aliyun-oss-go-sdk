package oss

import (
	"math"
	"net/http"
	"strings"

	. "gopkg.in/check.v1"
)

type OssErrorSuite struct{}

var _ = Suite(&OssErrorSuite{})

func (s *OssErrorSuite) TestCheckCRCHasCRCInResp(c *C) {
	headers := http.Header{
		"Expires":              {"-1"},
		"Content-Length":       {"0"},
		"Content-Encoding":     {"gzip"},
		"X-Oss-Hash-Crc64ecma": {"0"},
	}

	resp := &Response{
		StatusCode: 200,
		Headers:    headers,
		Body:       nil,
		ClientCRC:  math.MaxUint64,
		ServerCRC:  math.MaxUint64,
	}

	err := CheckCRC(resp, "test")
	c.Assert(err, IsNil)
}

func (s *OssErrorSuite) TestCheckCRCNotHasCRCInResp(c *C) {
	headers := http.Header{
		"Expires":          {"-1"},
		"Content-Length":   {"0"},
		"Content-Encoding": {"gzip"},
	}

	resp := &Response{
		StatusCode: 200,
		Headers:    headers,
		Body:       nil,
		ClientCRC:  math.MaxUint64,
		ServerCRC:  math.MaxUint32,
	}

	err := CheckCRC(resp, "test")
	c.Assert(err, IsNil)
}

func (s *OssErrorSuite) TestCheckCRCCNegative(c *C) {
	headers := http.Header{
		"Expires":              {"-1"},
		"Content-Length":       {"0"},
		"Content-Encoding":     {"gzip"},
		"X-Oss-Hash-Crc64ecma": {"0"},
	}

	resp := &Response{
		StatusCode: 200,
		Headers:    headers,
		Body:       nil,
		ClientCRC:  0,
		ServerCRC:  math.MaxUint64,
	}

	err := CheckCRC(resp, "test")
	c.Assert(err, NotNil)
	testLogger.Println("error:", err)
}

func (s *OssErrorSuite) TestCheckDownloadCRC(c *C) {
	err := CheckDownloadCRC(0xFBF9D9603A6FA020, 0xFBF9D9603A6FA020)
	c.Assert(err, IsNil)

	err = CheckDownloadCRC(0, 0)
	c.Assert(err, IsNil)

	err = CheckDownloadCRC(0xDB6EFFF26AA94946, 0)
	c.Assert(err, NotNil)
	testLogger.Println("error:", err)
}

func (s *OssErrorSuite) TestServiceErrorEndPoint(c *C) {
	xmlBody := `<?xml version="1.0" encoding="UTF-8"?>
	<Error>
	  <Code>AccessDenied</Code>
	  <Message>The bucket you visit is not belong to you.</Message>
	  <RequestId>5C1B5E9BD79A6B9B6466166E</RequestId>
	  <HostId>oss-c-sdk-test-verify-b.oss-cn-shenzhen.aliyuncs.com</HostId>
	</Error>`
	serverError, _ := serviceErrFromXML([]byte(xmlBody), 403, "5C1B5E9BD79A6B9B6466166E")
	errMsg := serverError.Error()
	c.Assert(strings.Contains(errMsg, "Endpoint="), Equals, false)

	xmlBodyWithEndPoint := `<?xml version="1.0" encoding="UTF-8"?>
	<Error>
      <Code>AccessDenied</Code>
	  <Message>The bucket you are attempting to access must be addressed using the specified endpoint. Please send all future requests to this endpoint.</Message>
	  <RequestId>5C1B595ED51820B569C6A12F</RequestId>
	  <HostId>hello-hangzws.oss-cn-qingdao.aliyuncs.com</HostId>
	  <Bucket>hello-hangzws</Bucket>
	  <Endpoint>oss-cn-shenzhen.aliyuncs.com</Endpoint>
	</Error>`
	serverError, _ = serviceErrFromXML([]byte(xmlBodyWithEndPoint), 406, "5C1B595ED51820B569C6A12F")
	errMsg = serverError.Error()
	c.Assert(strings.Contains(errMsg, "Endpoint=oss-cn-shenzhen.aliyuncs.com"), Equals, true)
}
