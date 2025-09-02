package train

import (
	"context"
	"log"
	"time"
)

// ExampleUsage demonstrates how to use the train service
func ExampleUsage() {
	// Create a new train service
	service := NewService()

	// Set authentication credentials (you'll get these from the actual API response)
	// These would typically come from a login flow or session management
	service.SetAuthCredentials("003c204f-01ca-4820-85cc-925fa66a6c41",
		"_ga=GA1.1.2112044518.1734322567; __stripe_mid=aa188b56-afd0-4e79-9731-b08f6971b6d837b937")

	// Example 1: Search for trains
	ctx := context.Background()
	searchParams := TrainSearchParams{
		From: "Toshkent",
		To:   "Samarqand",
		Date: time.Date(2025, 9, 2, 0, 0, 0, 0, time.UTC),
	}

	trains, err := service.FindAvailableTrains(ctx, searchParams)
	if err != nil {
		log.Printf("Error searching trains: %v", err)
		return
	}

	log.Printf("Found %d available trains", len(trains))
	for _, train := range trains {
		log.Printf("Train: %s", service.FormatTrainInfo(train))
		log.Printf("Total free seats: %d, Min price: %d UZS", train.GetTotalFreeSeats(), train.GetMinPrice())
	}

	// Example 2: Create a ticket alert
	alert := TicketAlert{
		ID:        "alert-123",
		UserID:    123456789,
		ChatID:    123456789,
		From:      "Toshkent",
		To:        "Samarqand",
		Date:      time.Date(2025, 9, 2, 0, 0, 0, 0, time.UTC),
		SeatTypes: []string{"O'rindiqli", "Kupe"},
		MinPrice:  0,
		MaxPrice:  500000, // 500,000 UZS
		IsActive:  true,
		CreatedAt: time.Now(),
	}

	// Check if any trains match the alert criteria
	matchingTrains, err := service.CheckTicketAvailability(ctx, alert)
	if err != nil {
		log.Printf("Error checking ticket availability: %v", err)
		return
	}

	if len(matchingTrains) > 0 {
		log.Printf("Alert triggered! Found %d matching trains:", len(matchingTrains))
		for _, train := range matchingTrains {
			log.Printf("Matching train: %s", service.FormatTrainInfo(train))
		}
	} else {
		log.Printf("No trains match the alert criteria yet")
	}
}

// CreateSampleRequest creates a sample request that matches the curl command you provided
func CreateSampleRequest() *SearchTrainsRequest {
	return &SearchTrainsRequest{
		Directions: Directions{
			Forward: &Journey{
				Date:           "2025-09-02",
				DepStationCode: "2900000", // Tashkent
				ArvStationCode: "2900700", // Samarkand
			},
		},
	}
}

// Real station codes from Uzbekistan Railways
var CommonStationCodes = map[string]string{
	"Andijon":   "2900680",
	"Buxoro":    "2900800",
	"Guliston":  "2900850",
	"Jizzax":    "2900720",
	"Margilon":  "2900920",
	"Namangan":  "2900940",
	"Navoiy":    "2900930",
	"Nukus":     "2900970",
	"Pop":       "2900693",
	"Qarshi":    "2900750",
	"Qoqon":     "2900880",
	"Samarqand": "2900700",
	"Termiz":    "2900255",
	"Toshkent":  "2900000",
	"Urgench":   "2900790",
	"Xiva":      "2900172",
}
