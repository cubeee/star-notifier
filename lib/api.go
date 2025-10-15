package lib

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strconv"
	"time"
)

type Star struct {
	Location       int
	CalledLocation string
	MappedLocation *StarLocation
	Tier           int
	World          int
	CalledAt       int64
	MinTime        int64
	MaxTime        int64
	DepleteTime    int64
}

type StarsResponse struct {
	World          int     `json:"world"`
	Location       int     `json:"location"`
	CalledLocation string  `json:"calledLocation"`
	CalledAt       float64 `json:"calledAt"`
	Tier           int     `json:"tier"`
	MinTime        int64   `json:"minTime"`
	MaxTime        int64   `json:"maxTime"`
}

func GetStars() (*[]*Star, error) {
	response, err := getStars(ApiUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to get stars: %f", err)
	}
	now := time.Now().Unix()
	var stars []*Star
	for _, star := range *response {
		depleteTime := int64(star.CalledAt) + int64(star.Tier*420)
		if depleteTime < now {
			continue
		}

		if slices.Contains(ExcludedWorlds, strconv.Itoa(star.World)) {
			continue
		}

		if !slices.Contains(AllowedLocations, strconv.Itoa(star.Location)) {
			continue
		}

		mappedLocation := GetStarLocation(star.CalledLocation)
		if mappedLocation == nil {
			continue
		}

		stars = append(stars, &Star{
			Location:       star.Location,
			CalledLocation: star.CalledLocation,
			MappedLocation: mappedLocation,
			Tier:           star.Tier,
			World:          star.World,
			CalledAt:       int64(star.CalledAt),
			MinTime:        star.MinTime,
			MaxTime:        star.MaxTime,
			DepleteTime:    int64(star.CalledAt) + int64(star.Tier*420),
		})
	}
	return &stars, nil
}

func getStars(url string) (*[]*StarsResponse, error) {
	now := strconv.FormatInt(time.Now().UnixMilli(), 10)
	client := http.Client{
		Timeout: time.Second * time.Duration(ApiTimeout),
	}
	req, err := http.NewRequest(http.MethodGet, url+"?timestamp="+now, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %f", err)
	}
	if len(ApiUserAgent) > 0 {
		req.Header.Add("User-Agent", ApiUserAgent)
	}
	if len(ApiReferer) > 0 {
		req.Header.Add("Referer", ApiReferer)
	}
	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get stars: %f", err)
	}

	if res.Body != nil {
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				panic(err)
			}
		}(res.Body)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %f", err)
	}

	var stars []*StarsResponse
	err = json.Unmarshal(body, &stars)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response body: %f", err)
	}

	return &stars, nil
}
