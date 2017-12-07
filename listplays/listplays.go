package main

import (
	"encoding/binary"
	"encoding/json"
	"log"

	"github.com/boltdb/bolt"
)

const boltDatabaseFileName = "../playlist/songs.bolt"

// PlayedSong is a track played
type PlayedSong struct {
	ScanTime    string
	Station     string
	PlayDate    string
	PlayTime    string
	TrackTitle  string
	TrackArtist string
	ContentID   string
}

func main() {
	plays, err := FetchPlays()
	if err != nil {
		log.Println(err)
	}

	log.Println(len(plays))
}

// FetchPlays pulls the plays out of the boltdb
func FetchPlays() ([]PlayedSong, error) {
	db, err := bolt.Open(boltDatabaseFileName, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	var play *PlayedSong

	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("plays"))

		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			id := binary.BigEndian.Uint64(k)
			if err != nil {
				log.Println(err)
			}
			log.Println(id)

			err := json.Unmarshal(v, &play)
			if err != nil {
				log.Println(err)
			}
			log.Println(play.TrackTitle)
			//fmt.Printf("key=%s, value=%s\n", k, v)
		}

		return nil
	})
	return nil, err
}
