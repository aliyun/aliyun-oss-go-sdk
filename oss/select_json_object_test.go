package oss

import (
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	. "gopkg.in/check.v1"
)

type OssSelectJsonSuite struct {
	cloudBoxControlClient *Client
	client                *Client
	bucket                *Bucket
}

var _ = Suite(&OssSelectJsonSuite{})

func (s *OssSelectJsonSuite) SetUpSuite(c *C) {
	bucketName := bucketNamePrefix + RandLowStr(6)
	if cloudboxControlEndpoint == "" {
		client, err := New(endpoint, accessID, accessKey)
		c.Assert(err, IsNil)
		s.client = client
		s.client.Config.LogLevel = Error // Debug
		err = s.client.CreateBucket(bucketName)
		c.Assert(err, IsNil)
		bucket, err := s.client.Bucket(bucketName)
		c.Assert(err, IsNil)
		s.bucket = bucket
	} else {
		client, err := New(cloudboxEndpoint, accessID, accessKey)
		c.Assert(err, IsNil)
		s.client = client

		controlClient, err := New(cloudboxControlEndpoint, accessID, accessKey)
		c.Assert(err, IsNil)
		s.cloudBoxControlClient = controlClient
		controlClient.CreateBucket(bucketName)

		bucket, err := s.client.Bucket(bucketName)
		c.Assert(err, IsNil)
		s.bucket = bucket
	}

	testLogger.Println("test select json started")
}

func (s *OssSelectJsonSuite) TearDownSuite(c *C) {
	// Delete objects
	marker := Marker("")
	for {
		lor, err := s.bucket.ListObjects(marker)
		c.Assert(err, IsNil)
		for _, object := range lor.Objects {
			err = s.bucket.DeleteObject(object.Key)
			c.Assert(err, IsNil)
		}
		marker = Marker(lor.NextMarker)
		if !lor.IsTruncated {
			break
		}
	}

	// Delete bucket
	if s.cloudBoxControlClient != nil {
		err := s.cloudBoxControlClient.DeleteBucket(s.bucket.BucketName)
		c.Assert(err, IsNil)
	} else {
		err := s.client.DeleteBucket(s.bucket.BucketName)
		c.Assert(err, IsNil)
	}

	testLogger.Println("test select json completed")
}

func (s *OssSelectJsonSuite) SetUpTest(c *C) {
	testLogger.Println("test func", c.TestName(), "start")
}

func (s *OssSelectJsonSuite) TearDownTest(c *C) {
	testLogger.Println("test func", c.TestName(), "succeed")
}

func (s *OssSelectJsonSuite) TestCreateSelectJsonObjectMeta(c *C) {
	key := "sample_json_lines.json"
	err := s.bucket.PutObjectFromFile(key, "../sample/sample_json_lines.json")
	c.Assert(err, IsNil)
	jsonMeta := JsonMetaRequest{
		InputSerialization: InputSerialization{
			JSON: JSON{
				JSONType: "LINES",
			},
		},
	}
	res, err := s.bucket.CreateSelectJsonObjectMeta(key, jsonMeta)
	c.Assert(err, IsNil)
	c.Assert(res.RowsCount, Equals, int64(100))

	err = s.bucket.DeleteObject(key)
	c.Assert(err, IsNil)
}

func (s *OssSelectJsonSuite) TestSelectJsonDocument(c *C) {
	key := "sample_json.json"
	err := s.bucket.PutObjectFromFile(key, "../sample/sample_json.json")
	c.Assert(err, IsNil)
	selReq := SelectRequest{}
	selReq.Expression = "select * from ossobject.objects[*] where party = 'Democrat'"
	selReq.InputSerializationSelect.JsonBodyInput.JSONType = "DOCUMENT"
	selReq.OutputSerializationSelect.JsonBodyOutput.RecordDelimiter = ","

	var responseHeader http.Header
	body, err := s.bucket.SelectObject(key, selReq, GetResponseHeader(&responseHeader))
	c.Assert(err, IsNil)
	defer body.Close()
	requestId := GetRequestId(responseHeader)
	c.Assert(len(requestId) > 0, Equals, true)

	p := make([]byte, 512)
	n, err := body.Read(p)
	c.Assert(err, IsNil)
	c.Assert(n, Equals, 512)
	p1 := make([]byte, 3)
	_, err = body.Read(p1)
	c.Assert(err, IsNil)
	rets, err := ioutil.ReadAll(body)
	c.Assert(err, IsNil)
	str, err := readJsonDocument("../sample/sample_json.json")
	c.Assert(err, IsNil)
	c.Assert(string(p)+string(p1)+string(rets), Equals, escaped_slashs(str))

	err = s.bucket.DeleteObject(key)
	c.Assert(err, IsNil)
}

func (s *OssSelectJsonSuite) TestSelectJsonLines(c *C) {
	key := "sample_json_lines.json"
	err := s.bucket.PutObjectFromFile(key, "../sample/sample_json_lines.json")
	c.Assert(err, IsNil)
	selReq := SelectRequest{}
	selReq.Expression = "select * from ossobject where party = 'Democrat'"
	selReq.OutputSerializationSelect.JsonBodyOutput.RecordDelimiter = ","
	selReq.InputSerializationSelect.JsonBodyInput.JSONType = "LINES"

	var responseHeader http.Header
	body, err := s.bucket.SelectObject(key, selReq, GetResponseHeader(&responseHeader))
	c.Assert(err, IsNil)
	defer body.Close()

	requestId := GetRequestId(responseHeader)
	c.Assert(len(requestId) > 0, Equals, true)

	rets, err := ioutil.ReadAll(body)
	c.Assert(err, IsNil)
	str, err := readJsonDocument("../sample/sample_json.json")
	c.Assert(string(rets), Equals, escaped_slashs(str))

	err = s.bucket.DeleteObject(key)
	c.Assert(err, IsNil)
}

func (s *OssSelectJsonSuite) TestSelectJsonLinesIntoFile(c *C) {
	key := "sample_json_lines.json"
	err := s.bucket.PutObjectFromFile(key, "../sample/sample_json_lines.json")
	c.Assert(err, IsNil)

	jsonMeta := JsonMetaRequest{
		InputSerialization: InputSerialization{
			JSON: JSON{
				JSONType: "LINES",
			},
		},
	}
	res, err := s.bucket.CreateSelectJsonObjectMeta(key, jsonMeta)
	c.Assert(err, IsNil)
	c.Assert(res.RowsCount, Equals, int64(100))

	selReq := SelectRequest{}
	selReq.Expression = "select * from ossobject where party = 'Democrat'"
	selReq.OutputSerializationSelect.JsonBodyOutput.RecordDelimiter = ","
	selReq.InputSerializationSelect.JsonBodyInput.JSONType = "LINES"

	var responseHeader http.Header
	outfile := "sample_json_out.json"
	err = s.bucket.SelectObjectIntoFile(key, outfile, selReq, GetResponseHeader(&responseHeader))
	c.Assert(err, IsNil)
	requestId := GetRequestId(responseHeader)
	c.Assert(len(requestId) > 0, Equals, true)

	_, err = os.Stat(outfile)
	c.Assert(err, IsNil)
	err = os.Remove(outfile)
	c.Assert(err, IsNil)

	err = s.bucket.DeleteObject(key)
	c.Assert(err, IsNil)
}

func (s *OssSelectJsonSuite) TestSelectJsonDocumentIntoFile(c *C) {
	key := "sample_json_lines.json"
	err := s.bucket.PutObjectFromFile(key, "../sample/sample_json.json")
	c.Assert(err, IsNil)

	selReq := SelectRequest{}
	selReq.Expression = "select * from ossobject.objects[*] where party = 'Democrat'"
	selReq.OutputSerializationSelect.JsonBodyOutput.RecordDelimiter = ","
	selReq.InputSerializationSelect.JsonBodyInput.JSONType = "DOCUMENT"

	var responseHeader http.Header
	outfile := "sample_json_out.json"
	err = s.bucket.SelectObjectIntoFile(key, outfile, selReq, GetResponseHeader(&responseHeader))
	c.Assert(err, IsNil)
	requestId := GetRequestId(responseHeader)
	c.Assert(len(requestId) > 0, Equals, true)

	_, err = os.Stat(outfile)
	c.Assert(err, IsNil)
	err = os.Remove(outfile)
	c.Assert(err, IsNil)

	err = s.bucket.DeleteObject(key)
	c.Assert(err, IsNil)
}

func (s *OssSelectJsonSuite) TestSelectJsonLinesLike(c *C) {
	key := "sample_json_lines.json"
	err := s.bucket.PutObjectFromFile(key, "../sample/sample_json_lines.json")
	c.Assert(err, IsNil)
	selReq := SelectRequest{}
	selReq.Expression = "select person.firstname, person.lastname from ossobject where person.birthday like '1959%'"
	selReq.OutputSerializationSelect.JsonBodyOutput.RecordDelimiter = ","
	selReq.InputSerializationSelect.JsonBodyInput.JSONType = "LINES"

	jsonMeta := JsonMetaRequest{
		InputSerialization: InputSerialization{
			JSON: JSON{
				JSONType: "LINES",
			},
		},
	}
	res, err := s.bucket.CreateSelectJsonObjectMeta(key, jsonMeta)
	c.Assert(err, IsNil)
	c.Assert(res.RowsCount, Equals, int64(100))

	var responseHeader http.Header
	body, err := s.bucket.SelectObject(key, selReq, GetResponseHeader(&responseHeader))
	c.Assert(err, IsNil)
	defer body.Close()

	requestId := GetRequestId(responseHeader)
	c.Assert(len(requestId) > 0, Equals, true)

	rets, err := ioutil.ReadAll(body)
	c.Assert(err, IsNil)
	str, err := readJsonLinesLike("../sample/sample_json.json")
	c.Assert(string(rets), Equals, str)

	err = s.bucket.DeleteObject(key)
	c.Assert(err, IsNil)
}

func (s *OssSelectJsonSuite) TestSelectJsonLinesRange(c *C) {
	key := "sample_json_lines.json"
	err := s.bucket.PutObjectFromFile(key, "../sample/sample_json_lines.json")
	c.Assert(err, IsNil)
	jsonMeta := JsonMetaRequest{
		InputSerialization: InputSerialization{
			JSON: JSON{
				JSONType: "LINES",
			},
		},
	}
	res, err := s.bucket.CreateSelectJsonObjectMeta(key, jsonMeta)
	c.Assert(err, IsNil)
	c.Assert(res.RowsCount, Equals, int64(100))

	selReq := SelectRequest{}
	selReq.Expression = "select person.firstname as aaa as firstname, person.lastname, extra from ossobject'"
	selReq.OutputSerializationSelect.JsonBodyOutput.RecordDelimiter = ","
	selReq.InputSerializationSelect.JsonBodyInput.JSONType = "LINES"
	selReq.InputSerializationSelect.JsonBodyInput.Range = "0-1"

	var responseHeader http.Header
	body, err := s.bucket.SelectObject(key, selReq, GetResponseHeader(&responseHeader))
	c.Assert(err, IsNil)
	defer body.Close()

	requestId := GetRequestId(responseHeader)
	c.Assert(len(requestId) > 0, Equals, true)

	rets, err := ioutil.ReadAll(body)
	c.Assert(err, IsNil)
	str, err := readJsonLinesRange("../sample/sample_json.json", 0, 2)
	c.Assert(string(rets), Equals, escaped_slashs(str))

	err = s.bucket.DeleteObject(key)
	c.Assert(err, IsNil)
}

func (s *OssSelectJsonSuite) TestSelectJsonDocumentIntAggregation(c *C) {
	key := "sample_json.json"
	err := s.bucket.PutObjectFromFile(key, "../sample/sample_json.json")
	c.Assert(err, IsNil)

	selReq := SelectRequest{}
	selReq.Expression = `
	select 
		avg(cast(person.cspanid as int)), max(cast(person.cspanid as int)), 
		min(cast(person.cspanid as int)) 
	from 
		ossobject.objects[*] 
	where 
		person.cspanid = 1011723
	`
	selReq.OutputSerializationSelect.JsonBodyOutput.RecordDelimiter = ","
	selReq.InputSerializationSelect.JsonBodyInput.JSONType = "Document"

	body, err := s.bucket.SelectObject(key, selReq)
	c.Assert(err, IsNil)
	defer body.Close()

	rets, err := ioutil.ReadAll(body)
	c.Assert(err, IsNil)
	c.Assert(string(rets), Equals, "{\"_1\":1011723,\"_2\":1011723,\"_3\":1011723},")

	err = s.bucket.DeleteObject(key)
	c.Assert(err, IsNil)
}

func (s *OssSelectJsonSuite) TestSelectJsonDocumentFloatAggregation(c *C) {
	key := "sample_json.json"
	err := s.bucket.PutObjectFromFile(key, "../sample/sample_json.json")
	c.Assert(err, IsNil)

	selReq := SelectRequest{}
	selReq.Expression = `
	select 
		avg(cast(person.cspanid as double)), max(cast(person.cspanid as double)), 
		min(cast(person.cspanid as double)) 
	from 
		ossobject.objects[*]
	`
	selReq.OutputSerializationSelect.JsonBodyOutput.RecordDelimiter = ","
	selReq.InputSerializationSelect.JsonBodyInput.JSONType = "Document"

	body, err := s.bucket.SelectObject(key, selReq)
	c.Assert(err, IsNil)
	defer body.Close()

	rets, err := ioutil.ReadAll(body)
	c.Assert(err, IsNil)
	testLogger.Println(string(rets))
	// avg, max, min, err := readJsonFloatAggregation("../sample/sample_json.json")
	// fmt.Println(string(rets), "\n", avg, max, min)
	// retsArr := strings.Split(string(rets), ":")
	// s1 := strconv.FormatFloat(avg, 'f', 6, 64) + ","
	// s1 += strconv.FormatFloat(max, 'f', 6, 64) + ","
	// s1 += strconv.FormatFloat(min, 'f', 6, 64) + ","
	// retS := ""
	// l := len(retsArr[1])
	// vv, err := strconv.ParseFloat(retsArr[1][:l-35], 64)
	// c.Assert(err, IsNil)
	// retS += strconv.FormatFloat(vv, 'f', 6, 64) + ","
	// l = len(retsArr[2])
	// vv, err = strconv.ParseFloat(retsArr[2][:l-6], 64)
	// c.Assert(err, IsNil)
	// retS += strconv.FormatFloat(vv, 'f', 6, 64) + ","
	// l = len(retsArr[3])
	// vv, err = strconv.ParseFloat(retsArr[3][:l-2], 64)
	// c.Assert(err, IsNil)
	// retS += strconv.FormatFloat(vv, 'f', 6, 64) + ","
	// c.Assert(retS, Equals, s1)

	err = s.bucket.DeleteObject(key)
	c.Assert(err, IsNil)
}

func (s *OssSelectJsonSuite) TestSelectJsonDocumentConcat(c *C) {
	key := "sample_json.json"
	err := s.bucket.PutObjectFromFile(key, "../sample/sample_json.json")
	c.Assert(err, IsNil)

	selReq := SelectRequest{}
	selReq.Expression = `
	select 
		person 
	from 
		ossobject.objects[*] 
	where 
		(person.firstname || person.lastname) = 'JohnKennedy'
	`
	selReq.OutputSerializationSelect.JsonBodyOutput.RecordDelimiter = ","
	selReq.InputSerializationSelect.JsonBodyInput.JSONType = "Document"

	body, err := s.bucket.SelectObject(key, selReq)
	c.Assert(err, IsNil)
	defer body.Close()

	rets, err := ioutil.ReadAll(body)
	c.Assert(err, IsNil)
	str, err := readJsonDocumentConcat("../sample/sample_json.json")
	c.Assert(err, IsNil)
	c.Assert(string(rets), Equals, escaped_slashs(str))

	err = s.bucket.DeleteObject(key)
	c.Assert(err, IsNil)
}

func (s *OssSelectJsonSuite) TestSelectJsonComplicateConcat(c *C) {
	key := "sample_json.json"
	err := s.bucket.PutObjectFromFile(key, "../sample/sample_json_lines.json")
	c.Assert(err, IsNil)

	selReq := SelectRequest{}
	selReq.Expression = `
	select 
		person.firstname, person.lastname, congress_numbers 
	from
		ossobject 
	where
		startdate > '2017-01-01' and 
		senator_rank = 'junior' or 
		state = 'CA' and 
		party = 'Republican'
	`
	selReq.OutputSerializationSelect.JsonBodyOutput.RecordDelimiter = ","
	selReq.InputSerializationSelect.JsonBodyInput.JSONType = "LINES"

	body, err := s.bucket.SelectObject(key, selReq)
	c.Assert(err, IsNil)
	defer body.Close()

	rets, err := ioutil.ReadAll(body)
	c.Assert(err, IsNil)
	str, err := readJsonComplicateConcat("../sample/sample_json.json")
	c.Assert(err, IsNil)
	c.Assert(string(rets), Equals, str)

	err = s.bucket.DeleteObject(key)
	c.Assert(err, IsNil)
}

func (s *OssSelectJsonSuite) TestSelectJsonLineInvalidSql(c *C) {
	key := "sample_json.json"
	err := s.bucket.PutObjectFromFile(key, "../sample/sample_json_lines.json")
	c.Assert(err, IsNil)

	selReq := SelectRequest{}
	selReq.OutputSerializationSelect.JsonBodyOutput.RecordDelimiter = ","
	selReq.InputSerializationSelect.JsonBodyInput.JSONType = "LINES"

	selReq.Expression = `select * from ossobject where avg(cast(person.birthday as int)) > 2016`
	_, err = s.bucket.SelectObject(key, selReq)
	c.Assert(err, NotNil)

	selReq.Expression = ``
	_, err = s.bucket.SelectObject(key, selReq)
	c.Assert(err, NotNil)

	selReq.Expression = `select person.lastname || person.firstname from ossobject`
	_, err = s.bucket.SelectObject(key, selReq)
	c.Assert(err, NotNil)

	selReq.Expression = `select * from ossobject group by person.firstname`
	_, err = s.bucket.SelectObject(key, selReq)
	c.Assert(err, NotNil)

	selReq.Expression = `select * from ossobject order by _1`
	_, err = s.bucket.SelectObject(key, selReq)
	c.Assert(err, NotNil)

	selReq.Expression = `select * from ossobject oss join s3object s3 on oss.CityName = s3.CityName`
	_, err = s.bucket.SelectObject(key, selReq)
	c.Assert(err, NotNil)

	err = s.bucket.PutObjectFromFile(key, "../sample/sample_json.json")
	c.Assert(err, IsNil)
	selReq.InputSerializationSelect.JsonBodyInput.JSONType = "DOCUMENT"
	selReq.Expression = `select _1 from ossobject.objects[*]`
	body, err := s.bucket.SelectObject(key, selReq)
	c.Assert(err, IsNil)
	defer body.Close()

	err = s.bucket.DeleteObject(key)
	c.Assert(err, IsNil)
}

func (s *OssSelectJsonSuite) TestSelectJsonParseNumAsString(c *C) {
	key := "sample_json.json"
	content := "{\"a\":123456789.123456789}"
	err := s.bucket.PutObject(key, strings.NewReader(content))
	c.Assert(err, IsNil)

	selReq := SelectRequest{}
	selReq.Expression = `select a from ossobject where cast(a as decimal) = 123456789.1234567890`
	bo := true
	selReq.InputSerializationSelect.JsonBodyInput.ParseJSONNumberAsString = &bo
	selReq.InputSerializationSelect.JsonBodyInput.JSONType = "DOCUMENT"

	body, err := s.bucket.SelectObject(key, selReq)
	c.Assert(err, IsNil)
	defer body.Close()

	rets, err := ioutil.ReadAll(body)
	c.Assert(err, IsNil)
	c.Assert(string(rets), Equals, "{\"a\":123456789.123456789}\n")

	err = s.bucket.DeleteObject(key)
	c.Assert(err, IsNil)
}

func escaped_slashs(value string) string {
	return strings.Replace(value, "/", "\\/", -1)
}
