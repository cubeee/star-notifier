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

	checker := lib.StarsNotifier{}
	go checker.MonitorStars()
}

func saveDb(database *db.Database) {
	err := database.Save()
	if err != nil {
		panic(err)
	}
}
