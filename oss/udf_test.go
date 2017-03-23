// udf test

package oss

import (
    "io"
    "io/ioutil"
    "os"
	"strings"
	"time"
    "math/rand"
    "fmt"

	. "gopkg.in/check.v1"
)

type OssUDFSuite struct {
    client  *Client
	udf     *UDF
	bucket *Bucket
}

var _ = Suite(&OssUDFSuite{})

var (
    udfNamePrefix = "go-sdk-test-udf-"  
    udfBucketName = "go-sdk-test-bucket-udf"
    udfLogPath = "go-sdk-test-udf-log" 
    letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
    imagePath = "test_resource/udf-go-pingpong.tar.gz" 
    upgradeImagePath = "test_resource/udf-go-pingpong-upgrade.tar.gz" 
    sleepSecond = time.Second
    BuildImageSleepSecond int64 = 40
    BuildAppSleepSecond int64 = 300 
    UpgradeAppSleepSecond int64 = 300 
    ResizeAppSleepSecond int64 = 300 
    ImageDeleteSleepSecond int64 = 100 
    AppDeleteSleepSecond int64 = 100 
)

func randStr(n int) string {
    b := make([]rune, n)
    r := rand.New(rand.NewSource(time.Now().UnixNano()))
    for i := range b {
        b[i] = letters[r.Intn(len(letters))]
    }
    return string(b)
}

func randLowStr(n int) string {
    return strings.ToLower(randStr(n))
}

// Run once when the suite starts running
func (s *OssUDFSuite) SetUpSuite(c *C) {
    testLogger.Println("test udf started")
    client, err := New(udfEndpoint, udfAccessID, udfAccessKey)
    c.Assert(err, IsNil)
    s.client = client

    udf, err := s.client.UDF()
    c.Assert(err, IsNil)
    s.udf = udf 
    testLogger.Println("test udf started")

    udfNameTest := udfNamePrefix + "ABC" 
    udfConfig := UDFConfiguration{UDFName:udfNameTest} 
    err = s.udf.CreateUDF(udfConfig)
    c.Assert(err, NotNil)

	s.client.CreateBucket(udfBucketName)
	time.Sleep(5 * time.Second)

    bucket, err := s.client.Bucket(udfBucketName)
    c.Assert(err, IsNil)
    s.bucket = bucket
}

// Run before each test or benchmark starts running
func (s *OssUDFSuite) TearDownSuite(c *C) {
    // Delete all test udf 
    s.DeleteAllUDFs(c)
    s.DeleteObjects(c)
}

// Run after each test or benchmark runs
func (s *OssUDFSuite) SetUpTest(c *C) {
}

// Run once after all tests or benchmarks have finished running
func (s *OssUDFSuite) TearDownTest(c *C) {
}

func (s *OssUDFSuite) DeleteAllUDFs(c *C) {
    ldr, err := s.udf.ListUDFs()
    c.Assert(err, IsNil)
    for i := 0; i < len(ldr.UDFs); i++ {
        udfName := ldr.UDFs[i].UDFName
        if !strings.HasPrefix(udfName, udfNamePrefix) {
            continue
        }
        err = s.udf.DeleteUDFApplication(udfName)  
        if err != nil {
            fmt.Println("delete app fail", udfName, err)
            guair, err := s.udf.GetUDFApplicationInfo(udfName)
            c.Assert(err, IsNil)
            fmt.Println("app status:", guair)
        } else {
            fmt.Println("delete app success", udfName)
        }
        err = s.udf.DeleteUDFImage(udfName)
        if err != nil {
            fmt.Println("delete image fail", udfName)
        } else {
            fmt.Println("delete image success", udfName)
        }
        err = s.udf.DeleteUDF(udfName)
        if err != nil {
            fmt.Println("delete udf fail", udfName)
        } else {
            fmt.Println("delete udf success", udfName)
        }
    }
}

func (s *OssUDFSuite) DeleteObjects(c *C) {
	// Delete Objects
	lor, err := s.bucket.ListObjects()
	c.Assert(err, IsNil)

	for _, object := range lor.Objects {
		err = s.bucket.DeleteObject(object.Key)
		c.Assert(err, IsNil)
	}
}

func (s *OssUDFSuite) WaitForImageStatusChange(udfName string, imageNum int, statusList []string, aheadSleepTime int64, c *C) GetUDFImageInfoResult {
    var guir GetUDFImageInfoResult
    var err error
    fmt.Println("waiting for image status change")
    if aheadSleepTime > 0 {
        time.Sleep(time.Duration(aheadSleepTime)*sleepSecond)
    }

    num := 0
    for num < 100 {
        guir, err = s.udf.GetUDFImageInfo(udfName)
        c.Assert(err, IsNil)
        c.Assert(len(guir.UDFImages), Equals, imageNum)
        i := 0
        for ; i < len(guir.UDFImages); {
            if s.stringContains(guir.UDFImages[i].Status, statusList) { 
                break
            }
            i++
        }
        if i == len(guir.UDFImages) {
            break
        }
        fmt.Println("waiting")
        time.Sleep(20*sleepSecond)
        num++
    } 
    c.Assert(num < 100, Equals, true)
    for i := 0; i < len(guir.UDFImages); i++ {
        c.Assert(guir.UDFImages[i].CanonicalRegion != "", Equals, true)
    }
    fmt.Println("wait end")
    return guir
}

func (s *OssUDFSuite) WaitForApplicationStatusChange(udfName string, statusList []string, aheadSleepTime int64, c *C) GetUDFApplicationInfoResult {
    var guair GetUDFApplicationInfoResult
    var err error
    fmt.Println("waiting for application status change")
    if aheadSleepTime > 0 {
        time.Sleep(time.Duration(aheadSleepTime)*sleepSecond)
    }

    num := 0
    for num < 100 {
        guair, err = s.udf.GetUDFApplicationInfo(udfName)
        c.Assert(err, IsNil)
        //c.Assert(guair.UDFName, Equals, udfName)
        c.Assert(guair.Region != "", Equals, true)
        if !s.stringContains(guair.Status, statusList) { 
            break
        }
        fmt.Println("waiting application")
        time.Sleep(20*sleepSecond)
        num++
    } 
    c.Assert(num < 100, Equals, true)
    fmt.Println("wait end")
    return guair
}

func (s *OssUDFSuite) stringContains(str string, sl []string) bool {
    for _, sv := range sl {
        if str == sv {
            return true
        }
    }
    return false
}

func (s *OssUDFSuite) WaitForImageDeleteEnd(udfName string, aheadSleepTime int64, c *C) {
    var guir GetUDFImageInfoResult 
    var err error
    fmt.Println("waiting for image delete")
    if aheadSleepTime > 0 {
        time.Sleep(time.Duration(aheadSleepTime)*sleepSecond)
    }

    num := 0
    for num < 100 {
        guir, err = s.udf.GetUDFImageInfo(udfName)
        c.Assert(err, IsNil)
        if len(guir.UDFImages) == 0 {
            break
        }
        fmt.Println("waiting")
        time.Sleep(20*sleepSecond)
        num++
    } 
    c.Assert(num < 100, Equals, true)
    fmt.Println("wait end")
}

func (s *OssUDFSuite) WaitForApplicationDeleteEnd(udfName string, aheadSleepTime int64, c *C) {
    var err error
    fmt.Println("waiting for application delete")
    if aheadSleepTime > 0 {
        time.Sleep(time.Duration(aheadSleepTime)*sleepSecond)
    }

    num := 0
    for num < 100 {
        _, err = s.udf.GetUDFApplicationInfo(udfName)
        if err != nil {
            if err.(ServiceError).Code == "NoSuchUdfApplication" {
                break
            }
        }
        fmt.Println("waiting application")
        time.Sleep(20*sleepSecond)
        num++
    } 
    c.Assert(num < 100, Equals, true)
    fmt.Println("wait end")
}

func (s *OssUDFSuite) getAppFromListAppsResult(udfName string, luar ListUDFApplicationsResult) (bool, GetUDFApplicationInfoResult) {
    var guar GetUDFApplicationInfoResult
    for _, uar := range luar.UDFApplications {
        if uar.UDFName == udfName {
            return true, uar
        }
    }
    return false, guar 
}

func (s *OssUDFSuite) checkAppNotInListAppsResult(udfName string, luar ListUDFApplicationsResult, c *C) {
    exist, _ := s.getAppFromListAppsResult(udfName, luar)
    c.Assert(exist, Equals, false)
}

func (s *OssUDFSuite) readBody(body io.ReadCloser) (string, error) {
    data, err := ioutil.ReadAll(body)
    body.Close()
    if err != nil {
        return "", err
    }
    return string(data), nil
}

func (s *OssUDFSuite) readFile(fileName string, c *C) string {
    f, err := ioutil.ReadFile(fileName)
    c.Assert(err, IsNil)
    return string(f)
}


func (s *OssUDFSuite) testLogNotExist(udfName string, t time.Time, c *C) {
    os.Remove(udfLogPath)
    _, err := os.Stat(udfLogPath)
    c.Assert(err, NotNil)
    c.Assert(os.IsNotExist(err), Equals, true)

    // get log
    _, err = s.udf.GetObjectApplicationLog(udfName)
    c.Assert(err, NotNil)

    _, err = s.udf.GetObjectApplicationLog(udfName, UDFSince(time.Now()))
    c.Assert(err, NotNil)

    _, err = s.udf.GetObjectApplicationLog(udfName, UDFTail(10))
    c.Assert(err, NotNil)

    _, err = s.udf.GetObjectApplicationLog(udfName, UDFSince(time.Now()), UDFTail(10))
    c.Assert(err, NotNil)

    // get log to file
    err = s.udf.GetObjectApplicationLogToFile(udfName, udfLogPath)
    c.Assert(err, NotNil)

    err = s.udf.GetObjectApplicationLogToFile(udfName, udfLogPath, UDFSince(time.Now()))
    c.Assert(err, NotNil)

    err = s.udf.GetObjectApplicationLogToFile(udfName, udfLogPath, UDFTail(10))
    c.Assert(err, NotNil)

    err = s.udf.GetObjectApplicationLogToFile(udfName, udfLogPath, UDFSince(t), UDFTail(10))
    c.Assert(err, NotNil)

    _, err = os.Stat(udfLogPath)
    c.Assert(err, NotNil)
    c.Assert(os.IsNotExist(err), Equals, true)
}

func (s *OssUDFSuite) testLog(udfName string, t time.Time, c *C) {
    os.Remove(udfLogPath)

    // get log
    body, err := s.udf.GetObjectApplicationLog(udfName)
    c.Assert(err, IsNil)
    str, err := s.readBody(body)
    c.Assert(err, IsNil)
    c.Assert(str != "", Equals, true)

    // get log to file
    err = s.udf.GetObjectApplicationLogToFile(udfName, udfLogPath)
    c.Assert(err, IsNil)
    str = s.readFile(udfLogPath, c)
    c.Assert(str != "", Equals, true)

    // with since
    body, err = s.udf.GetObjectApplicationLog(udfName, UDFSince(time.Now()))
    c.Assert(err, IsNil)
    str, err = s.readBody(body)
    c.Assert(err, IsNil)
 
    // with since
    body, err = s.udf.GetObjectApplicationLog(udfName, UDFSince(t))
    c.Assert(err, IsNil)
    str, err = s.readBody(body)
    c.Assert(err, IsNil)
    c.Assert(str != "", Equals, true)
 
    err = s.udf.GetObjectApplicationLogToFile(udfName, udfLogPath, UDFSince(t))
    c.Assert(err, IsNil)
    str = s.readFile(udfLogPath, c)
    c.Assert(str, Equals, str)
    c.Assert(str != "", Equals, true)

    body, err = s.udf.GetObjectApplicationLog(udfName, UDFTail(10))
    c.Assert(err, IsNil)
    str, err = s.readBody(body)
    c.Assert(err, IsNil)
    c.Assert(str != "", Equals, true)

    err = s.udf.GetObjectApplicationLogToFile(udfName, udfLogPath, UDFTail(10))
    c.Assert(err, IsNil)
    str = s.readFile(udfLogPath, c)
    c.Assert(str, Equals, str)
    c.Assert(str != "", Equals, true)
    
    body, err = s.udf.GetObjectApplicationLog(udfName, UDFSince(t), UDFTail(10))
    c.Assert(err, IsNil)
    str, err = s.readBody(body)
    c.Assert(err, IsNil)
    c.Assert(str != "", Equals, true)

    err = s.udf.GetObjectApplicationLogToFile(udfName, udfLogPath, UDFSince(t), UDFTail(10))
    c.Assert(err, IsNil)
    str = s.readFile(udfLogPath, c)
    c.Assert(str != "", Equals, true)
}

func (s *OssUDFSuite) TestCreateUDF(c *C) {
    udfNameTest := udfNamePrefix + "ABC" 
    udfConfig := UDFConfiguration{UDFName:udfNameTest} 
    err := s.udf.CreateUDF(udfConfig)
    c.Assert(err, NotNil)

    udfNameTest = udfNamePrefix + randLowStr(5)
    udfNameTest1 := udfNameTest + "1" 
    udfConfig = UDFConfiguration{UDFName:udfNameTest} 

    err = s.udf.DeleteUDF(udfNameTest)

    err = s.udf.DeleteUDF(udfNameTest1)

    err = s.udf.CreateUDF(udfConfig)
    c.Assert(err, IsNil)

    err = s.udf.CreateUDF(udfConfig)
    c.Assert(err, IsNil)

    udfConfig = UDFConfiguration{UDFName:udfNameTest, UDFID:"123"}

    err = s.udf.CreateUDF(udfConfig)
    c.Assert(err, NotNil)

    udfConfig = UDFConfiguration{UDFName:udfNameTest, UDFDescription:"UDF帮助"}
    err = s.udf.CreateUDF(udfConfig)
    c.Assert(err, IsNil)

    udfConfig = UDFConfiguration{UDFName:udfNameTest, UDFID:"123", UDFDescription:"UDF帮助"}
    err = s.udf.CreateUDF(udfConfig)
    c.Assert(err, NotNil)

    // get udf 
    gur, err := s.udf.GetUDF(udfNameTest)
    c.Assert(err, IsNil)
    c.Assert(gur.UDFName, Equals, udfNameTest)
    c.Assert(gur.UDFID != "", Equals, true)
    c.Assert(gur.OwnerID != "", Equals, true)
    c.Assert(gur.UDFDescription, Equals, "")
    c.Assert(gur.ACL, Equals, "private")
    c.Assert(gur.CreationDate.Unix() <= time.Now().Unix(), Equals, true)
    t1 := gur.CreationDate.Unix() 

    // delete udf 
    err = s.udf.DeleteUDF(udfNameTest)
    c.Assert(err, IsNil)

    // get udf
    gur, err = s.udf.GetUDF(udfNameTest)
    c.Assert(err, NotNil)

    // create udf with id
    // invalid udfid
    udfID := "1"
    udfConfig = UDFConfiguration{UDFName:udfNameTest, UDFID:udfID, UDFDescription:udfNameTest + "的描述"}
    err = s.udf.CreateUDF(udfConfig)
    c.Assert(err, NotNil)
    
    // not exist udfid
    udfID = "11111111111111111111111111111111" 
    udfConfig = UDFConfiguration{UDFName:udfNameTest, UDFID:udfID, UDFDescription:udfNameTest + "的描述"}
    err = s.udf.CreateUDF(udfConfig)
    c.Assert(err, NotNil)

    udfConfig = UDFConfiguration{UDFName:udfNameTest, UDFDescription:udfNameTest + "的描述"}
    err = s.udf.CreateUDF(udfConfig)
    c.Assert(err, IsNil)

    gur, err = s.udf.GetUDF(udfNameTest)
    c.Assert(err, IsNil)
    c.Assert(gur.UDFName, Equals, udfNameTest)
    c.Assert(gur.OwnerID != "", Equals, true)
    c.Assert(gur.UDFDescription, Equals, udfNameTest + "的描述")
    c.Assert(gur.ACL, Equals, "private")
    c.Assert(gur.CreationDate.Unix() >= t1, Equals, true)
    c.Assert(gur.UDFID != "", Equals, true)
    udfID = gur.UDFID
    t1 = gur.CreationDate.Unix()

    // list udf
    ldr, err := s.udf.ListUDFs(UDFID(udfID))
    c.Assert(err, IsNil)
    c.Assert(len(ldr.UDFs), Equals, 1)
    c.Assert(ldr.UDFs[0].UDFName, Equals, udfNameTest)
    c.Assert(ldr.UDFs[0].UDFID, Equals, udfID)
    c.Assert(ldr.UDFs[0].OwnerID != "", Equals, true)
    c.Assert(ldr.UDFs[0].UDFDescription, Equals, udfNameTest + "的描述")
    c.Assert(ldr.UDFs[0].ACL, Equals, "private")
    c.Assert(ldr.UDFs[0].CreationDate.Unix(), Equals, t1)

    time.Sleep(5)

    udfConfig = UDFConfiguration{UDFName:udfNameTest1, UDFID:udfID, UDFDescription:"%&#++(@jfs)"+ udfNameTest1 + "的描述"}
    err = s.udf.CreateUDF(udfConfig)
    c.Assert(err, IsNil)

    gur, err = s.udf.GetUDF(udfNameTest1)
    c.Assert(err, IsNil)
    c.Assert(gur.UDFName, Equals, udfNameTest1)
    c.Assert(gur.UDFID, Equals, udfID)
    c.Assert(gur.OwnerID != "", Equals, true)
    c.Assert(gur.UDFDescription, Equals, "%&#++(@jfs)"+ udfNameTest1 + "的描述")
    c.Assert(gur.ACL, Equals, "private")
    c.Assert(gur.CreationDate.Unix() >= t1, Equals, true)

    // list udf
    ldr, err = s.udf.ListUDFs(UDFID(udfID))
    c.Assert(err, IsNil)
    c.Assert(len(ldr.UDFs), Equals, 2)
    c.Assert(ldr.UDFs[0].UDFName, Equals, udfNameTest)
    c.Assert(ldr.UDFs[1].UDFName, Equals, udfNameTest1)
    c.Assert(ldr.UDFs[0].UDFDescription, Equals, udfNameTest + "的描述")
    c.Assert(ldr.UDFs[1].UDFDescription, Equals, "%&#++(@jfs)"+ udfNameTest1 + "的描述")
    for i := 0; i < 2; i++ {
        c.Assert(ldr.UDFs[0].UDFID, Equals, udfID)
        c.Assert(ldr.UDFs[0].OwnerID != "", Equals, true)
        c.Assert(ldr.UDFs[0].ACL, Equals, "private")
        c.Assert(ldr.UDFs[0].CreationDate.Unix(), Equals, t1)
    }

    // delete udf
    err = s.udf.DeleteUDF(udfNameTest)
    c.Assert(err, IsNil)

    err = s.udf.DeleteUDF(udfNameTest1)
    c.Assert(err, IsNil)

    // list udf
    ldr, err = s.udf.ListUDFs(UDFID(udfID))
    c.Assert(err, IsNil)
    c.Assert(len(ldr.UDFs), Equals, 0)
}

func (s *OssUDFSuite) TestGetUDF(c *C) {
    // get not exist udf
    udfNameTest := udfNamePrefix + randLowStr(5)

    gur, err := s.udf.GetUDF(udfNameTest)
    c.Assert(err, NotNil)

    gur, err = s.udf.GetUDF("")
    c.Assert(err, NotNil)

    gur, err = s.udf.GetUDF("1")
    c.Assert(err, NotNil)

    udfConfig := UDFConfiguration{UDFName:udfNameTest} 
    err = s.udf.CreateUDF(udfConfig)
    c.Assert(err, IsNil)

    gur, err = s.udf.GetUDF(udfNameTest)
    c.Assert(err, IsNil)
    c.Assert(gur.UDFName, Equals, udfNameTest)
    c.Assert(gur.UDFID != "", Equals, true)
    c.Assert(gur.OwnerID != "", Equals, true)
    c.Assert(gur.UDFDescription, Equals, "")
    c.Assert(gur.ACL, Equals, "private")
    c.Assert(gur.CreationDate.Unix() <= time.Now().Unix(), Equals, true)

    err = s.udf.DeleteUDF(udfNameTest)
    c.Assert(err, IsNil)
}

func (s *OssUDFSuite) TestListUDFs(c *C) {
    s.DeleteAllUDFs(c)

    ldr, err := s.udf.ListUDFs()
    c.Assert(err, IsNil)
    num := len(ldr.UDFs)

    udfNameTest := udfNamePrefix + randLowStr(5)
    udfConfig := UDFConfiguration{UDFName:udfNameTest}
    err = s.udf.CreateUDF(udfConfig)
    c.Assert(err, IsNil)

    ldr, err = s.udf.ListUDFs()
    c.Assert(err, IsNil)
    c.Assert(len(ldr.UDFs), Equals, num + 1)
    udfID := ""
    for i := 0; i < len(ldr.UDFs); i++ {
        if ldr.UDFs[i].UDFName == udfNameTest {
            udfID = ldr.UDFs[i].UDFID
            break
        }
    }
    c.Assert(udfID != "", Equals, true)

    ldr, err = s.udf.ListUDFs(UDFID(udfID))
    c.Assert(err, IsNil)
    c.Assert(len(ldr.UDFs), Equals, 1)
    c.Assert(ldr.UDFs[0].UDFName, Equals, udfNameTest)
    c.Assert(ldr.UDFs[0].UDFID != "", Equals, true)

    udfNameTest1 := udfNamePrefix + randLowStr(5)
    udfConfig = UDFConfiguration{UDFName:udfNameTest1, UDFID:udfID}
    err = s.udf.CreateUDF(udfConfig)
    c.Assert(err, IsNil)

    ldr, err = s.udf.ListUDFs()
    c.Assert(err, IsNil)
    c.Assert(len(ldr.UDFs), Equals, num + 2)

    ldr, err = s.udf.ListUDFs(UDFID(udfID))
    c.Assert(err, IsNil)
    c.Assert(len(ldr.UDFs), Equals, 2)
    c.Assert(ldr.UDFs[0].UDFName == udfNameTest || ldr.UDFs[0].UDFName == udfNameTest1, Equals, true)
    c.Assert(ldr.UDFs[0].UDFID, Equals, udfID)
    c.Assert(ldr.UDFs[1].UDFName == udfNameTest || ldr.UDFs[1].UDFName == udfNameTest1, Equals, true)
    c.Assert(ldr.UDFs[1].UDFID, Equals, udfID)

    // create udf
    udfNameTest2 := udfNamePrefix + randLowStr(5)
    udfConfig = UDFConfiguration{UDFName:udfNameTest2}
    err = s.udf.CreateUDF(udfConfig)
    c.Assert(err, IsNil)

    ldr, err = s.udf.ListUDFs()
    c.Assert(err, IsNil)
    c.Assert(len(ldr.UDFs), Equals, num + 3)

    ldr, err = s.udf.ListUDFs(UDFID(udfID))
    c.Assert(err, IsNil)
    c.Assert(len(ldr.UDFs), Equals, 2)
    c.Assert(ldr.UDFs[0].UDFName == udfNameTest || ldr.UDFs[0].UDFName == udfNameTest1, Equals, true)
    c.Assert(ldr.UDFs[0].UDFID, Equals, udfID)
    c.Assert(ldr.UDFs[1].UDFName == udfNameTest || ldr.UDFs[1].UDFName == udfNameTest1, Equals, true)
    c.Assert(ldr.UDFs[1].UDFID, Equals, udfID)

    // list err udfid
    ldr, err = s.udf.ListUDFs(UDFID("1"))
    c.Assert(err, NotNil)

    ldr, err = s.udf.ListUDFs(UDFID(""))
    c.Assert(err, NotNil)

    // not exist udfid
    ldr, err = s.udf.ListUDFs(UDFID("11111111111111111111111111111111"))
    c.Assert(err, IsNil)
    c.Assert(len(ldr.UDFs), Equals, 0)

    s.DeleteAllUDFs(c)

    ldr, err = s.udf.ListUDFs()
    c.Assert(err, IsNil)
    c.Assert(len(ldr.UDFs), Equals, num)
}

func (s *OssUDFSuite) TestDeleteUDF(c *C) {
    // delete not exist udf
    err := s.udf.DeleteUDF("")
    c.Assert(err, NotNil)

    err = s.udf.DeleteUDF("测试")
    c.Assert(err, NotNil)

    err = s.udf.DeleteUDF("1")
    c.Assert(err, NotNil)

    err = s.udf.DeleteUDF("ABCD")
    c.Assert(err, NotNil)

    udfNameTest := udfNamePrefix + randLowStr(5)
    udfConfig := UDFConfiguration{UDFName:udfNameTest} 
    err = s.udf.CreateUDF(udfConfig)
    c.Assert(err, IsNil)

    gur, err := s.udf.GetUDF(udfNameTest)
    c.Assert(err, IsNil)
    c.Assert(gur.UDFName, Equals, udfNameTest)
    udfID := gur.UDFID

    // create another udf name
    udfNameTest1 := udfNamePrefix + randLowStr(5)
    udfConfig = UDFConfiguration{UDFName:udfNameTest1, UDFID:udfID} 
    err = s.udf.CreateUDF(udfConfig)
    c.Assert(err, IsNil)

    // delete udf 
    err = s.udf.DeleteUDF(udfNameTest)
    c.Assert(err, IsNil)

    // get udfid
    gur, err = s.udf.GetUDF(udfNameTest1)
    c.Assert(err, IsNil)
    c.Assert(gur.UDFName, Equals, udfNameTest1)
    c.Assert(gur.UDFID, Equals, udfID)

    // list udf
    ldr, err := s.udf.ListUDFs(UDFID(udfID))
    c.Assert(err, IsNil)
    c.Assert(len(ldr.UDFs), Equals, 1)
    c.Assert(ldr.UDFs[0].UDFName, Equals, udfNameTest1)
    c.Assert(ldr.UDFs[0].UDFID != "", Equals, true)

    // delete udf
    err = s.udf.DeleteUDF(udfNameTest1)
    c.Assert(err, IsNil)
}

func (s *OssUDFSuite) TestUDFImage(c *C) {
    udfNameTest := udfNamePrefix + randLowStr(6) 
    fmt.Println(udfNameTest)

    // delete udf
    err := s.udf.DeleteUDF(udfNameTest)

    // get not exist udf image
    _, err = s.udf.GetUDFImageInfo(udfNameTest)
    c.Assert(err, NotNil)

    // delete not exist udf image
    err = s.udf.DeleteUDFImage(udfNameTest)    
    c.Assert(err, NotNil)

    // create udf
    udfConfig := UDFConfiguration{UDFName:udfNameTest} 
    err = s.udf.CreateUDF(udfConfig)
    c.Assert(err, IsNil)
    
    // get udf
    guir, err := s.udf.GetUDFImageInfo(udfNameTest)
    c.Assert(err, IsNil)
    c.Assert(len(guir.UDFImages), Equals, 0)

    // upload invalid image
    err = s.udf.UploadUDFImageFromFile(udfNameTest, "udf_test.go", UDFImageDesc("错误的镜像"))
    c.Assert(err, IsNil)

    err = s.udf.UploadUDFImage(udfNameTest, strings.NewReader("abc")) 
    c.Assert(err, IsNil)

    // upload correct image
    err = s.udf.UploadUDFImageFromFile(udfNameTest, imagePath) 
    c.Assert(err, IsNil)

    fd, err := os.Open(imagePath)
    c.Assert(err, IsNil)
    defer fd.Close()

    err = s.udf.UploadUDFImage(udfNameTest, fd, UDFImageDesc("正确的镜像")) 
    c.Assert(err, IsNil)

    // get udf
    gur, err := s.udf.GetUDF(udfNameTest)
    c.Assert(err, IsNil)
    c.Assert(gur.UDFName, Equals, udfNameTest)
    udfID := gur.UDFID

    // list udfid
    ldr, err := s.udf.ListUDFs(UDFID(udfID))
    c.Assert(err, IsNil)
    c.Assert(len(ldr.UDFs), Equals, 1)
    c.Assert(ldr.UDFs[0].UDFName, Equals, udfNameTest)

    // check build status
    guir, err = s.udf.GetUDFImageInfo(udfNameTest)
    c.Assert(err, IsNil)
    c.Assert(len(guir.UDFImages), Equals, 4)
    c.Assert(guir.UDFImages[0].Version, Equals, int64(1))
    c.Assert(guir.UDFImages[0].Description, Equals, "错误的镜像")
    c.Assert(guir.UDFImages[1].Version, Equals, int64(2))
    c.Assert(guir.UDFImages[1].Description, Equals, "")
    c.Assert(guir.UDFImages[2].Version, Equals, int64(3))
    c.Assert(guir.UDFImages[2].Description, Equals, "")
    c.Assert(guir.UDFImages[3].Version, Equals, int64(4))
    c.Assert(guir.UDFImages[3].Description, Equals, "正确的镜像")
    for i := 0; i < len(guir.UDFImages); i++ {
        c.Assert(guir.UDFImages[i].Status, Equals, "building")
        c.Assert(guir.UDFImages[i].CreationDate.Unix() <= time.Now().Unix(), Equals, true)
    }
    
    // delete building image
    err = s.udf.DeleteUDFImage(udfNameTest)
    c.Assert(err, NotNil)

    // delete udf
    err = s.udf.DeleteUDF(udfNameTest)
    c.Assert(err, NotNil)

    // create alias udf
    udfNameTest1 := udfNameTest + "1"
    udfConfig = UDFConfiguration{UDFName:udfNameTest1, UDFID:udfID, UDFDescription:"别名"}
    err = s.udf.CreateUDF(udfConfig)
    c.Assert(err, IsNil)

    // list udfid
    ldr, err = s.udf.ListUDFs(UDFID(udfID))
    c.Assert(err, IsNil)
    c.Assert(len(ldr.UDFs), Equals, 2)
    c.Assert(ldr.UDFs[0].UDFName, Equals, udfNameTest)
    c.Assert(ldr.UDFs[1].UDFName, Equals, udfNameTest1)

    // create application error
    udfAppConfig := UDFAppConfiguration{ImageVersion:1, InstanceNum:1, Flavor:UDFAppFlavor{InstanceType:"ecs.n1.small"}}
    err = s.udf.CreateUDFApplication(udfNameTest, udfAppConfig)
    c.Assert(err, NotNil)

    udfAppConfig = UDFAppConfiguration{ImageVersion:1, InstanceNum:3, Flavor:UDFAppFlavor{InstanceType:"ecs.n1.small"}}
    err = s.udf.CreateUDFApplication(udfNameTest, udfAppConfig)
    c.Assert(err, NotNil)

    // get application, application not exist
    _, err = s.udf.GetUDFApplicationInfo(udfNameTest)
    c.Assert(err, NotNil)

    // delete alias udf 
    err = s.udf.DeleteUDF(udfNameTest1)
    c.Assert(err, IsNil)

    // wait for image build finish
    guir = s.WaitForImageStatusChange(udfNameTest, 4, []string{"building"}, BuildImageSleepSecond, c)
    c.Assert(guir.UDFImages[0].Status, Equals, "build_failed")
    c.Assert(guir.UDFImages[1].Status, Equals, "build_failed")
    c.Assert(guir.UDFImages[2].Status, Equals, "build_success")
    c.Assert(guir.UDFImages[3].Status, Equals, "build_success")

    // delete udf
    err = s.udf.DeleteUDF(udfNameTest)
    c.Assert(err, NotNil)

    // create application to build_failed one
    udfAppConfig = UDFAppConfiguration{ImageVersion:1, InstanceNum:1, Flavor:UDFAppFlavor{InstanceType:"ecs.n1.small"}}
    err = s.udf.CreateUDFApplication(udfNameTest, udfAppConfig)
    c.Assert(err, NotNil)

    // delete image
    err = s.udf.DeleteUDFImage(udfNameTest)    
    c.Assert(err, IsNil)

    // get image
    guir, err = s.udf.GetUDFImageInfo(udfNameTest)
    c.Assert(err, IsNil)
    c.Assert(len(guir.UDFImages), Equals, 4)
    c.Assert(guir.UDFImages[0].Version, Equals, int64(1))
    c.Assert(guir.UDFImages[0].Description, Equals, "错误的镜像")
    c.Assert(guir.UDFImages[1].Version, Equals, int64(2))
    c.Assert(guir.UDFImages[1].Description, Equals, "")
    c.Assert(guir.UDFImages[2].Version, Equals, int64(3))
    c.Assert(guir.UDFImages[2].Description, Equals, "")
    c.Assert(guir.UDFImages[3].Version, Equals, int64(4))
    c.Assert(guir.UDFImages[3].Description, Equals, "正确的镜像")
    for i := 0; i < len(guir.UDFImages); i++ {
        c.Assert(guir.UDFImages[0].Status, Equals, "deleting")
        c.Assert(guir.UDFImages[i].CanonicalRegion != "", Equals, true)
        c.Assert(guir.UDFImages[i].CreationDate.Unix() <= time.Now().Unix(), Equals, true)
    }

    // create alias udf
    udfNameTest1 = udfNameTest + "1"
    udfConfig = UDFConfiguration{UDFName:udfNameTest1, UDFID:udfID, UDFDescription:"别名"}
    err = s.udf.CreateUDF(udfConfig)
    c.Assert(err, IsNil)

    // list udfid
    ldr, err = s.udf.ListUDFs(UDFID(udfID))
    c.Assert(err, IsNil)
    c.Assert(len(ldr.UDFs), Equals, 2)
    c.Assert(ldr.UDFs[0].UDFName, Equals, udfNameTest)
    c.Assert(ldr.UDFs[1].UDFName, Equals, udfNameTest1)

    // delete image
    err = s.udf.DeleteUDFImage(udfNameTest)    
    c.Assert(err, NotNil)

    // delete alias udf 
    err = s.udf.DeleteUDF(udfNameTest1)
    c.Assert(err, IsNil)

    // delete udf when image in deleting
    err = s.udf.DeleteUDF(udfNameTest)
    c.Assert(err, NotNil)

    // create image again
    err = s.udf.UploadUDFImageFromFile(udfNameTest, imagePath) 
    c.Assert(err, NotNil)

    // delete udf
    err = s.udf.DeleteUDF(udfNameTest)
    c.Assert(err, NotNil)

    s.WaitForImageDeleteEnd(udfNameTest, ImageDeleteSleepSecond, c)

    // get image
    guir, err = s.udf.GetUDFImageInfo(udfNameTest)
    c.Assert(err, IsNil)
    c.Assert(len(guir.UDFImages), Equals, 0)

    // delete image again after image delete end
    err = s.udf.DeleteUDFImage(udfNameTest)
    c.Assert(err, IsNil)

    // delete udf
    err = s.udf.DeleteUDF(udfNameTest)
    c.Assert(err, IsNil)
}

func (s *OssUDFSuite) TestUDFApplication(c *C) {
    // Simultaneous start two udf for different test, reduce test time
    str := randLowStr(5)
    udfNameTestForUpgrade := udfNamePrefix + str + "upgrade"
    udfNameTestForResize := udfNamePrefix + str + "resize"

    tlog := time.Now()

    // 1.UDF Not Exist
    // get application
    _, err := s.udf.GetUDFApplicationInfo(udfNameTestForUpgrade)
    c.Assert(err, NotNil)

    // list application
    luar, err := s.udf.ListUDFApplications()
    c.Assert(err, IsNil)
    s.checkAppNotInListAppsResult(udfNameTestForUpgrade, luar, c)
    s.checkAppNotInListAppsResult(udfNameTestForResize, luar, c)

    // delete application
    err = s.udf.DeleteUDFApplication(udfNameTestForUpgrade)
    c.Assert(err, NotNil)

    // upgrade application  
    err = s.udf.UpgradeUDFApplication(udfNameTestForUpgrade, int64(2))
    c.Assert(err, NotNil)

    // create application to not exist udf
    udfAppConfig := UDFAppConfiguration{ImageVersion:1, InstanceNum:1, Flavor:UDFAppFlavor{InstanceType:"ecs.n1.small"}}
    err = s.udf.CreateUDFApplication(udfNameTestForUpgrade, udfAppConfig)
    c.Assert(err, NotNil)

    // resize application
    err = s.udf.ResizeUDFApplication(udfNameTestForResize, int64(2))
    c.Assert(err, NotNil)

    // test log
    s.testLogNotExist(udfNameTestForUpgrade, tlog, c)

    // 2.UDF Exist, Image Not Exist
    udfConfig := UDFConfiguration{UDFName:udfNameTestForUpgrade} 
    err = s.udf.CreateUDF(udfConfig)
    c.Assert(err, IsNil)

    // get udf 
    gur, err := s.udf.GetUDF(udfNameTestForUpgrade)
    c.Assert(err, IsNil)
    c.Assert(gur.UDFName, Equals, udfNameTestForUpgrade)
    c.Assert(gur.UDFID != "", Equals, true)
    udfID := gur.UDFID
 
    udfConfig = UDFConfiguration{UDFName:udfNameTestForResize} 
    err = s.udf.CreateUDF(udfConfig)
    c.Assert(err, IsNil)
 
    // create application
    udfAppConfig = UDFAppConfiguration{ImageVersion:1, InstanceNum:1, Flavor:UDFAppFlavor{InstanceType:"ecs.n1.small"}}
    err = s.udf.CreateUDFApplication(udfNameTestForUpgrade, udfAppConfig)
    c.Assert(err, NotNil)

    // get application
    _, err = s.udf.GetUDFApplicationInfo(udfNameTestForUpgrade)
    c.Assert(err, NotNil)

    // list application
    luar, err = s.udf.ListUDFApplications()
    c.Assert(err, IsNil)
    s.checkAppNotInListAppsResult(udfNameTestForUpgrade, luar, c)
    s.checkAppNotInListAppsResult(udfNameTestForResize, luar, c)

    // delete not exist application
    err = s.udf.DeleteUDFApplication(udfNameTestForUpgrade)
    c.Assert(err, IsNil)

    // update application
    err = s.udf.UpgradeUDFApplication(udfNameTestForUpgrade, int64(1))
    c.Assert(err, NotNil)

    // resize application
    err = s.udf.ResizeUDFApplication(udfNameTestForResize, int64(2))
    c.Assert(err, NotNil)

    // 3.Create Image, Image Status Building 
    // create image
    err = s.udf.UploadUDFImageFromFile(udfNameTestForUpgrade, imagePath, UDFImageDesc("镜像")) 
    c.Assert(err, IsNil)

    // check build status
    guir, err := s.udf.GetUDFImageInfo(udfNameTestForUpgrade)
    c.Assert(err, IsNil)
    c.Assert(len(guir.UDFImages), Equals, 1)
    c.Assert(guir.UDFImages[0].Version, Equals, int64(1))
    c.Assert(guir.UDFImages[0].Status, Equals, "building")
    c.Assert(guir.UDFImages[0].Description, Equals, "镜像")

    // create image
    err = s.udf.UploadUDFImageFromFile(udfNameTestForResize, imagePath, UDFImageDesc("resize镜像")) 
    c.Assert(err, IsNil)

    // check build status
    guir, err = s.udf.GetUDFImageInfo(udfNameTestForResize)
    c.Assert(err, IsNil)
    c.Assert(len(guir.UDFImages), Equals, 1)
    c.Assert(guir.UDFImages[0].Version, Equals, int64(1))
    c.Assert(guir.UDFImages[0].Status, Equals, "building")
    c.Assert(guir.UDFImages[0].Description, Equals, "resize镜像")

    // resize application
    err = s.udf.ResizeUDFApplication(udfNameTestForResize, int64(2))
    c.Assert(err, NotNil)

    // get application
    _, err = s.udf.GetUDFApplicationInfo(udfNameTestForUpgrade)
    c.Assert(err, NotNil)

    // list application
    luar, err = s.udf.ListUDFApplications()
    c.Assert(err, IsNil)
    s.checkAppNotInListAppsResult(udfNameTestForUpgrade, luar, c)
    s.checkAppNotInListAppsResult(udfNameTestForResize, luar, c)

    // update application
    err = s.udf.UpgradeUDFApplication(udfNameTestForUpgrade, int64(1))
    c.Assert(err, NotNil)

    // delete not exist application
    err = s.udf.DeleteUDFApplication(udfNameTestForUpgrade)
    c.Assert(err, IsNil)

    // create application with building status
    udfAppConfig = UDFAppConfiguration{ImageVersion:1, InstanceNum:1, Flavor:UDFAppFlavor{InstanceType:"ecs.n1.small"}}
    err = s.udf.CreateUDFApplication(udfNameTestForUpgrade, udfAppConfig)
    c.Assert(err, NotNil)

    // get application
    _, err = s.udf.GetUDFApplicationInfo(udfNameTestForUpgrade)
    c.Assert(err, NotNil)

    // list application
    luar, err = s.udf.ListUDFApplications()
    c.Assert(err, IsNil)
    s.checkAppNotInListAppsResult(udfNameTestForUpgrade, luar, c)
    s.checkAppNotInListAppsResult(udfNameTestForResize, luar, c)

    // delete not exist application
    err = s.udf.DeleteUDFApplication(udfNameTestForUpgrade)
    c.Assert(err, IsNil)

    // Wait For Image Success 
    guir = s.WaitForImageStatusChange(udfNameTestForUpgrade, 1, []string{"building"}, BuildImageSleepSecond, c)
    c.Assert(guir.UDFImages[0].Status, Equals, "build_success")

    guir = s.WaitForImageStatusChange(udfNameTestForResize, 1, []string{"building"}, 0, c)
    c.Assert(guir.UDFImages[0].Status, Equals, "build_success")

    // 4.Image Status Build Success, Application Status Starting

    // upgrade application, application not exist 
    err = s.udf.UpgradeUDFApplication(udfNameTestForUpgrade, int64(1))
    c.Assert(err, NotNil)

    // resize application, application not exist
    err = s.udf.ResizeUDFApplication(udfNameTestForResize, int64(2))
    c.Assert(err, NotNil)

    // test log
    s.testLogNotExist(udfNameTestForUpgrade, tlog, c)

    // Create Application
    // create application with error config 
    udfAppConfig = UDFAppConfiguration{ImageVersion:3, InstanceNum:1, Flavor:UDFAppFlavor{InstanceType:"ecs.n1.small"}}
    err = s.udf.CreateUDFApplication(udfNameTestForUpgrade, udfAppConfig)
    c.Assert(err, NotNil)

    udfAppConfig = UDFAppConfiguration{ImageVersion:0, InstanceNum:1, Flavor:UDFAppFlavor{InstanceType:"ecs.n1.small"}}
    err = s.udf.CreateUDFApplication(udfNameTestForUpgrade, udfAppConfig)
    c.Assert(err, NotNil)

    udfAppConfig = UDFAppConfiguration{ImageVersion:1, InstanceNum:11, Flavor:UDFAppFlavor{InstanceType:"ecs.n1.small"}}
    err = s.udf.CreateUDFApplication(udfNameTestForUpgrade, udfAppConfig)
    c.Assert(err, NotNil)

    // create correct application 
    udfAppConfig = UDFAppConfiguration{ImageVersion:1, InstanceNum:1, Flavor:UDFAppFlavor{InstanceType:"ecs.n1.small"}}
    err = s.udf.CreateUDFApplication(udfNameTestForUpgrade, udfAppConfig)
    c.Assert(err, IsNil)

    // get application
    guair, err := s.udf.GetUDFApplicationInfo(udfNameTestForUpgrade)
    c.Assert(err, IsNil)
    c.Assert(guair.UDFID, Equals, udfID)
    //c.Assert(guair.UDFName, Equals, udfNameTestForUpgrade)
    c.Assert(guair.Region != "", Equals, true)
    c.Assert(guair.ImageVersion, Equals, int64(1))
    c.Assert(guair.InstanceNum, Equals, int64(1))
    c.Assert(guair.Status == "creating" || guair.Status == "starting", Equals, true)
    c.Assert(guair.CreationDate.Unix() <= time.Now().Unix(), Equals, true)
    c.Assert(guair.Flavor.InstanceType, Equals, "ecs.n1.small")

    // create application for resize
    udfAppConfig = UDFAppConfiguration{ImageVersion:1, InstanceNum:1, Flavor:UDFAppFlavor{InstanceType:"ecs.n1.small"}}
    err = s.udf.CreateUDFApplication(udfNameTestForResize, udfAppConfig)
    c.Assert(err, IsNil)

    guair, err = s.udf.GetUDFApplicationInfo(udfNameTestForResize)
    c.Assert(err, IsNil)
    c.Assert(guair.UDFID != udfID, Equals, true)
    //c.Assert(guair.UDFName, Equals, udfNameTestForResize)
    c.Assert(guair.Region != "", Equals, true)
    c.Assert(guair.ImageVersion, Equals, int64(1))
    c.Assert(guair.InstanceNum, Equals, int64(1))
    c.Assert(guair.Status == "creating" || guair.Status == "starting", Equals, true)
    c.Assert(guair.CreationDate.Unix() <= time.Now().Unix(), Equals, true)
    c.Assert(guair.Flavor.InstanceType, Equals, "ecs.n1.small")

    // list application
    luar, err = s.udf.ListUDFApplications()
    c.Assert(err, IsNil)
    exist, guair1 := s.getAppFromListAppsResult(udfNameTestForUpgrade, luar)
    c.Assert(exist, Equals, true)
    c.Assert(guair1.UDFName, Equals, udfNameTestForUpgrade)
    c.Assert(guair1.Status == "creating" || guair1.Status == "starting", Equals, true)
    c.Assert(guair1.ImageVersion, Equals, int64(1))
    c.Assert(guair1.InstanceNum, Equals, int64(1))
    c.Assert(guair1.Region != "", Equals, true)
    c.Assert(guair1.CreationDate.Unix() <= time.Now().Unix(), Equals, true)
    c.Assert(guair1.Flavor.InstanceType, Equals, "ecs.n1.small")
    exist, guair2 := s.getAppFromListAppsResult(udfNameTestForResize, luar)
    c.Assert(exist, Equals, true)
    c.Assert(guair2.UDFName, Equals, udfNameTestForResize)
    c.Assert(guair2.Status == "creating" || guair1.Status == "starting", Equals, true)
    c.Assert(guair2.ImageVersion, Equals, int64(1))
    c.Assert(guair2.InstanceNum, Equals, int64(1))
    c.Assert(guair2.Region != "", Equals, true)
    c.Assert(guair2.CreationDate.Unix() <= time.Now().Unix(), Equals, true)
    c.Assert(guair2.Flavor.InstanceType, Equals, "ecs.n1.small")

    // delete application of starting
    err = s.udf.DeleteUDFApplication(udfNameTestForUpgrade)
    c.Assert(err, NotNil)

    // upgrade application to not exist image 
    err = s.udf.UpgradeUDFApplication(udfNameTestForUpgrade, int64(2))
    c.Assert(err, NotNil)

    // upgrade application to itself(image status build success) 
    err = s.udf.UpgradeUDFApplication(udfNameTestForUpgrade, int64(1))
    c.Assert(err, NotNil)

    // resize application of starting
    err = s.udf.ResizeUDFApplication(udfNameTestForResize, int64(2))
    c.Assert(err, NotNil)

    // Wait Application to running status
    guair = s.WaitForApplicationStatusChange(udfNameTestForUpgrade, []string{"starting", "creating"}, BuildAppSleepSecond, c)
    c.Assert(guair.Status, Equals, "running")
    c.Assert(guair.UDFID, Equals, udfID)
    c.Assert(guair.ImageVersion, Equals, int64(1))
    c.Assert(guair.InstanceNum, Equals, int64(1))

    guair = s.WaitForApplicationStatusChange(udfNameTestForResize, []string{"starting", "creating"}, 0, c)
    c.Assert(guair.Status, Equals, "running")
    c.Assert(guair.ImageVersion, Equals, int64(1))
    c.Assert(guair.InstanceNum, Equals, int64(1))

    // 5.Application Status Running

    // list application
    luar, err = s.udf.ListUDFApplications()
    c.Assert(err, IsNil)
    exist, guair1 = s.getAppFromListAppsResult(udfNameTestForUpgrade, luar)
    c.Assert(exist, Equals, true)
    c.Assert(guair1.UDFName, Equals, udfNameTestForUpgrade)
    c.Assert(guair1.Status, Equals, "running")
    c.Assert(guair1.ImageVersion, Equals, int64(1))
    c.Assert(guair1.InstanceNum, Equals, int64(1))
    c.Assert(guair1.Region != "", Equals, true)
    c.Assert(guair1.CreationDate.Unix() <= time.Now().Unix(), Equals, true)
    c.Assert(guair1.Flavor.InstanceType, Equals, "ecs.n1.small")
    exist, guair2 = s.getAppFromListAppsResult(udfNameTestForResize, luar)
    c.Assert(exist, Equals, true)
    c.Assert(guair2.UDFName, Equals, udfNameTestForResize)
    c.Assert(guair2.Status, Equals, "running")
    c.Assert(guair2.ImageVersion, Equals, int64(1))
    c.Assert(guair2.InstanceNum, Equals, int64(1))
    c.Assert(guair2.Region != "", Equals, true)
    c.Assert(guair2.CreationDate.Unix() <= time.Now().Unix(), Equals, true)
    c.Assert(guair2.Flavor.InstanceType, Equals, "ecs.n1.small")

    // test log
    // run application, put object
    objectName := objectNamePrefix + "udf-log"
    objectValue := "abc"
    err = s.bucket.PutObject(objectName, strings.NewReader(objectValue))
    c.Assert(err, IsNil)
    s.testLog(udfNameTestForUpgrade, tlog, c)

    // Test Upgrade
    // Create Another Image For Upgrade 
    err = s.udf.UploadUDFImageFromFile(udfNameTestForUpgrade, upgradeImagePath, UDFImageDesc("更新的镜像")) 
    c.Assert(err, IsNil)

    // get image info
    guir, err = s.udf.GetUDFImageInfo(udfNameTestForUpgrade)
    c.Assert(err, IsNil)
    c.Assert(len(guir.UDFImages), Equals, 2)
    c.Assert(guir.UDFImages[0].Version, Equals, int64(1))
    c.Assert(guir.UDFImages[0].Status, Equals, "build_success")
    c.Assert(guir.UDFImages[0].Description, Equals, "镜像")
    c.Assert(guir.UDFImages[1].Version, Equals, int64(2))
    c.Assert(guir.UDFImages[1].Status, Equals, "building")
    c.Assert(guir.UDFImages[1].Description, Equals, "更新的镜像")

    // upgrade application to image of building status
    err = s.udf.UpgradeUDFApplication(udfNameTestForUpgrade, int64(2))
    c.Assert(err, NotNil)

    // Test Resize
    // resize application of error instance num
    err = s.udf.ResizeUDFApplication(udfNameTestForResize, int64(11))
    c.Assert(err, NotNil)

    // normal resize
    err = s.udf.ResizeUDFApplication(udfNameTestForResize, int64(2))
    c.Assert(err, IsNil)
    tresize := time.Now().Unix() 

    // get application
    guair, err = s.udf.GetUDFApplicationInfo(udfNameTestForResize)
    c.Assert(err, IsNil)
    //c.Assert(guair.UDFName, Equals, udfNameTestForUpgrade)
    c.Assert(guair.ImageVersion, Equals, int64(1))
    c.Assert(guair.Status, Equals, "resizing")

    // delete application of resizing 
    err = s.udf.DeleteUDFApplication(udfNameTestForResize)
    c.Assert(err, NotNil)

    // list application
    luar, err = s.udf.ListUDFApplications()
    c.Assert(err, IsNil)
    exist, guair1 = s.getAppFromListAppsResult(udfNameTestForUpgrade, luar)
    c.Assert(exist, Equals, true)
    c.Assert(guair1.UDFName, Equals, udfNameTestForUpgrade)
    c.Assert(guair1.Status, Equals, "running")
    c.Assert(guair1.ImageVersion, Equals, int64(1))
    c.Assert(guair1.InstanceNum, Equals, int64(1))
    exist, guair2 = s.getAppFromListAppsResult(udfNameTestForResize, luar)
    c.Assert(exist, Equals, true)
    c.Assert(guair2.UDFName, Equals, udfNameTestForResize)
    c.Assert(guair2.Status, Equals, "resizing")
    c.Assert(guair2.ImageVersion, Equals, int64(1))

    // Wait until udfNameTestForUpgrade image version 2 to build_success
    guir = s.WaitForImageStatusChange(udfNameTestForUpgrade, 2, []string{"building"}, BuildImageSleepSecond, c)
    c.Assert(guir.UDFImages[0].Version, Equals, int64(1))
    c.Assert(guir.UDFImages[0].Status, Equals, "build_success")
    c.Assert(guir.UDFImages[1].Version, Equals, int64(2))
    c.Assert(guir.UDFImages[1].Status, Equals, "build_success")

    // upgrade application success 
    err = s.udf.UpgradeUDFApplication(udfNameTestForUpgrade, int64(2))
    c.Assert(err, IsNil)
    tupgrade := time.Now().Unix() 

    // get application
    guair, err = s.udf.GetUDFApplicationInfo(udfNameTestForUpgrade)
    c.Assert(err, IsNil)
    c.Assert(guair.UDFID, Equals, udfID)
    //c.Assert(guair.UDFName, Equals, udfNameTestForUpgrade)
    c.Assert(guair.Region != "", Equals, true)
    c.Assert(guair.InstanceNum, Equals, int64(1))
    c.Assert(guair.Status, Equals, "upgrading")
    c.Assert(guair.CreationDate.Unix() <= time.Now().Unix(), Equals, true)
    c.Assert(guair.Flavor.InstanceType, Equals, "ecs.n1.small")

    // list application
    luar, err = s.udf.ListUDFApplications()
    c.Assert(err, IsNil)
    exist, guair1 = s.getAppFromListAppsResult(udfNameTestForUpgrade, luar)
    c.Assert(exist, Equals, true)
    c.Assert(guair1.UDFName, Equals, udfNameTestForUpgrade)
    c.Assert(guair1.Status, Equals, "upgrading")
    c.Assert(guair1.InstanceNum, Equals, int64(1))
    exist, guair2 = s.getAppFromListAppsResult(udfNameTestForResize, luar)
    c.Assert(exist, Equals, true)
    c.Assert(guair2.UDFName, Equals, udfNameTestForResize)
    c.Assert(guair2.Status == "resizing" || guair2.Status == "running", Equals, true)
    c.Assert(guair2.ImageVersion, Equals, int64(1))

    // delete application of upgrading
    err = s.udf.DeleteUDFApplication(udfNameTestForUpgrade)
    c.Assert(err, NotNil)

    // Wait For Resize OK
    if luar.UDFApplications[0].Status == "resizing" {
        t := time.Now().Unix()
        guair = s.WaitForApplicationStatusChange(udfNameTestForResize, []string{"resizing"}, t - tresize, c)
        c.Assert(guair.Status, Equals, "running")
        //c.Assert(guair.UDFName, Equals, udfNameTestForResize)
        c.Assert(guair.ImageVersion, Equals, int64(1))
        c.Assert(guair.InstanceNum, Equals, int64(2))
    }

    // Wait For Upgrade OK
    t := time.Now().Unix()
    guair = s.WaitForApplicationStatusChange(udfNameTestForUpgrade, []string{"upgrading"}, t - tupgrade, c) 
    c.Assert(guair.Status, Equals, "running")
    //c.Assert(guair.UDFName, Equals, udfNameTestForUpgrade)
    c.Assert(guair.ImageVersion, Equals, int64(2))
    c.Assert(guair.InstanceNum, Equals, int64(1))

    // 6.Delete UDF 

    // delete image before delete app
    err = s.udf.DeleteUDFImage(udfNameTestForUpgrade) 
    c.Assert(err, NotNil)

    // Delete Application, Application in Deleting, forbid upload/delete/upgrade/resize App, forbid delete image, allow upload image  

    // delete application 
    err = s.udf.DeleteUDFApplication(udfNameTestForUpgrade)
    c.Assert(err, IsNil)

    // get application 
    guair, err = s.udf.GetUDFApplicationInfo(udfNameTestForUpgrade)
    c.Assert(err, IsNil)
    c.Assert(guair.UDFID, Equals, udfID)
    //c.Assert(guair.UDFName, Equals, udfNameTestForUpgrade)
    c.Assert(guair.ImageVersion, Equals, int64(2))
    c.Assert(guair.InstanceNum, Equals, int64(1))
    c.Assert(guair.Status, Equals, "deleting")

    // list application
    luar, err = s.udf.ListUDFApplications()
    c.Assert(err, IsNil)
    exist, guair1 = s.getAppFromListAppsResult(udfNameTestForUpgrade, luar)
    c.Assert(exist, Equals, true)
    c.Assert(guair1.UDFName, Equals, udfNameTestForUpgrade)
    c.Assert(guair1.Status, Equals, "deleting")
    c.Assert(guair1.ImageVersion, Equals, int64(2))
    c.Assert(guair1.InstanceNum, Equals, int64(1))
    exist, guair2 = s.getAppFromListAppsResult(udfNameTestForResize, luar)
    c.Assert(exist, Equals, true)
    c.Assert(guair2.UDFName, Equals, udfNameTestForResize)
    c.Assert(guair2.Status, Equals, "running")
    c.Assert(guair2.ImageVersion, Equals, int64(1))
    c.Assert(guair2.InstanceNum, Equals, int64(2))

    // create application in app deleting 
    udfAppConfig = UDFAppConfiguration{ImageVersion:1, InstanceNum:1, Flavor:UDFAppFlavor{InstanceType:"ecs.n1.small"}}
    err = s.udf.CreateUDFApplication(udfNameTestForUpgrade, udfAppConfig)
    c.Assert(err, NotNil)

    // delete application in app deleting 
    err = s.udf.DeleteUDFApplication(udfNameTestForUpgrade)
    c.Assert(err, NotNil)

    // upgrade application in app deleting
    err = s.udf.UpgradeUDFApplication(udfNameTestForUpgrade, int64(1))
    c.Assert(err, NotNil)
    err = s.udf.UpgradeUDFApplication(udfNameTestForUpgrade, int64(2))
    c.Assert(err, NotNil)

    // resize application in app deleting
    err = s.udf.ResizeUDFApplication(udfNameTestForUpgrade, int64(2))
    c.Assert(err, NotNil)
    err = s.udf.ResizeUDFApplication(udfNameTestForUpgrade, int64(3))
    c.Assert(err, NotNil)

    // delete image in app deleting
    err = s.udf.DeleteUDFImage(udfNameTestForUpgrade)
    c.Assert(err, NotNil)

    // delete udf fail
    err = s.udf.DeleteUDF(udfNameTestForUpgrade)
    c.Assert(err, NotNil)

    // list udf
    ldr, err := s.udf.ListUDFs(UDFID(udfID))
    c.Assert(err, IsNil)
    c.Assert(len(ldr.UDFs), Equals, 1)
    c.Assert(ldr.UDFs[0].UDFName, Equals, udfNameTestForUpgrade)
    c.Assert(ldr.UDFs[0].UDFID, Equals, udfID)

    // delete image resize
    err = s.udf.DeleteUDFImage(udfNameTestForResize)
    c.Assert(err, NotNil)

    // delete application resize
    err = s.udf.DeleteUDFApplication(udfNameTestForResize)
    c.Assert(err, IsNil)

    // delete image
    err = s.udf.DeleteUDFImage(udfNameTestForResize)
    c.Assert(err, NotNil)

    // delete udf fail
    err = s.udf.DeleteUDF(udfNameTestForResize)
    c.Assert(err, NotNil)

    gur, err = s.udf.GetUDF(udfNameTestForResize)
    c.Assert(err, IsNil)

    ts := time.Now().Unix()

    // 7.Wait Until Delete Application End 
    s.WaitForApplicationDeleteEnd(udfNameTestForUpgrade, AppDeleteSleepSecond - (time.Now().Unix() - ts), c)
    s.WaitForApplicationDeleteEnd(udfNameTestForResize, AppDeleteSleepSecond - (time.Now().Unix() - ts), c)

    // 8.Delete Image
    // delete image upgrade 
    err = s.udf.DeleteUDFImage(udfNameTestForUpgrade)
    c.Assert(err, IsNil)

    // delete image resize 
    err = s.udf.DeleteUDFImage(udfNameTestForResize)
    c.Assert(err, IsNil)

    // get image
    guir, err = s.udf.GetUDFImageInfo(udfNameTestForUpgrade)
    c.Assert(err, IsNil)
    c.Assert(len(guir.UDFImages), Equals, 1)
    c.Assert(guir.UDFImages[0].Status, Equals, "deleting")

    // resize application in app deleting
    err = s.udf.ResizeUDFApplication(udfNameTestForUpgrade, int64(0))
    c.Assert(err, NotNil)

    // 8.Wait Until Delete Image End
    s.WaitForImageDeleteEnd(udfNameTestForUpgrade, ImageDeleteSleepSecond - (time.Now().Unix() - ts), c)
    s.WaitForImageDeleteEnd(udfNameTestForResize, ImageDeleteSleepSecond - (time.Now().Unix() - ts), c)

    s.testLogNotExist(udfNameTestForUpgrade, tlog, c)

    // get image
    guir, err = s.udf.GetUDFImageInfo(udfNameTestForUpgrade)
    c.Assert(err, IsNil)
    c.Assert(len(guir.UDFImages), Equals, 0)

    // delete udf
    err = s.udf.DeleteUDF(udfNameTestForUpgrade)
    c.Assert(err, IsNil)

    // delete udf
    err = s.udf.DeleteUDF(udfNameTestForResize)
    c.Assert(err, IsNil)

    // get udf
    gur, err = s.udf.GetUDF(udfNameTestForUpgrade)
    c.Assert(err, NotNil)

    // list udf
    ldr, err = s.udf.ListUDFs(UDFID(udfID))
    c.Assert(err, IsNil)
    c.Assert(len(ldr.UDFs), Equals, 0)
}
