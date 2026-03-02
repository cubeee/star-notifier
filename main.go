package main

import (
	"fmt"
	"star-notifier/lib"
	"star-notifier/lib/db"
)

func main() {
	database, err := db.Load(fmt.Sprintf("%s/db.json", lib.DatabaseDirectory))
	if err != nil {
		panic(fmt.Errorf("failed to open db: %f", err))
	}
	defer saveDb(database)

	checker := &lib.StarsNotifier{
		Database:          database,
		LastDowntime:      nil,
		LastStarCheck:     0,
		LastListingUpdate: 0,
		NewStarsSeen:      0,
	}
	go checker.MonitorStars()

	web := lib.Web{
		Notifier: checker,
	}
	if err = web.Start(); err != nil {
		panic(err)
	}
}

func saveDb(database *db.Database) {
	err := database.Save()
	if err != nil {
		panic(err)
	}
}
