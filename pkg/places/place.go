package places

import (
	"time"

	"googlemaps.github.io/maps"
)

// Place holds info about a particular place
type Place struct {
	Name string
	// the full address
	FormattedAddress string
	Geometry         maps.AddressGeometry
	GooglePlaceID    string
	Categories       []string
	Tags             []string
	Published        time.Time
	Updated          time.Time
}
