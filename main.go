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

	checker := StarsNotifier{}
	go checker.MonitorStars()
}

func saveDb(database *db.Database) {
	err := database.Save()
	if err != nil {
		panic(err)
	}
}

type StarsNotifier struct {
	database          *db.Database
	stars             *[]*lib.Star
	lastStarCheck     int64
	lastListingUpdate int64
}

func (s *StarsNotifier) MonitorStars() {
	stars, forceUpdateListing, err := lib.GetStars()
	if err != nil {
		log.Println("Failed to get star list on start", err)
	}
	s.stars = stars

	for {
		now := time.Now().Unix()
		log.Println("Running cycle...", now)

		s.deleteOldStarMessages(s.database)

		if (now - s.lastStarCheck) >= int64(lib.SleepTime) {
			log.Println("Checking stars...")
			stars, forceUpdateListing, err = lib.GetStars()
			if err != nil {
				log.Println("failed to get stars:", err)
				s.waitLoop()
				s.lastStarCheck = now
				continue
			}

			listingUpdated := false
			if forceUpdateListing || (now-s.lastListingUpdate) >= int64(lib.ListingUpdateInterval*60) {
				if forceUpdateListing {
					log.Println("Force updating listing...")
				}
				err = s.updateListing(stars, s.database)
				if err != nil {
					log.Println("Failed to update listing", err)
					s.waitLoop()
					s.lastStarCheck = now
					listingUpdated = true
					continue
				} else {
					s.lastListingUpdate = now
				}
			}

			var newStars []*lib.Star

			for _, star := range *stars {
				if s.stars != nil && !slices.ContainsFunc(*s.stars, func(prev *lib.Star) bool {
					return star.CalledLocation == prev.CalledLocation && star.Location == prev.Location && star.World == prev.World
				}) {
					log.Println("- NEW STAR", *star)
					newStars = append(newStars, star)
				}
			}

			if len(newStars) > 0 {
				if !listingUpdated {
					if err = s.updateListing(stars, s.database); err != nil {
						log.Println("Failed to update listing after new star", err)
					}
					s.lastListingUpdate = now
				}

				err := lib.PostNewStars(&newStars, lib.WebhookUrls, now, s.database)
				if err != nil {
					log.Println("Failed to post new stars", err)
					s.waitLoop()
					s.lastStarCheck = now
					continue
				}
			}
			s.stars = stars
			s.lastStarCheck = now
		}

		s.waitLoop()
	}
}

func (s *StarsNotifier) deleteOldStarMessages(database *db.Database) {
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

func (s *StarsNotifier) updateListing(stars *[]*lib.Star, database *db.Database) error {
	err := lib.PostStarListing(stars, lib.WebhookUrls, database)
	return err
}

func (s *StarsNotifier) waitLoop() {
	time.Sleep(time.Duration(lib.SleepTime) * time.Second)
}
