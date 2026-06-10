package config

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"io"
	"net/http"
	"strings"
)

// Identifier type constants
const (
	IdentifierSha     string = "identifiersha"
	JsonSha                  = "jsonsha"
	NormalizedJsonSha        = "normalizedjsonsha"
	IdentifierString         = "identifierstring"
	SourceUrl                = "sourceurl"
	FileSha                  = "filesha"
)

// ContextOption represents context handling modes for JSON-LD processing.
// Stored as a string to match YAML config values ("strict", "https", "http", etc.).
type ContextOption = string

const (
	Strict            ContextOption = "strict"
	Https             ContextOption = "https"
	Http              ContextOption = "http"
	StandardizedHttps ContextOption = "standardizedHttps"
	StandardizedHttp  ContextOption = "standardizedHttp"
)

// AccceptContentType is the default accept content type for HTTP requests (note: legacy spelling preserved)
const AccceptContentType string = "application/ld+json, text/html"

// Sources holds configuration for a data source (website, API, etc.)
// Originally from Gleaner's internal/config/sources.go
type Sources struct {
	// Valid values for SourceType: sitemap, sitegraph, csv, googledrive, api, and robots
	SourceType      string `default:"sitemap"`
	Name            string
	Logo            string
	URL             string
	Headless        bool `default:"false"`
	PID             string
	ProperName      string
	Domain          string
	Active          bool                   `default:"true"`
	CredentialsFile string                 // do not want someone's google api key exposed
	Other           map[string]interface{} `mapstructure:",remain"`

	HeadlessWait      int           // if loading is slow, wait
	Delay             int64         // A domain-specific crawl delay value
	IdentifierPath    string        // JSON Path to the identifier
	ApiPageLimit      int
	IdentifierType    string
	FixContextOption  ContextOption // context handling mode
	AcceptContentType string        `default:"application/ld+json, text/html"`
	JsonProfile       string        // jsonprofile
}

// SourcesConfig holds the full sources configuration used by both
// Gleaner (harvesting) and Nabu (loading) operations.
type SourcesConfig struct {
	Sources               []Sources
	Objects               Objects
	ImplNetwork           ImplNetwork
	Context               map[string]string
	ContextMaps           []ContextMapping
}

// ContextMapping holds the JSON-LD mappings for cached context
type ContextMapping struct {
	Prefix string
	File   string
}

var sourcesConfigTemplate = map[string]interface{}{
	"sources":                []interface{}{},
	"objects":                ObjectTemplate,
	"implementation_network": implNetworkTemplate,
	"context":                map[string]string{},
	"contextmaps":           []interface{}{},
}

// ReadSourcesConfig reads a sources configuration file containing
// data source definitions, object prefixes, and implementation network settings.
func ReadSourcesConfig(filename string, cfgPath string) (*viper.Viper, error) {
	v := viper.New()
	for key, value := range sourcesConfigTemplate {
		v.SetDefault(key, value)
	}

	v.SetConfigName(fileNameWithoutExtTrimSuffix(filename))
	v.AddConfigPath(cfgPath)
	v.SetConfigType("yaml")
	v.AutomaticEnv()
	err := v.ReadInConfig()
	return v, err
}

// ReadSourcesConfigURL reads a sources configuration from a URL
func ReadSourcesConfigURL(configURL string) (*viper.Viper, error) {
	v := viper.New()
	for key, value := range sourcesConfigTemplate {
		v.SetDefault(key, value)
	}

	log.Printf("Reading sources config from URL: %v\n", configURL)

	resp, err := http.Get(configURL)
	if err != nil {
		return v, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return v, fmt.Errorf("HTTP request failed with status code %v", resp.StatusCode)
	}

	configData, err := io.ReadAll(resp.Body)
	if err != nil {
		return v, err
	}

	reader := strings.NewReader(string(configData))

	v.SetConfigType("yaml")
	v.AutomaticEnv()

	err = v.ReadConfig(reader)
	if err != nil {
		fmt.Printf("Error reading sources config from URL: %v\n", err)
		return v, err
	}

	return v, err
}

// GetSources retrieves the list of sources from the config
func GetSources(v1 *viper.Viper) ([]Sources, error) {
	var subtreeKey = "sources"
	var cfg []Sources

	err := v1.UnmarshalKey(subtreeKey, &cfg)
	if err != nil {
		log.Fatal("error when parsing ", subtreeKey, " config: ", err)
	}
	for i, s := range cfg {
		cfg[i] = populateSourceDefaults(s)
	}
	return cfg, err
}

// GetActiveSources returns only active sources from the config
func GetActiveSources(v1 *viper.Viper) ([]Sources, error) {
	var activeSources []Sources

	sources, err := GetSources(v1)
	if err != nil {
		return nil, err
	}
	for _, s := range sources {
		if s.Active {
			activeSources = append(activeSources, s)
		}
	}
	return activeSources, err
}

func populateSourceDefaults(s Sources) Sources {
	if s.SourceType == "" {
		s.SourceType = "sitemap"
	}
	if s.AcceptContentType == "" {
		s.AcceptContentType = "application/ld+json, text/html"
	}
	if s.JsonProfile == "" {
		s.JsonProfile = "application/ld+json"
	}
	s.URL = strings.TrimSpace(s.URL)
	return s
}

// GetSourceByType returns sources matching a specific source type
func GetSourceByType(sources []Sources, key string) []Sources {
	var sourcesSlice []Sources
	for _, s := range sources {
		if s.SourceType == key {
			sourcesSlice = append(sourcesSlice, s)
		}
	}
	return sourcesSlice
}

// GetActiveSourceByType returns active sources matching a specific source type
func GetActiveSourceByType(sources []Sources, key string) []Sources {
	var sourcesSlice []Sources
	for _, s := range sources {
		if s.SourceType == key && s.Active {
			sourcesSlice = append(sourcesSlice, s)
		}
	}
	return sourcesSlice
}

// GetActiveSourceByHeadless returns active sources filtered by headless setting
func GetActiveSourceByHeadless(sources []Sources, headless bool) []Sources {
	var sourcesSlice []Sources
	for _, s := range sources {
		if s.Headless == headless && s.Active {
			sourcesSlice = append(sourcesSlice, s)
		}
	}
	return sourcesSlice
}

// GetSourceByName returns a specific source by name
func GetSourceByName(sources []Sources, name string) (*Sources, error) {
	for i := 0; i < len(sources); i++ {
		if sources[i].Name == name {
			return &sources[i], nil
		}
	}
	return nil, fmt.Errorf("unable to find a source with name %s", name)
}

// PruneSources filters sources to only include those in the useSources list
func PruneSources(v1 *viper.Viper, useSources []string) (*viper.Viper, error) {
	var finalSources []Sources
	allSources, err := GetSources(v1)
	if err != nil {
		log.Fatal("error retrieving sources: ", err)
	}
	for _, s := range allSources {
		if containsString(useSources, s.Name) {
			s.Active = true
			finalSources = append(finalSources, s)
		}
	}
	if len(finalSources) > 0 {
		v1.Set("sources", finalSources)
		return v1, err
	}
	return v1, fmt.Errorf("cannot find a source with the given names")
}

func containsString(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}

// SourceToNabuPrefix converts sources to the object prefix paths used for summoned/milled data
func SourceToNabuPrefix(sources []Sources, useMilled bool) []string {
	jsonld := "summoned"
	if useMilled {
		jsonld = "milled"
	}
	var prefixes []string
	for _, s := range sources {
		switch s.SourceType {
		case "sitemap":
			prefixes = append(prefixes, fmt.Sprintf("%s/%s", jsonld, s.Name))
		case "sitegraph":
			prefixes = append(prefixes, fmt.Sprintf("%s/%s", "summoned", s.Name))
		case "googledrive":
			prefixes = append(prefixes, fmt.Sprintf("%s/%s", jsonld, s.Name))
		}
	}
	return prefixes
}

// SourceToNabuProv converts sources to the prov prefix paths
func SourceToNabuProv(sources []Sources) []string {
	var prefixes []string
	for _, s := range sources {
		switch s.SourceType {
		case "sitemap", "sitegraph", "googledrive":
			prefixes = append(prefixes, "prov/"+s.Name)
		}
	}
	return prefixes
}

// ReadGleanerConfig reads a gleaner-style configuration file for backward compatibility.
// This supports the combined config format used by Gleaner's gleanerconfig.yaml.
func ReadGleanerConfig(filename string, cfgDir string) (*viper.Viper, error) {
	v := viper.New()

	gleanerTemplate := map[string]interface{}{
		"minio":   MinioTemplate,
		"gleaner": map[string]string{},
		"context": map[string]string{},
		"contextmaps": map[string]string{},
		"summoner":    SummonerTemplate,
		"millers":     map[string]string{},
		"sources":     sourcesConfigTemplate["sources"],
	}

	for key, value := range gleanerTemplate {
		v.SetDefault(key, value)
	}

	v.SetConfigName(fileNameWithoutExtTrimSuffix(filename))
	v.AddConfigPath(cfgDir)
	v.SetConfigType("yaml")
	v.BindEnv("minio.address", "MINIO_ADDRESS")
	v.BindEnv("minio.port", "MINIO_PORT")
	v.BindEnv("minio.ssl", "MINIO_USE_SSL")
	v.BindEnv("minio.accesskey", "MINIO_ACCESS_KEY")
	v.BindEnv("minio.secretkey", "MINIO_SECRET_KEY")
	v.AutomaticEnv()
	err := v.ReadInConfig()
	if err != nil {
		log.Fatalf("cannot find config file '%v': %v", filename, err)
	}
	return v, err
}
