package rtsp

import (
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

/*
3.2.1 The WWW-Authenticate Response Header

   If a server receives a request for an access-protected object, and an
   acceptable Authorization header is not sent, the server responds with
   a "401 Unauthorized" status code, and a WWW-Authenticate header as
   per the framework defined above, which for the digest scheme is
   utilized as follows:

      challenge        =  "Digest" digest-challenge

      digest-challenge  = 1#( realm | [ domain ] | nonce |
                          [ opaque ] |[ stale ] | [ algorithm ] |
                          [ qop-options ] | [auth-param] )


      domain            = "domain" "=" <"> URI ( 1*SP URI ) <">
      URI               = absoluteURI | abs_path
      nonce             = "nonce" "=" nonce-value
      nonce-value       = quoted-string
      opaque            = "opaque" "=" quoted-string
      stale             = "stale" "=" ( "true" | "false" )
      algorithm         = "algorithm" "=" ( "MD5" | "MD5-sess" |
                           token )
      qop-options       = "qop" "=" <"> 1#qop-value <">
      qop-value         = "auth" | "auth-int" | token

   ref: https://tools.ietf.org/html/rfc2617#section-3.2.1
*/
type WWWAuthenticate struct {
	Realm     string `authfield:"realm"`
	Nonce     string `authfield:"nonce"`
	Opaque    string `authfield:"opaque"`
	Stale     string `authfield:"stale"`
	Algorithm string `authfield:"algorithm"`
	Qop       string `authfield:"qop"`
	Token     string `authfield:"token"`
}

func (this *WWWAuthenticate) Marshal() string {
	fields := reflect.ValueOf(this).Elem()
	str := ""
	for i := 0; i < fields.NumField(); i++ {
		fieldName := fields.Type().Field(i).Tag.Get("authfield")
		fieldValue := fields.Field(i).String()
		if len(fieldValue) != 0 {
			str += fmt.Sprintf("%s=\"%s\"", fieldName, fieldValue) + ", "
		}
	}
	return strings.TrimSpace(str[:len(str)-2])
}

func (this *WWWAuthenticate) Unmarshal(str string) {
	fields := reflect.ValueOf(this).Elem()
	for i := 0; i < fields.NumField(); i++ {
		fieldName := fields.Type().Field(i).Tag.Get("authfield")
		fieldValue := strings.Replace(strings.Replace(regexp.MustCompile(fieldName+"=\"[\\w|. ]*\"").FindString(str), fieldName+"=\"", "", -1), "\"", "", -1)
		fields.Field(i).SetString(fieldValue)
	}
}

/*
3.2.2 The Authorization Request Header

   The client is expected to retry the request, passing an Authorization
   header line, which is defined according to the framework above,
   utilized as follows.

       credentials      = "Digest" digest-response
       digest-response  = 1#( username | realm | nonce | digest-uri
                       | response | [ algorithm ] | [cnonce] |
                       [opaque] | [message-qop] |
                           [nonce-count]  | [auth-param] )

       username         = "username" "=" username-value
       username-value   = quoted-string
       digest-uri       = "uri" "=" digest-uri-value
       digest-uri-value = request-uri   ; As specified by HTTP/1.1
       message-qop      = "qop" "=" qop-value
       cnonce           = "cnonce" "=" cnonce-value
       cnonce-value     = nonce-value
       nonce-count      = "nc" "=" nc-value
       nc-value         = 8LHEX
       response         = "response" "=" request-digest
       request-digest = <"> 32LHEX <">
       LHEX             =  "0" | "1" | "2" | "3" |
                           "4" | "5" | "6" | "7" |
                           "8" | "9" | "a" | "b" |
                           "c" | "d" | "e" | "f"
    ref: https://tools.ietf.org/html/rfc2617#section-3.2.2
*/
type Authorization struct {
	Username   string `authfield:"username"`
	Realm      string `authfield:"realm"`
	Nonce      string `authfield:"nonce"`
	URI        string `authfield:"uri"`
	Response   string `authfield:"response"`
	Algorithm  string `authfield:"algorithm"`
	Cnonce     string `authfield:"cnonce"`
	Opaque     string `authfield:"opaque"`
	Qop        string `authfield:"qop"`
	NonceCount string `authfield:"nc"`
	Token      string `authfield:"token"`
}

func (this *Authorization) Marshal() string {
	fields := reflect.ValueOf(this).Elem()
	str := ""
	if len(this.Nonce) != 0 {
		str += "Digest "
		for i := 0; i < fields.NumField(); i++ {
			fieldName := fields.Type().Field(i).Tag.Get("authfield")
			fieldValue := fields.Field(i).String()
			if len(fieldValue) != 0 {
				str += fmt.Sprintf("%s=\"%s\", ", fieldName, fieldValue)
			}
		}
		return str[0 : len(str)-2]
	} else {
		str += "Basic " + this.Token
		return str
	}
}

func (this *Authorization) Unmarshal(str string) {
	fields := reflect.ValueOf(this).Elem()
	for i := 0; i < fields.NumField(); i++ {
		fieldName := fields.Type().Field(i).Tag.Get("authfield")
		fieldValue := strings.Replace(strings.Replace(regexp.MustCompile(fieldName+"=\"[\\w|.]*\"").FindString(str), fieldName+"=\"", "", -1), "\"", "", -1)
		fields.Field(i).SetString(fieldValue)
	}
}

/*
3.2.2.1 Request-Digest

   If the "qop" value is "auth" or "auth-int":

      request-digest  = <"> < KD ( H(A1),     unq(nonce-value)
                                          ":" nc-value
                                          ":" unq(cnonce-value)
                                          ":" unq(qop-value)
                                          ":" H(A2)
                                  ) <">

   If the "qop" directive is not present (this construction is for
   compatibility with RFC 2069):

      request-digest  =
                 <"> < KD ( H(A1), unq(nonce-value) ":" H(A2) ) >
   <">

   See below for the definitions for A1 and A2.

   ref: https://tools.ietf.org/html/rfc2617#section-3.2.2.1

3.2.2.2 A1

   If the "algorithm" directive's value is "MD5" or is unspecified, then
   A1 is:

      A1       = unq(username-value) ":" unq(realm-value) ":" passwd

   where

      passwd   = < user's password >

   If the "algorithm" directive's value is "MD5-sess", then A1 is
   calculated only once - on the first request by the client following
   receipt of a WWW-Authenticate challenge from the server.  It uses the
   server nonce from that challenge, and the first client nonce value to
   construct A1 as follows:

      A1       = H( unq(username-value) ":" unq(realm-value)
                     ":" passwd )
                     ":" unq(nonce-value) ":" unq(cnonce-value)

   This creates a 'session key' for the authentication of subsequent
   requests and responses which is different for each "authentication
   session", thus limiting the amount of material hashed with any one
   key.  (Note: see further discussion of the authentication session in

   section 3.3.) Because the server need only use the hash of the user
   credentials in order to create the A1 value, this construction could
   be used in conjunction with a third party authentication service so
   that the web server would not need the actual password value.  The
   specification of such a protocol is beyond the scope of this
   specification.

   ref: https://tools.ietf.org/html/rfc2617#section-3.2.2.2

3.2.2.3 A2

   If the "qop" directive's value is "auth" or is unspecified, then A2
   is:

      A2       = Method ":" digest-uri-value

   If the "qop" value is "auth-int", then A2 is:

      A2       = Method ":" digest-uri-value ":" H(entity-body)

   ref: https://tools.ietf.org/html/rfc2617#section-3.2.2.3
*/
func (this *Authorization) DigestAuth(authResponse WWWAuthenticate, username string, password string, uri string, method string) {
	this.Realm = authResponse.Realm
	this.Qop = authResponse.Qop
	this.Algorithm = authResponse.Algorithm
	this.Nonce = authResponse.Nonce
	this.URI = uri
	this.Username = username

	//Calculate H(A1)
	var A1 string
	A1 = this.Username + ":" + this.Realm + ":" + password
	A1 = fmt.Sprintf("%x", md5.Sum([]byte(A1)))
	if strings.EqualFold(strings.ToLower(this.Algorithm), "md5-sess") {
		this.Cnonce = "0a4f113b"
		this.NonceCount = "00000001"
		A1 += ":" + this.Nonce + ":" + this.Cnonce
		A1 = fmt.Sprintf("%x", md5.Sum([]byte(A1)))
	}

	//Calculate H(A2)
	var A2 string
	A2 = method + ":" + this.URI
	if strings.EqualFold(strings.ToLower(this.Qop), "auth-int") {
		//H(entity-body)
	}
	A2 = fmt.Sprintf("%x", md5.Sum([]byte(A2)))

	//Calculate response
	var response string
	response = A1 + ":" + this.Nonce + ":"
	if strings.EqualFold(strings.ToLower(this.Qop), "auth") || strings.EqualFold(strings.ToLower(this.Qop), "auth-int") {
		response += this.NonceCount + ":" + this.Cnonce + ":" + this.Qop + ":"
	}
	response += A2
	response = fmt.Sprintf("%x", md5.Sum([]byte(response)))
	this.Response = response
}

/*
2 Basic Authentication Scheme

   The "basic" authentication scheme is based on the model that the
   client must authenticate itself with a user-ID and a password for
   each realm.  The realm value should be considered an opaque string
   which can only be compared for equality with other realms on that
   server. The server will service the request only if it can validate
   the user-ID and password for the protection space of the Request-URI.
   There are no optional authentication parameters.

   For Basic, the framework above is utilized as follows:

      challenge   = "Basic" realm
      credentials = "Basic" basic-credentials

   Upon receipt of an unauthorized request for a URI within the
   protection space, the origin server MAY respond with a challenge like
   the following:

      WWW-Authenticate: Basic realm="WallyWorld"

   where "WallyWorld" is the string assigned by the server to identify
   the protection space of the Request-URI. A proxy may respond with the
   same challenge using the Proxy-Authenticate header field.

   To receive authorization, the client sends the userid and password,
   separated by a single colon (":") character, within a base64 [7]
   encoded string in the credentials.

      basic-credentials = base64-user-pass
      base64-user-pass  = <base64 [4] encoding of user-pass,
                       except not limited to 76 char/line>
      user-pass   = userid ":" password
      userid      = *<TEXT excluding ":">
      password    = *TEXT

   Userids might be case sensitive.

   If the user agent wishes to send the userid "Aladdin" and password
   "open sesame", it would use the following header field:

      Authorization: Basic QWxhZGRpbjpvcGVuIHNlc2FtZQ==

   A client SHOULD assume that all paths at or deeper than the depth of
   the last symbolic element in the path field of the Request-URI also
   are within the protection space specified by the Basic realm value of
   the current challenge. A client MAY preemptively send the
   corresponding Authorization header with requests for resources in
   that space without receipt of another challenge from the server.
   Similarly, when a client sends a request to a proxy, it may reuse a
   userid and password in the Proxy-Authorization header field without
   receiving another challenge from the proxy server. See section 4 for
   security considerations associated with Basic authentication.

   ref: https://tools.ietf.org/html/rfc2617#section-2
*/
func (this *Authorization) BasicAuth(username string, password string) {
	this.Token = base64.StdEncoding.EncodeToString([]byte(username + ":" + password))
}
