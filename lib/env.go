package lib

import (
	"os"
)

var (
	DatabaseDirectory     = GetEnv("DATABASE_DIRECTORY", "data")
	ApiUrl                = os.Getenv("STARS_API_URL")
	ApiTimeout            = GetEnvInt("STARS_API_TIMEOUT", 5)
	ApiUserAgent          = os.Getenv("STARS_API_USER_AGENT")
	ApiReferer            = os.Getenv("STARS_API_REFERER")
	AllowedLocations      = GetEnvList("ALLOWED_LOCATIONS", ",")
	SleepTime             = GetEnvInt("SLEEP_TIME_SECONDS", 30)
	ListingUpdateInterval = GetEnvInt("LISTING_UPDATE_INTERVAL", 1)
	MapWidth              = GetEnvInt("MAP_WIDTH", 512)
	MapHeight             = GetEnvInt("MAP_HEIGHT", 512)
	WebhookUrls           = GetEnvList("DISCORD_WEBHOOK_URLS", ",")
	ListingFooter         = GetEnv("LISTING_FOOTER", "")
	NewStarMessageMaxAge  = GetEnvInt("NEW_STAR_MESSAGE_MAX_AGE", 50)
)
