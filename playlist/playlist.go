package main

import (
	"log"

	"github.com/nryberg/music-tracker/playlist/fetchplays"
)

func main() {
	log.Println("Playlist Start")
	url := "https://kool108.iheart.com/music/recently-played/"
	url = "https://957bigfm.iheart.com/music/recently-played/"
	results, err := fetchplays.ReadTracks(url)
	if err != nil {
		panic(err)
	}
	log.Println(len(results))

}
