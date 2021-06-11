package sample

import (
	"fmt"
	"io/ioutil"
	
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

// SelectObjectSample shows how to get data from csv/json object by sql
func SelectObjectSample() {
	// Create a bucket
	bucket, err := GetTestBucket(bucketName)
	if err != nil {
		HandleError(err)
	}

	//
	// Create a Csv object
	//
	err = bucket.PutObjectFromFile(objectKey, localCsvFile)
	if err != nil {
		HandleError(err)
	}

	// Create Csv Meta
	csvMeta := oss.CsvMetaRequest{}
	ret, err := bucket.CreateSelectCsvObjectMeta(objectKey, csvMeta)
	if err != nil {
		HandleError(err)
	}
	fmt.Println("csv file meta:", ret)

	// case 1: Isn't NULL
	selReq := oss.SelectRequest{}
	selReq.Expression = "select Year, StateAbbr, CityName, PopulationCount from ossobject where CityName != ''"
	selReq.InputSerializationSelect.CsvBodyInput.FileHeaderInfo = "Use"

	body, err := bucket.SelectObject(objectKey, selReq)
	if err != nil {
		HandleError(err)
	}
	defer body.Close()

	databyte, err := ioutil.ReadAll(body)
	if err != nil {
		HandleError(err)
	}
	fmt.Println("some data in SelectCSVObject result:", string(databyte[:9]))

	// case 2: Like
	selReq = oss.SelectRequest{}
	selReq.Expression =  "select Year, StateAbbr, CityName, Short_Question_Text from ossobject where Measure like '%blood pressure%Years'"
	selReq.InputSerializationSelect.CsvBodyInput.FileHeaderInfo = "Use"
	body, err = bucket.SelectObject(objectKey, selReq)
	if err != nil {
		HandleError(err)
	}
	defer body.Close()

	databyte, err = ioutil.ReadAll(body)
	if err != nil {
		HandleError(err)
	}
	fmt.Println("some data in SelectCSVObject result:", string(databyte[:9]))

	// delete object
	err = bucket.DeleteObject(objectKey)
	if err != nil {
		HandleError(err)
	}

	//
	// Create a LINES json object
	//
	err = bucket.PutObjectFromFile(objectKey, localJSONLinesFile)
	if err != nil {
		HandleError(err)
	}

	// Create LINES JSON Meta
	jsonMeta := oss.JsonMetaRequest{
		InputSerialization: oss.InputSerialization {
			JSON: oss.JSON {
				JSONType:"LINES",
			},
		},
	}
	restSt, err := bucket.CreateSelectJsonObjectMeta(objectKey, jsonMeta)
	if err != nil {
		HandleError(err)
	}
	fmt.Println("csv json meta:", restSt)

	// case 1: sql where A=B
	selReq = oss.SelectRequest{}
	selReq.Expression = "select * from ossobject where party = 'Democrat'"
	selReq.OutputSerializationSelect.JsonBodyOutput.RecordDelimiter = ","
	selReq.InputSerializationSelect.JsonBodyInput.JSONType = "LINES"

	body, err = bucket.SelectObject(objectKey, selReq)
	if err != nil {
		HandleError(err)
	}
	defer body.Close()

	databyte, err = ioutil.ReadAll(body)
	if err != nil {
		HandleError(err)
	}
	fmt.Println("some data in SelectJsonObject result:", string(databyte[:9]))

	// case 2: LIKE
	selReq = oss.SelectRequest{}
	selReq.Expression = "select person.firstname, person.lastname from ossobject where person.birthday like '1959%'"
	selReq.OutputSerializationSelect.JsonBodyOutput.RecordDelimiter = ","
	selReq.InputSerializationSelect.JsonBodyInput.JSONType = "LINES"

	body, err = bucket.SelectObject(objectKey, selReq)
	if err != nil {
		HandleError(err)
	}
	defer body.Close()

	databyte, err = ioutil.ReadAll(body)
	if err != nil {
		HandleError(err)
	}
	fmt.Println("some data in SelectJsonObject result:", string(databyte[:9]))

	// delete object
	err = bucket.DeleteObject(objectKey)
	if err != nil {
		HandleError(err)
	}

	//
	// Create a Document json object
	//
	err = bucket.PutObjectFromFile(objectKey, localJSONFile)
	if err != nil {
		HandleError(err)
	}

	// case 1: int avg, max, min 
	selReq = oss.SelectRequest{}
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
	
	body, err = bucket.SelectObject(objectKey, selReq)
	if err != nil {
		HandleError(err)
	}
	defer body.Close()

	databyte, err = ioutil.ReadAll(body)
	if err != nil {
		HandleError(err)
	}
	fmt.Println("data:", string(databyte))

	// case 2: Concat
	selReq = oss.SelectRequest{}
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
	
	body, err = bucket.SelectObject(objectKey, selReq)
	if err != nil {
		HandleError(err)
	}
	defer body.Close()

	databyte, err = ioutil.ReadAll(body)
	if err != nil {
		HandleError(err)
	}
	fmt.Println("some data in SelectJsonObject result:", string(databyte[:9]))
	
	// Delete the object and bucket
	err = DeleteTestBucketAndObject(bucketName)
	if err != nil {
		HandleError(err)
	}

	fmt.Println("SelectObjectSample completed")
}
