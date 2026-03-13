package config

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"io"
	"net/http"
	"strings"
)

// ServicesConfig holds all service connection details (URLs, credentials, secrets).
// This is used by both Gleaner (harvesting) and Nabu (loading) operations.
type ServicesConfig struct {
	Minio     Minio
	Endpoints []EndPoint
	Summoner  Summoner
}

var servicesTemplate = map[string]interface{}{
	"minio":     MinioTemplate,
	"endpoints": []interface{}{},
	"summoner":  SummonerTemplate,
}

// ReadServicesConfig reads a services configuration file containing
// connection details for MinIO, SPARQL endpoints, headless browser, etc.
func ReadServicesConfig(filename string, cfgPath string) (*viper.Viper, error) {
	v := viper.New()
	for key, value := range servicesTemplate {
		v.SetDefault(key, value)
	}

	v.SetConfigName(fileNameWithoutExtTrimSuffix(filename))
	v.AddConfigPath(cfgPath)
	v.SetConfigType("yaml")
	v.BindEnv("minio.address", "MINIO_ADDRESS")
	v.BindEnv("minio.port", "MINIO_PORT")
	v.BindEnv("minio.ssl", "MINIO_USE_SSL")
	v.BindEnv("minio.accesskey", "MINIO_ACCESS_KEY")
	v.BindEnv("minio.secretkey", "MINIO_SECRET_KEY")
	v.BindEnv("minio.bucket", "MINIO_BUCKET")
	v.AutomaticEnv()
	err := v.ReadInConfig()
	return v, err
}

// ReadServicesConfigURL reads a services configuration from a URL
func ReadServicesConfigURL(configURL string) (*viper.Viper, error) {
	v := viper.New()
	for key, value := range servicesTemplate {
		v.SetDefault(key, value)
	}

	log.Printf("Reading services config from URL: %v\n", configURL)

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
	v.BindEnv("minio.address", "MINIO_ADDRESS")
	v.BindEnv("minio.port", "MINIO_PORT")
	v.BindEnv("minio.ssl", "MINIO_USE_SSL")
	v.BindEnv("minio.accesskey", "MINIO_ACCESS_KEY")
	v.BindEnv("minio.secretkey", "MINIO_SECRET_KEY")
	v.BindEnv("minio.bucket", "MINIO_BUCKET")
	v.AutomaticEnv()

	err = v.ReadConfig(reader)
	if err != nil {
		fmt.Printf("Error reading services config from URL: %v\n", err)
		return v, err
	}

	return v, err
}
