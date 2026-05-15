package utils

import (
	"log"
	"net"

	"github.com/oschwald/geoip2-golang"
)

var geoipDB *geoip2.Reader

// InitGeoIP initializes the MaxMind GeoIP2 reader.
func InitGeoIP(dbPath string) error {
	var err error
	geoipDB, err = geoip2.Open(dbPath)
	if err != nil {
		log.Printf("Warning: failed to open GeoIP database at %s: %v", dbPath, err)
		return err
	}
	log.Printf("GeoIP database loaded successfully from %s", dbPath)
	return nil
}

// CloseGeoIP closes the GeoIP database.
func CloseGeoIP() {
	if geoipDB != nil {
		geoipDB.Close()
	}
}

// GeoLocation contains geographical information for an IP address.
type GeoLocation struct {
	Country   string
	Region    string
	City      string
	Latitude  float64
	Longitude float64
}

// LookupIP queries the GeoIP database for the given IP string.
func LookupIP(ipStr string) *GeoLocation {
	if geoipDB == nil || ipStr == "" || ipStr == "-" {
		return nil
	}

	ip := net.ParseIP(ipStr)
	if ip == nil {
		return nil
	}

	record, err := geoipDB.City(ip)
	if err != nil {
		return nil
	}

	geo := &GeoLocation{
		Latitude:  record.Location.Latitude,
		Longitude: record.Location.Longitude,
	}

	if record.Country.Names != nil {
		geo.Country = record.Country.Names["zh-CN"]
		if geo.Country == "" {
			geo.Country = record.Country.Names["en"]
		}
	}

	if len(record.Subdivisions) > 0 && record.Subdivisions[0].Names != nil {
		geo.Region = record.Subdivisions[0].Names["zh-CN"]
		if geo.Region == "" {
			geo.Region = record.Subdivisions[0].Names["en"]
		}
	}

	if record.City.Names != nil {
		geo.City = record.City.Names["zh-CN"]
		if geo.City == "" {
			geo.City = record.City.Names["en"]
		}
	}

	return geo
}
