package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"

	"github.com/lib/pq"

	"github.com/boltdb/bolt"
)

const (
	dbUser     = "psql_writer"
	dbPassword = "uoumgsC4xViNG7"
	dbName     = "music"
	dbHost     = "box"
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

	log.Println("Clearing load table")
	rows, err := ClearLoadTable()
	if err != nil {
		log.Println(err)
	} else {
		log.Println("Rows affected: ", rows)
	}

	log.Println("Fetching songs")
	plays, err := FetchPlays()
	if err != nil {
		log.Println(err)
	} else {
		log.Println("Found Plays:", len(plays))
	}

	log.Println("Pushing data to db")
	rows, err = PushPlaystoDb(plays)
	if err != nil {
		log.Println(err)
	} else {
		log.Println("Pushed rows:", rows)
	}

	log.Println("Updating fact tables")
	err = UpdateData()
	if err != nil {
		log.Println(err)
	}

	log.Println("Done")

}

// ClearLoadTable deletes all of the load table records
func ClearLoadTable() (int64, error) {
	dbinfo := fmt.Sprintf("user=%s password=%s dbname=%s host=%s sslmode=disable",
		dbUser, dbPassword, dbName, dbHost)
	db, err := sql.Open("postgres", dbinfo)
	if err != nil {
		log.Println(err)
	}
	defer db.Close()

	var rows int64

	sqlStatement := `DELETE FROM load ;`
	effect, err := db.Exec(sqlStatement)
	if err != nil {
		log.Println(err)
	}
	rows, err = effect.RowsAffected()

	return rows, err
}

// PushPlaystoDb sends the plays to the load table
func PushPlaystoDb(plays []PlayedSong) (int64, error) {
	dbinfo := fmt.Sprintf("user=%s password=%s dbname=%s host=%s sslmode=disable",
		dbUser, dbPassword, dbName, dbHost)
	db, err := sql.Open("postgres", dbinfo)
	if err != nil {
		log.Println(err)
	}
	defer db.Close()

	txn, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}

	var counter int64

	stmt, err := txn.Prepare(pq.CopyIn("load", "scantime", "station", "playdate", "playtime", "tracktitle", "trackartist", "contentid"))
	if err != nil {
		log.Fatal(err)
	}

	//err = db.QueryRow("INSERT INTO load(ScanTime) VALUES($1) returning ID;", "Scanned Time").Scan(&lastInsertID)

	// if err != nil {
	// 	log.Println(err)
	// }

	var play PlayedSong

	for _, play = range plays {

		_, err = stmt.Exec(play.ScanTime, play.Station, play.PlayDate, play.PlayTime, play.TrackTitle, play.TrackArtist, play.ContentID)
		if err != nil {
			log.Fatal(err)
		} else {
			counter++
		}

	}
	_, err = stmt.Exec()
	if err != nil {
		log.Fatal(err)
	}

	err = stmt.Close()
	if err != nil {
		log.Fatal(err)
	}

	err = txn.Commit()
	if err != nil {
		log.Fatal(err)
	}

	return counter, err
}

// FetchPlays pulls the plays out of the boltdb
func FetchPlays() ([]PlayedSong, error) {
	db, err := bolt.Open(boltDatabaseFileName, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	var play *PlayedSong
	var plays []PlayedSong

	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("plays"))

		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			//id := binary.BigEndian.Uint64(k)
			if err != nil {
				log.Println(err)
			}

			err := json.Unmarshal(v, &play)
			if err != nil {
				log.Println(err)
			}

			plays = append(plays, *play)
			//log.Println(id, play.PlayDate, play.PlayTime, play.TrackTitle)
			//fmt.Printf("key=%s, value=%s\n", k, v)
		}

		return nil
	})
	return plays, err
}

// UpdateData goes through the load table and updates the fact tables
func UpdateData() error {
	dbinfo := fmt.Sprintf("user=%s password=%s dbname=%s host=%s sslmode=disable",
		dbUser, dbPassword, dbName, dbHost)
	db, err := sql.Open("postgres", dbinfo)
	if err != nil {
		log.Println(err)
	}
	defer db.Close()

	var effect sql.Result
	// Update the artists

	sqlStatement := `INSERT INTO artist (name)
	SELECT trackartist
		FROM public.load as ld
		LEFT JOIN public.artist as art
			on ld.trackartist = art.name
		 WHERE art.name is null
		GROUP BY trackartist
		 ;`
	effect, err = db.Exec(sqlStatement)
	rows, err := effect.RowsAffected()
	if err != nil {
		log.Println(err)
	} else {
		log.Println(" - Artists inserted : ", rows)
	}

	sqlStatement = `INSERT INTO song (title, contentid)
	SELECT tracktitle, cast(ld.contentid as int) as contentid
		FROM public.load as ld
		LEFT JOIN public.song as sng
			on cast(ld.contentid as int) = sng.contentid
		 WHERE sng.title is null
		GROUP BY ld.tracktitle, ld.contentid
		;`

	effect, err = db.Exec(sqlStatement)
	rows, _ = effect.RowsAffected()
	if err != nil {
		log.Println(err)
	} else {
		log.Println(" - Songs inserted : ", rows)
	}

	sqlStatement = `INSERT INTO PLAY (scantime, station, contentid, playtime) 
    SELECT 
	TO_TIMESTAMP(scantime, 'day, dd-mon-yy HH24:MI:SS') as converteddate,
    station,
    cast(contentid as int) as contentid,
    TO_TIMESTAMP(CONCAT(playdate, ', 2017 ', playtime), 'day, mon-dd, yyyy HH:MI AM') as convertplaytime
	FROM public.load as ld
   ;`
	effect, err = db.Exec(sqlStatement)
	rows, _ = effect.RowsAffected()
	if err != nil {
		log.Println(err)
	} else {
		log.Println(" - Plays inserted : ", rows)
	}

	sqlStatement = `UPDATE song 
	SET artistid=arts.artistid
	FROM (SELECT cast(contentid as int) as contentid, art.id as artistid
		FROM public.load as ld
		JOIN public.artist as art
			ON ld.trackartist = art.name) as arts
	WHERE song.contentid = arts.contentid
	AND song.artistid IS NULL;`
	effect, err = db.Exec(sqlStatement)
	rows, _ = effect.RowsAffected()
	if err != nil {
		log.Println(err)
	} else {
		log.Println(" - Artist ID's updated : ", rows)
	}

	sqlStatement = `DELETE FROM public.play
	WHERE ID NOT IN (
	SELECT min(id) as id
	FROM public.play
    GROUP BY scantime, station, contentid, playtime)
	;`
	effect, err = db.Exec(sqlStatement)
	rows, _ = effect.RowsAffected()
	if err != nil {
		log.Println(err)
	} else {
		log.Println(" - Duplicates Removed : ", rows)
	}
	return err
}
