package cmd

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/kasuboski/neptune/pkg/places"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/kennygrant/sanitize"
	"github.com/sdomino/scribble"
)

const collection = "places"

var importCmd = &cobra.Command{
	Use:   "import <geojson|csv> -f file",
	Short: "import either a geojson or places csv file",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		format := args[0]
		if format != "geojson" && format != "csv" {
			log.Fatal("one of geojson or csv needs to be specified")
		}

		filePath := viper.GetString("file")

		bs, err := ioutil.ReadFile(filePath)
		if err != nil {
			log.Fatal(err)
		}

		ctx := context.Background()
		apiKey := viper.GetString("mapsKey")
		if apiKey == "" {
			log.Fatal("mapsKey is required")
		}

		var imported []*places.Place
		if format == "geojson" {
			imported, err = places.FromSavedJSON(bs)
			if err != nil {
				log.Fatal(err)
			}
		} else if format == "csv" {
			imported, err = places.FromSavedCSV(bs)
			if err != nil {
				log.Fatal(err)
			}

			err = places.AddPlaceDetailsForID(ctx, apiKey, imported)
			if err != nil {
				log.Fatal(err)
			}
		}

		tags, err := cmd.Flags().GetStringSlice("tags")
		if err == nil {
			for _, p := range imported {
				p.Tags = tags
			}
		}

		dir := "data/out/"
		db, err := scribble.New(dir, nil)
		if err != nil {
			log.Fatal(err)
		}

		// check if places already in db
		// if not add it
		for _, p := range imported {
			found := &places.Place{}
			n := resourceName(p)
			if err := db.Read(collection, n, found); os.IsNotExist(err) {
				err := db.Write(collection, n, p)
				if err != nil {
					log.Fatal(err)
				}
			} else {
				if err == nil {
					// found one without issue
					// add tags from imported p
					found.Tags = append(found.Tags, p.Tags...)
					db.Write(collection, n, found)
				}
			}
		}

		ps, err := readPlacesFromDB(db)
		if err != nil {
			log.Fatal(err)
		}

		err = places.AddPlaceDetails(ctx, apiKey, ps)
		if err != nil {
			log.Fatal(err)
		}

		err = writePlacesToDB(db, ps)
		if err != nil {
			log.Fatal(err)
		}
	},
}

func writePlacesToDB(db *scribble.Driver, ps []*places.Place) error {
	for _, p := range ps {
		res := resourceName(p)
		err := db.Write(collection, res, p)
		if err != nil {
			return err
		}
	}
	return nil
}

func readPlacesFromDB(db *scribble.Driver) ([]*places.Place, error) {
	records, err := db.ReadAll(collection)
	if err != nil {
		return nil, err
	}

	ps := []*places.Place{}
	for _, r := range records {
		p := &places.Place{}
		if err := json.Unmarshal([]byte(r), p); err != nil {
			return nil, err
		}
		ps = append(ps, p)
	}

	return ps, nil
}

func resourceName(p *places.Place) string {
	res := sanitize.BaseName(fmt.Sprintf("%s_%s", p.Name, p.FormattedAddress))
	ret := md5.Sum([]byte(res))
	return fmt.Sprintf("%x", ret)
}

func init() {
	importCmd.Flags().StringP("file", "f", "", "file to import")
	importCmd.MarkFlagRequired("file")

	importCmd.Flags().String("mapsKey", "", "The API key to access the Places API")
	importCmd.Flags().StringSliceP("tags", "t", []string{}, "tags to add to these places")

	if err := viper.BindPFlags(importCmd.Flags()); err != nil {
		log.Fatal(err)
	}

	rootCmd.AddCommand(importCmd)
}
