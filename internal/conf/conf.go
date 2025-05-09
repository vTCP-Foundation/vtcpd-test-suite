package conf

import (
	"log"

	"github.com/spf13/viper"
)

// ClusterSettings holds configuration for the test cluster.
type ClusterSettings struct {
	NodeImageName string `yaml:"nodeImageName"`
	NetworkName   string `yaml:"networkName"`
}

const (
	configFileName = "conf"
	configFileType = "yaml"
)

var (
	// effectiveConfig holds the final configuration after loading.
	effectiveConfig ClusterSettings
)

// init is automatically called once when the package is imported.
// It loads configurations using Viper with the following order of precedence (highest to lowest):
func init() {
	v := viper.New()

	// Configure to read from a configuration file
	v.SetConfigName(configFileName)
	v.SetConfigType(configFileType)
	v.AddConfigPath("../") // Look for config in the current directory

	// Attempt to read the configuration file
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Printf("INFO: Configuration file (%s.%s) not found.", configFileName, configFileType)
		} else {
			log.Printf("WARNING: Error reading configuration file: %v. ", err)
		}
	} else {
		log.Printf("INFO: Successfully loaded configuration from %s.%s", configFileName, configFileType)
	}

	if err := v.Unmarshal(&effectiveConfig); err != nil {
		log.Printf("ERROR: Failed to unmarshal config: %v", err)
	}

	log.Printf("INFO: Final test configuration: %+v", effectiveConfig)
}

// GetConfig returns the loaded and processed cluster configuration.
func GetConfig() ClusterSettings {
	return effectiveConfig
}
