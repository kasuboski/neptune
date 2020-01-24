package places

import (
	"context"
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

// ImportFromSavedJSON reads in geojson formatted bytes and returns them in a standardized Place
func ImportFromSavedJSON(apiKey string, bs []byte) ([]*Place, error) {
	fc, err := geojson.UnmarshalFeatureCollection(bs)
	if err != nil {
		return nil, err
	}

	c, err := maps.NewClient(maps.WithAPIKey(apiKey))
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	places := make([]*Place, 0, len(fc.Features))
	fieldMask := []maps.PlaceSearchFieldMask{
		maps.PlaceSearchFieldMaskName,
		maps.PlaceSearchFieldMaskPlaceID,
		maps.PlaceSearchFieldMaskFormattedAddress,
		maps.PlaceSearchFieldMaskGeometry,
		maps.PlaceSearchFieldMaskTypes,
	}
	for _, feature := range fc.Features {
		title, err := getSearchQuery(feature)
		if err != nil {
			return nil, err
		}

		resp, err := c.FindPlaceFromText(ctx, &maps.FindPlaceFromTextRequest{
			Input:              title,
			InputType:          maps.FindPlaceFromTextInputTypeTextQuery,
			LocationBias:       maps.FindPlaceFromTextLocationBiasCircular,
			LocationBiasRadius: 20,
			LocationBiasCenter: &maps.LatLng{
				Lat: feature.Geometry.Point[1],
				Lng: feature.Geometry.Point[0],
			},
			Fields: fieldMask,
		})
		if err != nil {
			return nil, err
		}

		if len(resp.Candidates) > 1 {
			return nil, fmt.Errorf("found more than one place for %v", resp.Candidates)
		}

		place := &Place{}
		err = place.FromPlacesSearchResult(resp.Candidates[0])
		if err != nil {
			return nil, err
		}

		places = append(places, place)
	}

	return places, nil
}

// FromPlacesSearchResult will fill in p with the values from s
func (p *Place) FromPlacesSearchResult(s maps.PlacesSearchResult) error {
	p.Name = s.Name
	p.FormattedAddress = s.FormattedAddress
	p.GooglePlaceID = s.PlaceID
	p.Geometry = s.Geometry
	p.Categories = s.Types

	return nil
}

func getSearchQuery(f *geojson.Feature) (string, error) {
	title, err := f.PropertyString("Title")
	if err != nil {
		return "", err
	}

	location, ok := f.Properties["Location"]
	if !ok {
		return title, nil
	}

	lmap, ok := location.(map[string]interface{})
	if !ok {
		return title, nil
	}

	address, ok := lmap["Address"]
	if !ok {
		return title, nil
	}

	addrStr, ok := address.(string)
	if !ok {
		return title, nil
	}

	return fmt.Sprintf("%s %s", title, addrStr), nil
}
