package train

import "strings"

// Station represents a railway station with its details
type StationInfo struct {
	Code        string   `json:"code"`
	Name        string   `json:"name"`
	NameUz      string   `json:"nameUz"`      // Uzbek name
	NameEn      string   `json:"nameEn"`      // English name
	Aliases     []string `json:"aliases"`     // Alternative spellings
	Region      string   `json:"region"`      // Region/Province
	IsActive    bool     `json:"isActive"`    // Whether station is active
	IsMajor     bool     `json:"isMajor"`     // Major station flag
	Coordinates *LatLng  `json:"coordinates"` // GPS coordinates (optional)
}

// LatLng represents GPS coordinates
type LatLng struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// GetAllStations returns all known railway stations in Uzbekistan
func GetAllStations() []StationInfo {
	return []StationInfo{
		{
			Code:     "2900680",
			Name:     "Andijon",
			NameUz:   "Andijon",
			NameEn:   "Andijan",
			Aliases:  []string{"andijan", "andizhan"},
			Region:   "Andijon",
			IsActive: true,
			IsMajor:  true,
		},
		{
			Code:     "2900800",
			Name:     "Buxoro",
			NameUz:   "Buxoro",
			NameEn:   "Bukhara",
			Aliases:  []string{"bukhara", "bokhara"},
			Region:   "Buxoro",
			IsActive: true,
			IsMajor:  true,
		},
		{
			Code:     "2900850",
			Name:     "Guliston",
			NameUz:   "Guliston",
			NameEn:   "Gulistan",
			Aliases:  []string{"gulistan"},
			Region:   "Sirdaryo",
			IsActive: true,
			IsMajor:  false,
		},
		{
			Code:     "2900720",
			Name:     "Jizzax",
			NameUz:   "Jizzax",
			NameEn:   "Jizzakh",
			Aliases:  []string{"jizzakh", "djizak"},
			Region:   "Jizzax",
			IsActive: true,
			IsMajor:  true,
		},
		{
			Code:     "2900920",
			Name:     "Margilon",
			NameUz:   "Margilon",
			NameEn:   "Margilan",
			Aliases:  []string{"margilan", "marghilan"},
			Region:   "Farg'ona",
			IsActive: true,
			IsMajor:  true,
		},
		{
			Code:     "2900940",
			Name:     "Namangan",
			NameUz:   "Namangan",
			NameEn:   "Namangan",
			Aliases:  []string{},
			Region:   "Namangan",
			IsActive: true,
			IsMajor:  true,
		},
		{
			Code:     "2900930",
			Name:     "Navoiy",
			NameUz:   "Navoiy",
			NameEn:   "Navoi",
			Aliases:  []string{"navoi", "navoiy"},
			Region:   "Navoiy",
			IsActive: true,
			IsMajor:  true,
		},
		{
			Code:     "2900970",
			Name:     "Nukus",
			NameUz:   "Nukus",
			NameEn:   "Nukus",
			Aliases:  []string{},
			Region:   "Qoraqalpog'iston",
			IsActive: true,
			IsMajor:  true,
		},
		{
			Code:     "2900693",
			Name:     "Pop",
			NameUz:   "Pop",
			NameEn:   "Pop",
			Aliases:  []string{},
			Region:   "Namangan",
			IsActive: true,
			IsMajor:  false,
		},
		{
			Code:     "2900750",
			Name:     "Qarshi",
			NameUz:   "Qarshi",
			NameEn:   "Karshi",
			Aliases:  []string{"karshi", "qarshi"},
			Region:   "Qashqadaryo",
			IsActive: true,
			IsMajor:  true,
		},
		{
			Code:     "2900880",
			Name:     "Qo'qon",
			NameUz:   "Qo'qon",
			NameEn:   "Kokand",
			Aliases:  []string{"kokand", "qoqon", "kokhand"},
			Region:   "Farg'ona",
			IsActive: true,
			IsMajor:  true,
		},
		{
			Code:     "2900700",
			Name:     "Samarqand",
			NameUz:   "Samarqand",
			NameEn:   "Samarkand",
			Aliases:  []string{"samarkand", "samarqand"},
			Region:   "Samarqand",
			IsActive: true,
			IsMajor:  true,
		},
		{
			Code:     "2900255",
			Name:     "Termiz",
			NameUz:   "Termiz",
			NameEn:   "Termez",
			Aliases:  []string{"termez", "termiz"},
			Region:   "Surxondaryo",
			IsActive: true,
			IsMajor:  true,
		},
		{
			Code:     "2900000",
			Name:     "Toshkent",
			NameUz:   "Toshkent",
			NameEn:   "Tashkent",
			Aliases:  []string{"tashkent", "toshkent"},
			Region:   "Toshkent",
			IsActive: true,
			IsMajor:  true,
		},
		{
			Code:     "2900790",
			Name:     "Urgench",
			NameUz:   "Urganch",
			NameEn:   "Urgench",
			Aliases:  []string{"urganch", "urgench"},
			Region:   "Xorazm",
			IsActive: true,
			IsMajor:  true,
		},
		{
			Code:     "2900172",
			Name:     "Xiva",
			NameUz:   "Xiva",
			NameEn:   "Khiva",
			Aliases:  []string{"khiva", "xiva"},
			Region:   "Xorazm",
			IsActive: true,
			IsMajor:  true,
		},
	}
}

// GetStationByCode returns station information by code
func GetStationByCode(code string) *StationInfo {
	stations := GetAllStations()
	for _, station := range stations {
		if station.Code == code {
			return &station
		}
	}
	return nil
}

// GetStationByName returns station information by name (case-insensitive)
func GetStationByName(name string) *StationInfo {
	stations := GetAllStations()
	lowerName := strings.ToLower(name)

	for _, station := range stations {
		// Check main names
		if strings.ToLower(station.Name) == lowerName ||
			strings.ToLower(station.NameUz) == lowerName ||
			strings.ToLower(station.NameEn) == lowerName {
			return &station
		}

		// Check aliases
		for _, alias := range station.Aliases {
			if strings.ToLower(alias) == lowerName {
				return &station
			}
		}
	}
	return nil
}

// GetMajorStations returns only major railway stations
func GetMajorStations() []StationInfo {
	stations := GetAllStations()
	var majorStations []StationInfo

	for _, station := range stations {
		if station.IsMajor {
			majorStations = append(majorStations, station)
		}
	}

	return majorStations
}

// GetStationsByRegion returns stations in a specific region
func GetStationsByRegion(region string) []StationInfo {
	stations := GetAllStations()
	var regionStations []StationInfo

	lowerRegion := strings.ToLower(region)
	for _, station := range stations {
		if strings.ToLower(station.Region) == lowerRegion {
			regionStations = append(regionStations, station)
		}
	}

	return regionStations
}
