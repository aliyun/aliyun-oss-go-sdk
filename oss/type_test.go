package oss

import (
	"encoding/xml"
	"log"
	"net/url"
	"sort"
	"strings"

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

	res.DeletedObjectsDetail = []DeletedKeyInfo{{Key: ""}}
	err = decodeDeleteObjectsResult(&res)
	c.Assert(err, IsNil)
	c.Assert(res.DeletedObjectsDetail[0].Key, Equals, "")

	res.DeletedObjectsDetail = []DeletedKeyInfo{{Key: goURLStr}, {Key: chnURLStr}}
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

	sort.Sort(UploadParts(parts))
	c.Assert(len(parts), Equals, 0)

	parts = []UploadPart{
		{PartNumber: 5, ETag: "E5"},
		{PartNumber: 1, ETag: "E1"},
		{PartNumber: 4, ETag: "E4"},
		{PartNumber: 2, ETag: "E2"},
		{PartNumber: 3, ETag: "E3"},
	}

	sort.Sort(UploadParts(parts))

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
	c.Assert(err, IsNil)

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
	c.Assert(err, IsNil)

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
	c.Assert(err, IsNil)

	rule = LifecycleRule{
		ID:     "ruleID",
		Prefix: "prefix",
		Status: "Enabled",
	}
	rules = []LifecycleRule{rule}
	err = verifyLifecycleRules(rules)
	c.Assert(err, IsNil)

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

// test get meta query statsu result
func (s *OssTypeSuite) TestGetMetaQueryStatusResult(c *C) {
	var res GetMetaQueryStatusResult
	xmlData := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<MetaQueryStatus>
  <State>Running</State>
  <Phase>FullScanning</Phase>
  <CreateTime>2021-08-02T10:49:17.289372919+08:00</CreateTime>
  <UpdateTime>2021-08-02T10:49:17.289372919+08:00</UpdateTime>
</MetaQueryStatus>`)
	err := xml.Unmarshal(xmlData, &res)
	c.Assert(err, IsNil)
	c.Assert(res.State, Equals, "Running")
	c.Assert(res.Phase, Equals, "FullScanning")
	c.Assert(res.CreateTime, Equals, "2021-08-02T10:49:17.289372919+08:00")
	c.Assert(res.UpdateTime, Equals, "2021-08-02T10:49:17.289372919+08:00")
}

// test do meta query request xml
func (s *OssTypeSuite) TestDoMetaQueryRequest(c *C) {
	var res MetaQuery
	xmlData := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<MetaQuery>
  <NextToken>MTIzNDU2Nzg6aW1tdGVzdDpleGFtcGxlYnVja2V0OmRhdGFzZXQwMDE6b3NzOi8vZXhhbXBsZWJ1Y2tldC9zYW1wbGVvYmplY3QxLmpwZw==</NextToken>
  <MaxResults>5</MaxResults>
  <Query>{"Field": "Size","Value": "1048576","Operation": "gt"}</Query>
  <Sort>Size</Sort>
  <Order>asc</Order>
  <Aggregations>
    <Aggregation>
      <Field>Size</Field>
      <Operation>sum</Operation>
    </Aggregation>
    <Aggregation>
      <Field>Size</Field>
      <Operation>max</Operation>
    </Aggregation>
  </Aggregations>
</MetaQuery>`)
	err := xml.Unmarshal(xmlData, &res)
	c.Assert(err, IsNil)
	c.Assert(res.NextToken, Equals, "MTIzNDU2Nzg6aW1tdGVzdDpleGFtcGxlYnVja2V0OmRhdGFzZXQwMDE6b3NzOi8vZXhhbXBsZWJ1Y2tldC9zYW1wbGVvYmplY3QxLmpwZw==")
	c.Assert(res.MaxResults, Equals, int64(5))
	c.Assert(res.Query, Equals, `{"Field": "Size","Value": "1048576","Operation": "gt"}`)
	c.Assert(res.Sort, Equals, "Size")
	c.Assert(res.Order, Equals, "asc")
	c.Assert(res.Aggregations[0].Field, Equals, "Size")
	c.Assert(res.Aggregations[1].Operation, Equals, "max")
}

// test do meta query result
func (s *OssTypeSuite) TestDoMetaQueryResult(c *C) {
	var res DoMetaQueryResult
	xmlData := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<MetaQuery>
  <NextToken>MTIzNDU2Nzg6aW1tdGVzdDpleGFtcGxlYnVja2V0OmRhdGFzZXQwMDE6b3NzOi8vZXhhbXBsZWJ1Y2tldC9zYW1wbGVvYmplY3QxLmpwZw==</NextToken>
  <Files>
    <File>
      <Filename>exampleobject.txt</Filename>
      <Size>120</Size>
      <FileModifiedTime>2021-06-29T14:50:13.011643661+08:00</FileModifiedTime>
      <OSSObjectType>Normal</OSSObjectType>
      <OSSStorageClass>Standard</OSSStorageClass>
      <ObjectACL>default</ObjectACL>
      <ETag>"fba9dede5f27731c9771645a3986****"</ETag>
      <OSSCRC64>4858A48BD1466884</OSSCRC64>
      <OSSTaggingCount>2</OSSTaggingCount>
      <OSSTagging>
        <Tagging>
          <Key>owner</Key>
          <Value>John</Value>
        </Tagging>
        <Tagging>
          <Key>type</Key>
          <Value>document</Value>
        </Tagging>
      </OSSTagging>
      <OSSUserMeta>
        <UserMeta>
          <Key>x-oss-meta-location</Key>
          <Value>hangzhou</Value>
        </UserMeta>
      </OSSUserMeta>
    </File>
  </Files>
</MetaQuery>`)
	err := xml.Unmarshal(xmlData, &res)
	c.Assert(err, IsNil)
	c.Assert(res.NextToken, Equals, "MTIzNDU2Nzg6aW1tdGVzdDpleGFtcGxlYnVja2V0OmRhdGFzZXQwMDE6b3NzOi8vZXhhbXBsZWJ1Y2tldC9zYW1wbGVvYmplY3QxLmpwZw==")
	c.Assert(res.Files[0].Filename, Equals, "exampleobject.txt")
	c.Assert(res.Files[0].Size, Equals, int64(120))
	c.Assert(res.Files[0].FileModifiedTime, Equals, "2021-06-29T14:50:13.011643661+08:00")
	c.Assert(res.Files[0].OssObjectType, Equals, "Normal")
	c.Assert(res.Files[0].OssCRC64, Equals, "4858A48BD1466884")
	c.Assert(res.Files[0].OssObjectType, Equals, "Normal")
	c.Assert(res.Files[0].OssStorageClass, Equals, "Standard")
	c.Assert(res.Files[0].ObjectACL, Equals, "default")
	c.Assert(res.Files[0].OssTagging[1].Key, Equals, "type")
	c.Assert(res.Files[0].OssTagging[1].Value, Equals, "document")
	c.Assert(res.Files[0].OssUserMeta[0].Value, Equals, "hangzhou")

	// test aggregations
	xmlData = []byte(`<?xml version="1.0" encoding="UTF-8"?>
<MetaQuery>
    <NextToken></NextToken>
    <Aggregations>
        <Aggregation>
            <Field>Size</Field>
            <Operation>sum</Operation>
            <Value>839794720</Value>
        </Aggregation>
        <Aggregation>
            <Field>Size</Field>
            <Operation>group</Operation>
            <Groups>
                <Group>
                    <Value>518</Value>
                    <Count>1</Count>
                </Group>
                <Group>
                    <Value>581</Value>
                    <Count>1</Count>
                </Group>
            </Groups>
        </Aggregation>
    </Aggregations>
</MetaQuery>
`)
	err = xml.Unmarshal(xmlData, &res)
	c.Assert(err, IsNil)
	c.Assert(res.NextToken, Equals, "")
	c.Assert(res.Aggregations[0].Field, Equals, "Size")
	c.Assert(res.Aggregations[0].Operation, Equals, "sum")
	c.Assert(res.Aggregations[0].Value, Equals, float64(839794720))
	c.Assert(res.Aggregations[1].Operation, Equals, "group")
	c.Assert(res.Aggregations[1].Groups[1].Value, Equals, "581")
	c.Assert(res.Aggregations[1].Groups[1].Count, Equals, int64(1))
}

// test get bucket stat result
func (s *OssTypeSuite) TestGetBucketStatResult(c *C) {
	var res GetBucketStatResult
	xmlData := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<BucketStat>
  <Storage>1600</Storage>
  <ObjectCount>230</ObjectCount>
  <MultipartUploadCount>40</MultipartUploadCount>
  <LiveChannelCount>4</LiveChannelCount>
  <LastModifiedTime>1643341269</LastModifiedTime>
  <StandardStorage>430</StandardStorage>
  <StandardObjectCount>66</StandardObjectCount>
  <InfrequentAccessStorage>2359296</InfrequentAccessStorage>
  <InfrequentAccessRealStorage>360</InfrequentAccessRealStorage>
  <InfrequentAccessObjectCount>54</InfrequentAccessObjectCount>
  <ArchiveStorage>2949120</ArchiveStorage>
  <ArchiveRealStorage>450</ArchiveRealStorage>
  <ArchiveObjectCount>74</ArchiveObjectCount>
  <ColdArchiveStorage>2359296</ColdArchiveStorage>
  <ColdArchiveRealStorage>360</ColdArchiveRealStorage>
  <ColdArchiveObjectCount>36</ColdArchiveObjectCount>
</BucketStat>`)
	err := xml.Unmarshal(xmlData, &res)
	c.Assert(err, IsNil)
	c.Assert(res.Storage, Equals, int64(1600))
	c.Assert(res.ObjectCount, Equals, int64(230))
	c.Assert(res.MultipartUploadCount, Equals, int64(40))
	c.Assert(res.LiveChannelCount, Equals, int64(4))
	c.Assert(res.LastModifiedTime, Equals, int64(1643341269))
	c.Assert(res.StandardStorage, Equals, int64(430))
	c.Assert(res.StandardObjectCount, Equals, int64(66))
	c.Assert(res.InfrequentAccessStorage, Equals, int64(2359296))
	c.Assert(res.InfrequentAccessRealStorage, Equals, int64(360))
	c.Assert(res.InfrequentAccessObjectCount, Equals, int64(54))
	c.Assert(res.ArchiveStorage, Equals, int64(2949120))
	c.Assert(res.ArchiveRealStorage, Equals, int64(450))
	c.Assert(res.ArchiveObjectCount, Equals, int64(74))
	c.Assert(res.ColdArchiveStorage, Equals, int64(2359296))
	c.Assert(res.ColdArchiveRealStorage, Equals, int64(360))
	c.Assert(res.ColdArchiveObjectCount, Equals, int64(36))
}

// test delete object struct turn to xml string
func (s *OssTypeSuite) TestDeleteObjectToXml(c *C) {
	versionIds := make([]DeleteObject, 0)
	versionIds = append(versionIds, DeleteObject{Key: "\f", VersionId: "1111"})
	dxml := deleteXML{}
	dxml.Objects = versionIds
	dxml.Quiet = false
	str := marshalDeleteObjectToXml(dxml)
	str2 := "<Delete><Quiet>false</Quiet><Object><Key>&#x0C;</Key><VersionId>1111</VersionId></Object></Delete>"
	c.Assert(str, Equals, str2)

	versionIds = append(versionIds, DeleteObject{Key: "A ' < > \" & ~ ` ! @ # $ % ^ & * ( ) [] {} - _ + = / | \\ ? . , : ; A", VersionId: "2222"})
	dxml.Objects = versionIds
	dxml.Quiet = false
	str = marshalDeleteObjectToXml(dxml)
	str2 = "<Delete><Quiet>false</Quiet><Object><Key>&#x0C;</Key><VersionId>1111</VersionId></Object><Object><Key>A &#39; &lt; &gt; &#34; &amp; ~ ` ! @ # $ % ^ &amp; * ( ) [] {} - _ + = / | \\ ? . , : ; A</Key><VersionId>2222</VersionId></Object></Delete>"
	c.Assert(str, Equals, str2)

	objects := make([]DeleteObject, 0)
	objects = append(objects, DeleteObject{Key: "\v"})
	dxml.Objects = objects
	dxml.Quiet = true
	str = marshalDeleteObjectToXml(dxml)
	str2 = "<Delete><Quiet>true</Quiet><Object><Key>&#x0B;</Key></Object></Delete>"
	c.Assert(str, Equals, str2)
}

// test access monitor
func (s *OssTypeSuite) TestAccessMonitor(c *C) {
	var res GetBucketAccessMonitorResult
	xmlData := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<AccessMonitorConfiguration>
<Status>Enabled</Status>
</AccessMonitorConfiguration>
`)
	err := xml.Unmarshal(xmlData, &res)
	c.Assert(err, IsNil)
	c.Assert(res.Status, Equals, "Enabled")

	// test aggregations
	xmlData = []byte(`<?xml version="1.0" encoding="UTF-8"?>
<AccessMonitorConfiguration>
<Status>Disabled</Status>
</AccessMonitorConfiguration>
`)
	err = xml.Unmarshal(xmlData, &res)
	c.Assert(res.Status, Equals, "Disabled")

	var req PutBucketAccessMonitor
	req.Status = "Enabled"
	bs, err := xml.Marshal(req)
	c.Assert(err, IsNil)
	c.Assert(string(bs), Equals, `<AccessMonitorConfiguration><Status>Enabled</Status></AccessMonitorConfiguration>`)

	req.Status = "Disabled"
	bs, err = xml.Marshal(req)
	c.Assert(err, IsNil)
	c.Assert(string(bs), Equals, `<AccessMonitorConfiguration><Status>Disabled</Status></AccessMonitorConfiguration>`)
}

// test bucket info result with access monitor
func (s *OssTypeSuite) TestBucketInfoWithAccessMonitor(c *C) {
	var res GetBucketInfoResult
	xmlData := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<BucketInfo>
  <Bucket>
    <AccessMonitor>Enabled</AccessMonitor>
    <CreationDate>2013-07-31T10:56:21.000Z</CreationDate>
    <ExtranetEndpoint>oss-cn-hangzhou.aliyuncs.com</ExtranetEndpoint>
    <IntranetEndpoint>oss-cn-hangzhou-internal.aliyuncs.com</IntranetEndpoint>
    <Location>oss-cn-hangzhou</Location>
    <StorageClass>Standard</StorageClass>
    <TransferAcceleration>Disabled</TransferAcceleration>
    <CrossRegionReplication>Disabled</CrossRegionReplication>
    <Name>oss-example</Name>
    <Owner>
      <DisplayName>username</DisplayName>
      <ID>27183473914</ID>
    </Owner>
    <Comment>test</Comment>
  </Bucket>
</BucketInfo>
`)
	err := xml.Unmarshal(xmlData, &res)
	c.Assert(err, IsNil)
	c.Assert(res.BucketInfo.AccessMonitor, Equals, "Enabled")
	c.Assert(res.BucketInfo.CreationDate.Format("2006-01-02 15:04:05 +0000 UTC"), Equals, "2013-07-31 10:56:21 +0000 UTC")
	c.Assert(res.BucketInfo.ExtranetEndpoint, Equals, "oss-cn-hangzhou.aliyuncs.com")
	c.Assert(res.BucketInfo.IntranetEndpoint, Equals, "oss-cn-hangzhou-internal.aliyuncs.com")
	c.Assert(res.BucketInfo.Location, Equals, "oss-cn-hangzhou")
	c.Assert(res.BucketInfo.StorageClass, Equals, "Standard")
	c.Assert(res.BucketInfo.TransferAcceleration, Equals, "Disabled")
	c.Assert(res.BucketInfo.CrossRegionReplication, Equals, "Disabled")
	c.Assert(res.BucketInfo.Name, Equals, "oss-example")
	c.Assert(res.BucketInfo.Owner.ID, Equals, "27183473914")
	c.Assert(res.BucketInfo.Owner.DisplayName, Equals, "username")
	xmlData = []byte(`<?xml version="1.0" encoding="UTF-8"?>
<BucketInfo>
  <Bucket>
    <AccessMonitor>Disabled</AccessMonitor>
    <CreationDate>2013-07-31T10:56:21.000Z</CreationDate>
    <ExtranetEndpoint>oss-cn-hangzhou.aliyuncs.com</ExtranetEndpoint>
    <IntranetEndpoint>oss-cn-hangzhou-internal.aliyuncs.com</IntranetEndpoint>
    <Location>oss-cn-hangzhou</Location>
    <StorageClass>Standard</StorageClass>
    <TransferAcceleration>Disabled</TransferAcceleration>
    <CrossRegionReplication>Disabled</CrossRegionReplication>
    <Name>oss-example</Name>
  </Bucket>
</BucketInfo>
`)
	err = xml.Unmarshal(xmlData, &res)
	c.Assert(err, IsNil)
	c.Assert(res.BucketInfo.AccessMonitor, Equals, "Disabled")
}

func (s *OssTypeSuite) TestValidateLifeCycleRulesWithAccessTime(c *C) {
	expiration := LifecycleExpiration{
		Days:              30,
		CreatedBeforeDate: "2015-11-11T00:00:00.000Z",
	}
	isTrue := true
	isFalse := false
	rule := LifecycleRule{
		ID:         "ruleID",
		Prefix:     "prefix",
		Status:     "Enabled",
		Expiration: &expiration,
		Transitions: []LifecycleTransition{
			{
				Days:         30,
				StorageClass: StorageIA,
				IsAccessTime: &isFalse,
			},
		},
	}
	rules := []LifecycleRule{rule}
	err := verifyLifecycleRules(rules)
	c.Assert(err, IsNil)

	rule = LifecycleRule{
		ID:         "ruleID",
		Prefix:     "prefix",
		Status:     "Enabled",
		Expiration: &expiration,
		Transitions: []LifecycleTransition{
			{
				Days:                 30,
				StorageClass:         StorageIA,
				IsAccessTime:         &isTrue,
				ReturnToStdWhenVisit: &isFalse,
			},
		},
	}
	rules = []LifecycleRule{rule}
	err = verifyLifecycleRules(rules)
	c.Assert(err, IsNil)

	rule = LifecycleRule{
		ID:         "ruleID",
		Prefix:     "prefix",
		Status:     "Enabled",
		Expiration: &expiration,
		Transitions: []LifecycleTransition{
			{
				Days:                 30,
				StorageClass:         StorageIA,
				IsAccessTime:         &isTrue,
				ReturnToStdWhenVisit: &isTrue,
			},
		},
	}
	rules = []LifecycleRule{rule}
	err = verifyLifecycleRules(rules)
	c.Assert(err, IsNil)

	rule = LifecycleRule{
		ID:         "ruleID",
		Prefix:     "prefix",
		Status:     "Enabled",
		Expiration: &expiration,
		NonVersionTransitions: []LifecycleVersionTransition{
			{
				NoncurrentDays:       10,
				StorageClass:         StorageIA,
				IsAccessTime:         &isTrue,
				ReturnToStdWhenVisit: &isTrue,
			},
		},
	}
	rules = []LifecycleRule{rule}
	err = verifyLifecycleRules(rules)
	c.Assert(err, IsNil)

	abortMPU := LifecycleAbortMultipartUpload{
		CreatedBeforeDate: "2015-11-11T00:00:00.000Z",
	}
	rule = LifecycleRule{
		ID:                   "ruleID",
		Prefix:               "prefix",
		Status:               "Enabled",
		AbortMultipartUpload: &abortMPU,
		NonVersionTransitions: []LifecycleVersionTransition{
			{
				NoncurrentDays:       10,
				StorageClass:         StorageIA,
				IsAccessTime:         &isTrue,
				ReturnToStdWhenVisit: &isTrue,
			},
		},
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
		NonVersionTransitions: []LifecycleVersionTransition{
			{
				NoncurrentDays:       10,
				StorageClass:         StorageIA,
				IsAccessTime:         &isTrue,
				ReturnToStdWhenVisit: &isFalse,
			},
		},
	}
	rules = []LifecycleRule{rule}
	err = verifyLifecycleRules(rules)
	c.Assert(err, IsNil)
}

func (s *OssTypeSuite) TestLifeCycleRules(c *C) {
	expiration := LifecycleExpiration{
		Days:              30,
		CreatedBeforeDate: "2015-11-11T00:00:00.000Z",
	}
	isTrue := true
	isFalse := false
	rule := LifecycleRule{
		ID:         "ruleID",
		Prefix:     "prefix",
		Status:     "Enabled",
		Expiration: &expiration,
		Transitions: []LifecycleTransition{
			{
				Days:         30,
				StorageClass: StorageIA,
				IsAccessTime: &isFalse,
			},
		},
	}
	rules := []LifecycleRule{rule}
	err := verifyLifecycleRules(rules)
	c.Assert(err, IsNil)

	rule = LifecycleRule{
		ID:         "ruleID",
		Prefix:     "prefix",
		Status:     "Enabled",
		Expiration: &expiration,
		Transitions: []LifecycleTransition{
			{
				Days:                 30,
				StorageClass:         StorageIA,
				IsAccessTime:         &isTrue,
				ReturnToStdWhenVisit: &isFalse,
			},
		},
	}
	rules = []LifecycleRule{rule}
	err = verifyLifecycleRules(rules)
	c.Assert(err, IsNil)

	rule = LifecycleRule{
		ID:         "ruleID",
		Prefix:     "prefix",
		Status:     "Enabled",
		Expiration: &expiration,
		Transitions: []LifecycleTransition{
			{
				Days:                 30,
				StorageClass:         StorageIA,
				IsAccessTime:         &isTrue,
				ReturnToStdWhenVisit: &isTrue,
			},
		},
	}
	rules = []LifecycleRule{rule}
	err = verifyLifecycleRules(rules)
	c.Assert(err, IsNil)

	rule = LifecycleRule{
		ID:         "ruleID",
		Prefix:     "prefix",
		Status:     "Enabled",
		Expiration: &expiration,
		NonVersionTransitions: []LifecycleVersionTransition{
			{
				NoncurrentDays:       10,
				StorageClass:         StorageIA,
				IsAccessTime:         &isTrue,
				ReturnToStdWhenVisit: &isTrue,
			},
		},
	}
	rules = []LifecycleRule{rule}
	err = verifyLifecycleRules(rules)
	c.Assert(err, IsNil)

	abortMPU := LifecycleAbortMultipartUpload{
		CreatedBeforeDate: "2015-11-11T00:00:00.000Z",
	}
	rule = LifecycleRule{
		ID:                   "ruleID",
		Prefix:               "prefix",
		Status:               "Enabled",
		AbortMultipartUpload: &abortMPU,
		NonVersionTransitions: []LifecycleVersionTransition{
			{
				NoncurrentDays:       10,
				StorageClass:         StorageIA,
				IsAccessTime:         &isTrue,
				ReturnToStdWhenVisit: &isTrue,
			},
		},
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
		NonVersionTransitions: []LifecycleVersionTransition{
			{
				NoncurrentDays:       10,
				StorageClass:         StorageIA,
				IsAccessTime:         &isTrue,
				ReturnToStdWhenVisit: &isFalse,
			},
		},
	}
	rules = []LifecycleRule{rule}
	err = verifyLifecycleRules(rules)
	c.Assert(err, IsNil)
}

func (s *OssTypeSuite) TestGetBucketLifecycleResult(c *C) {
	xmlData := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<LifecycleConfiguration>
  <Rule>
    <ID>RuleID</ID>
    <Prefix>logs</Prefix>
    <Status>Enabled</Status>
    <Filter>
      <Not>
        <Prefix>logs1</Prefix>
        <Tag><Key>key1</Key><Value>value1</Value></Tag>
        </Not>
    </Filter>
    <Expiration>
      <Days>100</Days>
    </Expiration>
    <Transition>
      <Days>30</Days>
      <StorageClass>Archive</StorageClass>
    </Transition>
  </Rule>
</LifecycleConfiguration>
`)
	var res GetBucketLifecycleResult
	err := xml.Unmarshal(xmlData, &res)
	c.Assert(err, IsNil)
	c.Assert(res.Rules[0].ID, Equals, "RuleID")
	c.Assert(res.Rules[0].Filter.Not[0].Prefix, Equals, "logs1")
	c.Assert(res.Rules[0].Filter.Not[0].Tag.Key, Equals, "key1")
	c.Assert(res.Rules[0].Filter.Not[0].Tag.Value, Equals, "value1")

	xmlData = []byte(`<?xml version="1.0" encoding="UTF-8"?>
<LifecycleConfiguration>
  <Rule>
    <ID>test2</ID>
    <Prefix>logs</Prefix>
    <Status>Enabled</Status>
    <Filter>
      <Not>
        <Prefix>logs-demo</Prefix>
      </Not>
      <Not>
        <Prefix>abc/not1/</Prefix>
        <Tag>
          <Key>notkey1</Key>
          <Value>notvalue1</Value>
        </Tag>
      </Not>
      <Not>
        <Prefix>abc/not2/</Prefix>
        <Tag>
          <Key>notkey2</Key>
          <Value>notvalue2</Value>
        </Tag>
      </Not>
    </Filter>
    <Expiration>
      <Days>100</Days>
    </Expiration>
    <Transition>
      <Days>30</Days>
      <StorageClass>Archive</StorageClass>
    </Transition>
  </Rule>
</LifecycleConfiguration>
`)
	var res1 GetBucketLifecycleResult
	err = xml.Unmarshal(xmlData, &res1)
	c.Assert(err, IsNil)
	c.Assert(res1.Rules[0].ID, Equals, "test2")
	c.Assert(res1.Rules[0].Filter.Not[0].Prefix, Equals, "logs-demo")
	c.Assert(res1.Rules[0].Filter.Not[1].Prefix, Equals, "abc/not1/")
	c.Assert(res1.Rules[0].Filter.Not[1].Tag.Key, Equals, "notkey1")
	c.Assert(res1.Rules[0].Filter.Not[1].Tag.Value, Equals, "notvalue1")
	c.Assert(res1.Rules[0].Filter.Not[2].Prefix, Equals, "abc/not2/")
	c.Assert(res1.Rules[0].Filter.Not[2].Tag.Key, Equals, "notkey2")
	c.Assert(res1.Rules[0].Filter.Not[2].Tag.Value, Equals, "notvalue2")

	xmlData = []byte(`<?xml version="1.0" encoding="UTF-8"?><LifecycleConfiguration>
  <Rule>
    <ID>r1</ID>
    <Prefix>abc/</Prefix>
    <Filter>
      <ObjectSizeGreaterThan>500</ObjectSizeGreaterThan>
      <ObjectSizeLessThan>64000</ObjectSizeLessThan>
      <Not>
        <Prefix>abc/not1/</Prefix>
        <Tag>
          <Key>notkey1</Key>
          <Value>notvalue1</Value>
        </Tag>
      </Not>
      <Not>
        <Prefix>abc/not2/</Prefix>
        <Tag>
          <Key>notkey2</Key>
          <Value>notvalue2</Value>
        </Tag>
      </Not>
    </Filter>
  </Rule>
  <Rule>
    <ID>r2</ID>
    <Prefix>def/</Prefix>
    <Filter>
      <ObjectSizeGreaterThan>500</ObjectSizeGreaterThan>
      <Not>
        <Prefix>def/not1/</Prefix>
      </Not>
      <Not>
        <Prefix>def/not2/</Prefix>
        <Tag>
          <Key>notkey2</Key>
          <Value>notvalue2</Value>
        </Tag>
      </Not>
    </Filter>
  </Rule>
</LifecycleConfiguration>
`)

	var res2 GetBucketLifecycleResult
	err = xml.Unmarshal(xmlData, &res2)
	testLogger.Println(res2.Rules[1].Filter)
	c.Assert(err, IsNil)
	c.Assert(res2.Rules[0].ID, Equals, "r1")
	c.Assert(res2.Rules[0].Prefix, Equals, "abc/")
	c.Assert(*(res2.Rules[0].Filter.ObjectSizeGreaterThan), Equals, int64(500))
	c.Assert(*(res2.Rules[0].Filter.ObjectSizeLessThan), Equals, int64(64000))
	c.Assert(res2.Rules[0].Filter.Not[0].Prefix, Equals, "abc/not1/")
	c.Assert(res2.Rules[0].Filter.Not[0].Tag.Key, Equals, "notkey1")
	c.Assert(res2.Rules[0].Filter.Not[0].Tag.Value, Equals, "notvalue1")

	c.Assert(res2.Rules[0].Filter.Not[1].Prefix, Equals, "abc/not2/")
	c.Assert(res2.Rules[0].Filter.Not[1].Tag.Key, Equals, "notkey2")
	c.Assert(res2.Rules[0].Filter.Not[1].Tag.Value, Equals, "notvalue2")

	c.Assert(res2.Rules[1].ID, Equals, "r2")
	c.Assert(res2.Rules[1].Prefix, Equals, "def/")
	c.Assert(*(res2.Rules[1].Filter.ObjectSizeGreaterThan), Equals, int64(500))
	c.Assert(res2.Rules[1].Filter.ObjectSizeLessThan, IsNil)
	c.Assert(res2.Rules[1].Filter.Not[0].Prefix, Equals, "def/not1/")

	c.Assert(res2.Rules[1].Filter.Not[1].Prefix, Equals, "def/not2/")
	c.Assert(res2.Rules[1].Filter.Not[1].Tag.Key, Equals, "notkey2")
	c.Assert(res2.Rules[1].Filter.Not[1].Tag.Value, Equals, "notvalue2")
}
func (s *OssTypeSuite) TestLifecycleConfiguration(c *C) {
	expiration := LifecycleExpiration{
		Days:              30,
		CreatedBeforeDate: "2015-11-11T00:00:00.000Z",
	}
	isTrue := true
	isFalse := false
	greater := int64(500)
	less := int64(645000)
	filter := LifecycleFilter{
		ObjectSizeGreaterThan: &greater,
		ObjectSizeLessThan:    &less,
	}
	rule0 := LifecycleRule{
		ID:         "r0",
		Prefix:     "prefix0",
		Status:     "Enabled",
		Expiration: &expiration,
	}
	rule1 := LifecycleRule{
		ID:         "r1",
		Prefix:     "prefix1",
		Status:     "Enabled",
		Expiration: &expiration,
		Transitions: []LifecycleTransition{
			{
				Days:         30,
				StorageClass: StorageIA,
				IsAccessTime: &isFalse,
			},
		},
		Filter: &filter,
	}

	abortMPU := LifecycleAbortMultipartUpload{
		Days:              30,
		CreatedBeforeDate: "2015-11-11T00:00:00.000Z",
	}
	rule2 := LifecycleRule{
		ID:                   "r3",
		Prefix:               "prefix3",
		Status:               "Enabled",
		Expiration:           &expiration,
		AbortMultipartUpload: &abortMPU,
		NonVersionTransitions: []LifecycleVersionTransition{
			{
				NoncurrentDays:       10,
				StorageClass:         StorageIA,
				IsAccessTime:         &isTrue,
				ReturnToStdWhenVisit: &isFalse,
			},
		},
	}
	tag := Tag{
		Key:   "key1",
		Value: "val1",
	}
	filter3 := LifecycleFilter{
		ObjectSizeGreaterThan: &greater,
		ObjectSizeLessThan:    &less,
		Not: []LifecycleFilterNot{
			{
				Tag: &tag,
			},
		},
	}
	rule3 := LifecycleRule{
		ID:                   "r4",
		Prefix:               "",
		Status:               "Enabled",
		Expiration:           &expiration,
		AbortMultipartUpload: &abortMPU,
		NonVersionTransitions: []LifecycleVersionTransition{
			{
				NoncurrentDays:       10,
				StorageClass:         StorageIA,
				IsAccessTime:         &isTrue,
				ReturnToStdWhenVisit: &isFalse,
			},
		},
		Filter: &filter3,
	}
	rules := []LifecycleRule{rule0, rule1, rule2, rule3}
	config := LifecycleConfiguration{
		Rules: rules,
	}
	xmlData, err := xml.Marshal(config)
	testLogger.Println(string(xmlData))
	c.Assert(err, IsNil)
	c.Assert(string(xmlData), Equals, "<LifecycleConfiguration><Rule><ID>r0</ID><Prefix>prefix0</Prefix><Status>Enabled</Status><Expiration><Days>30</Days><CreatedBeforeDate>2015-11-11T00:00:00.000Z</CreatedBeforeDate></Expiration></Rule><Rule><ID>r1</ID><Prefix>prefix1</Prefix><Status>Enabled</Status><Expiration><Days>30</Days><CreatedBeforeDate>2015-11-11T00:00:00.000Z</CreatedBeforeDate></Expiration><Transition><Days>30</Days><StorageClass>IA</StorageClass><IsAccessTime>false</IsAccessTime></Transition><Filter><ObjectSizeGreaterThan>500</ObjectSizeGreaterThan><ObjectSizeLessThan>645000</ObjectSizeLessThan></Filter></Rule><Rule><ID>r3</ID><Prefix>prefix3</Prefix><Status>Enabled</Status><Expiration><Days>30</Days><CreatedBeforeDate>2015-11-11T00:00:00.000Z</CreatedBeforeDate></Expiration><AbortMultipartUpload><Days>30</Days><CreatedBeforeDate>2015-11-11T00:00:00.000Z</CreatedBeforeDate></AbortMultipartUpload><NoncurrentVersionTransition><NoncurrentDays>10</NoncurrentDays><StorageClass>IA</StorageClass><IsAccessTime>true</IsAccessTime><ReturnToStdWhenVisit>false</ReturnToStdWhenVisit></NoncurrentVersionTransition></Rule><Rule><ID>r4</ID><Prefix></Prefix><Status>Enabled</Status><Expiration><Days>30</Days><CreatedBeforeDate>2015-11-11T00:00:00.000Z</CreatedBeforeDate></Expiration><AbortMultipartUpload><Days>30</Days><CreatedBeforeDate>2015-11-11T00:00:00.000Z</CreatedBeforeDate></AbortMultipartUpload><NoncurrentVersionTransition><NoncurrentDays>10</NoncurrentDays><StorageClass>IA</StorageClass><IsAccessTime>true</IsAccessTime><ReturnToStdWhenVisit>false</ReturnToStdWhenVisit></NoncurrentVersionTransition><Filter><Not><Prefix></Prefix><Tag><Key>key1</Key><Value>val1</Value></Tag></Not><ObjectSizeGreaterThan>500</ObjectSizeGreaterThan><ObjectSizeLessThan>645000</ObjectSizeLessThan></Filter></Rule></LifecycleConfiguration>")

	filter4 := LifecycleFilter{
		ObjectSizeGreaterThan: &greater,
		ObjectSizeLessThan:    &less,
		Not: []LifecycleFilterNot{
			{},
		},
	}
	rule4 := LifecycleRule{
		ID:         "r5",
		Prefix:     "",
		Status:     "Enabled",
		Expiration: &expiration,
		Filter:     &filter4,
	}
	rules4 := []LifecycleRule{rule4}
	config4 := LifecycleConfiguration{
		Rules: rules4,
	}
	xmlData4, err := xml.Marshal(config4)
	testLogger.Println(string(xmlData4))
	c.Assert(err, IsNil)
	c.Assert(string(xmlData4), Equals, "<LifecycleConfiguration><Rule><ID>r5</ID><Prefix></Prefix><Status>Enabled</Status><Expiration><Days>30</Days><CreatedBeforeDate>2015-11-11T00:00:00.000Z</CreatedBeforeDate></Expiration><Filter><Not><Prefix></Prefix></Not><ObjectSizeGreaterThan>500</ObjectSizeGreaterThan><ObjectSizeLessThan>645000</ObjectSizeLessThan></Filter></Rule></LifecycleConfiguration>")
}

// Test Bucket Resource Group
func (s *OssTypeSuite) TestBucketResourceGroup(c *C) {
	var res GetBucketResourceGroupResult
	xmlData := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<BucketResourceGroupConfiguration>
  <ResourceGroupId>rg-xxxxxx</ResourceGroupId>
</BucketResourceGroupConfiguration>`)
	err := xml.Unmarshal(xmlData, &res)
	c.Assert(err, IsNil)
	c.Assert(res.ResourceGroupId, Equals, "rg-xxxxxx")

	xmlData = []byte(`<?xml version="1.0" encoding="UTF-8"?>
<BucketResourceGroupConfiguration>
  <ResourceGroupId></ResourceGroupId>
</BucketResourceGroupConfiguration>`)
	err = xml.Unmarshal(xmlData, &res)
	c.Assert(err, IsNil)
	c.Assert(res.ResourceGroupId, Equals, "")

	resource := PutBucketResourceGroup{
		ResourceGroupId: "rg-xxxxxx",
	}

	bs, err := xml.Marshal(resource)
	c.Assert(err, IsNil)

	c.Assert(string(bs), Equals, "<BucketResourceGroupConfiguration><ResourceGroupId>rg-xxxxxx</ResourceGroupId></BucketResourceGroupConfiguration>")

	resource = PutBucketResourceGroup{
		ResourceGroupId: "",
	}

	bs, err = xml.Marshal(resource)
	c.Assert(err, IsNil)

	c.Assert(string(bs), Equals, "<BucketResourceGroupConfiguration><ResourceGroupId></ResourceGroupId></BucketResourceGroupConfiguration>")
}

func (s *OssTypeSuite) TestListBucketCnameResult(c *C) {
	xmlData := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<ListCnameResult>
  <Bucket>targetbucket</Bucket>
  <Owner>testowner</Owner>
  <Cname>
    <Domain>example.com</Domain>
    <LastModified>2021-09-15T02:35:07.000Z</LastModified>
    <Status>Enabled</Status>
    <Certificate>
      <Type>CAS</Type>
      <CertId>493****-cn-hangzhou</CertId>
      <Status>Enabled</Status>
      <CreationDate>Wed, 15 Sep 2021 02:35:06 GMT</CreationDate>
      <Fingerprint>DE:01:CF:EC:7C:A7:98:CB:D8:6E:FB:1D:97:EB:A9:64:1D:4E:**:**</Fingerprint>
      <ValidStartDate>Tue, 12 Apr 2021 10:14:51 GMT</ValidStartDate>
      <ValidEndDate>Mon, 4 May 2048 10:14:51 GMT</ValidEndDate>
    </Certificate>
  </Cname>
  <Cname>
    <Domain>example.org</Domain>
    <LastModified>2021-09-15T02:34:58.000Z</LastModified>
    <Status>Enabled</Status>
  </Cname>
  <Cname>
    <Domain>example.edu</Domain>
    <LastModified>2021-09-15T02:50:34.000Z</LastModified>
    <Status>Enabled</Status>
  </Cname>
</ListCnameResult>`)
	var res ListBucketCnameResult
	err := xml.Unmarshal(xmlData, &res)
	c.Assert(err, IsNil)
	c.Assert(res.Bucket, Equals, "targetbucket")
	c.Assert(res.Owner, Equals, "testowner")
	c.Assert(res.Cname[0].Domain, Equals, "example.com")
	c.Assert(res.Cname[0].LastModified, Equals, "2021-09-15T02:35:07.000Z")
	c.Assert(res.Cname[0].Status, Equals, "Enabled")
	c.Assert(res.Cname[0].Certificate.Type, Equals, "CAS")
	c.Assert(res.Cname[0].Certificate.CreationDate, Equals, "Wed, 15 Sep 2021 02:35:06 GMT")
	c.Assert(res.Cname[0].Certificate.ValidEndDate, Equals, "Mon, 4 May 2048 10:14:51 GMT")
	c.Assert(res.Cname[1].LastModified, Equals, "2021-09-15T02:34:58.000Z")
	c.Assert(res.Cname[1].Domain, Equals, "example.org")

	c.Assert(res.Cname[2].Domain, Equals, "example.edu")
	c.Assert(res.Cname[2].LastModified, Equals, "2021-09-15T02:50:34.000Z")

}

func (s *OssTypeSuite) TestPutBucketCname(c *C) {
	var putCnameConfig PutBucketCname
	putCnameConfig.Cname = "www.aliyun.com"

	bs, err := xml.Marshal(putCnameConfig)
	c.Assert(err, IsNil)
	c.Assert(string(bs), Equals, "<BucketCnameConfiguration><Cname><Domain>www.aliyun.com</Domain></Cname></BucketCnameConfiguration>")
	certificate := "-----BEGIN CERTIFICATE----- MIIDhDCCAmwCCQCFs8ixARsyrDANBgkqhkiG9w0BAQsFADCBgzELMAkGA1UEBhMC **** -----END CERTIFICATE-----"
	privateKey := "-----BEGIN CERTIFICATE----- MIIDhDCCAmwCCQCFs8ixARsyrDANBgkqhkiG9w0BAQsFADCBgzELMAkGA1UEBhMC **** -----END CERTIFICATE-----"
	var CertificateConfig CertificateConfiguration
	CertificateConfig.CertId = "493****-cn-hangzhou"
	CertificateConfig.Certificate = certificate
	CertificateConfig.PrivateKey = privateKey
	CertificateConfig.Force = true
	putCnameConfig.CertificateConfiguration = &CertificateConfig

	bs, err = xml.Marshal(putCnameConfig)
	c.Assert(err, IsNil)

	testLogger.Println(string(bs))
	c.Assert(string(bs), Equals, "<BucketCnameConfiguration><Cname><Domain>www.aliyun.com</Domain><CertificateConfiguration><CertId>493****-cn-hangzhou</CertId><Certificate>-----BEGIN CERTIFICATE----- MIIDhDCCAmwCCQCFs8ixARsyrDANBgkqhkiG9w0BAQsFADCBgzELMAkGA1UEBhMC **** -----END CERTIFICATE-----</Certificate><PrivateKey>-----BEGIN CERTIFICATE----- MIIDhDCCAmwCCQCFs8ixARsyrDANBgkqhkiG9w0BAQsFADCBgzELMAkGA1UEBhMC **** -----END CERTIFICATE-----</PrivateKey><Force>true</Force></CertificateConfiguration></Cname></BucketCnameConfiguration>")

	var config CertificateConfiguration
	config.DeleteCertificate = true
	putCnameConfig2 := PutBucketCname{
		Cname:                    "www.aliyun.com",
		CertificateConfiguration: &config,
	}

	bs, err = xml.Marshal(putCnameConfig2)
	c.Assert(err, IsNil)
	c.Assert(string(bs), Equals, "<BucketCnameConfiguration><Cname><Domain>www.aliyun.com</Domain><CertificateConfiguration><DeleteCertificate>true</DeleteCertificate></CertificateConfiguration></Cname></BucketCnameConfiguration>")
}

// Test Bucket Style
func (s *OssTypeSuite) TestBucketStyle(c *C) {
	var res GetBucketStyleResult
	xmlData := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<Style>
 <Name>imageStyle</Name>
 <Content>image/resize,p_50</Content>
 <CreateTime>Wed, 20 May 2020 12:07:15 GMT</CreateTime>
 <LastModifyTime>Wed, 21 May 2020 12:07:15 GMT</LastModifyTime>
</Style>`)
	err := xml.Unmarshal(xmlData, &res)
	c.Assert(err, IsNil)
	c.Assert(res.Name, Equals, "imageStyle")
	c.Assert(res.Content, Equals, "image/resize,p_50")
	c.Assert(res.CreateTime, Equals, "Wed, 20 May 2020 12:07:15 GMT")
	c.Assert(res.LastModifyTime, Equals, "Wed, 21 May 2020 12:07:15 GMT")

	var list GetBucketListStyleResult
	xmlData = []byte(`<?xml version="1.0" encoding="UTF-8"?>
<StyleList>
 <Style>
 <Name>imageStyle</Name>
 <Content>image/resize,p_50</Content>
 <CreateTime>Wed, 20 May 2020 12:07:15 GMT</CreateTime>
 <LastModifyTime>Wed, 21 May 2020 12:07:15 GMT</LastModifyTime>
 </Style>
 <Style>
 <Name>imageStyle1</Name>
 <Content>image/resize,w_200</Content>
 <CreateTime>Wed, 20 May 2020 12:08:04 GMT</CreateTime>
 <LastModifyTime>Wed, 21 May 2020 12:08:04 GMT</LastModifyTime>
 </Style>
 <Style>
 <Name>imageStyle3</Name>
 <Content>image/resize,w_300</Content>
 <CreateTime>Fri, 12 Mar 2021 06:19:13 GMT</CreateTime>
 <LastModifyTime>Fri, 13 Mar 2021 06:27:21 GMT</LastModifyTime>
 </Style>
</StyleList>`)
	err = xml.Unmarshal(xmlData, &list)
	c.Assert(err, IsNil)
	c.Assert(list.Style[0].Name, Equals, "imageStyle")
	c.Assert(list.Style[0].Content, Equals, "image/resize,p_50")
	c.Assert(list.Style[0].CreateTime, Equals, "Wed, 20 May 2020 12:07:15 GMT")
	c.Assert(list.Style[0].LastModifyTime, Equals, "Wed, 21 May 2020 12:07:15 GMT")

	c.Assert(err, IsNil)
	c.Assert(list.Style[1].Name, Equals, "imageStyle1")
	c.Assert(list.Style[2].Content, Equals, "image/resize,w_300")
	c.Assert(list.Style[1].CreateTime, Equals, "Wed, 20 May 2020 12:08:04 GMT")
	c.Assert(list.Style[2].LastModifyTime, Equals, "Fri, 13 Mar 2021 06:27:21 GMT")
}

func (s *OssTypeSuite) TestBucketReplication(c *C) {
	// case 1:test PutBucketReplication
	enabled := "enabled"
	prefix1 := "prefix_1"
	prefix2 := "prefix_2"
	keyId := "c4d49f85-ee30-426b-a5ed-95e9139d"
	prefixSet := ReplicationRulePrefix{[]*string{&prefix1, &prefix2}}
	reqReplication := PutBucketReplication{
		Rule: []ReplicationRule{
			{
				RTC:       &enabled,
				PrefixSet: &prefixSet,
				Action:    "all",
				Destination: &ReplicationRuleDestination{
					Bucket:       "srcBucket",
					Location:     "oss-cn-hangzhou",
					TransferType: "oss_acc",
				},
				HistoricalObjectReplication: "disabled",
			},
		},
	}
	xmlData, err := xml.Marshal(reqReplication)
	c.Assert(err, IsNil)
	c.Assert(string(xmlData), Equals, "<ReplicationConfiguration><Rule><RTC><Status>enabled</Status></RTC><PrefixSet><Prefix>prefix_1</Prefix><Prefix>prefix_2</Prefix></PrefixSet><Action>all</Action><Destination><Bucket>srcBucket</Bucket><Location>oss-cn-hangzhou</Location><TransferType>oss_acc</TransferType></Destination><HistoricalObjectReplication>disabled</HistoricalObjectReplication></Rule></ReplicationConfiguration>")

	prefixSet = ReplicationRulePrefix{[]*string{&prefix1, &prefix2}}
	reqReplication = PutBucketReplication{
		Rule: []ReplicationRule{
			{
				RTC:       &enabled,
				PrefixSet: &prefixSet,
				Action:    "all",
				Destination: &ReplicationRuleDestination{
					Bucket:       "srcBucket",
					Location:     "oss-cn-hangzhou",
					TransferType: "oss_acc",
				},
				HistoricalObjectReplication: "disabled",
				SyncRole:                    "aliyunramrole",
				SourceSelectionCriteria:     &enabled,
				EncryptionConfiguration:     &keyId,
			},
		},
	}
	xmlData, err = xml.Marshal(reqReplication)
	testLogger.Println(string(xmlData))
	c.Assert(err, IsNil)
	c.Assert(string(xmlData), Equals, "<ReplicationConfiguration><Rule><RTC><Status>enabled</Status></RTC><PrefixSet><Prefix>prefix_1</Prefix><Prefix>prefix_2</Prefix></PrefixSet><Action>all</Action><Destination><Bucket>srcBucket</Bucket><Location>oss-cn-hangzhou</Location><TransferType>oss_acc</TransferType></Destination><HistoricalObjectReplication>disabled</HistoricalObjectReplication><SyncRole>aliyunramrole</SyncRole><SourceSelectionCriteria><SseKmsEncryptedObjects><Status>enabled</Status></SseKmsEncryptedObjects></SourceSelectionCriteria><EncryptionConfiguration><ReplicaKmsKeyID>c4d49f85-ee30-426b-a5ed-95e9139d</ReplicaKmsKeyID></EncryptionConfiguration></Rule></ReplicationConfiguration>")

	reqReplication = PutBucketReplication{
		Rule: []ReplicationRule{
			{},
		},
	}

	xmlData, err = xml.Marshal(reqReplication)
	testLogger.Println(string(xmlData))
	c.Assert(err, IsNil)
	c.Assert(string(xmlData), Equals, "<ReplicationConfiguration><Rule></Rule></ReplicationConfiguration>")

	reqReplication = PutBucketReplication{
		Rule: []ReplicationRule{},
	}

	xmlData, err = xml.Marshal(reqReplication)
	c.Assert(err, IsNil)
	c.Assert(string(xmlData), Equals, "<ReplicationConfiguration></ReplicationConfiguration>")

	reqReplication = PutBucketReplication{}

	xmlData, err = xml.Marshal(reqReplication)
	c.Assert(err, IsNil)
	c.Assert(string(xmlData), Equals, "<ReplicationConfiguration></ReplicationConfiguration>")

	// case 2: test GetBucketReplicationResult
	xmlData = []byte(`<?xml version="1.0" encoding="UTF-8"?>
<ReplicationConfiguration>
  <Rule>
    <ID>test_replication_1</ID>
    <PrefixSet>
      <Prefix>source1</Prefix>
      <Prefix>video</Prefix>
    </PrefixSet>
    <Action>PUT</Action>
    <Destination>
      <Bucket>destbucket</Bucket>
      <Location>oss-cn-beijing</Location>
      <TransferType>oss_acc</TransferType>
    </Destination>
    <Status>doing</Status>
    <HistoricalObjectReplication>enabled</HistoricalObjectReplication>
    <SyncRole>aliyunramrole</SyncRole>
    <RTC>
      <Status>enabled</Status>
    </RTC>
	<SourceSelectionCriteria>
         <SseKmsEncryptedObjects>
           <Status>Enabled</Status>
         </SseKmsEncryptedObjects>
      </SourceSelectionCriteria>
      <EncryptionConfiguration>
           <ReplicaKmsKeyID>c4d49f85-ee30-426b-a5ed-95e9139d****</ReplicaKmsKeyID>
      </EncryptionConfiguration>
  </Rule>
</ReplicationConfiguration>`)
	var reqResult GetBucketReplicationResult
	err = xml.Unmarshal(xmlData, &reqResult)
	c.Assert(err, IsNil)
	rule := reqResult.Rule[0]
	c.Assert(rule.ID, Equals, "test_replication_1")
	c.Assert(*rule.PrefixSet.Prefix[0], Equals, "source1")
	c.Assert(*rule.PrefixSet.Prefix[1], Equals, "video")
	c.Assert(rule.Action, Equals, "PUT")
	c.Assert(rule.Destination.Bucket, Equals, "destbucket")
	c.Assert(rule.Destination.Location, Equals, "oss-cn-beijing")
	c.Assert(rule.Destination.TransferType, Equals, "oss_acc")

	c.Assert(rule.Status, Equals, "doing")
	c.Assert(rule.HistoricalObjectReplication, Equals, "enabled")
	c.Assert(rule.SyncRole, Equals, "aliyunramrole")
	c.Assert(*rule.RTC, Equals, "enabled")

	c.Assert(*rule.SourceSelectionCriteria, Equals, "Enabled")
	c.Assert(*rule.EncryptionConfiguration, Equals, "c4d49f85-ee30-426b-a5ed-95e9139d****")
}

func (s *OssTypeSuite) TestBucketRtc(c *C) {
	enabled := "enabled"
	id := "300c8809-fe50-4966-bbfa-******"
	reqRtc := PutBucketRTC{
		RTC: &enabled,
		ID:  id,
	}

	xmlData, err := xml.Marshal(reqRtc)
	c.Assert(err, IsNil)
	c.Assert(string(xmlData), Equals, "<ReplicationRule><RTC><Status>enabled</Status></RTC><ID>300c8809-fe50-4966-bbfa-******</ID></ReplicationRule>")

	reqRtc = PutBucketRTC{}
	xmlData, err = xml.Marshal(reqRtc)
	c.Assert(err, IsNil)
	c.Assert(string(xmlData), Equals, "<ReplicationRule></ReplicationRule>")
}

func (s *OssTypeSuite) TestGetBucketReplicationLocationResult(c *C) {
	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<ReplicationLocation>
  <Location>oss-ap-northeast-1</Location>
  <Location>oss-ap-northeast-2</Location>
  <Location>oss-ap-south-1</Location>
  <Location>oss-ap-southeast-1</Location>
  <Location>oss-ap-southeast-2</Location>
  <Location>oss-ap-southeast-3</Location>
  <Location>oss-ap-southeast-5</Location>
  <Location>oss-ap-southeast-6</Location>
  <Location>oss-ap-southeast-7</Location>
  <Location>oss-cn-beijing</Location>
  <Location>oss-cn-chengdu</Location>
  <Location>oss-cn-fuzhou</Location>
  <Location>oss-cn-guangzhou</Location>
  <Location>oss-cn-heyuan</Location>
  <Location>oss-cn-hongkong</Location>
  <Location>oss-cn-huhehaote</Location>
  <Location>oss-cn-nanjing</Location>
  <Location>oss-cn-qingdao</Location>
  <Location>oss-cn-shanghai</Location>
  <Location>oss-cn-shenzhen</Location>
  <Location>oss-cn-wulanchabu</Location>
  <Location>oss-cn-zhangjiakou</Location>
  <Location>oss-eu-central-1</Location>
  <Location>oss-eu-west-1</Location>
  <Location>oss-me-central-1</Location>
  <Location>oss-me-east-1</Location>
  <Location>oss-rus-west-1</Location>
  <Location>oss-us-east-1</Location>
  <Location>oss-us-west-1</Location>
  <LocationTransferTypeConstraint>
    <LocationTransferType>
      <Location>oss-cn-hongkong</Location>
      <TransferTypes>
        <Type>oss_acc</Type>
      </TransferTypes>
    </LocationTransferType>
    <LocationTransferType>
      <Location>oss-eu-central-1</Location>
      <TransferTypes>
        <Type>oss_acc</Type>
      </TransferTypes>
    </LocationTransferType>
    <LocationTransferType>
      <Location>oss-ap-southeast-7</Location>
      <TransferTypes>
        <Type>oss_acc</Type>
      </TransferTypes>
    </LocationTransferType>
    <LocationTransferType>
      <Location>oss-ap-southeast-6</Location>
      <TransferTypes>
        <Type>oss_acc</Type>
      </TransferTypes>
    </LocationTransferType>
    <LocationTransferType>
      <Location>oss-ap-southeast-5</Location>
      <TransferTypes>
        <Type>oss_acc</Type>
      </TransferTypes>
    </LocationTransferType>
    <LocationTransferType>
      <Location>oss-eu-west-1</Location>
      <TransferTypes>
        <Type>oss_acc</Type>
      </TransferTypes>
    </LocationTransferType>
    <LocationTransferType>
      <Location>oss-rus-west-1</Location>
      <TransferTypes>
        <Type>oss_acc</Type>
      </TransferTypes>
    </LocationTransferType>
    <LocationTransferType>
      <Location>oss-ap-southeast-2</Location>
      <TransferTypes>
        <Type>oss_acc</Type>
      </TransferTypes>
    </LocationTransferType>
    <LocationTransferType>
      <Location>oss-ap-southeast-1</Location>
      <TransferTypes>
        <Type>oss_acc</Type>
      </TransferTypes>
    </LocationTransferType>
    <LocationTransferType>
      <Location>oss-me-central-1</Location>
      <TransferTypes>
        <Type>oss_acc</Type>
      </TransferTypes>
    </LocationTransferType>
    <LocationTransferType>
      <Location>oss-ap-south-1</Location>
      <TransferTypes>
        <Type>oss_acc</Type>
      </TransferTypes>
    </LocationTransferType>
    <LocationTransferType>
      <Location>oss-us-east-1</Location>
      <TransferTypes>
        <Type>oss_acc</Type>
      </TransferTypes>
    </LocationTransferType>
    <LocationTransferType>
      <Location>oss-ap-northeast-1</Location>
      <TransferTypes>
        <Type>oss_acc</Type>
      </TransferTypes>
    </LocationTransferType>
    <LocationTransferType>
      <Location>oss-me-east-1</Location>
      <TransferTypes>
        <Type>oss_acc</Type>
      </TransferTypes>
    </LocationTransferType>
    <LocationTransferType>
      <Location>oss-ap-northeast-2</Location>
      <TransferTypes>
        <Type>oss_acc</Type>
      </TransferTypes>
    </LocationTransferType>
    <LocationTransferType>
      <Location>oss-ap-southeast-3</Location>
      <TransferTypes>
        <Type>oss_acc</Type>
      </TransferTypes>
    </LocationTransferType>
    <LocationTransferType>
      <Location>oss-us-west-1</Location>
      <TransferTypes>
        <Type>oss_acc</Type>
      </TransferTypes>
    </LocationTransferType>
  </LocationTransferTypeConstraint>
  <LocationRTCConstraint>
    <Location>oss-cn-beijing</Location>
    <Location>oss-cn-qingdao</Location>
    <Location>oss-cn-shanghai</Location>
    <Location>oss-cn-shenzhen</Location>
    <Location>oss-cn-zhangjiakou</Location>
  </LocationRTCConstraint>
</ReplicationLocation>`
	var repResult GetBucketReplicationLocationResult
	err := xmlUnmarshal(strings.NewReader(xmlData), &repResult)
	c.Assert(err, IsNil)
	c.Assert(repResult.Location[0], Equals, "oss-ap-northeast-1")
	c.Assert(repResult.Location[9], Equals, "oss-cn-beijing")
	testLogger.Println(repResult.LocationTransferType[1].Location)
	c.Assert(repResult.LocationTransferType[1].Location, Equals, "oss-eu-central-1")
	c.Assert(repResult.LocationTransferType[1].TransferTypes, Equals, "oss_acc")
	c.Assert(repResult.RTCLocation[2], Equals, "oss-cn-shanghai")
}

func (s *OssTypeSuite) TestGetBucketReplicationProgressResult(c *C) {
	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<ReplicationProgress>
 <Rule>
   <ID>test_replication_1</ID>
   <PrefixSet>
    <Prefix>source_image</Prefix>
    <Prefix>video</Prefix>
   </PrefixSet>
   <Action>PUT</Action>
   <Destination>
    <Bucket>target-bucket</Bucket>
    <Location>oss-cn-beijing</Location>
    <TransferType>oss_acc</TransferType>
   </Destination>
   <Status>doing</Status>
   <HistoricalObjectReplication>enabled</HistoricalObjectReplication>
   <Progress>
    <HistoricalObject>0.85</HistoricalObject>
    <NewObject>2015-09-24T15:28:14.000Z</NewObject>
   </Progress>
 </Rule>
</ReplicationProgress>`
	var repResult GetBucketReplicationProgressResult
	err := xmlUnmarshal(strings.NewReader(xmlData), &repResult)
	c.Assert(err, IsNil)
	c.Assert(repResult.Rule[0].ID, Equals, "test_replication_1")
	c.Assert(*repResult.Rule[0].PrefixSet.Prefix[0], Equals, "source_image")
	c.Assert(*repResult.Rule[0].PrefixSet.Prefix[1], Equals, "video")
	c.Assert(repResult.Rule[0].Action, Equals, "PUT")
	c.Assert(repResult.Rule[0].Destination.Bucket, Equals, "target-bucket")
	c.Assert(repResult.Rule[0].Destination.Location, Equals, "oss-cn-beijing")
	c.Assert(repResult.Rule[0].Destination.TransferType, Equals, "oss_acc")
	c.Assert(repResult.Rule[0].Status, Equals, "doing")
	c.Assert(repResult.Rule[0].HistoricalObjectReplication, Equals, "enabled")
	c.Assert((*repResult.Rule[0].Progress).HistoricalObject, Equals, "0.85")
	c.Assert((*repResult.Rule[0].Progress).NewObject, Equals, "2015-09-24T15:28:14.000Z")
}

func (s *OssTypeSuite) TestGetBucketRefererResult(c *C) {
	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<RefererConfiguration>
  <AllowEmptyReferer>true</AllowEmptyReferer>
  <RefererList>
  </RefererList>
</RefererConfiguration>`
	var repResult GetBucketRefererResult
	err := xmlUnmarshal(strings.NewReader(xmlData), &repResult)
	c.Assert(err, IsNil)
	c.Assert(repResult.AllowEmptyReferer, Equals, true)
	c.Assert(repResult.AllowTruncateQueryString, IsNil)
	c.Assert(repResult.RefererList, IsNil)
	c.Assert(repResult.RefererBlacklist, IsNil)

	xmlData = `<?xml version="1.0" encoding="UTF-8"?>
<RefererConfiguration>
<AllowEmptyReferer>true</AllowEmptyReferer>
<AllowTruncateQueryString>false</AllowTruncateQueryString>
    <RefererList>
        <Referer>http://www.aliyun.com</Referer>
        <Referer>https://www.aliyun.com</Referer>
        <Referer>http://www.*.com</Referer>
        <Referer>https://www.?.aliyuncs.com</Referer>
    </RefererList>
</RefererConfiguration>`

	var repResult0 GetBucketRefererResult
	err = xmlUnmarshal(strings.NewReader(xmlData), &repResult0)
	c.Assert(err, IsNil)
	c.Assert(repResult0.AllowEmptyReferer, Equals, true)
	c.Assert(*repResult0.AllowTruncateQueryString, Equals, false)
	c.Assert(len(repResult0.RefererList), Equals, 4)
	c.Assert(repResult0.RefererList[1], Equals, "https://www.aliyun.com")
	c.Assert(repResult0.RefererBlacklist, IsNil)

	xmlData = `<?xml version="1.0" encoding="UTF-8"?>
<RefererConfiguration>
  <AllowEmptyReferer>false</AllowEmptyReferer>
  <AllowTruncateQueryString>false</AllowTruncateQueryString>
  <RefererList>
        <Referer>http://www.aliyun.com</Referer>
        <Referer>https://www.aliyun.com</Referer>
        <Referer>http://www.*.com</Referer>
        <Referer>https://www.?.aliyuncs.com</Referer>
  </RefererList>
  <RefererBlacklist>
        <Referer>http://www.refuse.com</Referer>
        <Referer>https://*.hack.com</Referer>
        <Referer>http://ban.*.com</Referer>
    	  <Referer>https://www.?.deny.com</Referer>
  </RefererBlacklist>
</RefererConfiguration>`
	var repResult1 GetBucketRefererResult
	err = xmlUnmarshal(strings.NewReader(xmlData), &repResult1)
	c.Assert(err, IsNil)
	c.Assert(repResult1.AllowEmptyReferer, Equals, false)
	c.Assert(*repResult1.AllowTruncateQueryString, Equals, false)
	c.Assert(len(repResult1.RefererList), Equals, 4)
	c.Assert(repResult1.RefererList[3], Equals, "https://www.?.aliyuncs.com")

	c.Assert(len(repResult1.RefererList), Equals, 4)
	c.Assert((repResult1.RefererBlacklist).Referer[2], Equals, "http://ban.*.com")
}

func (s *OssTypeSuite) TestRefererXML(c *C) {
	xmlData := `<RefererConfiguration><AllowEmptyReferer>true</AllowEmptyReferer><RefererList></RefererList></RefererConfiguration>`
	var setBucketReferer RefererXML
	setBucketReferer.AllowEmptyReferer = true
	setBucketReferer.RefererList = []string{}
	xmlBody, err := xml.Marshal(setBucketReferer)
	log.Println(string(xmlBody))
	c.Assert(err, IsNil)
	c.Assert(string(xmlBody), Equals, xmlData)

	xmlData1 := `<RefererConfiguration><AllowEmptyReferer>true</AllowEmptyReferer><AllowTruncateQueryString>false</AllowTruncateQueryString><RefererList><Referer>http://www.aliyun.com</Referer><Referer>https://www.aliyun.com</Referer><Referer>http://www.*.com</Referer><Referer>https://www.?.aliyuncs.com</Referer></RefererList></RefererConfiguration>`

	setBucketReferer.AllowEmptyReferer = true
	boolFalse := false
	setBucketReferer.AllowTruncateQueryString = &boolFalse
	setBucketReferer.RefererList = []string{
		"http://www.aliyun.com",
		"https://www.aliyun.com",
		"http://www.*.com",
		"https://www.?.aliyuncs.com",
	}
	c.Log(setBucketReferer)
	xmlBody, err = xml.Marshal(setBucketReferer)
	c.Assert(err, IsNil)
	c.Assert(string(xmlBody), Equals, xmlData1)

	xmlData2 := `<RefererConfiguration><AllowEmptyReferer>false</AllowEmptyReferer><AllowTruncateQueryString>false</AllowTruncateQueryString><RefererList><Referer>http://www.aliyun.com</Referer><Referer>https://www.aliyun.com</Referer><Referer>http://www.*.com</Referer><Referer>https://www.?.aliyuncs.com</Referer></RefererList><RefererBlacklist><Referer>http://www.refuse.com</Referer><Referer>https://*.hack.com</Referer><Referer>http://ban.*.com</Referer><Referer>https://www.?.deny.com</Referer></RefererBlacklist></RefererConfiguration>`

	setBucketReferer.AllowEmptyReferer = false
	setBucketReferer.AllowTruncateQueryString = &boolFalse
	setBucketReferer.RefererList = []string{
		"http://www.aliyun.com",
		"https://www.aliyun.com",
		"http://www.*.com",
		"https://www.?.aliyuncs.com",
	}
	referer1 := "http://www.refuse.com"
	referer2 := "https://*.hack.com"
	referer3 := "http://ban.*.com"
	referer4 := "https://www.?.deny.com"
	setBucketReferer.RefererBlacklist = &RefererBlacklist{
		[]string{
			referer1, referer2, referer3, referer4,
		},
	}
	xmlBody, err = xml.Marshal(setBucketReferer)
	c.Assert(err, IsNil)
	c.Assert(string(xmlBody), Equals, xmlData2)
}

func (s *OssTypeSuite) TestDescribeRegionsResult(c *C) {
	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<RegionInfoList>
  <RegionInfo>
     <Region>oss-cn-hangzhou</Region>
     <InternetEndpoint>oss-cn-hangzhou.aliyuncs.com</InternetEndpoint>
     <InternalEndpoint>oss-cn-hangzhou-internal.aliyuncs.com</InternalEndpoint>
     <AccelerateEndpoint>oss-accelerate.aliyuncs.com</AccelerateEndpoint>  
  </RegionInfo>
  <RegionInfo>
     <Region>oss-cn-shanghai</Region>
     <InternetEndpoint>oss-cn-shanghai.aliyuncs.com</InternetEndpoint>
     <InternalEndpoint>oss-cn-shanghai-internal.aliyuncs.com</InternalEndpoint>
     <AccelerateEndpoint>oss-accelerate.aliyuncs.com</AccelerateEndpoint>  
  </RegionInfo>
</RegionInfoList>`
	var repResult DescribeRegionsResult
	err := xmlUnmarshal(strings.NewReader(xmlData), &repResult)
	c.Assert(err, IsNil)
	c.Assert(repResult.Regions[0].Region, Equals, "oss-cn-hangzhou")
	c.Assert(repResult.Regions[0].InternetEndpoint, Equals, "oss-cn-hangzhou.aliyuncs.com")
	c.Assert(repResult.Regions[0].InternalEndpoint, Equals, "oss-cn-hangzhou-internal.aliyuncs.com")
	c.Assert(repResult.Regions[0].AccelerateEndpoint, Equals, "oss-accelerate.aliyuncs.com")

	c.Assert(repResult.Regions[1].Region, Equals, "oss-cn-shanghai")
	c.Assert(repResult.Regions[1].InternetEndpoint, Equals, "oss-cn-shanghai.aliyuncs.com")
	c.Assert(repResult.Regions[1].InternalEndpoint, Equals, "oss-cn-shanghai-internal.aliyuncs.com")
	c.Assert(repResult.Regions[1].AccelerateEndpoint, Equals, "oss-accelerate.aliyuncs.com")
}

func (s *OssTypeSuite) TestAsyncProcessResult(c *C) {
	jsonData := `{"EventId":"10C-1XqxdjCRx3x7gRim3go1yLUVWgm","RequestId":"B8AD6942-BBDE-571D-A9A9-525A4C34B2B3","TaskId":"MediaConvert-58a8f19f-697f-4f8d-ae2c-0d7b15bef68d"}`
	var repResult AsyncProcessObjectResult
	err := jsonUnmarshal(strings.NewReader(jsonData), &repResult)
	c.Assert(err, IsNil)
	c.Assert(repResult.EventId, Equals, "10C-1XqxdjCRx3x7gRim3go1yLUVWgm")
	c.Assert(repResult.RequestId, Equals, "B8AD6942-BBDE-571D-A9A9-525A4C34B2B3")
	c.Assert(repResult.TaskId, Equals, "MediaConvert-58a8f19f-697f-4f8d-ae2c-0d7b15bef68d")
}

func (s *OssTypeSuite) TestPutResponseHeader(c *C) {
	reqHeader := PutBucketResponseHeader{
		Rule: []ResponseHeaderRule{
			{
				Name: "name1",
				Filters: ResponseHeaderRuleFilters{
					[]string{
						"Put", "GetObject",
					},
				},
				HideHeaders: ResponseHeaderRuleHeaders{
					[]string{
						"Last-Modified",
					},
				},
			},
		},
	}
	xmlData, err := xml.Marshal(reqHeader)
	c.Assert(err, IsNil)
	c.Assert(string(xmlData), Equals, "<ResponseHeaderConfiguration><Rule><Name>name1</Name><Filters><Operation>Put</Operation><Operation>GetObject</Operation></Filters><HideHeaders><Header>Last-Modified</Header></HideHeaders></Rule></ResponseHeaderConfiguration>")

	reqHeader = PutBucketResponseHeader{
		Rule: []ResponseHeaderRule{
			{
				Name: "name1",
				Filters: ResponseHeaderRuleFilters{
					[]string{
						"Put", "GetObject",
					},
				},
				HideHeaders: ResponseHeaderRuleHeaders{
					[]string{
						"Last-Modified",
					},
				},
			},
			{
				Name: "name2",
				Filters: ResponseHeaderRuleFilters{
					[]string{
						"*",
					},
				},
				HideHeaders: ResponseHeaderRuleHeaders{
					[]string{
						"Last-Modified",
					},
				},
			},
		},
	}
	xmlData, err = xml.Marshal(reqHeader)
	c.Assert(err, IsNil)
	c.Assert(string(xmlData), Equals, "<ResponseHeaderConfiguration><Rule><Name>name1</Name><Filters><Operation>Put</Operation><Operation>GetObject</Operation></Filters><HideHeaders><Header>Last-Modified</Header></HideHeaders></Rule><Rule><Name>name2</Name><Filters><Operation>*</Operation></Filters><HideHeaders><Header>Last-Modified</Header></HideHeaders></Rule></ResponseHeaderConfiguration>")

	reqHeader = PutBucketResponseHeader{
		Rule: []ResponseHeaderRule{
			{
				Name: "name1",
			},
		},
	}
	xmlData, err = xml.Marshal(reqHeader)
	c.Assert(err, IsNil)
	c.Assert(string(xmlData), Equals, "<ResponseHeaderConfiguration><Rule><Name>name1</Name><Filters></Filters><HideHeaders></HideHeaders></Rule></ResponseHeaderConfiguration>")
}

func (s *OssTypeSuite) TestGetResponseHeaderResult(c *C) {
	xmlData := `<ResponseHeaderConfiguration>
    <Rule>
        <Name>rule1</Name>
        <Filters>
            <Operation>Put</Operation>
            <Operation>GetObject</Operation>
        </Filters>
        <HideHeaders>
           	<Header>Last-Modified</Header>
			<Header>x-oss-request-id</Header>
        </HideHeaders>
    </Rule>
	<Rule>
        <Name>rule2</Name>
        <Filters>
            <Operation>*</Operation>
        </Filters>
        <HideHeaders>
           	<Header>Last-Modified</Header>
			<Header>x-oss-request-id</Header>
        </HideHeaders>
    </Rule>
</ResponseHeaderConfiguration>`
	var repResult GetBucketResponseHeaderResult
	err := xmlUnmarshal(strings.NewReader(xmlData), &repResult)
	c.Assert(err, IsNil)
	c.Assert(repResult.Rule[0].Name, Equals, "rule1")
	c.Assert(repResult.Rule[0].Filters.Operation[0], Equals, "Put")
	c.Assert(repResult.Rule[0].HideHeaders.Header[0], Equals, "Last-Modified")

	c.Assert(repResult.Rule[1].Name, Equals, "rule2")
	c.Assert(repResult.Rule[1].Filters.Operation[0], Equals, "*")
	c.Assert(repResult.Rule[1].HideHeaders.Header[0], Equals, "Last-Modified")
}

func (s *OssTypeSuite) TestGetBucketCORSResult(c *C) {
	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<CORSConfiguration>
    <CORSRule>
      <AllowedOrigin>*</AllowedOrigin>
      <AllowedMethod>PUT</AllowedMethod>
      <AllowedMethod>GET</AllowedMethod>
      <AllowedHeader>Authorization</AllowedHeader>
    </CORSRule>
    <CORSRule>
      <AllowedOrigin>http://example.com</AllowedOrigin>
      <AllowedOrigin>http://example.net</AllowedOrigin>
      <AllowedMethod>GET</AllowedMethod>
      <AllowedHeader>Authorization</AllowedHeader>
      <ExposeHeader>x-oss-test</ExposeHeader>
      <ExposeHeader>x-oss-test1</ExposeHeader>
      <MaxAgeSeconds>100</MaxAgeSeconds>
    </CORSRule>
    <ResponseVary>false</ResponseVary>
</CORSConfiguration>`
	var repResult GetBucketCORSResult
	err := xmlUnmarshal(strings.NewReader(xmlData), &repResult)
	c.Assert(err, IsNil)
	c.Assert(repResult.CORSRules[0].AllowedOrigin[0], Equals, "*")
	c.Assert(repResult.CORSRules[0].AllowedMethod[0], Equals, "PUT")
	c.Assert(repResult.CORSRules[0].AllowedMethod[1], Equals, "GET")
	c.Assert(repResult.CORSRules[0].AllowedHeader[0], Equals, "Authorization")

	c.Assert(repResult.CORSRules[1].AllowedOrigin[0], Equals, "http://example.com")
	c.Assert(repResult.CORSRules[1].AllowedOrigin[1], Equals, "http://example.net")
	c.Assert(repResult.CORSRules[1].AllowedMethod[0], Equals, "GET")
	c.Assert(repResult.CORSRules[1].AllowedHeader[0], Equals, "Authorization")
	c.Assert(repResult.CORSRules[1].ExposeHeader[0], Equals, "x-oss-test")
	c.Assert(repResult.CORSRules[1].ExposeHeader[1], Equals, "x-oss-test1")
	c.Assert(repResult.CORSRules[1].MaxAgeSeconds, Equals, int(100))
	c.Assert(*repResult.ResponseVary, Equals, false)
}

func (s *OssTypeSuite) TestPutBucketCORS(c *C) {
	isTrue := true
	rule1 := CORSRule{
		AllowedOrigin: []string{"*"},
		AllowedMethod: []string{"PUT", "GET", "POST"},
		AllowedHeader: []string{},
		ExposeHeader:  []string{},
		MaxAgeSeconds: 100,
	}

	rule2 := CORSRule{
		AllowedOrigin: []string{"http://www.a.com", "http://www.b.com"},
		AllowedMethod: []string{"GET"},
		AllowedHeader: []string{"Authorization"},
		ExposeHeader:  []string{"x-oss-test", "x-oss-test1"},
		MaxAgeSeconds: 100,
	}

	put := PutBucketCORS{}
	put.CORSRules = []CORSRule{rule1, rule2}
	put.ResponseVary = &isTrue

	bs, err := xml.Marshal(put)
	c.Assert(err, IsNil)
	c.Assert(string(bs), Equals, "<CORSConfiguration><CORSRule><AllowedOrigin>*</AllowedOrigin><AllowedMethod>PUT</AllowedMethod><AllowedMethod>GET</AllowedMethod><AllowedMethod>POST</AllowedMethod><MaxAgeSeconds>100</MaxAgeSeconds></CORSRule><CORSRule><AllowedOrigin>http://www.a.com</AllowedOrigin><AllowedOrigin>http://www.b.com</AllowedOrigin><AllowedMethod>GET</AllowedMethod><AllowedHeader>Authorization</AllowedHeader><ExposeHeader>x-oss-test</ExposeHeader><ExposeHeader>x-oss-test1</ExposeHeader><MaxAgeSeconds>100</MaxAgeSeconds></CORSRule><ResponseVary>true</ResponseVary></CORSConfiguration>")
}
