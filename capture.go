package main

import (
	"github.com/deepch/vdk/av"
	"github.com/deepch/vdk/codec/h264parser"
	"github.com/deepch/vdk/format/rtsp"
	"log"
	"time"
)

func Capture(addr string, ch chan Frame) {
	session, err := rtsp.Dial(addr)
	if err != nil {
		panic(err)
	}
	session.RtpKeepAliveTimeout = 10 * time.Second

	codecs, err := session.Streams()
	if err != nil {
		panic(err)
	}
	for i, t := range codecs {
		log.Println("Stream", i, "is of type", t.Type().String())
	}
	if codecs[0].Type() != av.H264 {
		panic("RTSP feed must begin with a H264 codec")
	}
	if len(codecs) != 1 {
		log.Println("Ignoring all but the first stream.")
	}

	var previousTime time.Duration

	annexbNALUStartCode := func() []byte { return []byte{0x00, 0x00, 0x00, 0x01} }

	for {
		pkt, err := session.ReadPacket()
		if err != nil {
			break
		}

		if pkt.Idx != 0 {
			//audio or other stream, skip it
			continue
		}

		pkt.Data = pkt.Data[4:]

		// For every key-frame pre-pend the SPS and PPS
		if pkt.IsKeyFrame {
			pkt.Data = append(annexbNALUStartCode(), pkt.Data...)
			pkt.Data = append(codecs[0].(h264parser.CodecData).PPS(), pkt.Data...)
			pkt.Data = append(annexbNALUStartCode(), pkt.Data...)
			pkt.Data = append(codecs[0].(h264parser.CodecData).SPS(), pkt.Data...)
			pkt.Data = append(annexbNALUStartCode(), pkt.Data...)
		}

		bufferDuration := pkt.Time - previousTime
		previousTime = pkt.Time

		ch <- Frame{Data: pkt.Data, Duration: bufferDuration}
	}

	if err = session.Close(); err != nil {
		log.Println("session Close error", err)
	}
}
