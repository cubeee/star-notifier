package lib

import (
	"log"
	"slices"
	"star-notifier/lib/db"
	"time"
)

type StarsNotifier struct {
	Database          *db.Database
	stars             *[]*Star
	LastStarCheck     int64
	LastListingUpdate int64
	LastDowntime      *int64
}

func (s *StarsNotifier) MonitorStars() {
	stars, forceUpdateListing, err := GetStars(s.LastDowntime)
	if err != nil {
		log.Println("Failed to get star list on start", err)
	}
	s.stars = stars

	for {
		now := time.Now().Unix()
		log.Println("Running cycle...", now)

		s.deleteOldStarMessages()

		if (now - s.LastStarCheck) >= int64(SleepTime) {
			log.Println("Checking stars...")
			stars, forceUpdateListing, err = GetStars(s.LastDowntime)
			if err != nil {
				log.Println("failed to get stars:", err)
				s.waitLoop()
				s.LastStarCheck = now
				continue
			}

			listingUpdated := false
			if forceUpdateListing || (now-s.LastListingUpdate) >= int64(ListingUpdateInterval*60) {
				log.Println("Updating listings... forced:", forceUpdateListing)
				err = s.updateListing(stars, s.Database)
				if err != nil {
					log.Println("Failed to update listing", err)
					s.waitLoop()
					s.LastStarCheck = now
					listingUpdated = true
					continue
				} else {
					s.LastListingUpdate = now
				}
			}

			var newStars []*Star

			for _, star := range *stars {
				if s.stars != nil && !slices.ContainsFunc(*s.stars, func(prev *Star) bool {
					return star.CalledLocation == prev.CalledLocation && star.Location == prev.Location && star.World == prev.World
				}) {
					log.Println("- NEW STAR", *star)
					newStars = append(newStars, star)
				}
			}

			if len(newStars) > 0 {
				if !listingUpdated {
					if err = s.updateListing(stars, s.Database); err != nil {
						log.Println("Failed to update listing after new star", err)
					}
					s.LastListingUpdate = now
				}

				err := PostNewStars(&newStars, WebhookUrls, now, s.Database)
				if err != nil {
					log.Println("Failed to post new stars", err)
					s.waitLoop()
					s.LastStarCheck = now
					continue
				}
			}
			s.stars = stars
			s.LastStarCheck = now
		}

		s.waitLoop()
	}
}

func (s *StarsNotifier) deleteOldStarMessages() {
	oldMessages := s.Database.GetOldNewStarMessages(NewStarMessageMaxAge)
	if len(*oldMessages) == 0 {
		return
	}

	log.Printf("Removing %d old message(s)\n", len(*oldMessages))
	for _, message := range *oldMessages {
		err := DeleteMessage(message.WebhookUrl, message.MessageId)
		if err != nil {
			log.Printf("failed to delete message %s from %s: %f\n", message.MessageId, message.WebhookUrl, err)
		}
		time.Sleep(1 * time.Second)
	}
	s.Database.RemoveNewStarMessages(oldMessages)
	s.Database.SaveUnsafe()
}

func (s *StarsNotifier) updateListing(stars *[]*Star, database *db.Database) error {
	err := PostStarListing(stars, WebhookUrls, database)
	return err
}

func (s *StarsNotifier) waitLoop() {
	time.Sleep(time.Duration(SleepTime) * time.Second)
}
