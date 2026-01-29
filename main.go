package main

import (
	"fmt"
	"log"
	"slices"
	"star-notifier/lib"
	"star-notifier/lib/db"
	"time"
)

func main() {
	database, err := db.Load(fmt.Sprintf("%s/db.json", lib.DatabaseDirectory))
	if err != nil {
		panic(fmt.Errorf("failed to open db: %f", err))
	}
	defer saveDb(database)

	lastListingUpdate := int64(0)
	lastStarCheck := int64(0)

	stars, forceUpdateListing, err := lib.GetStars()
	if err != nil {
		log.Println("Failed to get star list on start", err)
	}
	previousStars := stars

	for {
		now := time.Now().Unix()
		log.Println("Running cycle...", now)

		deleteOldStarMessages(database)

		if (now - lastStarCheck) >= int64(lib.SleepTime) {
			log.Println("Checking stars...")
			stars, forceUpdateListing, err = lib.GetStars()
			if err != nil {
				log.Println("failed to get stars:", err)
				waitLoop()
				lastStarCheck = now
				continue
			}

			listingUpdated := false
			if forceUpdateListing || (now-lastListingUpdate) >= int64(lib.ListingUpdateInterval*60) {
				if forceUpdateListing {
					log.Println("Force updating listing...")
				}
				err = updateListing(stars, database)
				if err != nil {
					log.Println("Failed to update listing", err)
					waitLoop()
					lastStarCheck = now
					listingUpdated = true
					continue
				} else {
					lastListingUpdate = now
				}
			}

			var newStars []*lib.Star

			for _, star := range *stars {
				if previousStars != nil && !slices.ContainsFunc(*previousStars, func(prev *lib.Star) bool {
					return star.CalledLocation == prev.CalledLocation && star.Location == prev.Location && star.World == prev.World
				}) {
					log.Println("- NEW STAR", *star)
					newStars = append(newStars, star)
				}
			}

			if len(newStars) > 0 {
				if !listingUpdated {
					if err = updateListing(stars, database); err != nil {
						log.Println("Failed to update listing after new star", err)
					}
					lastListingUpdate = now
				}

				err := lib.PostNewStars(&newStars, lib.WebhookUrls, now, database)
				if err != nil {
					log.Println("Failed to post new stars", err)
					waitLoop()
					lastStarCheck = now
					continue
				}
			}
			previousStars = stars
			lastStarCheck = now
		}

		waitLoop()
	}
}

func deleteOldStarMessages(database *db.Database) {
	oldMessages := database.GetOldNewStarMessages(lib.NewStarMessageMaxAge)
	if len(*oldMessages) == 0 {
		return
	}

	log.Printf("Removing %d old message(s)\n", len(*oldMessages))
	for _, message := range *oldMessages {
		err := lib.DeleteMessage(message.WebhookUrl, message.MessageId)
		if err != nil {
			log.Printf("failed to delete message %s from %s: %f\n", message.MessageId, message.WebhookUrl, err)
		}
		time.Sleep(1 * time.Second)
	}
	database.RemoveNewStarMessages(oldMessages)
	database.SaveUnsafe()
}

func updateListing(stars *[]*lib.Star, database *db.Database) error {
	err := lib.PostStarListing(stars, lib.WebhookUrls, database)
	return err
}

func waitLoop() {
	time.Sleep(time.Duration(lib.SleepTime) * time.Second)
}

func saveDb(database *db.Database) {
	err := database.Save()
	if err != nil {
		panic(err)
	}
}
