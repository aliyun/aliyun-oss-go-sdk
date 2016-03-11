package oss

import (
	. "gopkg.in/check.v1"
)

type OssUtilsSuite struct{}

var _ = Suite(&OssUtilsSuite{})

func (s *OssUtilsSuite) TestUtilsTime(c *C) {
	c.Assert(GetNowSec() > 1448597674, Equals, true)
	c.Assert(GetNowNanoSec() > 1448597674000000000, Equals, true)
	c.Assert(len(GetNowGMT()), Equals, len("Fri, 27 Nov 2015 04:14:34 GMT"))
}

func (s *OssUtilsSuite) TestUtilsSplitFile(c *C) {
	localFile := "../sample/BingWallpaper-2015-11-07.jpg"

	// Num
	parts, err := SplitFileByPartNum(localFile, 4)
	c.Assert(err, IsNil)
	c.Assert(len(parts), Equals, 4)
	testLogger.Println("parts 4:", parts)
	for i, part := range parts {
		c.Assert(part.Number, Equals, i+1)
		c.Assert(part.Offset, Equals, int64(i*120512))
		c.Assert(part.Size, Equals, int64(120512))
	}

	parts, err = SplitFileByPartNum(localFile, 5)
	c.Assert(err, IsNil)
	c.Assert(len(parts), Equals, 5)
	testLogger.Println("parts 5:", parts)
	for i, part := range parts {
		c.Assert(part.Number, Equals, i+1)
		c.Assert(part.Offset, Equals, int64(i*96409))
	}

	_, err = SplitFileByPartNum(localFile, 10001)
	c.Assert(err, NotNil)

	_, err = SplitFileByPartNum(localFile, 0)
	c.Assert(err, NotNil)

	_, err = SplitFileByPartNum(localFile, -1)
	c.Assert(err, NotNil)

	_, err = SplitFileByPartNum("notexist", 1024)
	c.Assert(err, NotNil)

	// Size
	parts, err = SplitFileByPartSize(localFile, 120512)
	c.Assert(err, IsNil)
	c.Assert(len(parts), Equals, 4)
	testLogger.Println("parts 4:", parts)
	for i, part := range parts {
		c.Assert(part.Number, Equals, i+1)
		c.Assert(part.Offset, Equals, int64(i*120512))
		c.Assert(part.Size, Equals, int64(120512))
	}

	parts, err = SplitFileByPartSize(localFile, 96409)
	c.Assert(err, IsNil)
	c.Assert(len(parts), Equals, 6)
	testLogger.Println("parts 6:", parts)
	for i, part := range parts {
		c.Assert(part.Number, Equals, i+1)
		c.Assert(part.Offset, Equals, int64(i*96409))
	}

	_, err = SplitFileByPartSize(localFile, 0)
	c.Assert(err, NotNil)

	_, err = SplitFileByPartSize(localFile, -1)
	c.Assert(err, NotNil)

	_, err = SplitFileByPartSize(localFile, 10)
	c.Assert(err, NotNil)

	_, err = SplitFileByPartSize("noexist", 120512)
	c.Assert(err, NotNil)
}

func (s *OssUtilsSuite) TestUtilsFileExt(c *C) {
	c.Assert(TypeByExtension("test.txt"), Equals, "text/plain; charset=utf-8")
	c.Assert(TypeByExtension("test.jpg"), Equals, "image/jpeg")
	c.Assert(TypeByExtension("test.pdf"), Equals, "application/pdf")
	c.Assert(TypeByExtension("test"), Equals, "")
	c.Assert(TypeByExtension("/root/dir/test.txt"), Equals, "text/plain; charset=utf-8")
	c.Assert(TypeByExtension("root/dir/test.txt"), Equals, "text/plain; charset=utf-8")
	c.Assert(TypeByExtension("root\\dir\\test.txt"), Equals, "text/plain; charset=utf-8")
	c.Assert(TypeByExtension("D:\\work\\dir\\test.txt"), Equals, "text/plain; charset=utf-8")
}

func (s *OssUtilsSuite) TestGetPartEnd(c *C) {
	end := GetPartEnd(3, 10, 3)
	c.Assert(end, Equals, int64(5))

	end = GetPartEnd(9, 10, 3)
	c.Assert(end, Equals, int64(9))

	end = GetPartEnd(7, 10, 3)
	c.Assert(end, Equals, int64(9))
}
