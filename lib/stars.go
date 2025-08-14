package lib

import (
	"strings"
)

type StarLocation struct {
	X int
	Y int
}

var StarLocations map[string]StarLocation

// TODO: load from json?
func init() {
	StarLocations = make(map[string]StarLocation)
	StarLocations["south east"] = StarLocation{X: 1745, Y: 2954}
	StarLocations["hunter"] = StarLocation{X: 1487, Y: 3090}
	StarLocations["colosseum"] = StarLocation{X: 1773, Y: 3102}
	StarLocations["salvager overlook"] = StarLocation{X: 1627, Y: 3275}
	StarLocations["aldarin"] = StarLocation{X: 1422, Y: 2874}
	StarLocations["custodia"] = StarLocation{X: 1290, Y: 3411}
}

func GetStarLocation(calledLocation string) *StarLocation {
	for trigger, location := range StarLocations {
		if strings.Contains(strings.ToLower(calledLocation), trigger) {
			return &location
		}
	}
	return nil
}
