package places

import (
	"context"
	"fmt"
	"sort"
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

// AddPlaceDetails will look up the Place on the Google Places API to fill in details.
// It will not look it up if the Place already has the GooglePlaceID set.
func AddPlaceDetails(ctx context.Context, apiKey string, places []*Place) error {
	c, err := maps.NewClient(maps.WithAPIKey(apiKey))
	if err != nil {
		return err
	}

	fieldMask := []maps.PlaceSearchFieldMask{
		maps.PlaceSearchFieldMaskName,
		maps.PlaceSearchFieldMaskPlaceID,
		maps.PlaceSearchFieldMaskFormattedAddress,
		maps.PlaceSearchFieldMaskGeometry,
		maps.PlaceSearchFieldMaskTypes,
	}
	for _, p := range places {
		if p.GooglePlaceID != "" {
			// skip places that are already filled in (saves an API call)
			continue
		}

		title, err := getSearchQuery(p)
		if err != nil {
			return err
		}

		resp, err := c.FindPlaceFromText(ctx, &maps.FindPlaceFromTextRequest{
			Input:              title,
			InputType:          maps.FindPlaceFromTextInputTypeTextQuery,
			LocationBias:       maps.FindPlaceFromTextLocationBiasCircular,
			LocationBiasRadius: 20,
			LocationBiasCenter: &p.Geometry.Location,
			Fields:             fieldMask,
		})
		if err != nil {
			return err
		}

		if len(resp.Candidates) > 1 {
			return fmt.Errorf("found more than one place for %v", resp.Candidates)
		}

		err = p.fromPlacesSearchResult(resp.Candidates[0])
		if err != nil {
			return err
		}
	}

	return nil
}

// AddPlaceDetailsForID will look up the Place on the Google Places API using the GooglePlaceID to fill in details.
func AddPlaceDetailsForID(ctx context.Context, apiKey string, places []*Place) error {
	c, err := maps.NewClient(maps.WithAPIKey(apiKey))
	if err != nil {
		return err
	}

	fieldMask := []maps.PlaceDetailsFieldMask{
		maps.PlaceDetailsFieldMaskName,
		maps.PlaceDetailsFieldMaskPlaceID,
		maps.PlaceDetailsFieldMaskFormattedAddress,
		maps.PlaceDetailsFieldMaskGeometry,
		maps.PlaceDetailsFieldMaskTypes,
	}
	for _, p := range places {

		resp, err := c.PlaceDetails(ctx, &maps.PlaceDetailsRequest{
			PlaceID: p.GooglePlaceID,
			Fields:  fieldMask,
		})
		if err != nil {
			return err
		}

		err = p.fromPlaceDetailsResult(resp)
		if err != nil {
			return err
		}
	}

	return nil
}

// fromPlacesSearchResult will fill in p with the values from s
// it's assuming you already knew the name and address
func (p *Place) fromPlacesSearchResult(s maps.PlacesSearchResult) error {
	// don't change the name or formatted address because these are used for the db ID
	p.GooglePlaceID = s.PlaceID
	// overwrite the geometry from places API since it has more info
	p.Geometry = s.Geometry
	p.Categories = s.Types

	return nil
}

// fromPlacesSearchResult will fill in p with the values from s
// it's assuming you already knew the PlaceID
func (p *Place) fromPlaceDetailsResult(s maps.PlaceDetailsResult) error {
	p.FormattedAddress = s.FormattedAddress
	// overwrite the geometry from places API since it has more info
	p.Geometry = s.Geometry
	p.Categories = s.Types

	return nil
}

func getSearchQuery(p *Place) (string, error) {
	name := p.Name
	if name == "" {
		return "", fmt.Errorf("place must have a name")
	}

	addr := p.FormattedAddress
	if addr == "" {
		return "", fmt.Errorf("place must have a formatted address")
	}

	return fmt.Sprintf("%s %s", name, addr), nil
}

func DistanceToPlaceFrom(ctx context.Context, c maps.Client, f string, p *Place) (int, error) {
	req := maps.DirectionsRequest{Origin: f, Destination: p.FormattedAddress}
	route, _, err := c.Directions(ctx, &req)
	if err != nil {
		return 0, err
	}

	distance := 0
	if len(route) == 0 {
		// If no route can be found just assume 0 distance
		return distance, nil
	}
	for _, leg := range route[0].Legs {
		distance += leg.Meters
	}

	return distance, nil
}

// DistinctCategories returns a sorted list of distinct categories
func DistinctCategories(ps []*Place) (cats []string) {
	for _, p := range ps {
		cat := p.Categories[0]
		if !contains(cats, cat) {
			cats = append(cats, cat)
		}
	}

	sort.Strings(cats)
	return cats
}

// contains checks if s contains str
func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}
