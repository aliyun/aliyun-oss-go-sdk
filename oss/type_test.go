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
	var res DeleteObjectVersionsResult
	err := decodeDeleteObjectsResult(&res)
	c.Assert(err, IsNil)

	res.DeletedObjectsDetail = []DeletedKeyInfo{DeletedKeyInfo{Key: ""}}
	err = decodeDeleteObjectsResult(&res)
	c.Assert(err, IsNil)
	c.Assert(res.DeletedObjectsDetail[0].Key, Equals, "")

	res.DeletedObjectsDetail = []DeletedKeyInfo{DeletedKeyInfo{Key: goURLStr}, DeletedKeyInfo{Key: chnURLStr}}
	err = decodeDeleteObjectsResult(&res)
	c.Assert(err, IsNil)
	c.Assert(res.DeletedObjectsDetail[0].Key, Equals, goStr)
	c.Assert(res.DeletedObjectsDetail[1].Key, Equals, chnStr)
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

func (s *OssTypeSuite) TestValidateLifecleRules(c *C) {
	expiration := LifecycleExpiration{
		Days:              30,
		CreatedBeforeDate: "2015-11-11T00:00:00.000Z",
	}
	rule := LifecycleRule{
		ID:         "ruleID",
		Prefix:     "prefix",
		Status:     "Enabled",
		Expiration: &expiration,
	}
	rules := []LifecycleRule{rule}
	err := verifyLifecycleRules(rules)
	c.Assert(err, IsNil)

	expiration = LifecycleExpiration{
		Date:              "2015-11-11T00:00:00.000Z",
		CreatedBeforeDate: "2015-11-11T00:00:00.000Z",
	}
	rule = LifecycleRule{
		ID:         "ruleID",
		Prefix:     "prefix",
		Status:     "Enabled",
		Expiration: &expiration,
	}
	rules = []LifecycleRule{rule}
	err = verifyLifecycleRules(rules)
	c.Assert(err, IsNil)

	expiration = LifecycleExpiration{
		Days:              0,
		CreatedBeforeDate: "",
		Date:              "",
	}
	rule = LifecycleRule{
		ID:         "ruleID",
		Prefix:     "prefix",
		Status:     "Enabled",
		Expiration: &expiration,
	}
	rules = []LifecycleRule{rule}
	err = verifyLifecycleRules(rules)
	c.Assert(err, IsNil)

	abortMPU := LifecycleAbortMultipartUpload{
		Days:              30,
		CreatedBeforeDate: "2015-11-11T00:00:00.000Z",
	}
	rule = LifecycleRule{
		ID:                   "ruleID",
		Prefix:               "prefix",
		Status:               "Enabled",
		AbortMultipartUpload: &abortMPU,
	}
	rules = []LifecycleRule{rule}
	err = verifyLifecycleRules(rules)
	c.Assert(err, NotNil)

	abortMPU = LifecycleAbortMultipartUpload{
		Days:              0,
		CreatedBeforeDate: "",
	}
	rule = LifecycleRule{
		ID:                   "ruleID",
		Prefix:               "prefix",
		Status:               "Enabled",
		AbortMultipartUpload: &abortMPU,
	}
	rules = []LifecycleRule{rule}
	err = verifyLifecycleRules(rules)
	c.Assert(err, NotNil)

	transition := LifecycleTransition{
		Days:              30,
		CreatedBeforeDate: "2015-11-11T00:00:00.000Z",
		StorageClass:      StorageIA,
	}
	rule = LifecycleRule{
		ID:          "ruleID",
		Prefix:      "prefix",
		Status:      "Enabled",
		Transitions: []LifecycleTransition{transition},
	}
	rules = []LifecycleRule{rule}
	err = verifyLifecycleRules(rules)
	c.Assert(err, NotNil)

	transition = LifecycleTransition{
		Days:              0,
		CreatedBeforeDate: "",
		StorageClass:      StorageIA,
	}
	rule = LifecycleRule{
		ID:          "ruleID",
		Prefix:      "prefix",
		Status:      "Enabled",
		Transitions: []LifecycleTransition{transition},
	}
	rules = []LifecycleRule{rule}
	err = verifyLifecycleRules(rules)
	c.Assert(err, NotNil)

	transition = LifecycleTransition{
		Days:         30,
		StorageClass: StorageStandard,
	}
	rule = LifecycleRule{
		ID:          "ruleID",
		Prefix:      "prefix",
		Status:      "Enabled",
		Transitions: []LifecycleTransition{transition},
	}
	rules = []LifecycleRule{rule}
	err = verifyLifecycleRules(rules)
	c.Assert(err, NotNil)

	transition = LifecycleTransition{
		CreatedBeforeDate: "2015-11-11T00:00:00.000Z",
		StorageClass:      StorageStandard,
	}
	rule = LifecycleRule{
		ID:          "ruleID",
		Prefix:      "prefix",
		Status:      "Enabled",
		Transitions: []LifecycleTransition{transition},
	}
	rules = []LifecycleRule{rule}
	err = verifyLifecycleRules(rules)
	c.Assert(err, NotNil)

	transition1 := LifecycleTransition{
		Days:         30,
		StorageClass: StorageIA,
	}
	transition2 := LifecycleTransition{
		Days:         60,
		StorageClass: StorageArchive,
	}
	transition3 := LifecycleTransition{
		Days:         100,
		StorageClass: StorageArchive,
	}
	rule = LifecycleRule{
		ID:          "ruleID",
		Prefix:      "prefix",
		Status:      "Enabled",
		Transitions: []LifecycleTransition{transition1, transition2, transition3},
	}
	rules = []LifecycleRule{rule}
	err = verifyLifecycleRules(rules)
	c.Assert(err, NotNil)

	rule = LifecycleRule{
		ID:     "ruleID",
		Prefix: "prefix",
		Status: "Enabled",
	}
	rules = []LifecycleRule{rule}
	err = verifyLifecycleRules(rules)
	c.Assert(err, NotNil)

	rules = []LifecycleRule{}
	err1 := verifyLifecycleRules(rules)
	c.Assert(err1, NotNil)

	expiration = LifecycleExpiration{
		Days: 30,
	}
	rule = LifecycleRule{
		ID:         "ruleID",
		Prefix:     "prefix",
		Status:     "Enabled",
		Expiration: &expiration,
	}
	rules = []LifecycleRule{rule}
	err = verifyLifecycleRules(rules)
	c.Assert(err, IsNil)

	expiration = LifecycleExpiration{
		CreatedBeforeDate: "2015-11-11T00:00:00.000Z",
	}
	rule = LifecycleRule{
		ID:         "ruleID",
		Prefix:     "prefix",
		Status:     "Enabled",
		Expiration: &expiration,
	}
	rules = []LifecycleRule{rule}
	err = verifyLifecycleRules(rules)
	c.Assert(err, IsNil)

	abortMPU = LifecycleAbortMultipartUpload{
		Days: 30,
	}
	rule = LifecycleRule{
		ID:                   "ruleID",
		Prefix:               "prefix",
		Status:               "Enabled",
		AbortMultipartUpload: &abortMPU,
	}
	rules = []LifecycleRule{rule}
	err = verifyLifecycleRules(rules)
	c.Assert(err, IsNil)

	abortMPU = LifecycleAbortMultipartUpload{
		CreatedBeforeDate: "2015-11-11T00:00:00.000Z",
	}
	rule = LifecycleRule{
		ID:                   "ruleID",
		Prefix:               "prefix",
		Status:               "Enabled",
		AbortMultipartUpload: &abortMPU,
	}
	rules = []LifecycleRule{rule}
	err = verifyLifecycleRules(rules)
	c.Assert(err, IsNil)

	expiration = LifecycleExpiration{
		Days: 30,
	}
	abortMPU = LifecycleAbortMultipartUpload{
		Days: 30,
	}
	rule = LifecycleRule{
		ID:                   "ruleID",
		Prefix:               "prefix",
		Status:               "Enabled",
		Expiration:           &expiration,
		AbortMultipartUpload: &abortMPU,
	}
	rules = []LifecycleRule{rule}
	err = verifyLifecycleRules(rules)
	c.Assert(err, IsNil)

	expiration = LifecycleExpiration{
		CreatedBeforeDate: "2015-11-11T00:00:00.000Z",
	}
	abortMPU = LifecycleAbortMultipartUpload{
		Days: 30,
	}
	transition = LifecycleTransition{
		Days:         30,
		StorageClass: StorageIA,
	}
	rule = LifecycleRule{
		ID:                   "ruleID",
		Prefix:               "prefix",
		Status:               "Enabled",
		Expiration:           &expiration,
		AbortMultipartUpload: &abortMPU,
		Transitions:          []LifecycleTransition{transition},
	}
	rules = []LifecycleRule{rule}
	err = verifyLifecycleRules(rules)
	c.Assert(err, IsNil)

	expiration = LifecycleExpiration{
		CreatedBeforeDate: "2015-11-11T00:00:00.000Z",
	}
	abortMPU = LifecycleAbortMultipartUpload{
		Days: 30,
	}
	transition1 = LifecycleTransition{
		Days:         30,
		StorageClass: StorageIA,
	}
	transition2 = LifecycleTransition{
		Days:         60,
		StorageClass: StorageArchive,
	}
	rule = LifecycleRule{
		ID:                   "ruleID",
		Prefix:               "prefix",
		Status:               "Enabled",
		Expiration:           &expiration,
		AbortMultipartUpload: &abortMPU,
		Transitions:          []LifecycleTransition{transition1, transition2},
	}
	rules = []LifecycleRule{rule}
	err = verifyLifecycleRules(rules)
	c.Assert(err, IsNil)
}
