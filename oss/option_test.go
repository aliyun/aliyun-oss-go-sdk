package oss

import (
	"context"
	"net/http"
	"time"

	. "gopkg.in/check.v1"
)

type OssOptionSuite struct{}

var _ = Suite(&OssOptionSuite{})

type optionTestCase struct {
	option Option
	key    string
	value  string
}

var headerTestcases = []optionTestCase{
	{
		option: Meta("User", "baymax"),
		key:    "X-Oss-Meta-User",
		value:  "baymax",
	},
	{
		option: ACL(ACLPrivate),
		key:    "X-Oss-Acl",
		value:  "private",
	},
	{
		option: ContentType("plain/text"),
		key:    "Content-Type",
		value:  "plain/text",
	},
	{
		option: CacheControl("no-cache"),
		key:    "Cache-Control",
		value:  "no-cache",
	},
	{
		option: ContentDisposition("Attachment; filename=example.txt"),
		key:    "Content-Disposition",
		value:  "Attachment; filename=example.txt",
	},
	{
		option: ContentEncoding("gzip"),
		key:    "Content-Encoding",
		value:  "gzip",
	},
	{
		option: Expires(pastDate),
		key:    "Expires",
		value:  pastDate.Format(http.TimeFormat),
	},
	{
		option: Range(0, 9),
		key:    "Range",
		value:  "bytes=0-9",
	},
	{
		option: Origin("localhost"),
		key:    "Origin",
		value:  "localhost",
	},
	{
		option: CopySourceRange(0, 9),
		key:    "X-Oss-Copy-Source-Range",
		value:  "bytes=0-8",
	},
	{
		option: IfModifiedSince(pastDate),
		key:    "If-Modified-Since",
		value:  pastDate.Format(http.TimeFormat),
	},
	{
		option: IfUnmodifiedSince(futureDate),
		key:    "If-Unmodified-Since",
		value:  futureDate.Format(http.TimeFormat),
	},
	{
		option: IfMatch("xyzzy"),
		key:    "If-Match",
		value:  "xyzzy",
	},
	{
		option: IfNoneMatch("xyzzy"),
		key:    "If-None-Match",
		value:  "xyzzy",
	},
	{
		option: CopySource("bucket_name", "object_name"),
		key:    "X-Oss-Copy-Source",
		value:  "/bucket_name/object_name",
	},
	{
		option: CopySourceIfModifiedSince(pastDate),
		key:    "X-Oss-Copy-Source-If-Modified-Since",
		value:  pastDate.Format(http.TimeFormat),
	},
	{
		option: CopySourceIfUnmodifiedSince(futureDate),
		key:    "X-Oss-Copy-Source-If-Unmodified-Since",
		value:  futureDate.Format(http.TimeFormat),
	},
	{
		option: CopySourceIfMatch("xyzzy"),
		key:    "X-Oss-Copy-Source-If-Match",
		value:  "xyzzy",
	},
	{
		option: CopySourceIfNoneMatch("xyzzy"),
		key:    "X-Oss-Copy-Source-If-None-Match",
		value:  "xyzzy",
	},
	{
		option: MetadataDirective(MetaCopy),
		key:    "X-Oss-Metadata-Directive",
		value:  "COPY",
	},
	{
		option: ServerSideEncryption("AES256"),
		key:    "X-Oss-Server-Side-Encryption",
		value:  "AES256",
	},
	{
		option: ObjectACL(ACLPrivate),
		key:    "X-Oss-Object-Acl",
		value:  "private",
	},
	{
		option: ObjectStorageClass(StorageStandard),
		key:    "X-Oss-Storage-Class",
		value:  "Standard",
	},
	{
		option: Callback("JTdCJTIyY2FsbGJhY2tVcmwlMjIlM0ElMjJleGFtcGxlLmNvbS9pbmRleC5odG1sJTIyJTdE"),
		key:    "X-Oss-Callback",
		value:  "JTdCJTIyY2FsbGJhY2tVcmwlMjIlM0ElMjJleGFtcGxlLmNvbS9pbmRleC5odG1sJTIyJTdE",
	},
	{
		option: CallbackVar("JTdCJTIyeCUzQXZhcjElMjIlM0ElMjJ2YWx1ZTElMjIlMkMlMjJ4JTNBdmFyMiUyMiUzQSUyMnZhbHVlMiUyMiU3RA=="),
		key:    "X-Oss-Callback-Var",
		value:  "JTdCJTIyeCUzQXZhcjElMjIlM0ElMjJ2YWx1ZTElMjIlMkMlMjJ4JTNBdmFyMiUyMiUzQSUyMnZhbHVlMiUyMiU3RA==",
	},
	{
		option: ContentLanguage("zh-CN"),
		key:    "Content-Language",
		value:  "zh-CN",
	},
	{
		option: ServerSideEncryptionKeyID("xossekid"),
		key:    "X-Oss-Server-Side-Encryption-Key-Id",
		value:  "xossekid",
	},
}

func (s *OssOptionSuite) TestHeaderOptions(c *C) {
	for _, testcase := range headerTestcases {
		headers := make(map[string]optionValue)
		err := testcase.option(headers)
		c.Assert(err, IsNil)

		expected, actual := testcase.value, headers[testcase.key].Value
		c.Assert(expected, Equals, actual)
	}
}

var paramTestCases = []optionTestCase{
	{
		option: Delimiter("/"),
		key:    "delimiter",
		value:  "/",
	},
	{
		option: Marker("abc"),
		key:    "marker",
		value:  "abc",
	},
	{
		option: MaxKeys(150),
		key:    "max-keys",
		value:  "150",
	},
	{
		option: Prefix("fun"),
		key:    "prefix",
		value:  "fun",
	},
	{
		option: EncodingType("ascii"),
		key:    "encoding-type",
		value:  "ascii",
	},
	{
		option: MaxUploads(100),
		key:    "max-uploads",
		value:  "100",
	},
	{
		option: KeyMarker("abc"),
		key:    "key-marker",
		value:  "abc",
	},
	{
		option: UploadIDMarker("xyz"),
		key:    "upload-id-marker",
		value:  "xyz",
	},
	{
		option: MaxParts(1000),
		key:    "max-parts",
		value:  "1000",
	},
	{
		option: PartNumberMarker(1),
		key:    "part-number-marker",
		value:  "1",
	},
	{
		option: Process("image/format,png"),
		key:    "x-oss-process",
		value:  "image/format,png",
	},
}

func (s *OssOptionSuite) TestParamOptions(c *C) {
	for _, testcase := range paramTestCases {
		params := make(map[string]optionValue)
		err := testcase.option(params)
		c.Assert(err, IsNil)

		expected, actual := testcase.value, params[testcase.key].Value
		c.Assert(expected, Equals, actual)
	}
}

func (s *OssOptionSuite) TestHandleOptions(c *C) {
	headers := make(map[string]string)
	options := []Option{}

	for _, testcase := range headerTestcases {
		options = append(options, testcase.option)
	}

	err := handleOptions(headers, options)
	c.Assert(err, IsNil)

	for _, testcase := range headerTestcases {
		expected, actual := testcase.value, headers[testcase.key]
		c.Assert(expected, Equals, actual)
	}

	options = []Option{IfMatch(""), nil}
	headers = map[string]string{}
	err = handleOptions(headers, options)
	c.Assert(err, IsNil)
	c.Assert(len(headers), Equals, 1)
}

func (s *OssOptionSuite) TestHandleParams(c *C) {
	client, err := New(endpoint, accessID, accessKey)
	c.Assert(err, IsNil)

	options := []Option{}

	for _, testcase := range paramTestCases {
		options = append(options, testcase.option)
	}

	params, err := GetRawParams(options)
	c.Assert(err, IsNil)

	out := client.Conn.getURLParams(params)
	c.Assert(len(out), Equals, 191)

	options = []Option{KeyMarker(""), nil}

	params, err = GetRawParams(options)
	c.Assert(err, IsNil)

	out = client.Conn.getURLParams(params)
	c.Assert(out, Equals, "key-marker")
}

func (s *OssOptionSuite) TestFindOption(c *C) {
	options := []Option{}

	for _, testcase := range headerTestcases {
		options = append(options, testcase.option)
	}

	str, err := FindOption(options, "X-Oss-Acl", "")
	c.Assert(err, IsNil)
	c.Assert(str, Equals, "private")

	str, err = FindOption(options, "MyProp", "")
	c.Assert(err, IsNil)
	c.Assert(str, Equals, "")
}

func (s *OssOptionSuite) TestDeleteOption(c *C) {
	options := []Option{VersionId("123"), VersionIdMarker("456"), KeyMarker("789")}
	str, err := FindOption(options, "versionId", "")
	c.Assert(str, Equals, "123")
	c.Assert(err, IsNil)

	skipOption := DeleteOption(options, "versionId")
	str, err = FindOption(skipOption, "versionId", "")
	c.Assert(str, Equals, "")

	str, err = FindOption(skipOption, "version-id-marker", "")
	c.Assert(str, Equals, "456")

	str, err = FindOption(skipOption, "key-marker", "")
	c.Assert(str, Equals, "789")

}

func (s *OssOptionSuite) TestWithContext(c *C) {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	options := []Option{WithContext(ctx)}
	ctxArg, _ := FindOption(options, contextArg, nil)

	c.Assert(ctxArg, NotNil)
	c.Assert(ctxArg.(context.Context), Equals, ctx)

	options = []Option{}
	ctxArg, _ = FindOption(options, contextArg, nil)
	c.Assert(ctxArg, Equals, nil)
	ctxnil, _ := ctxArg.(context.Context)
	c.Assert(ctxnil, Equals, nil)
}
