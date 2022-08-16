package oss

import (
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	. "gopkg.in/check.v1"
)

type OssSelectCsvSuite struct {
	cloudBoxControlClient *Client
	client                *Client
	bucket                *Bucket
}

var _ = Suite(&OssSelectCsvSuite{})

func (s *OssSelectCsvSuite) SetUpSuite(c *C) {
	bucketName := bucketNamePrefix + RandLowStr(6)
	if cloudboxControlEndpoint == "" {
		client, err := New(endpoint, accessID, accessKey)
		c.Assert(err, IsNil)
		s.client = client
		s.client.Config.LogLevel = Error // Debug
		// s.client.Config.Timeout = 5
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

	testLogger.Println("test select csv started")
}

func (s *OssSelectCsvSuite) TearDownSuite(c *C) {
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

	testLogger.Println("test select csv completed")
}

func (s *OssSelectCsvSuite) SetUpTest(c *C) {
	testLogger.Println("test func", c.TestName(), "start")
}

func (s *OssSelectCsvSuite) TearDownTest(c *C) {
	testLogger.Println("test func", c.TestName(), "succeed")
}

// TestCreateSelectObjectMeta
func (s *OssSelectCsvSuite) TestCreateSelectCsvObjectMeta(c *C) {
	key := "sample_data.csv"
	localCsvFile := "../sample/sample_data.csv"
	err := s.bucket.PutObjectFromFile(key, localCsvFile)
	c.Assert(err, IsNil)
	csvMeta := CsvMetaRequest{}
	var bo bool
	csvMeta.OverwriteIfExists = &bo
	res, err := s.bucket.CreateSelectCsvObjectMeta(key, csvMeta)
	c.Assert(err, IsNil)
	l, err := readCsvLine(localCsvFile)
	c.Assert(err, IsNil)
	c.Assert(res.RowsCount, Equals, int64(l))

	bo = true
	csvMeta.OverwriteIfExists = &bo
	csvMeta.InputSerialization.CSV.RecordDelimiter = "\n"
	csvMeta.InputSerialization.CSV.FieldDelimiter = ","
	csvMeta.InputSerialization.CSV.QuoteCharacter = "\""
	res, err = s.bucket.CreateSelectCsvObjectMeta(key, csvMeta)
	c.Assert(err, IsNil)
	c.Assert(res.RowsCount, Equals, int64(l))

	err = s.bucket.DeleteObject(key)
	c.Assert(err, IsNil)
}

func (s *OssSelectCsvSuite) TestSelectCsvObjectIsEmpty(c *C) {
	key := "sample_data.csv"
	localCsvFile := "../sample/sample_data.csv"
	err := s.bucket.PutObjectFromFile(key, localCsvFile)
	c.Assert(err, IsNil)
	csvMeta := CsvMetaRequest{}
	_, err = s.bucket.CreateSelectCsvObjectMeta(key, csvMeta)
	c.Assert(err, IsNil)
	selReq := SelectRequest{}
	selReq.Expression = "select Year, StateAbbr, CityName, PopulationCount from ossobject where CityName != ''"
	selReq.InputSerializationSelect.CsvBodyInput.FileHeaderInfo = "Use"

	body, err := s.bucket.SelectObject(key, selReq)
	c.Assert(err, IsNil)
	defer body.Close()

	p := make([]byte, 512)
	n, err := body.Read(p)
	c.Assert(err, IsNil)
	c.Assert(n, Equals, 512)
	p1 := make([]byte, 3)
	_, err = body.Read(p1)
	c.Assert(err, IsNil)
	rets, err := ioutil.ReadAll(body)
	c.Assert(err, IsNil)
	str, err := readCsvIsEmpty(localCsvFile)
	c.Assert(err, IsNil)
	c.Assert(string(p)+string(p1)+string(rets), Equals, str)

	err = s.bucket.DeleteObject(key)
	c.Assert(err, IsNil)
}

func (s *OssSelectCsvSuite) TestSelectObjectIntoFile(c *C) {
	var bo bool = true
	key := "sample_data.csv"
	localCsvFile := "../sample/sample_data.csv"
	err := s.bucket.PutObjectFromFile(key, localCsvFile)
	c.Assert(err, IsNil)
	csvMeta := CsvMetaRequest{
		InputSerialization: InputSerialization{
			CSV: CSV{
				RecordDelimiter: "\n",
				FieldDelimiter:  ",",
				QuoteCharacter:  "\"",
			},
		},
		OverwriteIfExists: &bo,
	}
	res, err := s.bucket.CreateSelectCsvObjectMeta(key, csvMeta)
	c.Assert(err, IsNil)
	l, err := readCsvLine(localCsvFile)
	c.Assert(err, IsNil)
	c.Assert(res.RowsCount, Equals, int64(l))

	selReq := SelectRequest{
		Expression: "select * from ossobject",
		InputSerializationSelect: InputSerializationSelect{
			CsvBodyInput: CSVSelectInput{
				FileHeaderInfo:   "None",
				CommentCharacter: "#",
				RecordDelimiter:  "\n",
				FieldDelimiter:   ",",
				QuoteCharacter:   "\"",
				Range:            "",
			},
		},
	}
	outfile := "sample_data_out.csv"
	err = s.bucket.SelectObjectIntoFile(key, outfile, selReq)
	c.Assert(err, IsNil)

	fd1, err := os.Open(outfile)
	c.Assert(err, IsNil)
	defer fd1.Close()
	fd2, err := os.Open(localCsvFile)
	c.Assert(err, IsNil)
	defer fd2.Close()
	str1, err := ioutil.ReadAll(fd1)
	c.Assert(err, IsNil)
	str2, err := ioutil.ReadAll(fd2)
	c.Assert(err, IsNil)
	c.Assert(string(str1), Equals, string(str2))

	err = os.Remove(outfile)
	c.Assert(err, IsNil)
	err = s.bucket.DeleteObject(key)
	c.Assert(err, IsNil)
}

func (s *OssSelectCsvSuite) TestSelectCsvObjectRange(c *C) {
	key := "sample_data.csv"
	localCsvFile := "../sample/sample_data.csv"
	err := s.bucket.PutObjectFromFile(key, localCsvFile)
	c.Assert(err, IsNil)
	csvMeta := CsvMetaRequest{}
	_, err = s.bucket.CreateSelectCsvObjectMeta(key, csvMeta)
	c.Assert(err, IsNil)
	selReq := SelectRequest{}
	selReq.Expression = "select Year,StateAbbr, CityName, Short_Question_Text from ossobject"
	selReq.InputSerializationSelect.CsvBodyInput.FileHeaderInfo = "Use"
	selReq.InputSerializationSelect.CsvBodyInput.Range = "0-2"
	body, err := s.bucket.SelectObject(key, selReq)
	c.Assert(err, IsNil)
	defer body.Close()
	rets, err := ioutil.ReadAll(body)

	str, err := readCsvRange(localCsvFile, 0, 2)
	c.Assert(err, IsNil)
	c.Assert(string(rets), Equals, str)

	err = s.bucket.DeleteObject(key)
	c.Assert(err, IsNil)
}

func (s *OssSelectCsvSuite) TestSelectCsvObjectLike(c *C) {
	key := "sample_data.csv"
	localCsvFile := "../sample/sample_data.csv"
	err := s.bucket.PutObjectFromFile(key, localCsvFile)
	c.Assert(err, IsNil)
	selReq := SelectRequest{}
	selReq.Expression = "select Year, StateAbbr, CityName, Short_Question_Text from ossobject where Measure like '%blood pressure%Years'"
	selReq.InputSerializationSelect.CsvBodyInput.FileHeaderInfo = "Use"
	ret, err := s.bucket.SelectObject(key, selReq)
	c.Assert(err, IsNil)
	defer ret.Close()
	ts, err := ioutil.ReadAll(ret)
	c.Assert(err, IsNil)
	str, err := readCsvLike(localCsvFile)
	c.Assert(err, IsNil)
	c.Assert(string(ts), Equals, str)

	err = s.bucket.DeleteObject(key)
	c.Assert(err, IsNil)
}

func (s *OssSelectCsvSuite) TestSelectCsvObjectIntAggregation(c *C) {
	key := "sample_data.csv"
	localCsvFile := "../sample/sample_data.csv"
	err := s.bucket.PutObjectFromFile(key, localCsvFile)
	c.Assert(err, IsNil)
	selReq := SelectRequest{}
	selReq.Expression = `select avg(cast(year as int)), max(cast(year as int)), min(cast(year as int)) from ossobject where year = 2015`
	selReq.InputSerializationSelect.CsvBodyInput.FileHeaderInfo = "Use"
	ret, err := s.bucket.SelectObject(key, selReq)
	c.Assert(err, IsNil)
	defer ret.Close()
	ts, err := ioutil.ReadAll(ret)
	c.Assert(err, IsNil)

	c.Assert(string(ts), Equals, "2015,2015,2015\n")

	err = s.bucket.DeleteObject(key)
	c.Assert(err, IsNil)
}

func (s *OssSelectCsvSuite) TestSelectCsvObjectFloatAggregation(c *C) {
	key := "sample_data.csv"
	localCsvFile := "../sample/sample_data.csv"
	err := s.bucket.PutObjectFromFile(key, localCsvFile)
	c.Assert(err, IsNil)
	selReq := SelectRequest{}
	selReq.Expression = `select avg(cast(data_value as double)), max(cast(data_value as double)), sum(cast(data_value as double)) from ossobject`
	selReq.InputSerializationSelect.CsvBodyInput.FileHeaderInfo = "Use"
	ret, err := s.bucket.SelectObject(key, selReq)
	c.Assert(err, IsNil)
	defer ret.Close()
	ts, err := ioutil.ReadAll(ret)
	strR := string(ts)
	c.Assert(err, IsNil)

	avg, max, sum, err := readCsvFloatAgg(localCsvFile)
	c.Assert(err, IsNil)

	s1 := strconv.FormatFloat(avg, 'f', 5, 32) + ","
	s1 += strconv.FormatFloat(max, 'f', 5, 32) + ","
	s1 += strconv.FormatFloat(sum, 'f', 5, 32) + ","
	retS := ""
	for _, v := range strings.Split(strR[:len(strR)-1], ",") {
		vv, err := strconv.ParseFloat(v, 64)
		c.Assert(err, IsNil)
		retS += strconv.FormatFloat(vv, 'f', 5, 32) + ","
	}
	c.Assert(s1, Equals, retS)

	err = s.bucket.DeleteObject(key)
	c.Assert(err, IsNil)
}

func (s *OssSelectCsvSuite) TestSelectCsvObjectConcat(c *C) {
	key := "sample_data.csv"
	localCsvFile := "../sample/sample_data.csv"
	err := s.bucket.PutObjectFromFile(key, localCsvFile)
	c.Assert(err, IsNil)
	selReq := SelectRequest{}
	selReq.Expression = `select Year,StateAbbr, CityName, Short_Question_Text from ossobject where (data_value || data_value_unit) = '14.8%'`
	selReq.InputSerializationSelect.CsvBodyInput.FileHeaderInfo = "Use"
	ret, err := s.bucket.SelectObject(key, selReq)
	c.Assert(err, IsNil)
	defer ret.Close()
	ts, err := ioutil.ReadAll(ret)
	c.Assert(err, IsNil)

	str, err := readCsvConcat(localCsvFile)
	c.Assert(err, IsNil)
	c.Assert(string(ts), Equals, str)

	err = s.bucket.DeleteObject(key)
	c.Assert(err, IsNil)
}

func (s *OssSelectCsvSuite) TestSelectCsvObjectComplicateConcat(c *C) {
	key := "sample_data.csv"
	localCsvFile := "../sample/sample_data.csv"
	err := s.bucket.PutObjectFromFile(key, localCsvFile)
	c.Assert(err, IsNil)
	selReq := SelectRequest{}
	selReq.Expression = `
	select 
		Year,StateAbbr, CityName, Short_Question_Text, data_value, 
		data_value_unit, category, high_confidence_limit 
	from 
		ossobject 
	where 
		data_value > 14.8 and 
		data_value_unit = '%' or 
		Measure like '%18 Years' and 
		Category = 'Unhealthy Behaviors' or 
		high_confidence_limit > 70.0 `

	selReq.InputSerializationSelect.CsvBodyInput.FileHeaderInfo = "Use"
	ret, err := s.bucket.SelectObject(key, selReq)
	c.Assert(err, IsNil)
	defer ret.Close()
	ts, err := ioutil.ReadAll(ret)
	c.Assert(err, IsNil)

	str, err := readCsvComplicateCondition(localCsvFile)
	c.Assert(err, IsNil)
	c.Assert(string(ts), Equals, str)

	err = s.bucket.DeleteObject(key)
	c.Assert(err, IsNil)
}

func (s *OssSelectCsvSuite) TestSelectCsvObjectInvalidSql(c *C) {
	key := "sample_data.csv"
	localCsvFile := "../sample/sample_data.csv"
	err := s.bucket.PutObjectFromFile(key, localCsvFile)
	c.Assert(err, IsNil)
	selReq := SelectRequest{}
	selReq.Expression = `select * from ossobject where avg(cast(year as int)) > 2016`
	selReq.InputSerializationSelect.CsvBodyInput.FileHeaderInfo = "Use"
	_, err = s.bucket.SelectObject(key, selReq)
	c.Assert(err, NotNil)

	selReq.Expression = ``
	_, err = s.bucket.SelectObject(key, selReq)
	c.Assert(err, NotNil)

	selReq.Expression = `select year || CityName from ossobject`
	_, err = s.bucket.SelectObject(key, selReq)
	c.Assert(err, NotNil)

	selReq.Expression = `select * from ossobject group by CityName`
	_, err = s.bucket.SelectObject(key, selReq)
	c.Assert(err, NotNil)

	selReq.Expression = `select * from ossobject order by _1`
	_, err = s.bucket.SelectObject(key, selReq)
	c.Assert(err, NotNil)

	selReq.Expression = `select * from ossobject oss join s3object s3 on oss.CityName = s3.CityName`
	_, err = s.bucket.SelectObject(key, selReq)
	c.Assert(err, NotNil)

	selReq.Expression = `select _1 from ossobject`
	ret, err := s.bucket.SelectObject(key, selReq)
	c.Assert(err, IsNil)
	defer ret.Close()
	_, err = ioutil.ReadAll(ret)
	c.Assert(err, IsNil)

	err = s.bucket.DeleteObject(key)
	c.Assert(err, IsNil)
}

func (s *OssSelectCsvSuite) TestSelectCsvObjectWithOutputDelimiters(c *C) {
	key := "sample_data.csv"
	content := "abc,def\n"
	err := s.bucket.PutObject(key, strings.NewReader(content))
	c.Assert(err, IsNil)
	selReq := SelectRequest{}
	selReq.Expression = `select _1, _2 from ossobject `
	selReq.OutputSerializationSelect.CsvBodyOutput.RecordDelimiter = "\r\n"
	selReq.OutputSerializationSelect.CsvBodyOutput.FieldDelimiter = "|"

	ret, err := s.bucket.SelectObject(key, selReq)
	c.Assert(err, IsNil)
	defer ret.Close()
	ts, err := ioutil.ReadAll(ret)
	c.Assert(err, IsNil)
	c.Assert(string(ts), Equals, "abc|def\r\n")

	err = s.bucket.DeleteObject(key)
	c.Assert(err, IsNil)
}

func (s *OssSelectCsvSuite) TestSelectCsvObjectWithCrc(c *C) {
	key := "sample_data.csv"
	content := "abc,def\n"
	err := s.bucket.PutObject(key, strings.NewReader(content))
	c.Assert(err, IsNil)
	selReq := SelectRequest{}
	selReq.Expression = `select * from ossobject`
	bo := true
	selReq.OutputSerializationSelect.EnablePayloadCrc = &bo

	ret, err := s.bucket.SelectObject(key, selReq)
	c.Assert(err, IsNil)
	defer ret.Close()
	ts, err := ioutil.ReadAll(ret)
	c.Assert(err, IsNil)
	c.Assert(string(ts), Equals, content)

	err = s.bucket.DeleteObject(key)
	c.Assert(err, IsNil)
}

func (s *OssSelectCsvSuite) TestSelectCsvObjectWithSkipPartialData(c *C) {
	key := "sample_data.csv"
	content := "abc,def\nefg\n"
	err := s.bucket.PutObject(key, strings.NewReader(content))
	c.Assert(err, IsNil)
	selReq := SelectRequest{}
	selReq.Expression = `select _1, _2 from ossobject`
	bo := true
	selReq.SelectOptions.SkipPartialDataRecord = &bo
	ret, err := s.bucket.SelectObject(key, selReq)
	c.Assert(err, IsNil)
	defer ret.Close()
	ts, err := ioutil.ReadAll(ret)
	c.Assert(err, IsNil)
	c.Assert(string(ts), Equals, "abc,def\n")

	err = s.bucket.DeleteObject(key)
	c.Assert(err, IsNil)
}

func (s *OssSelectCsvSuite) TestSelectCsvObjectWithOutputRaw(c *C) {
	key := "sample_data.csv"
	content := "abc,def\n"
	err := s.bucket.PutObject(key, strings.NewReader(content))
	c.Assert(err, IsNil)
	selReq := SelectRequest{}
	selReq.Expression = `select _1 from ossobject`
	bo := true
	selReq.OutputSerializationSelect.OutputRawData = &bo

	ret, err := s.bucket.SelectObject(key, selReq)
	c.Assert(err, IsNil)
	defer ret.Close()
	ts, err := ioutil.ReadAll(ret)
	c.Assert(err, IsNil)
	c.Assert(string(ts), Equals, "abc\n")

	err = s.bucket.DeleteObject(key)
	c.Assert(err, IsNil)
}

func (s *OssSelectCsvSuite) TestSelectCsvObjectWithKeepColumns(c *C) {
	key := "sample_data.csv"
	content := "abc,def\n"
	err := s.bucket.PutObject(key, strings.NewReader(content))
	c.Assert(err, IsNil)
	selReq := SelectRequest{}
	selReq.Expression = `select _1 from ossobject`
	bo := true
	selReq.OutputSerializationSelect.KeepAllColumns = &bo

	ret, err := s.bucket.SelectObject(key, selReq)
	c.Assert(err, IsNil)
	defer ret.Close()
	ts, err := ioutil.ReadAll(ret)
	c.Assert(err, IsNil)
	c.Assert(string(ts), Equals, "abc,\n")

	err = s.bucket.DeleteObject(key)
	c.Assert(err, IsNil)
}

func (s *OssSelectCsvSuite) TestSelectCsvObjectWithOutputHeader(c *C) {
	key := "sample_data.csv"
	content := "name,job\nabc,def\n"
	err := s.bucket.PutObject(key, strings.NewReader(content))
	c.Assert(err, IsNil)
	selReq := SelectRequest{}
	selReq.Expression = `select name from ossobject`
	bo := true
	selReq.OutputSerializationSelect.OutputHeader = &bo
	selReq.InputSerializationSelect.CsvBodyInput.FileHeaderInfo = "Use"

	ret, err := s.bucket.SelectObject(key, selReq)
	c.Assert(err, IsNil)
	defer ret.Close()
	ts, err := ioutil.ReadAll(ret)
	c.Assert(err, IsNil)
	c.Assert(string(ts), Equals, "name\nabc\n")

	err = s.bucket.DeleteObject(key)
	c.Assert(err, IsNil)
}

func (s *OssSelectCsvSuite) TestSelectCsvObjectRead(c *C) {
	key := "sample_data.csv"
	content := "name,job\nabc,def\n"
	err := s.bucket.PutObject(key, strings.NewReader(content))
	c.Assert(err, IsNil)
	selReq := SelectRequest{}
	selReq.Expression = `select name from ossobject`
	bo := true
	selReq.OutputSerializationSelect.OutputHeader = &bo
	selReq.InputSerializationSelect.CsvBodyInput.FileHeaderInfo = "Use"
	selReq.OutputSerializationSelect.EnablePayloadCrc = &bo

	ret, err := s.bucket.SelectObject(key, selReq)
	c.Assert(err, IsNil)
	defer ret.Close()

	// case 1: read length > data length
	p := make([]byte, 512)
	n, err := ret.Read(p[:20])
	if err != nil && err != io.EOF {
		c.Assert(err, IsNil)
	}
	c.Assert(string(p[:n]), Equals, "name\nabc\n")
	ts, err := ioutil.ReadAll(ret)
	c.Assert(err, IsNil)
	c.Assert(string(ts), Equals, "")

	// case 2: read length = data length
	ret, err = s.bucket.SelectObject(key, selReq)
	c.Assert(err, IsNil)
	defer ret.Close()
	n, err = ret.Read(p[:9])
	if err != nil && err != io.EOF {
		c.Assert(err, IsNil)
	}
	c.Assert(string(p[:n]), Equals, "name\nabc\n")
	ts, err = ioutil.ReadAll(ret)
	c.Assert(err, IsNil)
	c.Assert(string(ts), Equals, "")

	// case 3: read length > one frame length and read length < two frame, (this data = 2 * frame length)
	ret, err = s.bucket.SelectObject(key, selReq)
	c.Assert(err, IsNil)
	defer ret.Close()
	n, err = ret.Read(p[:7])
	if err != nil && err != io.EOF {
		c.Assert(err, IsNil)
	}
	c.Assert(string(p[:n]), Equals, "name\nab")
	ts, err = ioutil.ReadAll(ret)
	c.Assert(err, IsNil)
	c.Assert(string(ts), Equals, "c\n")

	// case 4: read length = a frame length (this data = 2 * frame length)
	ret, err = s.bucket.SelectObject(key, selReq)
	c.Assert(err, IsNil)
	defer ret.Close()
	n, err = ret.Read(p[:5])
	if err != nil && err != io.EOF {
		c.Assert(err, IsNil)
	}
	c.Assert(string(p[:n]), Equals, "name\n")
	ts, err = ioutil.ReadAll(ret)
	c.Assert(err, IsNil)
	c.Assert(string(ts), Equals, "abc\n")

	// case 5: read length < a frame length (this data = 2 * frame length)
	ret, err = s.bucket.SelectObject(key, selReq)
	c.Assert(err, IsNil)
	defer ret.Close()
	n, err = ret.Read(p[:3])
	if err != nil && err != io.EOF {
		c.Assert(err, IsNil)
	}
	c.Assert(string(p[:n]), Equals, "nam")
	ts, err = ioutil.ReadAll(ret)
	c.Assert(err, IsNil)
	c.Assert(string(ts), Equals, "e\nabc\n")

	err = s.bucket.DeleteObject(key)
	c.Assert(err, IsNil)
}

// OssProgressListener is the progress listener
type OssSelectProgressListener struct {
}

// ProgressChanged handles progress event
func (listener *OssSelectProgressListener) ProgressChanged(event *ProgressEvent) {
	switch event.EventType {
	case TransferStartedEvent:
		testLogger.Printf("Transfer Started.\n")
	case TransferDataEvent:
		testLogger.Printf("Transfer Data, This time consumedBytes: %d \n", event.ConsumedBytes)
	case TransferCompletedEvent:
		testLogger.Printf("Transfer Completed, This time consumedBytes: %d.\n", event.ConsumedBytes)
	case TransferFailedEvent:
		testLogger.Printf("Transfer Failed, This time consumedBytes: %d.\n", event.ConsumedBytes)
	default:
	}
}

func (s *OssSelectCsvSuite) TestSelectCsvObjectConcatProgress(c *C) {
	key := "sample_data.csv"
	localCsvFile := "../sample/sample_data.csv"
	err := s.bucket.PutObjectFromFile(key, localCsvFile)
	c.Assert(err, IsNil)
	selReq := SelectRequest{}
	selReq.Expression = `select Year,StateAbbr, CityName, Short_Question_Text from ossobject where (data_value || data_value_unit) = '14.8%'`
	selReq.InputSerializationSelect.CsvBodyInput.FileHeaderInfo = "Use"
	ret, err := s.bucket.SelectObject(key, selReq, Progress(&OssSelectProgressListener{}))
	c.Assert(err, IsNil)
	defer ret.Close()
	ts, err := ioutil.ReadAll(ret)
	c.Assert(err, IsNil)

	str, err := readCsvConcat(localCsvFile)
	c.Assert(err, IsNil)
	c.Assert(string(ts), Equals, str)

	err = s.bucket.DeleteObject(key)
	c.Assert(err, IsNil)
}
