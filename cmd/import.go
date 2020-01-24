package cmd

import (
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/kasuboski/neptune/pkg/places"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/kennygrant/sanitize"
	"github.com/sdomino/scribble"
)

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

		apiKey := viper.GetString("mapsKey")
		if apiKey == "" {
			log.Fatal("mapsKey is required")
		}
		// places, err := places.ImportFromSavedJSON(apiKey, bs)
		places, err := places.FromSavedJSON(bs)
		if err != nil {
			log.Fatal(err)
		}

		// pretty.Println(places)
		dir := "data/out/"
		db, err := scribble.New(dir, nil)
		if err != nil {
			log.Fatal(err)
		}

		for _, p := range places {
			res := resourceName(p)
			err := db.Write("places", res, p)
			if err != nil {
				log.Fatal(err)
			}
		}
	},
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

	if err := viper.BindPFlags(importCmd.Flags()); err != nil {
		log.Fatal(err)
	}

	rootCmd.AddCommand(importCmd)
}
