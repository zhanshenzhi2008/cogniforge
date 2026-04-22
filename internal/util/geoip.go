package util

import (
	"fmt"
	"net/netip"
	"path/filepath"
	"sync"

	"github.com/oschwald/geoip2-golang/v2"
)

type GeoIP struct {
	db *geoip2.Reader
}

var (
	instance *GeoIP
	once     sync.Once
)

// InitGeoIP initializes the GeoIP database from the given path.
// Returns an error if the database cannot be opened.
func InitGeoIP(dbPath string) error {
	var initErr error
	once.Do(func() {
		db, err := geoip2.Open(dbPath)
		if err != nil {
			initErr = fmt.Errorf("failed to open GeoIP database: %w", err)
			return
		}
		instance = &GeoIP{db: db}
	})
	return initErr
}

// GetLocation returns the location string for an IP address.
// Returns empty string if the IP cannot be found or is invalid.
func GetLocation(ipString string) string {
	if instance == nil || instance.db == nil {
		return ""
	}

	ip, err := netip.ParseAddr(ipString)
	if err != nil {
		return ""
	}

	record, err := instance.db.City(ip)
	if err != nil {
		return ""
	}

	if !record.HasData() {
		return ""
	}

	// Build location string: City, Country
	location := ""
	if record.City.Names.SimplifiedChinese != "" {
		location = record.City.Names.SimplifiedChinese
	} else if record.City.Names.English != "" {
		location = record.City.Names.English
	}

	if record.Country.Names.SimplifiedChinese != "" {
		if location != "" {
			location += ", " + record.Country.Names.SimplifiedChinese
		} else {
			location = record.Country.Names.SimplifiedChinese
		}
	} else if record.Country.Names.English != "" {
		if location != "" {
			location += ", " + record.Country.Names.English
		} else {
			location = record.Country.Names.English
		}
	}

	return location
}

// Close closes the GeoIP database connection.
func Close() {
	if instance != nil && instance.db != nil {
		instance.db.Close()
	}
}

// DefaultGeoIPPath returns the default path for the GeoLite2 database.
func DefaultGeoIPPath() string {
	return filepath.Join("data", "geoip", "GeoLite2-City.mmdb")
}
