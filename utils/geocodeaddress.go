package utils

import (
	"RideUP/models"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

func GeocodeAddress(address string) (float64, float64, error) {
	baseURL := "https://nominatim.openstreetmap.org/search"
	params := url.Values{}
	params.Set("q", address)
	params.Set("format", "json")
	params.Set("limit", "1")

	// Création d'un client HTTP personnalisé
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Création de la requête
	req, err := http.NewRequest("GET", baseURL+"?"+params.Encode(), nil)
	if err != nil {
		return 0, 0, err
	}

	// CRITIQUE : Ajout du User-Agent requis par Nominatim
	req.Header.Set("User-Agent", "RideUP/1.0 (ride-up-app)")

	// Exécution de la requête
	resp, err := client.Do(req)
	if err != nil {
		return 0, 0, fmt.Errorf("erreur requête nominatim: %v", err)
	}
	defer resp.Body.Close()

	// Vérification du status code
	if resp.StatusCode != http.StatusOK {
		return 0, 0, fmt.Errorf("nominatim status: %d", resp.StatusCode)
	}

	var result models.NominatimResponseCoord
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, 0, fmt.Errorf("erreur décodage JSON: %v", err)
	}

	if len(result) == 0 {
		return 0, 0, fmt.Errorf("adresse introuvable")
	}

	lat, err := strconv.ParseFloat(result[0].Lat, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("erreur parsing latitude: %v", err)
	}

	lon, err := strconv.ParseFloat(result[0].Lon, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("erreur parsing longitude: %v", err)
	}

	log.Printf("Géocodage réussi: %s -> [%f, %f]", address, lat, lon)
	return lat, lon, nil
}
