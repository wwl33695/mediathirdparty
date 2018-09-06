package core

type LiveInputLayer interface {
	Open(uri string, stream *LiveStream)
	Receiving() bool
	Retry() bool
	Opened() bool
	Close()
}

// type VodInputLayer interface {
// 	Open(filename string, stream *VodStream, session *VodSession) error
// 	Close() error
// 	Start()
// }
