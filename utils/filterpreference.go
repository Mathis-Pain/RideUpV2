package utils

import (
	"RideUP/models"
	"math"
)

// distanceKm calcule la distance entre deux points GPS en kilomètres
func distanceKm(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371 // rayon de la Terre en km
	dLat := (lat2 - lat1) * math.Pi / 180
	dLon := (lon2 - lon1) * math.Pi / 180

	lat1Rad := lat1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Sin(dLon/2)*math.Sin(dLon/2)*math.Cos(lat1Rad)*math.Cos(lat2Rad)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return R * c
}

// FilterPreference filtre les événements selon la préférence (rayon en km) de l'utilisateur
func FilterPreference(events []models.Event, userLat, userLon float64, radiusKm int) []models.Event {
	var filtered []models.Event

	for _, e := range events {
		d := distanceKm(userLat, userLon, e.Latitude, e.Longitude)
		if d <= float64(radiusKm) {
			filtered = append(filtered, e)
		}
	}

	return filtered
}
