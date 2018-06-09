package cosmicDB

import (
	"github.com/boltdb/bolt"
	"cosmicio/cosmicStruct"
	"time"
	"log"
	"encoding/json"
)

//Variables
var mainDB *bolt.DB

///Inits or loads all cosmic.io databases
func LoadDatabases() {
	var err error
	mainDB,err = bolt.Open("main.db",0600,nil) //Open highscores database
	if err!=nil {panic(err)}

	//Check database structure
	mainDB.Update(func (tx *bolt.Tx) error {
		tx.CreateBucketIfNotExists([]byte("highscores"))
		tx.CreateBucketIfNotExists([]byte("users"))
		return nil
	})

	//Make backup of database
 	x := func() {
	mainDB.View(func (tx *bolt.Tx) error {
		log.Println("CosmicDB Generating database backup")
		deleteFileIfExists("main.db.backup")
		err := tx.CopyFile("main.db.backup",0600)
		log.Println("CosmicDB Backup ended")
		return err
	})}
	go x()

	log.Println("CosmicDB Databases loaded")
}

func UpdateHighscores(ships *[]cosmicStruct.PlayerShip) {
	//Create entry array
	highScores := make([]highScoreEntry,0,len(*ships)+1)
	for _,ship := range *ships {
		highScoreEntry := highScoreEntry{
			Nick: ship.Username,
			Score: ship.Score,
			Date: time.Now(),
		}
		highScores := append(highScores,highScoreEntry)
		_ = highScores //Workaround
	}

	//Update database
	mainDB.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("highscores")) //Get bucket reference
		for i,_ := range highScores {
			id,_ :=bucket.NextSequence() //Get an id for entry
			data,_ := json.Marshal(highScores[i]) //Get JSON []bytes of iterating highscore entry
			bucket.Put(uitob(id),data) //Put highscore into database
		}
		return nil
	})
	mainDB.Sync() //Sync changes with filesystem
}

type highScoreEntry struct {
	Nick string
	Score int
	Date time.Time
}