package main

import (
	"fmt"
	"github.com/pion/webrtc/v4"
)

func HandlePeer(
	offer webrtc.SessionDescription,
	config *IceRestreamConfig,
	track *BroadcastTrack,
	candidates []webrtc.ICECandidateInit,
) (*webrtc.SessionDescription, []*webrtc.ICECandidateInit, error) {
	peer, err := CreatePeer(config)
	if err != nil {
		return nil, nil, err
	}

	rtpSender, err := peer.AddTrack(track.Inner())
	if err != nil {
		return nil, nil, err
	}

	go func() {
		rtcpBuf := make([]byte, 1500)
		for {
			if _, _, rtcpErr := rtpSender.Read(rtcpBuf); rtcpErr != nil {
				return
			}
		}
	}()

	err = peer.SetRemoteDescription(offer)
	if err != nil {
		return nil, nil, err
	}

	peer.OnConnectionStateChange(func(state webrtc.PeerConnectionState) {
		fmt.Printf("Connection state changed: %s\n", state.String())
	})

	for _, candidate := range candidates {
		_ = peer.AddICECandidate(candidate)
	}

	candichan := make(chan *webrtc.ICECandidateInit)
	peer.OnICECandidate(func(candidate *webrtc.ICECandidate) {
		if candidate != nil {
			init := candidate.ToJSON()
			candichan <- &init
		} else {
			candichan <- nil
		}
	})

	answer, err := peer.CreateAnswer(nil)
	if err != nil {
		return nil, nil, err
	}

	err = peer.SetLocalDescription(answer)
	if err != nil {
		return nil, nil, err
	}

	var answerCandidates []*webrtc.ICECandidateInit
	for {
		candidate := <-candichan
		if candidate != nil {
			answerCandidates = append(answerCandidates, candidate)
		} else {
			break
		}
	}

	return peer.LocalDescription(), answerCandidates, nil
}
