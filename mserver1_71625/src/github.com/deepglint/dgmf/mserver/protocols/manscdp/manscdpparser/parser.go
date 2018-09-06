package manscdpparser

import (
	"bufio"
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"

	"io"
	"strings"

	"github.com/deepglint/dgmf/mserver/protocols/manscdp/manscdpbase"
	"github.com/deepglint/dgmf/mserver/protocols/sip/log"
)

// A Parser converts the raw bytes of SIP messages into base.SipMessage objects.
// It allows
type Parser interface {
	// Implements io.Writer. Queues the given bytes to be parsed.
	// If the parser has terminated due to a previous fatal error, it will return n=0 and an appropriate error.
	// Otherwise, it will return n=len(p) and err=nil.
	// Note that err=nil does not indicate that the data provided is valid - simply that the data was successfully queued for parsing.
	Write(p []byte) (n int, err error)

	Stop()
}

// Parse a SIP message by creating a parser on the fly.
// This is more costly than reusing a parser, but is necessary when we do not
// have a guarantee that all messages coming over a connection are from the
// same endpoint (e.g. UDP).
func ParseManscdp(msgData []byte) (manscdpbase.Manscdp, error) {
	output := make(chan manscdpbase.Manscdp, 0)
	errors := make(chan error, 0)
	parser := NewParser(output, errors)
	parser.Write(msgData)
	select {
	case msg := <-output:
		return msg, nil
	case err := <-errors:
		return nil, err
	}
}

// Create a new Parser.
//
// Parsed SIP messages will be sent down the 'output' chan provided.
// Any errors which force the parser to terminate will be sent down the 'errs' chan provided.
//
// If streamed=false, each Write call to the parser should contain data for one complete SIP message.

// If streamed=true, Write calls can contain a portion of a full SIP message.
// The end of one message and the start of the next may be provided in a single call to Write.
// When streamed=true, all SIP messages provided must have a Content-Length header.
// SIP messages without a Content-Length will cause the parser to permanently stop, and will result in an error on the errs chan.

// 'streamed' should be set to true whenever the caller cannot reliably identify the starts and ends of messages from the transport frames,
// e.g. when using streamed protocols such as TCP.
func NewParser(output chan<- manscdpbase.Manscdp, errs chan<- error) Parser {
	p := parser{}

	p.output = output
	p.errs = errs

	// Create a managed buffer to allow message data to be asynchronously provided to the parser, and
	// to allow the parser to block until enough data is available to parse.
	p.input = newParserBuffer()

	// Wait for input a line at a time, and produce SipMessages to send down p.output.
	go p.parse()

	return &p
}

type parser struct {
	inputdata   []byte
	input       *parserBuffer
	output      chan<- manscdpbase.Manscdp
	errs        chan<- error
	terminalErr error
	stopped     bool
}

// Consume input lines one at a time, producing base.SipMessage objects and sending them down p.output.
func (p *parser) parse() {

	var secondLine string

	// Parse the StartLine.

	startLine, err := p.input.NextLine()
	if err != nil {
		fmt.Println("Parser %p stopped", p)
		p.errs <- err
		return

	}
	if !strings.Contains(startLine, "<?xml version=\"1.0\"?>") {
		err = errors.New("manscdp格式错误")
		fmt.Println("manscdp格式错误 Parser %p stopped", p)
		p.errs <- err
		return
	}
	secondLine, err = p.input.NextLine()
	if err != nil {
		err = errors.New("manscdp格式错误")
		fmt.Println("manscdp格式错误 Parser %p stopped", p)
		return
	}

	switch {
	case strings.Contains(secondLine, "<Query>"):
		var result manscdpbase.Query
		err = xml.Unmarshal(p.inputdata, &result)
		if err != nil {
			fmt.Println("转换成 Query 出现错误 Parser %p stopped", p)
			p.errs <- err
		}
		// fmt.Println("转换成 Query 出现错误 Parser %v stopped", *result)
		p.output <- &result
	case strings.Contains(secondLine, "<Control>"):
		var result manscdpbase.Control
		err = xml.Unmarshal(p.inputdata, &result)
		if err != nil {
			fmt.Println("转换成 Control 出现错误 Parser %p stopped", string(p.inputdata))
			p.errs <- err
		}
		// fmt.Println("转换成 Control 出现错误 Parser %v stopped", *result)
		p.output <- &result

	default:
		p.errs <- errors.New("无法解析manscdp协议")

	}

	// if secondLine != "<Query>" {
	// 	fmt.Println("进入查询分支")
	// }

	// if strings.Replace(strings.Replace(secondLine, " ", "", -1), "\n", "", -1) == "<Query>" {
	// 	fmt.Println("进入查询分支")
	// } else {
	// 	fmt.Println("没有找到匹配")
	// }

	fmt.Println("解析manscdp成功")

	return
}

// Stop parser processing, and allow all resources to be garbage collected.
// The parser will not release its resources until Stop() is called,
// even if the parser object itself is garbage collected.
func (p *parser) Stop() {
	log.Debug("Stopping parser %p", p)
	p.stopped = true
	p.input.Stop()
	log.Debug("Parser %p stopped", p)
}

func (p *parser) Write(data []byte) (n int, err error) {
	if p.terminalErr != nil {
		// The parser has stopped due to a terminal error. Return it.
		log.Fine("Parser %p ignores %d new bytes due to previous terminal error: %s", p, len(data), p.terminalErr.Error())
		return 0, p.terminalErr
	} else if p.stopped {
		return 0, fmt.Errorf("Cannot write data to stopped parser %p", p)
	}
	p.inputdata = data
	p.input.Write(data)

	return len(data), nil
}

// parserBuffer is a specialized buffer for use in the parser package.
// It is written to via the non-blocking Write.
// It exposes various blocking read methods, which wait until the requested
// data is avaiable, and then return it.
type parserBuffer struct {
	io.Writer
	buffer bytes.Buffer

	// Wraps parserBuffer.pipeReader
	reader *bufio.Reader

	// Don't access this directly except when closing.
	pipeReader *io.PipeReader
}

// Create a new parserBuffer object (see struct comment for object details).
// Note that resources owned by the parserBuffer may not be able to be GCed
// until the Dispose() method is called.
func newParserBuffer() *parserBuffer {
	var pb parserBuffer
	pb.pipeReader, pb.Writer = io.Pipe()
	pb.reader = bufio.NewReader(pb.pipeReader)
	return &pb
}

// Block until the buffer contains at least one CRLF-terminated line.
// Return the line, excluding the terminal CRLF, and delete it from the buffer.
// Returns an error if the parserbuffer has been stopped.
func (pb *parserBuffer) NextLine() (response string, err error) {
	var buffer bytes.Buffer
	var data string

	data, err = pb.reader.ReadString('\n')
	buffer.WriteString(data)
	response = buffer.String()
	response = response[:len(response)-1]
	return
}

// Block until the buffer contains at least n characters.
// Return precisely those n characters, then delete them from the buffer.
func (pb *parserBuffer) NextChunk(n int) (response string, err error) {
	var data []byte = make([]byte, n)

	var read int
	for total := 0; total < n; {
		read, err = pb.reader.Read(data[total:])
		total += read
		if err != nil {
			return
		}
	}

	response = string(data)
	log.Debug("Parser buffer returns chunk '%s'", response)
	return
}

// Stop the parser buffer.
func (pb *parserBuffer) Stop() {
	pb.pipeReader.Close()
}
