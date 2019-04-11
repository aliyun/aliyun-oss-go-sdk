package oss

import (
	"net/url"
	"sort"

	. "gopkg.in/check.v1"
)

type OssTypeSuite struct{}

var _ = Suite(&OssTypeSuite{})

var (
	goStr     = "go go + go <> go"
	chnStr    = "试问闲情几许"
	goURLStr  = url.QueryEscape(goStr)
	chnURLStr = url.QueryEscape(chnStr)
)

func (s *OssTypeSuite) TestDecodeDeleteObjectsResult(c *C) {
	var res DeleteObjectsResult
	err := decodeDeleteObjectsResult(&res)
	c.Assert(err, IsNil)

	res.DeletedObjects = []string{""}
	err = decodeDeleteObjectsResult(&res)
	c.Assert(err, IsNil)
	c.Assert(res.DeletedObjects[0], Equals, "")

	res.DeletedObjects = []string{goURLStr, chnURLStr}
	err = decodeDeleteObjectsResult(&res)
	c.Assert(err, IsNil)
	c.Assert(res.DeletedObjects[0], Equals, goStr)
	c.Assert(res.DeletedObjects[1], Equals, chnStr)
}

func (s *OssTypeSuite) TestDecodeListObjectsResult(c *C) {
	var res ListObjectsResult
	err := decodeListObjectsResult(&res)
	c.Assert(err, IsNil)

	res = ListObjectsResult{}
	err = decodeListObjectsResult(&res)
	c.Assert(err, IsNil)

	res = ListObjectsResult{Prefix: goURLStr, Marker: goURLStr,
		Delimiter: goURLStr, NextMarker: goURLStr,
		Objects:        []ObjectProperties{{Key: chnURLStr}},
		CommonPrefixes: []string{chnURLStr}}

	err = decodeListObjectsResult(&res)
	c.Assert(err, IsNil)

	c.Assert(res.Prefix, Equals, goStr)
	c.Assert(res.Marker, Equals, goStr)
	c.Assert(res.Delimiter, Equals, goStr)
	c.Assert(res.NextMarker, Equals, goStr)
	c.Assert(res.Objects[0].Key, Equals, chnStr)
	c.Assert(res.CommonPrefixes[0], Equals, chnStr)
}

func (s *OssTypeSuite) TestDecodeListMultipartUploadResult(c *C) {
	res := ListMultipartUploadResult{}
	err := decodeListMultipartUploadResult(&res)
	c.Assert(err, IsNil)

	res = ListMultipartUploadResult{Prefix: goURLStr, KeyMarker: goURLStr,
		Delimiter: goURLStr, NextKeyMarker: goURLStr,
		Uploads: []UncompletedUpload{{Key: chnURLStr}}}

	err = decodeListMultipartUploadResult(&res)
	c.Assert(err, IsNil)

	c.Assert(res.Prefix, Equals, goStr)
	c.Assert(res.KeyMarker, Equals, goStr)
	c.Assert(res.Delimiter, Equals, goStr)
	c.Assert(res.NextKeyMarker, Equals, goStr)
	c.Assert(res.Uploads[0].Key, Equals, chnStr)
}

func (s *OssTypeSuite) TestSortUploadPart(c *C) {
	parts := []UploadPart{}

	sort.Sort(uploadParts(parts))
	c.Assert(len(parts), Equals, 0)

	parts = []UploadPart{
		{PartNumber: 5, ETag: "E5"},
		{PartNumber: 1, ETag: "E1"},
		{PartNumber: 4, ETag: "E4"},
		{PartNumber: 2, ETag: "E2"},
		{PartNumber: 3, ETag: "E3"},
	}

	sort.Sort(uploadParts(parts))

	c.Assert(parts[0].PartNumber, Equals, 1)
	c.Assert(parts[0].ETag, Equals, "E1")
	c.Assert(parts[1].PartNumber, Equals, 2)
	c.Assert(parts[1].ETag, Equals, "E2")
	c.Assert(parts[2].PartNumber, Equals, 3)
	c.Assert(parts[2].ETag, Equals, "E3")
	c.Assert(parts[3].PartNumber, Equals, 4)
	c.Assert(parts[3].ETag, Equals, "E4")
	c.Assert(parts[4].PartNumber, Equals, 5)
	c.Assert(parts[4].ETag, Equals, "E5")
}

func (s *OssTypeSuite) TestNewLifecleRuleByDays(c *C) {
	_, err := NewLifecycleRuleByDays("rule1", "one", true, 30, LRTExpriration)
	c.Assert(err, IsNil)

	_, err = NewLifecycleRuleByDays("rule2", "two", true, 30, LRTAbortMultiPartUpload)
	c.Assert(err, IsNil)

	_, err = NewLifecycleRuleByDays("rule3", "three", true, 30, LRTTransition, StorageIA)
	c.Assert(err, IsNil)

	_, err = NewLifecycleRuleByDays("rule4", "four", true, 30, LRTTransition, StorageArchive)
	c.Assert(err, IsNil)

	// expiration lifecycle type, set storage class type
	_, err = NewLifecycleRuleByDays("rule5", "five", true, 30, LRTExpriration, StorageIA)
	c.Assert(err, NotNil)

	// abort multipart upload lifecycle type, set storage class type
	_, err = NewLifecycleRuleByDays("rule6", "six", true, 30, LRTAbortMultiPartUpload, StorageIA)
	c.Assert(err, NotNil)

	// transition lifecycle type, the value of storage class type is StorageStandard
	_, err = NewLifecycleRuleByDays("rule7", "seven", true, 30, LRTTransition, StorageStandard)
	c.Assert(err, NotNil)

	// transition lifecycle type, do not set storage class type
	_, err = NewLifecycleRuleByDays("rule8", "eight", true, 30, LRTTransition)
	c.Assert(err, NotNil)

	// transition lifecycle type，set two storage class type
	_, err = NewLifecycleRuleByDays("rule9", "nine", true, 30, LRTTransition, StorageIA, StorageArchive)
	c.Assert(err, NotNil)
}

func (s *OssTypeSuite) TestNewLifecycleRuleByCreateBeforeDate(c *C) {
	_, err := NewLifecycleRuleByCreateBeforeDate("rule1", "one", true, 2019, 3, 30, LRTExpriration)
	c.Assert(err, IsNil)

	_, err = NewLifecycleRuleByCreateBeforeDate("rule2", "two", true, 2019, 3, 30, LRTAbortMultiPartUpload)
	c.Assert(err, IsNil)

	_, err = NewLifecycleRuleByCreateBeforeDate("rule3", "three", true, 2019, 3, 30, LRTTransition, StorageIA)
	c.Assert(err, IsNil)

	_, err = NewLifecycleRuleByCreateBeforeDate("rule4", "four", true, 2019, 3, 30, LRTTransition, StorageArchive)
	c.Assert(err, IsNil)

	// expiration lifecycle type, set storage class type
	_, err = NewLifecycleRuleByCreateBeforeDate("rule5", "five", true, 2019, 3, 30, LRTExpriration, StorageIA)
	c.Assert(err, NotNil)

	// abort multipart upload lifecycle type, set storage class type
	_, err = NewLifecycleRuleByCreateBeforeDate("rule6", "six", true, 2019, 3, 30, LRTAbortMultiPartUpload, StorageIA)
	c.Assert(err, NotNil)

	// transition lifecycle type, the value of storage class type is StorageStandard
	_, err = NewLifecycleRuleByCreateBeforeDate("rule7", "seven", true, 2019, 3, 30, LRTTransition, StorageStandard)
	c.Assert(err, NotNil)

	// transition lifecycle type, do not set storage class type
	_, err = NewLifecycleRuleByCreateBeforeDate("rule8", "eight", true, 2019, 3, 30, LRTTransition)
	c.Assert(err, NotNil)

	// transition lifecycle type，set two storage class type
	_, err = NewLifecycleRuleByCreateBeforeDate("rule9", "nine", true, 2019, 3, 30, LRTTransition, StorageIA, StorageArchive)
	c.Assert(err, NotNil)
}
