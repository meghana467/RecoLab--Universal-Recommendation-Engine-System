package demos

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/recolab/backend/internal/gorse"
)

// SeedAll loads all data files from dataDir and pushes them to Gorse.
func SeedAll(client *gorse.Client, dataDir string) error {
	// Users
	usersPath := filepath.Join(dataDir, "users.json")
	usersData, err := os.ReadFile(usersPath)
	if err != nil {
		return fmt.Errorf("read users.json: %w", err)
	}
	var users []gorse.User
	if err := json.Unmarshal(usersData, &users); err != nil {
		return fmt.Errorf("parse users.json: %w", err)
	}
	if err := client.InsertUsers(users); err != nil {
		return fmt.Errorf("insert users: %w", err)
	}
	fmt.Printf("Seeded %d users\n", len(users))

	// Items
	itemFiles := []string{"news_items.json", "product_items.json", "expo_booths.json"}
	for _, f := range itemFiles {
		path := filepath.Join(dataDir, f)
		data, err := os.ReadFile(path)
		if err != nil {
			fmt.Printf("Warning: skip %s: %v\n", f, err)
			continue
		}
		var items []gorse.Item
		if err := json.Unmarshal(data, &items); err != nil {
			fmt.Printf("Warning: parse %s: %v\n", f, err)
			continue
		}
		if err := client.InsertItems(items); err != nil {
			fmt.Printf("Warning: insert %s: %v\n", f, err)
			continue
		}
		fmt.Printf("Seeded %d items from %s\n", len(items), f)
	}

	// Feedback in batches
	feedbackPath := filepath.Join(dataDir, "feedback.json")
	feedbackData, err := os.ReadFile(feedbackPath)
	if err != nil {
		return fmt.Errorf("read feedback.json: %w", err)
	}
	var feedback []gorse.Feedback
	if err := json.Unmarshal(feedbackData, &feedback); err != nil {
		return fmt.Errorf("parse feedback.json: %w", err)
	}

	batchSize := 200
	for i := 0; i < len(feedback); i += batchSize {
		end := i + batchSize
		if end > len(feedback) {
			end = len(feedback)
		}
		if err := client.InsertFeedback(feedback[i:end]); err != nil {
			return fmt.Errorf("insert feedback batch %d: %w", i/batchSize, err)
		}
	}
	fmt.Printf("Seeded %d feedback events\n", len(feedback))

	return nil
}
