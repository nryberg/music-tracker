package main

import (
	"log"
	"net/http"

	"github.com/yhat/scrape"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// PlayedSong is a track played
type PlayedSong struct {
	trackTitle  string
	trackArtist string
	playTime    string
	playDate    string
	contentID   string
}

func main() {
	log.Println("Playlist Start")
	url := "https://kool108.iheart.com/music/recently-played/"
	url = "https://957bigfm.iheart.com/music/recently-played/"
	results, err := ReadTracks(url)
	if err != nil {
		panic(err)
	}
	log.Println(len(results))

	PlaysToStdout(results)

}

// PlaysToStdout sends them to stdout
func PlaysToStdout([]PlayedSong) {
	log.Println("Done")
}

// ReadTracks accepts the url for fetching data and goes and gets it
func ReadTracks(url string) ([]PlayedSong, error) {
	var playResults []PlayedSong

	res, err := http.Get(url)
	if err != nil {
		panic(err.Error())
	}

	if err != nil {
		panic(err.Error())
	}

	root, err := html.Parse(res.Body)
	if err != nil {
		panic(err)
	}

	var playedDate string

	// Get the play date
	playDateNode, ok := scrape.Find(root, scrape.ByClass("playlist-date-header"))
	if ok {
		dateNode, ok := scrape.Find(playDateNode, scrape.ByTag(atom.Span))
		if ok {
			playedDate = scrape.Text(dateNode)
		}

	}

	var played PlayedSong
	// grab all articles and print them
	plays := scrape.FindAll(root, scrape.ByClass("playlist-track-container"))
	for _, play := range plays {
		played.playDate = playedDate
		played.contentID = scrape.Attr(play, "data-contentid")
		trackTitleNode, ok := scrape.Find(play, scrape.ByClass("track-title"))
		if ok {
			played.trackTitle = scrape.Text(trackTitleNode)
		}
		trackArtistNode, ok := scrape.Find(play, scrape.ByClass("track-artist"))
		if ok {
			played.trackArtist = scrape.Text(trackArtistNode)
		}

		playListTrackTimeNode, ok := scrape.Find(play, scrape.ByClass("playlist-track-time"))
		if ok {
			timeNode, ok := scrape.Find(playListTrackTimeNode, scrape.ByTag(atom.Span))
			if ok {
				played.playTime = scrape.Text(timeNode)
			}
		}
		playResults = append(playResults, played)
	}
	return playResults, err
}
