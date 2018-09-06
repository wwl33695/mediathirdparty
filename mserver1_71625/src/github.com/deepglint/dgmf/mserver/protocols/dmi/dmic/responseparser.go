package dmic

import (
	"strings"
)

func ParseResponseHead(response string) (errCode string) {

	pos := strings.Index(response, " ")
	pos1 := strings.Index(response[pos+1:], " ")
	errCode = response[pos+1 : pos+pos1+1]

	return
}

func ParseResponseCSeq(response string) (cseq string) {

//	println(response)
	pos := strings.Index(response, "CSeq: ")
//	println(pos)
	pos += 6
	pos += strings.Index(response[pos:], " ")
//	println(pos)
	pos1 := strings.Index(response[pos+1:], "\r\n")

	cseq = response[pos+1 : pos+pos1+1]
//	println(cseq)

	return
}