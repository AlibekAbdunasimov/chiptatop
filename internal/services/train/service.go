package train

import (
	"context"
	"fmt"
	"log"
	"strings"
)

// Service provides train ticket search and monitoring functionality
type Service struct {
	client   *Client
	stations map[string]Station // Cache for station lookup
}

// NewService creates a new train service with default language (Uzbek)
func NewService() *Service {
	return NewServiceWithLanguage(LanguageUzbek)
}

// NewServiceWithLanguage creates a new train service with specified language
func NewServiceWithLanguage(language string) *Service {
	return &Service{
		client:   NewClient(language),
		stations: make(map[string]Station),
	}
}

// SetAuthCredentials sets authentication credentials for API requests
func (s *Service) SetAuthCredentials(xsrfToken, cookies string) {
	s.client.SetAuthHeaders(xsrfToken, cookies)
}

// InitializeCredentials automatically obtains fresh Railway.uz API credentials
func (s *Service) InitializeCredentials(ctx context.Context) error {
	return s.client.InitializeCredentials(ctx)
}

// SetLanguage changes the language for API requests
func (s *Service) SetLanguage(language string) {
	s.client.SetLanguage(language)
}

// GetLanguage returns the current language setting
func (s *Service) GetLanguage() string {
	return s.client.GetLanguage()
}

// SearchTrains searches for available trains between stations
func (s *Service) SearchTrains(ctx context.Context, params TrainSearchParams) (*SearchTrainsResponse, error) {
	// Convert user-friendly params to API request format
	req := &SearchTrainsRequest{
		Directions: Directions{
			Forward: &Journey{
				Date:           params.Date.Format("2006-01-02"),
				DepStationCode: s.GetStationCode(params.From),
				ArvStationCode: s.GetStationCode(params.To),
			},
		},
	}

	log.Printf("Searching trains from %s to %s on %s", params.From, params.To, params.Date.Format("2006-01-02"))

	response, err := s.client.SearchTrains(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to search trains: %w", err)
	}

	if response.Error != nil {
		return nil, fmt.Errorf("API error: %s - %s", response.Error.Code, response.Error.Message)
	}

	if response.Data == nil {
		return nil, fmt.Errorf("no data received from API")
	}

	return response, nil
}

// FindAvailableTrains returns only trains with available seats
func (s *Service) FindAvailableTrains(ctx context.Context, params TrainSearchParams) ([]Train, error) {
	response, err := s.SearchTrains(ctx, params)
	if err != nil {
		return nil, err
	}

	if response.Data == nil || response.Data.Directions.Forward == nil {
		return []Train{}, nil
	}

	var availableTrains []Train
	for _, train := range response.Data.Directions.Forward.Trains {
		if train.HasAvailableSeats() {
			availableTrains = append(availableTrains, train)
		}
	}

	return availableTrains, nil
}

// CheckTicketAvailability checks if tickets are available for the given alert criteria
func (s *Service) CheckTicketAvailability(ctx context.Context, alert TicketAlert) ([]Train, error) {
	params := TrainSearchParams{
		From: alert.From,
		To:   alert.To,
		Date: alert.Date,
	}

	trains, err := s.FindAvailableTrains(ctx, params)
	if err != nil {
		return nil, err
	}

	var matchingTrains []Train
	for _, train := range trains {
		if s.matchesAlertCriteria(train, alert) {
			matchingTrains = append(matchingTrains, train)
		}
	}

	return matchingTrains, nil
}

// FormatTrainInfo formats train information for display
func (s *Service) FormatTrainInfo(train Train) string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("🚂 *%s* (%s)\n", train.Brand, train.Number))
	builder.WriteString(fmt.Sprintf("📍 %s → %s\n", train.SubRoute.DepStationName, train.SubRoute.ArvStationName))
	builder.WriteString(fmt.Sprintf("🕐 %s - %s (%s)\n", train.GetDepartureTime(), train.GetArrivalTime(), train.TimeOnWay))
	builder.WriteString(fmt.Sprintf("📅 %s\n", train.GetDate()))
	builder.WriteString(fmt.Sprintf("🚄 Route: %s → %s\n", train.OriginRoute.DepStationName, train.OriginRoute.ArvStationName))

	if len(train.Cars) > 0 {
		builder.WriteString("\n💺 *Seat types and prices:*\n")
		for _, car := range train.Cars {
			// Show car type with total seats and price
			if len(car.Tariffs) > 0 {
				// Use the first tariff price as representative for this car type
				price := s.formatPrice(car.Tariffs[0].Tariff)
				builder.WriteString(fmt.Sprintf("*%s* (%d total seats): %s UZS\n",
					car.Type, car.FreeSeats, price))
			}
		}
	}

	return builder.String()
}

// FormatSearchResults formats multiple trains for display
func (s *Service) FormatSearchResults(trains []Train) string {
	if len(trains) == 0 {
		return "❌ No trains found for your search criteria."
	}

	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("🚂 *Found %d train(s):*\n\n", len(trains)))

	for i, train := range trains {
		builder.WriteString(s.FormatTrainInfo(train))
		if i < len(trains)-1 {
			builder.WriteString("\n" + strings.Repeat("─", 30) + "\n\n")
		}
	}

	return builder.String()
}

// GetStationCode returns the station code for a given station name or code
func (s *Service) GetStationCode(stationNameOrCode string) string {
	// Real station codes from Uzbekistan railways
	stationCodes := map[string]string{
		"andijon":   "2900680",
		"andijan":   "2900680", // Alternative spelling
		"buxoro":    "2900800",
		"bukhara":   "2900800", // Alternative spelling
		"guliston":  "2900850",
		"jizzax":    "2900720",
		"jizzakh":   "2900720", // Alternative spelling
		"margilon":  "2900920",
		"margilan":  "2900920", // Alternative spelling
		"namangan":  "2900940",
		"navoiy":    "2900930",
		"nukus":     "2900970",
		"pop":       "2900693",
		"qarshi":    "2900750",
		"qoqon":     "2900880",
		"kokand":    "2900880", // Alternative spelling
		"samarqand": "2900700",
		"samarkand": "2900700", // Alternative spelling
		"termiz":    "2900255",
		"termez":    "2900255", // Alternative spelling
		"toshkent":  "2900000",
		"tashkent":  "2900000", // Alternative spelling
		"urgench":   "2900790",
		"xiva":      "2900172",
		"khiva":     "2900172", // Alternative spelling
	}

	// If it's already a code (starts with numbers), return as is
	if len(stationNameOrCode) > 0 && stationNameOrCode[0] >= '0' && stationNameOrCode[0] <= '9' {
		return stationNameOrCode
	}

	// Try to find by name (case-insensitive)
	lowerName := strings.ToLower(strings.TrimSpace(stationNameOrCode))
	if code, exists := stationCodes[lowerName]; exists {
		return code
	}

	// If not found, return as is (might be a valid code we don't know about)
	return stationNameOrCode
}

// matchesAlertCriteria checks if a train matches the alert criteria
func (s *Service) matchesAlertCriteria(train Train, alert TicketAlert) bool {
	for _, car := range train.Cars {
		for _, tariff := range car.Tariffs {
			if tariff.FreeSeats == 0 {
				continue
			}

			// Check seat type preference
			if len(alert.SeatTypes) > 0 {
				found := false
				for _, preferredType := range alert.SeatTypes {
					if strings.EqualFold(car.Type, preferredType) ||
						strings.EqualFold(tariff.ClassServiceType, preferredType) {
						found = true
						break
					}
				}
				if !found {
					continue
				}
			}

			// Check price range (convert float64 to int for comparison)
			price := float64(tariff.Tariff)
			if alert.MinPrice > 0 && price < alert.MinPrice {
				continue
			}
			if alert.MaxPrice > 0 && price > alert.MaxPrice {
				continue
			}

			// If we reach here, this tariff matches the criteria
			return true
		}
	}

	return false
}

// formatPrice formats price with thousands separator
func (s *Service) formatPrice(price int) string {
	priceStr := fmt.Sprintf("%d", price)
	n := len(priceStr)
	if n <= 3 {
		return priceStr
	}

	var result strings.Builder
	for i, digit := range priceStr {
		if i > 0 && (n-i)%3 == 0 {
			result.WriteString(" ")
		}
		result.WriteRune(digit)
	}
	return result.String()
}

// GetStationSuggestions returns station name suggestions for autocomplete
func (s *Service) GetStationSuggestions(query string) []string {
	stations := []string{
		"Andijon", "Buxoro", "Guliston", "Jizzax", "Margilon",
		"Namangan", "Navoiy", "Nukus", "Pop", "Qarshi",
		"Qoqon", "Samarqand", "Termiz", "Toshkent", "Urgench", "Xiva",
	}

	if query == "" {
		return stations
	}

	var suggestions []string
	lowerQuery := strings.ToLower(query)

	for _, station := range stations {
		if strings.Contains(strings.ToLower(station), lowerQuery) {
			suggestions = append(suggestions, station)
		}
	}

	return suggestions
}
