package places

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"regexp"

	"github.com/jszwec/csvutil"
)

// FromSavedCSV reads in csv formatted bytes and returns them in a standardized Place
// The returned Place will be missing some fields
func FromSavedCSV(bs []byte) ([]*Place, error) {
	var in []savedplace
	if err := csvutil.Unmarshal(bs, &in); err != nil {
		return nil, err
	}

	ps := make([]*Place, 0, len(in))
	for _, sp := range in {
		id, err := lookupPlaceID(sp)
		if err != nil {
			return nil, err
		}

		p := &Place{
			Name:          sp.Title,
			GooglePlaceID: id,
		}
		ps = append(ps, p)
	}
	return ps, nil
}

func lookupPlaceID(sp savedplace) (string, error) {
	name := sp.Title
	url := sp.URL
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	reader := bufio.NewReader(resp.Body)
	var line string
	for {
		line, err = reader.ReadString('\n')
		if err != nil {
			break
		}

		// turns out this was too limiting for names with special characters
		// leaving here because would like to filter somehow
		// if !strings.Contains(line, name) {
		// 	// ignore lines that don't have the name of the place
		// 	continue
		// }

		found := id.FindStringSubmatch(line)
		if len(found) > 1 {
			return found[1], nil
		}
	}

	if err != io.EOF {
		return "", err
	}

	return "", fmt.Errorf("didn't find id for %s", name)
}

// pretty sure this is a big assumption
// there is at least one other format https://developers.google.com/places/web-service/place-id#id-overview
var id = regexp.MustCompile(`\"(ChI[\w-]+)\\`)

type savedplace struct {
	Title   string `csv:"Title"`
	Note    string `csv:"Note"`
	URL     string `csv:"URL"`
	Comment string `csv:"Comment"`
}
