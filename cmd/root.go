package cmd

import (
	"fmt"
	"os"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	// Used for flags.
	cfgFile string
	dirPath string
	mapsKey string

	rootCmd = &cobra.Command{
		Use:   "neptune",
		Short: "Store your places to help you navigate where to go",
		Long:  `Neptune lets you import places from google exports in both geojson and csv formats.`,
	}
)

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.neptune.yaml)")
	rootCmd.PersistentFlags().StringVar(&dirPath, "dir", "data/out/", "directory for storing places (default is data/out/)")
	rootCmd.PersistentFlags().StringVar(&mapsKey, "mapsKey", "", "The API key to access the Places API")

}

func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".neptune" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".neptune")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
