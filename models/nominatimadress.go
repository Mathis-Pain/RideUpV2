package models

type NominatimAddress struct {
	HouseNumber string `json:"house_number"`
	Road        string `json:"road"`
	Postcode    string `json:"postcode"`
	City        string `json:"city"`
	Country     string `json:"country"`
}
