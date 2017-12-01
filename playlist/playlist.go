package main

import (
	"bufio"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/boltdb/bolt"
	"github.com/yhat/scrape"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

const boltDatabaseFileName = "songs.bolt"

// PlayedSong is a track played
type PlayedSong struct {
	scanTime    string
	station     string
	playDate    string
	playTime    string
	trackTitle  string
	trackArtist string
	contentID   string
}

func main() {

	log.Println("Playlist Starter")

	stations, err := GetStations("stations.txt")

	plays, err := ReadTracksTest()
	if err != nil {
		log.Println(err)
	}

	log.Println(len(plays))
	// if err != nil {
	// 	log.Println(err)
	// } else {
	// 	IterateStations(stations)
	// }

}

// IterateStations walks list of stations
func IterateStations(stations []string) {
	const baseURLPrefix = "https://"
	const baseURLPostfix = ".iheart.com/music/recently-played/"

	var url string

	for _, station := range stations {
		log.Println(station)
		url = baseURLPrefix + station + baseURLPostfix
		log.Println(url)
		results, err := ReadTracks(url, station)
		if err != nil {
			panic(err)
		}
		log.Println(len(results))

		PlaysToStdout(results)
		err = SaveData(results)
		if err != nil {
			log.Println(err)
		}
	}
}

// SaveData pushes the results to a bolt database
func SaveData(plays []PlayedSong) error {
	db, err := bolt.Open(boltDatabaseFileName, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("plays"))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		for _, play := range plays {
			id, _ := b.NextSequence()
			//u.ID = int(id)
			buf, err := json.Marshal(play)
			if err != nil {
				return err
			}

			// Persist bytes to users bucket.
			return b.Put(itob(id), buf)
		}
		return nil
	})
	return err
}

func itob(v uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(v))
	return b
}

// GetStations pulls the current list of stations
func GetStations(filePath string) ([]string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		log.Println(err)
	}
	defer f.Close()

	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		//log.Printf(os.Stderr, err)
		log.Println(err)
	}

	return lines, err
}

// PlaysToStdout sends them to stdout
func PlaysToStdout(plays []PlayedSong) {
	//var play PlayedSong
	for _, play := range plays {
		log.Println(play.station, play.trackTitle)
	}
}

// ReadTracksTest uses the saved copy of results for testing only
func ReadTracksTest() ([]PlayedSong, error) {
	var playResults []PlayedSong

	file, err := os.Open("sample.html")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	root, err := html.Parse(file)
	if err != nil {
		panic(err)
	}

	var scanTime string
	t := time.Now()
	scanTime = t.Format(time.RFC850)

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
		played.scanTime = scanTime
		played.playDate = playedDate
		played.station = station
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

// ReadTracks accepts the url for fetching data and goes and gets it
func ReadTracks(url string, station string) ([]PlayedSong, error) {
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

	var scanTime string
	t := time.Now()
	scanTime = t.Format(time.RFC850)

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
		played.scanTime = scanTime
		played.playDate = playedDate
		played.station = station
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
