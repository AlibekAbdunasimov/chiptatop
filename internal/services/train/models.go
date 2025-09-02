package train

import "time"

// SearchTrainsRequest represents the request structure for searching trains
type SearchTrainsRequest struct {
	Directions Directions `json:"directions"`
}

// Directions contains forward and optionally return journey information
type Directions struct {
	Forward *Journey `json:"forward"`
	Return  *Journey `json:"return,omitempty"`
}

// Journey represents a single journey direction
type Journey struct {
	Date           string `json:"date"`           // Format: "2025-09-02"
	DepStationCode string `json:"depStationCode"` // Departure station code
	ArvStationCode string `json:"arvStationCode"` // Arrival station code
}

// SearchTrainsResponse represents the response from the train search API
type SearchTrainsResponse struct {
	Data  *TrainSearchData `json:"data,omitempty"`
	Error *APIError        `json:"error,omitempty"`
}

// TrainSearchData contains the main response data
type TrainSearchData struct {
	Directions DirectionsResponse `json:"directions"`
}

// DirectionsResponse contains forward and return journey trains
type DirectionsResponse struct {
	Forward *DirectionTrains `json:"forward,omitempty"`
	Return  *DirectionTrains `json:"return,omitempty"`
}

// DirectionTrains contains trains for a specific direction
type DirectionTrains struct {
	Trains []Train `json:"trains"`
}

// Train represents a train with its details (matching actual API response)
type Train struct {
	Type          string    `json:"type"`          // e.g., "СКРСТ", "СК", "ск"
	Number        string    `json:"number"`        // e.g., "778Ф"
	DepartureDate string    `json:"departureDate"` // e.g., "02.09.2025 06:03"
	ArrivalDate   string    `json:"arrivalDate"`   // e.g., "02.09.2025 08:21"
	TimeOnWay     string    `json:"timeOnWay"`     // e.g., "02:18"
	Brand         string    `json:"brand"`         // e.g., "Afrosiyob", "Sharq"
	OriginRoute   RouteInfo `json:"originRoute"`   // Full route info
	SubRoute      SubRoute  `json:"subRoute"`      // Searched segment
	Cars          []Car     `json:"cars"`          // Available cars/seats
	TrainID       *string   `json:"trainId"`       // Can be null
	Comment       *string   `json:"comment"`       // Can be null
}

// RouteInfo contains the full route information
type RouteInfo struct {
	DepStationName string `json:"depStationName"` // e.g., "Toshkent Markaziy"
	ArvStationName string `json:"arvStationName"` // e.g., "Buxoro"
}

// SubRoute contains the searched segment information
type SubRoute struct {
	DepStationName string `json:"depStationName"` // e.g., "TOSHKENT"
	DepStationCode string `json:"depStationCode"` // e.g., "2900000"
	ArvStationName string `json:"arvStationName"` // e.g., "SAMARQAND"
	ArvStationCode string `json:"arvStationCode"` // e.g., "2900700"
}

// Car represents a train car with available seats
type Car struct {
	Type      string   `json:"type"`      // e.g., "O'rindiqli", "Plaskartli", "Kupe"
	FreeSeats int      `json:"freeSeats"` // Total free seats in this car type
	Tariffs   []Tariff `json:"tariffs"`   // Different service classes
}

// Tariff represents pricing and availability for a specific service class
type Tariff struct {
	ClassServiceType string `json:"classServiceType"` // e.g., "1В", "2Е", "3П"
	FreeSeats        int    `json:"freeSeats"`        // Available seats in this class
	Tariff           int    `json:"tariff"`           // Price in UZS (as integer)
}

// Station represents a railway station
type Station struct {
	Code    string `json:"code"`
	Name    string `json:"name"`
	City    string `json:"city"`
	Region  string `json:"region"`
	Country string `json:"country"`
}

// Legacy structures for backward compatibility
type SeatClass struct {
	Type        string  `json:"type"`      // e.g., "ECONOMY", "BUSINESS", "LUXURY"
	Name        string  `json:"name"`      // Localized name
	Price       float64 `json:"price"`     // Price in UZS
	Currency    string  `json:"currency"`  // "UZS"
	Available   int     `json:"available"` // Number of available seats
	Total       int     `json:"total"`     // Total seats in this class
	IsAvailable bool    `json:"isAvailable"`
}

// APIError represents an error response from the API
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// TrainSearchParams represents user-friendly search parameters
type TrainSearchParams struct {
	From string    `json:"from"` // Station name or code
	To   string    `json:"to"`   // Station name or code
	Date time.Time `json:"date"` // Travel date
}

// TicketAlert represents a ticket availability alert
type TicketAlert struct {
	ID          string    `json:"id"`
	UserID      int64     `json:"userId"`      // Telegram user ID
	ChatID      int64     `json:"chatId"`      // Telegram chat ID
	From        string    `json:"from"`        // Departure station
	To          string    `json:"to"`          // Arrival station
	Date        time.Time `json:"date"`        // Travel date
	SeatTypes   []string  `json:"seatTypes"`   // Preferred seat classes
	MinPrice    float64   `json:"minPrice"`    // Minimum acceptable price
	MaxPrice    float64   `json:"maxPrice"`    // Maximum acceptable price
	IsActive    bool      `json:"isActive"`    // Whether alert is active
	CreatedAt   time.Time `json:"createdAt"`   // When alert was created
	LastChecked time.Time `json:"lastChecked"` // Last check time
	NotifyCount int       `json:"notifyCount"` // Number of notifications sent
}

// NotificationPayload represents data for sending notifications
type NotificationPayload struct {
	Alert TicketAlert `json:"alert"`
	Train Train       `json:"train"`
	Seats []SeatClass `json:"availableSeats"`
}

// Helper methods for Train struct

// GetDepartureTime extracts time from departureDate
func (t *Train) GetDepartureTime() string {
	if len(t.DepartureDate) >= 16 {
		return t.DepartureDate[11:16] // Extract "06:03" from "02.09.2025 06:03"
	}
	return t.DepartureDate
}

// GetArrivalTime extracts time from arrivalDate
func (t *Train) GetArrivalTime() string {
	if len(t.ArrivalDate) >= 16 {
		return t.ArrivalDate[11:16] // Extract "08:21" from "02.09.2025 08:21"
	}
	return t.ArrivalDate
}

// GetDate extracts date from departureDate
func (t *Train) GetDate() string {
	if len(t.DepartureDate) >= 10 {
		return t.DepartureDate[:10] // Extract "02.09.2025" from "02.09.2025 06:03"
	}
	return t.DepartureDate
}

// HasAvailableSeats checks if train has any available seats
func (t *Train) HasAvailableSeats() bool {
	for _, car := range t.Cars {
		for _, tariff := range car.Tariffs {
			if tariff.FreeSeats > 0 {
				return true
			}
		}
	}
	return false
}

// GetTotalFreeSeats returns total number of free seats across all cars
func (t *Train) GetTotalFreeSeats() int {
	total := 0
	for _, car := range t.Cars {
		total += car.FreeSeats
	}
	return total
}

// GetMinPrice returns the minimum price across all available tariffs
func (t *Train) GetMinPrice() int {
	minPrice := 0
	for _, car := range t.Cars {
		for _, tariff := range car.Tariffs {
			if tariff.FreeSeats > 0 && (minPrice == 0 || tariff.Tariff < minPrice) {
				minPrice = tariff.Tariff
			}
		}
	}
	return minPrice
}
