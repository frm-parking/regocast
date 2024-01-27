package main

import (
	"github.com/pion/webrtc/v4"
)

func CreatePeer(iceConfig *IceRestreamConfig) (*webrtc.PeerConnection, error) {
	peerConnectionConfig := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{iceConfig.Stun},
			},
			{
				URLs:           []string{iceConfig.Turn},
				Username:       iceConfig.TurnUsr,
				Credential:     iceConfig.TurnPwd,
				CredentialType: webrtc.ICECredentialTypePassword,
			},
		},
	}

	peerConnection, err := webrtc.NewPeerConnection(peerConnectionConfig)
	if err != nil {
		return nil, err
	}

	return peerConnection, nil
}
