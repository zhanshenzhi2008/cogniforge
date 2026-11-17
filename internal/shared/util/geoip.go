package util

import (
	"fmt"
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

// DefaultGeoIPPath returns the default path for the GeoLite2 database.
func DefaultGeoIPPath() string {
	return filepath.Join("data", "geoip", "GeoLite2-City.mmdb")
}
