// Package pushplays kicks the data out
package pushplays

import "log"

// PlayedSong is a track played
type PlayedSong struct {
	trackTitle  string
	trackArtist string
	playTime    string
	playDate    string
	contentID   string
}

// PlaysToStdout sends them to stdout
func PlaysToStdout([]PlayedSong) {
	log.Println("Done")
}
