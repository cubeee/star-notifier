package lib

import (
	"bytes"
	"encoding/json"
	"fmt"
	mapLib "github.com/cubeee/ent-notifier/lib"
	"image/png"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"slices"
	"sort"
	"star-notifier/lib/db"
	"strconv"
	"strings"
)

func getWebhookUrl(url string) (string, *string) {
	if strings.Contains(url, "=") {
		parts := strings.Split(url, "=")
		url := parts[0]
		roleId := parts[1]
		return url, &roleId
	}
	return url, nil
}

func PostStarListing(currentStars *[]*Star, webhookUrls []string, database *db.Database) error {
	listingMessage, err := createCurrentStarsMessage(currentStars)
	if err != nil {
		return fmt.Errorf("failed to create listing message: %f", err)
	}

	for _, url := range webhookUrls {
		if len(url) == 0 {
			continue
		}
		url, _ = getWebhookUrl(url)

		listingMessageId := database.GetListingMessage(url)
		if listingMessageId == nil || len(*listingMessageId) == 0 {
			log.Println("Listing message does not exist for url", url, "-- posting new message...")
			postListingMessage(listingMessage, url, database)
		} else {
			log.Println("Listing message stored for", url, "-- updating...")
			statusCode, err := editMessage(url, *listingMessageId, listingMessage)
			if err != nil {
				if statusCode == http.StatusNotFound {
					log.Println("Failed to edit listing message to", url, "-- posting new listing message")
					log.Println(err)
					postListingMessage(listingMessage, url, database)
				} else {
					log.Println("Failed to edit listing message to url", url, "--", err)
				}
			}
		}
	}
	return nil
}

func PostNewStars(stars *[]*Star, webhookUrls []string, timestamp int64, database *db.Database) error {
	for _, url := range webhookUrls {
		if len(url) == 0 {
			continue
		}
		url, roleId := getWebhookUrl(url)
		message, err := createNewStarMessage(stars, roleId)
		if err != nil {
			log.Println(fmt.Sprintf("failed to create new star message for %s: %f", url, err))
			continue
		}
		postNewStarMessage(message, url, timestamp, database)
	}
	return nil
}

func postListingMessage(message *DiscordMessage, webhookUrl string, database *db.Database) {
	messageId, err := postMessage(webhookUrl, message)
	if err != nil {
		log.Println("Failed to post listing webhook to url", webhookUrl, err)
	}
	if len(messageId) > 0 {
		database.SetListingMessage(webhookUrl, messageId)
		database.SaveUnsafe()
	}
}

func postNewStarMessage(message *DiscordMessage, webhookUrl string, timestamp int64, database *db.Database) {
	messageId, err := postMessage(webhookUrl, message)
	if err != nil {
		log.Println("Failed to post new star webhook to url", webhookUrl, err)
	}
	if len(messageId) > 0 {
		database.AddNewStarMessage(webhookUrl, messageId, timestamp)
		database.SaveUnsafe()
	}
}

func createNewStarMessage(stars *[]*Star, roleId *string) (*DiscordMessage, error) {
	var lines []string
	sort.Slice(*stars, func(a, b int) bool {
		return (*stars)[b].DepleteTime < (*stars)[a].DepleteTime
	})

	if roleId != nil {
		lines = append(lines, fmt.Sprintf("<@&%s>", *roleId))
	}

	for _, star := range *stars {
		lines = append(lines, fmt.Sprintf(
			"[NEW STAR] World %d, tier %d, %s (est. depletion: %s)",
			star.World,
			star.Tier,
			star.CalledLocation,
			fmt.Sprintf("<t:%d:R>", star.DepleteTime),
		))
	}

	lines = append(lines, "-# This is a temporary message to get your attention, use the listing")

	content := strings.Join(lines, "\n")

	message := &DiscordMessage{
		Content: content,
	}
	return message, nil
}

func createCurrentStarsMessage(stars *[]*Star) (*DiscordMessage, error) {
	content := ""
	footer := ""
	if len(ListingFooter) > 0 {
		footer = "\n-# " + ListingFooter
	}

	var files []DiscordFile
	var attachments []DiscordAttachment
	var starLocations []*StarLocation

	sort.Slice(*stars, func(a, b int) bool {
		return (*stars)[b].DepleteTime < (*stars)[a].DepleteTime
	})

	if len(*stars) == 0 {
		content += "No stars at the moment :(\n"
	}

	for _, star := range *stars {
		line := fmt.Sprintf(
			"[World %d, tier %d] %s (est. depletion: %s)",
			star.World,
			star.Tier,
			star.CalledLocation,
			fmt.Sprintf("<t:%d:R>", star.DepleteTime),
		)

		if len(content)+len(line)+len(footer) > 2000 {
			break
		}

		content += line + "\n"

		location := GetStarLocation(star.CalledLocation)
		if location != nil {
			if !slices.Contains(starLocations, location) {
				starLocations = append(starLocations, location)
			}
		}
	}

	for _, location := range starLocations {
		x := location.X
		y := location.Y
		fileId := fmt.Sprintf("%d_%d.png", x, y)
		imageName := fmt.Sprintf("map%d_%d.png", x, y)

		if mapImage, err := mapLib.CreateThumbnail(x, y, MapWidth, MapHeight); err == nil {
			buffer := new(bytes.Buffer)
			err = png.Encode(buffer, mapImage)
			if err != nil {
				continue
			}

			imageBuffer := buffer.Bytes()
			files = append(files, DiscordFile{
				Id:   &fileId,
				Name: imageName,
				Data: &imageBuffer,
			})
			attachments = append(attachments, DiscordAttachment{
				Id:          fileId,
				FileName:    imageName,
				Description: imageName,
			})
		}
	}

	if len(footer) > 0 {
		content = content + footer
	}

	message := &DiscordMessage{
		Content: content,
		Files:   &files,
	}

	return message, nil
}

func postMessage(webhookUrl string, message *DiscordMessage) (string, error) {
	url := fmt.Sprintf("%s?wait=true", webhookUrl)

	payload, contentType, err := encodeMessage(message)
	if err != nil {
		return "", fmt.Errorf("failed to encode message: %f", err)
	}

	resp, err := http.Post(url, contentType, payload)
	if err != nil {
		return "", fmt.Errorf("failed to post message: %f", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			panic(err)
		}
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %f", err)
	}

	var jsonBody map[string]json.RawMessage
	if err = json.Unmarshal(body, &jsonBody); err != nil {
		return "", fmt.Errorf("failed to unmarshal response body: %f", err)
	}

	var messageId string
	if err = json.Unmarshal(jsonBody["id"], &messageId); err != nil {
		return "", fmt.Errorf("failed to get message id from response body: %f, %v", err, string(body))
	}

	return messageId, nil
}

func editMessage(webhookUrl, messageId string, message *DiscordMessage) (int, error) {
	payload, contentType, err := encodeMessage(message)
	if err != nil {
		return -1, fmt.Errorf("failed to encode edit message: %f", err)
	}

	url := fmt.Sprintf("%s/messages/%s", webhookUrl, messageId)
	req, err := http.NewRequest(http.MethodPatch, url, payload)
	if err != nil {
		return -1, fmt.Errorf("failed to create edit message request: %f", err)
	}
	req.Header.Set("Content-Type", contentType)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return -1, fmt.Errorf("failed to edit message: %f", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return resp.StatusCode, fmt.Errorf("failed to edit webhook: %d", resp.StatusCode)
	}

	return resp.StatusCode, nil
}

func DeleteMessage(webhookUrl, messageId string) error {
	url := fmt.Sprintf("%s/messages/%s", webhookUrl, messageId)
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create delete message request: %f", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil || resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("failed to delete message: %f (%d)", err, resp.StatusCode)
	}
	return nil
}

func encodeMessage(message *DiscordMessage) (*bytes.Buffer, string, error) {
	payload := new(bytes.Buffer)

	if message.Files != nil {
		writer := multipart.NewWriter(payload)

		partWriter, err := writer.CreateFormField("payload_json")
		if err != nil {
			return nil, "", err
		}

		err = json.NewEncoder(partWriter).Encode(message)
		if err != nil {
			return nil, "", err
		}

		for index, file := range *message.Files {
			fileId := strconv.Itoa(index)
			if file.Id != nil {
				fileId = *file.Id
			}

			partWriter, err := writer.CreateFormFile(fmt.Sprintf("files[%v]", fileId), file.Name)
			if err != nil {
				return nil, "", err
			}

			_, err = partWriter.Write(*file.Data)
			if err != nil {
				return nil, "", err
			}
		}

		err = writer.Close()
		if err != nil {
			return nil, "", err
		}

		return payload, "multipart/form-data; boundary=" + writer.Boundary(), nil
	} else {
		err := json.NewEncoder(payload).Encode(message)
		return payload, "application/json", err
	}
}
