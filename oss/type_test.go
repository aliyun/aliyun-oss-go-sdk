package oss

import (
	"encoding/xml"
	. "gopkg.in/check.v1"
	"net/url"
	"sort"
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

func (s *OssTypeSuite) TestLifeCycleRulesWithFilter(c *C) {
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
      <Days>Days</Days>
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
	err = xml.Unmarshal(xmlData, &res)
	c.Assert(err, IsNil)
	c.Assert(res.Rules[0].ID, Equals, "test2")
	c.Assert(res.Rules[0].Filter.Not[0].Prefix, Equals, "logs-demo")
	c.Assert(res.Rules[0].Filter.Not[1].Prefix, Equals, "abc/not1/")
	c.Assert(res.Rules[0].Filter.Not[1].Tag.Key, Equals, "notkey1")
	c.Assert(res.Rules[0].Filter.Not[1].Tag.Value, Equals, "notvalue1")
	c.Assert(res.Rules[0].Filter.Not[2].Prefix, Equals, "abc/not2/")
	c.Assert(res.Rules[0].Filter.Not[2].Tag.Key, Equals, "notkey2")
	c.Assert(res.Rules[0].Filter.Not[2].Tag.Value, Equals, "notvalue2")
}
