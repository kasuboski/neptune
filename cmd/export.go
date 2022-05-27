package cmd

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/kasuboski/neptune/pkg/places"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"googlemaps.github.io/maps"

	"github.com/sdomino/scribble"
)

var exportCmd = &cobra.Command{
	Use:   "export <text>",
	Short: "export a text file of all places",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		format := args[0]
		if format != "text" {
			log.Fatal("please specify export format")
		}

		db, err := scribble.New(dirPath, nil)
		if err != nil {
			log.Fatal(err)
		}

		ps, err := places.ReadPlacesFromDB(db)
		if err != nil {
			log.Fatal(err)
		}

		location := viper.GetString("location")
		distance := viper.GetInt("distance") * 1000 // Convert km into meters
		if location != "" {
			ctx := context.Background()
			if mapsKey == "" {
				log.Fatal("mapsKey is required")
			}

			c, err := maps.NewClient(maps.WithAPIKey(mapsKey))
			if err != nil {
				log.Fatal(err)
			}

			// Remove all places if distance is grater then limit
			// TODO: Make this happen in parallel
			for i := len(ps) - 1; i >= 0; i-- {
				p := ps[i]
				d, err := places.DistanceToPlaceFrom(ctx, *c, location, p)
				if err != nil {
					log.Fatal(err)
				}
				if d > distance {
					ps = append(ps[:i], ps[i+1:]...)
				}
			}
		}

		// Export all remaining places to text category
		categories := places.DistinctCategories(ps)
		for _, cat := range categories {
			c := cases.Title(language.English)
			fmt.Printf("# %s\n", c.String(strings.Replace(cat, "_", " ", -1)))
			for _, p := range ps {
				if cat == p.Categories[0] {
					fmt.Printf("* %s\n", p.Name)
				}
			}
			fmt.Println("")
		}
	},
}

func init() {
	// TODO: Implement tags flag
	exportCmd.Flags().StringSliceP("tags", "t", []string{}, "tags to export")
	exportCmd.Flags().StringP("location", "l", "", "location to export places from")
	exportCmd.Flags().IntP("distance", "d", 200, "distance from location in km")

	if err := viper.BindPFlags(exportCmd.Flags()); err != nil {
		log.Fatal(err)
	}

	rootCmd.AddCommand(exportCmd)
}
