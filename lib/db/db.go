package db

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"slices"
	"time"
)

type NewStarMessage struct {
	WebhookUrl      string `json:"webhookUrl"`
	MessageId       string `json:"messageId"`
	PostedTimestamp int64  `json:"postedTimestamp"`
}

type Database struct {
	filePath string
	content  *DatabaseContent
}

type DatabaseContent struct {
	ListingMessages map[string]string `json:"listingMessages"`
	NewStarMessages []NewStarMessage  `json:"newStarMessages"`
}

func (db *Database) GetListingMessage(webhookUrl string) *string {
	messageId := db.content.ListingMessages[webhookUrl]
	return &messageId
}

func (db *Database) SetListingMessage(webhookUrl, messageId string) {
	db.content.ListingMessages[webhookUrl] = messageId
}

func (db *Database) AddNewStarMessage(webhookUrl, messageId string, timestamp int64) {
	db.content.NewStarMessages = append(db.content.NewStarMessages, NewStarMessage{
		WebhookUrl:      webhookUrl,
		MessageId:       messageId,
		PostedTimestamp: timestamp,
	})
}

func (db *Database) RemoveNewStarMessages(messages *[]*NewStarMessage) {
	db.content.NewStarMessages = slices.DeleteFunc(
		db.content.NewStarMessages,
		func(storedMessage NewStarMessage) bool {
			return slices.ContainsFunc(*messages, func(message *NewStarMessage) bool {
				return storedMessage.WebhookUrl == message.WebhookUrl && storedMessage.MessageId == message.MessageId
			})
		},
	)
}

func (db *Database) GetOldNewStarMessages(maxAge int) *[]*NewStarMessage {
	now := time.Now().Unix()
	old := slices.Collect(func(yield func(star *NewStarMessage) bool) {
		for _, newStarMessage := range db.content.NewStarMessages {
			age := now - newStarMessage.PostedTimestamp
			if age > int64(maxAge) {
				yield(&newStarMessage)
			}
		}
	})
	return &old
}

func Load(filePath string) (*Database, error) {
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_RDONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open db: %w", err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			panic(err)
		}
	}(file)

	fileContent, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read db file: %w", err)
	}

	var dbContent DatabaseContent
	if len(fileContent) == 0 {
		dbContent = createNewDatabaseContent()
	} else {
		err = json.Unmarshal(fileContent, &dbContent)
		if err != nil {
			fmt.Println(err)
			return nil, fmt.Errorf("failed to unmarshal db file content: %w", err)
		}
	}

	return &Database{
		filePath: filePath,
		content:  &dbContent,
	}, nil
}

func (db *Database) SaveUnsafe() {
	err := db.Save()
	if err != nil {
		panic(err)
	}
}

func (db *Database) Save() error {
	dbContent, err := json.Marshal(db.content)
	if err != nil {
		return fmt.Errorf("failed to marshal db contents: %w", err)
	}
	err = os.WriteFile(db.filePath, dbContent, 0644)
	if err != nil {
		return fmt.Errorf("failed to write db file: %w", err)
	}
	return nil
}

func createNewDatabaseContent() DatabaseContent {
	return DatabaseContent{
		ListingMessages: make(map[string]string),
		NewStarMessages: make([]NewStarMessage, 0),
	}
}
