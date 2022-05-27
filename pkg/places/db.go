package places

import (
	"crypto/md5"
	"encoding/json"
	"fmt"

	"github.com/kennygrant/sanitize"
	"github.com/sdomino/scribble"
)

const collection = "places"

func WritePlacesToDB(db *scribble.Driver, ps []*Place) error {
	for _, p := range ps {
		res := ResourceName(p)
		err := db.Write(collection, res, p)
		if err != nil {
			return err
		}
	}
	return nil
}

func ReadPlacesFromDB(db *scribble.Driver) ([]*Place, error) {
	records, err := db.ReadAll(collection)
	if err != nil {
		return nil, err
	}

	ps := []*Place{}
	for _, r := range records {
		p := &Place{}
		if err := json.Unmarshal([]byte(r), p); err != nil {
			return nil, err
		}
		ps = append(ps, p)
	}

	return ps, nil
}

func ResourceName(p *Place) string {
	res := sanitize.BaseName(fmt.Sprintf("%s_%s", p.Name, p.FormattedAddress))
	ret := md5.Sum([]byte(res))
	return fmt.Sprintf("%x", ret)
}
