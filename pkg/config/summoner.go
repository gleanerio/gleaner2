package config

import (
	"fmt"
	"github.com/spf13/viper"
	"strings"
)

// Summoner holds configuration for the web harvesting summoner
type Summoner struct {
	After          string
	Mode           string // full || diff
	Threads        int
	Delay          int64  // milliseconds (1000 = 1 second)
	Headless       string // URL for headless browser
	IdentifierType string // identifiersha, filesha, identifier
}

var SummonerTemplate = map[string]interface{}{
	"summoner": map[string]string{
		"after":          "",
		"mode":           "full",
		"threads":        "5",
		"delay":          "10000",
		"headless":       "http://127.0.0.1:9222",
		"identifiertype": "jsonsha",
	},
}

// ReadSummmonerConfig is the legacy name (triple 'm') preserved for compatibility
func ReadSummmonerConfig(viperSubtree *viper.Viper) (Summoner, error) {
	return ReadSummonerConfig(viperSubtree)
}

func ReadSummonerConfig(viperSubtree *viper.Viper) (Summoner, error) {
	var summoner Summoner
	for key, value := range SummonerTemplate {
		viperSubtree.SetDefault(key, value)
	}
	viperSubtree.BindEnv("headless", "GLEANER_HEADLESS_ENDPOINT")
	viperSubtree.BindEnv("threads", "GLEANER_THREADS")
	viperSubtree.BindEnv("mode", "GLEANER_MODE")

	viperSubtree.AutomaticEnv()
	err := viperSubtree.Unmarshal(&summoner)
	if err != nil {
		panic(fmt.Errorf("error when parsing summoner config: %v", err))
	}
	if strings.HasSuffix(summoner.Headless, "/") {
		panic(fmt.Errorf("headless URL should not end with /: %v", summoner.Headless))
	}
	return summoner, err
}
