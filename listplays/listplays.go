package main

import (
	"fmt"
	"log"

	"github.com/boltdb/bolt"
)

const boltDatabaseFileName = "../playlist/songs.bolt"

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

	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("plays"))

		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			fmt.Printf("key=%s, value=%s\n", k, v)
		}

		return nil
	})
	return nil, err
}
