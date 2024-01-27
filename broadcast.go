package main

import (
	"github.com/pion/webrtc/v4"
	"github.com/pion/webrtc/v4/pkg/media"
)

type BroadcastTrack struct {
	track *webrtc.TrackLocalStaticSample
}

func (t *BroadcastTrack) Write(frame Frame) error {
	err := t.track.WriteSample(media.Sample{
		Data:     frame.Data,
		Duration: frame.Duration,
	})

	return err
}

func (t *BroadcastTrack) Inner() *webrtc.TrackLocalStaticSample {
	return t.track
}

type TrackSet struct {
	tracks map[string]BroadcastTrack
}

func (ts *TrackSet) CreateTrack(id string) (*BroadcastTrack, error) {
	track, err := webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{MimeType: "video/h264"}, "video", "restream")
	if err != nil {
		return nil, err
	}

	broadcastTrack := BroadcastTrack{track: track}

	if ts.tracks == nil {
		ts.tracks = make(map[string]BroadcastTrack)
	}

	ts.tracks[id] = broadcastTrack

	return &broadcastTrack, nil
}

func (ts *TrackSet) Get(id string) *BroadcastTrack {
	if track, ok := ts.tracks[id]; ok {
		return &track
	} else {
		return nil
	}
}
