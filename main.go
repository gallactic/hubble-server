package main

import (
	"time"

	bc "github.com/gallactic/hubble_server/blockchain"
	config "github.com/gallactic/hubble_server/config"
	db "github.com/gallactic/hubble_server/database"
	ex "github.com/gallactic/hubble_server/explorer"
)

var explorerEngine ex.Explorer
var gConfig *config.Config

func main() {

	Init()
	SyncLoop()

}

//Init initializes engine
func Init() {
	gConfig, _ = config.LoadConfigFile(true)
	bcAdapter := bc.Gallactic{Config: gConfig}
	dbAdapter := db.Postgre{Config: gConfig}
	explorerEngine = ex.Explorer{BCAdapter: &bcAdapter, DBAdapter: &dbAdapter, Config: gConfig}

	explorerEngine.Init()
}

//SyncLoop goes in loop for syncing blockchain and database
func SyncLoop() {

	interval := time.Duration(gConfig.App.CheckingInterval)
	println("syncing every", interval, "miliseconds...")

	for {

		errUpdate := explorerEngine.Update()
		if errUpdate != nil {
			println("Updating engine erro: ", errUpdate.Error())
		}
		time.Sleep(interval * time.Millisecond)

	}
}
