package osscrypto

import (
	. "gopkg.in/check.v1"
)

func (s *OssCryptoBucketSuite) TestContentEncryptCipherError(c *C) {
	// crypto bucket
	masterRsaCipher, _ := CreateMasterRsa(matDesc, rsaPublicKey, rsaPrivateKey)
	contentProvider := CreateAesCtrCipher(masterRsaCipher)
	cc, err := contentProvider.ContentCipher()
	c.Assert(err, IsNil)

	var cipherData CipherData
	cipherData.RandomKeyIv(31, 15)

	_, err = cc.Clone(cipherData)
	c.Assert(err, NotNil)
}

func (s *OssCryptoBucketSuite) TestCreateCipherDataError(c *C) {
	// crypto bucket
	masterRsaCipher, _ := CreateMasterRsa(matDesc, "", "")
	contentProvider := CreateAesCtrCipher(masterRsaCipher)

	v := contentProvider.(aesCtrCipherBuilder)
	_, err := v.createCipherData()
	c.Assert(err, NotNil)
}

func (s *OssCryptoBucketSuite) TestContentCipherCDError(c *C) {
	var cd CipherData

	// crypto bucket
	masterRsaCipher, _ := CreateMasterRsa(matDesc, "", "")
	contentProvider := CreateAesCtrCipher(masterRsaCipher)

	v := contentProvider.(aesCtrCipherBuilder)
	_, err := v.contentCipherCD(cd)
	c.Assert(err, NotNil)
}
