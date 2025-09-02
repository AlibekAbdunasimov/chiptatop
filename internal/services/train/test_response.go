package train

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
)

// TestAPIResponse tests parsing of the actual API response
func TestAPIResponse() {
	// Sample response from the actual API
	sampleResponse := `{
		"data": {
			"directions": {
				"forward": {
					"trains": [
						{
							"type": "СКРСТ",
							"number": "778Ф",
							"departureDate": "02.09.2025 06:03",
							"timeOnWay": "02:18",
							"originRoute": {
								"depStationName": "Toshkent Markaziy",
								"arvStationName": "Buxoro"
							},
							"arrivalDate": "02.09.2025 08:21",
							"brand": "Afrosiyob",
							"cars": [
								{
									"type": "O'rindiqli",
									"freeSeats": 77,
									"tariffs": [
										{
											"classServiceType": "1В",
											"freeSeats": 11,
											"tariff": 545000
										},
										{
											"classServiceType": "2Е",
											"freeSeats": 66,
											"tariff": 270000
										}
									]
								}
							],
							"subRoute": {
								"depStationName": "TOSHKENT",
								"depStationCode": "2900000",
								"arvStationName": "SAMARQAND",
								"arvStationCode": "2900700"
							},
							"trainId": null,
							"comment": null
						}
					]
				}
			}
		},
		"error": null
	}`

	var response SearchTrainsResponse
	err := json.Unmarshal([]byte(sampleResponse), &response)
	if err != nil {
		log.Fatalf("Failed to parse response: %v", err)
	}

	if response.Data == nil {
		log.Fatal("No data in response")
	}

	if response.Data.Directions.Forward == nil {
		log.Fatal("No forward direction data")
	}

	trains := response.Data.Directions.Forward.Trains
	if len(trains) == 0 {
		log.Fatal("No trains found")
	}

	train := trains[0]
	fmt.Printf("Successfully parsed train data:\n")
	fmt.Printf("Train: %s %s\n", train.Brand, train.Number)
	fmt.Printf("Route: %s -> %s\n", train.SubRoute.DepStationName, train.SubRoute.ArvStationName)
	fmt.Printf("Time: %s - %s (%s)\n", train.GetDepartureTime(), train.GetArrivalTime(), train.TimeOnWay)
	fmt.Printf("Date: %s\n", train.GetDate())
	fmt.Printf("Has available seats: %v\n", train.HasAvailableSeats())
	fmt.Printf("Total free seats: %d\n", train.GetTotalFreeSeats())
	fmt.Printf("Min price: %d UZS\n", train.GetMinPrice())

	// Test formatting
	service := NewService()
	formatted := service.FormatTrainInfo(train)
	fmt.Printf("\nFormatted output:\n%s\n", formatted)
}

// TestStationCodes tests the station code mapping
func TestStationCodes() {
	service := NewService()

	testCases := []struct {
		input    string
		expected string
	}{
		{"Toshkent", "2900000"},
		{"toshkent", "2900000"},
		{"Tashkent", "2900000"}, // Alternative spelling
		{"Samarqand", "2900700"},
		{"samarkand", "2900700"}, // Alternative spelling
		{"Buxoro", "2900800"},
		{"2900000", "2900000"},           // Already a code
		{"Unknown City", "Unknown City"}, // Unknown city
	}

	fmt.Println("Testing station code mapping:")
	for _, tc := range testCases {
		result := service.GetStationCode(tc.input)
		status := "✅"
		if result != tc.expected {
			status = "❌"
		}
		fmt.Printf("%s %s -> %s (expected: %s)\n", status, tc.input, result, tc.expected)
	}
}

// TestPriceFormatting tests the price formatting function
func TestPriceFormatting() {
	service := NewService()

	testCases := []struct {
		input    int
		expected string
	}{
		{270000, "270 000"},
		{545000, "545 000"},
		{1000, "1 000"},
		{999, "999"},
		{1234567, "1 234 567"},
	}

	fmt.Println("\nTesting price formatting:")
	for _, tc := range testCases {
		result := service.formatPrice(tc.input)
		status := "✅"
		if result != tc.expected {
			status = "❌"
		}
		fmt.Printf("%s %d -> %s (expected: %s)\n", status, tc.input, result, tc.expected)
	}
}

// RunAllTests runs all test functions
func RunAllTests() {
	fmt.Println("=== Running Train Service Tests ===\n")

	fmt.Println("1. Testing API Response Parsing:")
	TestAPIResponse()

	fmt.Println("\n" + strings.Repeat("=", 50))
	TestStationCodes()

	fmt.Println("\n" + strings.Repeat("=", 50))
	TestPriceFormatting()

	fmt.Println("\n=== All Tests Completed ===")
}
