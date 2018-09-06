package rtsp

import (
	// "fmt"
	"strings"
	"testing"
)

type TestBook struct {
	Title     string `rtspfield:"title:"`
	Author    string `rtspfield:"author:"`
	Price     string `rtspfield:"price:"`
	Publisher string `rtspfield:"publisher:"`
}

func TestMarshalField(test *testing.T) {
	book := TestBook{
		Title:  "How to F*CK",
		Author: "LEO",
		Price:  "$123.456",
	}

	str := marshalField(&book)
	if !strings.EqualFold(str, "title: How to F*CK\r\nauthor: LEO\r\nprice: $123.456\r\n") {
		test.Error("Marshal field error")
	}
}

func TestUnmarshalField(test *testing.T) {
	var book TestBook
	str := "title: How to F*CK\r\nauthor: LEO\r\nprice: $123.456\r\n"
	unmarshalField(&book, str)

	if !strings.EqualFold(book.Author, "LEO") || !strings.EqualFold(book.Title, "How to F*CK") ||
		!strings.EqualFold(book.Price, "$123.456") || !strings.EqualFold(book.Publisher, "") {
		test.Error("Unmarshal field error")
	}
}

func TestRTSPGeneralHeaderMarshal(test *testing.T) {
	header := RTSPGeneralHeader{
		CacheControl: "no-cache",
		Connection:   "close",
		Date:         "Fri, 02 Sep 2016 03:04:17 GMT",
	}

	str := header.Marshal()
	if !strings.EqualFold(str, "Cache-Control: no-cache\r\nConnection: close\r\nDate: Fri, 02 Sep 2016 03:04:17 GMT\r\n") {
		test.Error("RTSPGeneralHeader marshal error")
	}

}

func TestRTSPGeneralHeaderUnmarshal(test *testing.T) {
	header := RTSPGeneralHeader{}
	header.Unmarshal("Cache-Control: no-cache\r\nConnection: close\r\nDate: Fri, 02 Sep 2016 03:04:17 GMT\r\n")
	if !strings.EqualFold(header.CacheControl, "no-cache") || !strings.EqualFold(header.Connection, "close") ||
		!strings.EqualFold(header.Date, "Fri, 02 Sep 2016 03:04:17 GMT") || !strings.EqualFold(header.Via, "") {
		test.Error("RTSPGeneralHeader unmarshal error")
	}
}

func TestRTSPEntityHeaderMarshal(test *testing.T) {
	header := RTSPEntityHeader{
		ContentBase:   "rtsp://192.168.5.12/leo/test/",
		ContentType:   "application/sdp",
		ContentLength: "905",
	}

	str := header.Marshal()
	if !strings.EqualFold(str, "Content-Base: rtsp://192.168.5.12/leo/test/\r\nContent-Length: 905\r\nContent-Type: application/sdp\r\n") {
		test.Error("RTSPEntityHeader marshal error")
	}
}

func TestRTSPEntityHeaderUnmarshal(test *testing.T) {
	header := RTSPEntityHeader{}
	header.Unmarshal("Content-Base: rtsp://192.168.5.12/leo/test/\r\nContent-Length: 905\r\nContent-Type: application/sdp\r\n")
	if !strings.EqualFold(header.ContentBase, "rtsp://192.168.5.12/leo/test/") ||
		!strings.EqualFold(header.ContentType, "application/sdp") ||
		!strings.EqualFold(header.ContentLength, "905") {
		test.Error("RTSPEntityHeader unmarshal error")
	}
}
