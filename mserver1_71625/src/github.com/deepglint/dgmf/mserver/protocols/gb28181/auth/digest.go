package auth

import (
	"crypto/md5"
	"encoding/hex"

	"github.com/deepglint/dgmf/mserver/protocols/sip/base"
)

//
func CreateDigestAuth(username string, password string, realm string, methord base.Method, uri string, nonce string) string {
	h := md5.New()
	h.Write([]byte(username + ":" + realm + ":" + password)) //
	h1 := hex.EncodeToString(h.Sum(nil))

	h = md5.New()
	h.Write([]byte((string)(methord) + ":" + uri))
	h2 := hex.EncodeToString(h.Sum(nil))

	h = md5.New()
	h.Write([]byte(h1 + ":" + nonce + ":" + h2))
	return hex.EncodeToString(h.Sum(nil))
}
