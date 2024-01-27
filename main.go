package main

import (
	"log"
)

func main() {
	config := LoadConfig()

	trackSet := TrackSet{}

	for _, stream := range config.Stream {
		track, err := trackSet.CreateTrack(stream.Id)
		if err != nil {
			panic(err)
		}

		addr := stream.Url

		go func() {
			frames := make(chan Frame)
			go Capture(addr, frames)

			for {
				frame := <-frames
				err = track.Write(frame)
				if err != nil {
					log.Printf("ERR: %s", err)
				}
			}
		}()
	}

	Serve(SignalingState{
		Config: &config,
		Tracks: &trackSet,
	})
}
