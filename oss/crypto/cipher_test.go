package osscrypto

import (
	"io"
	"strings"

	. "gopkg.in/check.v1"
)

func (s *OssCryptoBucketSuite) TestAesCtr(c *C) {
	var cipherData CipherData
	cipherData.RandomKeyIv(32, 16)
	cipher, _ := newAesCtr(cipherData)

	byteReader := strings.NewReader(RandLowStr(100))
	enReader := cipher.Encrypt(byteReader)
	encrypter := &CryptoEncrypter{Body: byteReader, Encrypter: enReader}
	encrypter.Close()
	buff := make([]byte, 10)
	n, err := encrypter.Read(buff)
	c.Assert(n, Equals, 0)
	c.Assert(err, Equals, io.EOF)

	deReader := cipher.Encrypt(byteReader)
	Decrypter := &CryptoDecrypter{Body: byteReader, Decrypter: deReader}
	Decrypter.Close()
	buff = make([]byte, 10)
	n, err = Decrypter.Read(buff)
	c.Assert(n, Equals, 0)
	c.Assert(err, Equals, io.EOF)

}
