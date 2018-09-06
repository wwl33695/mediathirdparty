package rtsp

import (
	// "fmt"
	"strings"
	"testing"
)

func TestWWWAuthenticateMarshal(test *testing.T) {
	auth := WWWAuthenticate{
		Realm: "deepglint.com",
		Nonce: "a92ea779a695be2ba5ebe73e1f1f486e",
	}

	if !strings.EqualFold(auth.Marshal(), "realm=\"deepglint.com\", nonce=\"a92ea779a695be2ba5ebe73e1f1f486e\"") {
		test.Error("WWWAuthenticate marshal error")
	}
}

func TestWWWAuthenticateUnmarshal(test *testing.T) {
	var auth WWWAuthenticate

	auth.Unmarshal("realm=\"deepglint.com\", nonce=\"a92ea779a695be2ba5ebe73e1f1f486e\"")
	if !strings.EqualFold(auth.Realm, "deepglint.com") {
		test.Error("WWWAuthenticate unmarshal error")
	}
	if !strings.EqualFold(auth.Nonce, "a92ea779a695be2ba5ebe73e1f1f486e") {
		test.Error("WWWAuthenticate unmarshal error")
	}
}

func TestAuthorizationMarshal(test *testing.T) {
	auth := Authorization{
		Username: "admin",
		Realm:    "deepglint.com",
		Nonce:    "a92ea779a695be2ba5ebe73e1f1f486e",
		URI:      "www.deepglint.com",
		Response: "a92ea779a695be2ba5ebe73e1f1f486e",
	}

	if !strings.EqualFold(auth.Marshal(), "Digest username=\"admin\", realm=\"deepglint.com\", nonce=\"a92ea779a695be2ba5ebe73e1f1f486e\", uri=\"www.deepglint.com\", response=\"a92ea779a695be2ba5ebe73e1f1f486e\"") {
		test.Error("Authorization marshal error")
	}
}

func TestAuthorizationUnmarshal(test *testing.T) {
	var auth Authorization

	auth.Unmarshal("Digest username=\"admin\", realm=\"deepglint.com\", nonce=\"a92ea779a695be2ba5ebe73e1f1f486e\", uri=\"www.deepglint.com\", response=\"a92ea779a695be2ba5ebe73e1f1f486e\"")
	if !strings.EqualFold(auth.Username, "admin") {
		test.Error("Authorization unmarshal error")
	}
	if !strings.EqualFold(auth.Realm, "deepglint.com") {
		test.Error("Authorization unmarshal error")
	}
	if !strings.EqualFold(auth.Nonce, "a92ea779a695be2ba5ebe73e1f1f486e") {
		test.Error("Authorization unmarshal error")
	}
	if !strings.EqualFold(auth.URI, "www.deepglint.com") {
		test.Error("Authorization unmarshal error")
	}
	if !strings.EqualFold(auth.Response, "a92ea779a695be2ba5ebe73e1f1f486e") {
		test.Error("Authorization unmarshal error")
	}
}

func TestDigestAuth(test *testing.T) {
	response := WWWAuthenticate{
		Realm: "deepglint.com",
		Nonce: "a92ea779a695be2ba5ebe73e1f1f486e",
	}
	request := Authorization{}
	request.DigestAuth(response, "admin", "deepglint123", "www.deepglint.com", "OPTIONS")
	if !strings.EqualFold(request.Response, "cee9280f981620c4a1c09e758e2c747a") {
		test.Error("Digest auth error")
	}
}

func TestBasicAuth(test *testing.T) {
	auth := Authorization{}
	auth.BasicAuth("admin", "deepglint123")
	if !strings.EqualFold(auth.Token, "YWRtaW46ZGVlcGdsaW50MTIz") {
		test.Error("Basic auth error")
	}
}
