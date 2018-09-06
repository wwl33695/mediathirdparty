package input

import (
	"errors"
	"net/url"
	"strings"

	"github.com/deepglint/dgmf/mserver/core"
)

func GetLiveInputCtx(uri string) (core.LiveInputLayer, error) {
	var inputCtx core.LiveInputLayer
	urlCtx, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}

	if strings.EqualFold(urlCtx.Scheme, "udp") {
		if !strings.Contains(urlCtx.Host, "127.0.0.1") && !strings.Contains(urlCtx.Host, "localhost") {
			err = errors.New(urlCtx.Host + " not support")
			return nil, err
		}
		parts := strings.Split(urlCtx.Host, ":")
		if len(parts) != 2 {
			err = errors.New("URI you provided is invalid")
			return nil, err
		}
		inputCtx = &UDPLiveInput{}
	} else if strings.EqualFold(urlCtx.Scheme, "rtsp") {
		inputCtx = &RTSPLiveInput{}
	} else if strings.EqualFold(urlCtx.Scheme, "file") {

	} else if strings.EqualFold(urlCtx.Scheme, "gb28181") {
		inputCtx = &GB28181LiveInput{}		
	} else {
		err = errors.New("LiveStream input must be udp, rtsp or file")
		return nil, err
	}

	return inputCtx, err
}
