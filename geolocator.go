package main

import (
	"encoding/json"
	"net/http"
)

const (
	mainGeolocatorURL   = "http://ip-api.com/json/"
	backupGeolocatorURL = "http://api.ipstack.com/"
)

// Geolocator defines the interface for an ip geolocator provider
type Geolocator interface {
	Geolocate(ip string) (string, error)
}

type geolocationProvider struct {
	Geolocators []Geolocator
}

// NewGeolocationProvider represents a geolocator aggregator
func NewGeolocationProvider(geos ...Geolocator) Geolocator {
	return geolocationProvider{Geolocators: geos}
}

func (g geolocationProvider) Geolocate(ip string) (string, error) {
	var country string
	var err error

	for _, locator := range g.Geolocators {
		country, err = locator.Geolocate(ip)
		if err == nil && ip != "" {
			return country, nil
		}
	}

	return country, err
}

type geolocateOption struct {
	Locator Geolocator
}

func (g geolocateOption) Apply(e FailedConnEvent) (FailedConnEvent, error) {
	country, err := g.Locator.Geolocate(e.IPAddress.String())
	if err != nil {
		return e, err
	}
	e.Country = country
	return e, nil
}

type ipAPI struct{}

type apiStack struct {
	AccessKey string
}

type ipAPIResponse struct {
	Country string `json:"country"`
}

func (c ipAPI) Geolocate(ip string) (string, error) {
	resp, err := http.Get(mainGeolocatorURL + ip)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	response := ipAPIResponse{}

	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return "", err
	}

	return response.Country, nil
}

type apiStackResponse struct {
	CountryName string `json:"country_name"`
}

func (c apiStack) Geolocate(ip string) (string, error) {
	resp, err := http.Get(backupGeolocatorURL + ip + "?access_key=" + c.AccessKey)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	response := apiStackResponse{}

	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return "", err
	}

	return response.CountryName, nil
}
