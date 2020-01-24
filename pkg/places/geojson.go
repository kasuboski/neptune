package places

import (
	"fmt"
	"time"

	geojson "github.com/paulmach/go.geojson"
	"googlemaps.github.io/maps"
)

// FromSavedJSON reads in geojson formatted bytes and returns them in a standardized Place
// The returned Place will be missing some fields
func FromSavedJSON(bs []byte) ([]*Place, error) {
	fc, err := geojson.UnmarshalFeatureCollection(bs)
	if err != nil {
		return nil, err
	}

	places := make([]*Place, 0, len(fc.Features))
	for _, f := range fc.Features {
		title, err := f.PropertyString("Title")
		if err != nil {
			return nil, err
		}

		location, ok := f.Properties["Location"]
		if !ok {
			return nil, fmt.Errorf("couldn't find Location for %s", title)
		}

		lmap, ok := location.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("couldn't parse Location for %s", title)
		}

		address, ok := lmap["Address"]
		if !ok {
			return nil, fmt.Errorf("couldn't find Address for %s", title)
		}

		addrStr, ok := address.(string)
		if !ok {
			return nil, fmt.Errorf("couldn't parse Address for %s", title)
		}

		pStr, err := f.PropertyString("Published")
		if err != nil {
			return nil, err
		}

		pt, err := time.Parse(time.RFC3339, pStr)
		if err != nil {
			return nil, err
		}

		uStr, err := f.PropertyString("Updated")
		if err != nil {
			return nil, err
		}

		ut, err := time.Parse(time.RFC3339, uStr)
		if err != nil {
			return nil, err
		}

		place := &Place{
			Name:             title,
			FormattedAddress: addrStr,
			Geometry: maps.AddressGeometry{
				Location: maps.LatLng{
					Lat: f.Geometry.Point[1],
					Lng: f.Geometry.Point[0],
				},
			},
			Published: pt,
			Updated:   ut,
		}
		places = append(places, place)
	}

	return places, nil
}
