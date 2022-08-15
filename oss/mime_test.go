package oss

import (
	. "gopkg.in/check.v1"
)

type OssMimeSuite struct{}

var _ = Suite(&OssMimeSuite{})

func (s *OssMimeSuite) TestTypeByExtension(c *C) {
	filePath := "demo.html"
	typ := TypeByExtension(filePath)
	c.Assert(typ, Equals, "text/html")

	filePath = "demo.htm"
	typ = TypeByExtension(filePath)
	c.Assert(typ, Equals, "text/html")

	filePath = "demo.txt"
	typ = TypeByExtension(filePath)
	c.Assert(typ, Equals, "text/plain")
	/*
		filePath = "demo.group"
		typ = TypeByExtension(filePath)
		c.Assert(typ, Equals, "text/x-ms-group")

		filePath = "demo.htc"
		typ = TypeByExtension(filePath)
		c.Assert(typ, Equals, "text/x-component")

		filePath = "demo.xsit"
		typ = TypeByExtension(filePath)
		c.Assert(typ, Equals, "text/xml")

		filePath = "demo.css"
		typ = TypeByExtension(filePath)
		c.Assert(typ, Equals, "text/css")

		filePath = "demo.rtx"
		typ = TypeByExtension(filePath)
		c.Assert(typ, Equals, "text/richtext")

		filePath = "demo.tsv"
		typ = TypeByExtension(filePath)
		c.Assert(typ, Equals, "text/tab-separated-values")

		filePath = "demo.jad"
		typ = TypeByExtension(filePath)
		c.Assert(typ, Equals, "text/vnd.sun.j2me.app-descriptor")

		filePath = "demo.wml"
		typ = TypeByExtension(filePath)
		c.Assert(typ, Equals, "text/vnd.wap.wml")

		filePath = "demo.wmls"
		typ = TypeByExtension(filePath)
		c.Assert(typ, Equals, "text/vnd.wap.wmlscript")

		filePath = "demo.etx"
		typ = TypeByExtension(filePath)
		c.Assert(typ, Equals, "text/x-setext")

		filePath = "demo.ics"
		typ = TypeByExtension(filePath)
		c.Assert(typ, Equals, "text/calendar")

		filePath = "demo.ifb"
		typ = TypeByExtension(filePath)
		c.Assert(typ, Equals, "text/calendar")
	*/
}
